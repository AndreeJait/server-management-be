package outbound

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"net"
	"time"

	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
)

type DockerConn struct {
	Engine outbound.DockerEngine
	Client *client.Client
}

func ConnectDocker(host string) (*DockerConn, error) {
	opts := []client.Opt{client.FromEnv}
	if host != "" {
		opts = append(opts, client.WithHost(host))
	}

	cli, err := client.NewClientWithOpts(opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	cli.NegotiateAPIVersion(context.Background())

	return &DockerConn{
		Engine: &dockerEngine{client: cli},
		Client: cli,
	}, nil
}

func DisconnectDocker(conn *DockerConn) func(ctx context.Context) error {
	return func(ctx context.Context) error {
		return conn.Client.Close()
	}
}

type dockerEngine struct {
	client *client.Client
}

func (d *dockerEngine) Ping(ctx context.Context) error {
	_, err := d.client.Ping(ctx)
	return err
}

func (d *dockerEngine) PullImage(ctx context.Context, img, authRegistry, authUsername, authPassword string, progressSink io.Writer) error {
	pullOpts := image.PullOptions{}
	if authRegistry != "" && authUsername != "" {
		auth := registry.AuthConfig{
			ServerAddress: authRegistry,
			Username:      authUsername,
			Password:      authPassword,
		}
		encoded, err := registry.EncodeAuthConfig(auth)
		if err != nil {
			return fmt.Errorf("failed to encode auth: %w", err)
		}
		pullOpts.RegistryAuth = encoded
	}

	reader, err := d.client.ImagePull(ctx, img, pullOpts)
	if err != nil {
		return fmt.Errorf("failed to pull image %s: %w", img, err)
	}
	defer reader.Close()

	if progressSink != nil {
		_, _ = io.Copy(progressSink, reader)
	} else {
		_, _ = io.Copy(io.Discard, reader)
	}
	return nil
}

func (d *dockerEngine) CreateContainer(ctx context.Context, params outbound.CreateContainerParams) (string, error) {
	portBindings := nat.PortMap{}
	exposedPorts := nat.PortSet{}
	for hostPort, containerPort := range params.Ports {
		port, err := nat.NewPort("tcp", containerPort)
		if err != nil {
			return "", fmt.Errorf("invalid container port %s: %w", containerPort, err)
		}
		exposedPorts[port] = struct{}{}
		portBindings[port] = []nat.PortBinding{{HostIP: "0.0.0.0", HostPort: hostPort}}
	}

	var env []string
	for k, v := range params.Env {
		env = append(env, fmt.Sprintf("%s=%s", k, v))
	}

	config := &container.Config{
		Image:        params.Image,
		Env:          env,
		ExposedPorts: exposedPorts,
	}

	if params.Labels != nil {
		config.Labels = make(map[string]string)
		for k, v := range params.Labels {
			config.Labels[k] = v
		}
	}

	if params.HealthCheck != nil {
		config.Healthcheck = &container.HealthConfig{
			Test:        params.HealthCheck.Test,
			Interval:    time.Duration(params.HealthCheck.IntervalSec) * time.Second,
			Timeout:     time.Duration(params.HealthCheck.TimeoutSec) * time.Second,
			Retries:     params.HealthCheck.Retries,
			StartPeriod: time.Duration(params.HealthCheck.StartPeriodSec) * time.Second,
		}
	}

	hostConfig := &container.HostConfig{
		PortBindings: portBindings,
		AutoRemove:   params.AutoRemove,
		RestartPolicy: container.RestartPolicy{
			Name: "unless-stopped",
		},
	}

	for _, vm := range params.VolumeMounts {
		mode := vm.Mode
		if mode == "" {
			mode = "rw"
		}
		hostConfig.Binds = append(hostConfig.Binds, fmt.Sprintf("%s:%s:%s", vm.HostPath, vm.ContainerPath, mode))
	}

	var networkingConfig *network.NetworkingConfig
	if params.Network != "" {
		networkingConfig = &network.NetworkingConfig{
			EndpointsConfig: map[string]*network.EndpointSettings{
				params.Network: {},
			},
		}
	}

	resp, err := d.client.ContainerCreate(ctx, config, hostConfig, networkingConfig, nil, params.Name)
	if err != nil {
		return "", fmt.Errorf("failed to create container: %w", err)
	}
	return resp.ID, nil
}

func (d *dockerEngine) StartContainer(ctx context.Context, containerID string) error {
	return d.client.ContainerStart(ctx, containerID, container.StartOptions{})
}

func (d *dockerEngine) StopContainer(ctx context.Context, containerID string, timeoutSeconds int) error {
	timeout := timeoutSeconds
	return d.client.ContainerStop(ctx, containerID, container.StopOptions{Timeout: &timeout})
}

func (d *dockerEngine) RemoveContainer(ctx context.Context, containerID string, force bool) error {
	return d.client.ContainerRemove(ctx, containerID, container.RemoveOptions{Force: force})
}

func (d *dockerEngine) InspectContainer(ctx context.Context, containerID string) (*outbound.ContainerInfo, error) {
	inspect, err := d.client.ContainerInspect(ctx, containerID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect container: %w", err)
	}

	info := &outbound.ContainerInfo{
		ID:      inspect.ID,
		Name:    inspect.Name,
		Running: inspect.State.Running,
		Status:  inspect.State.Status,
	}

	if inspect.State.Health != nil {
		info.Health = inspect.State.Health.Status
	}

	info.Ports = make(map[string]string)
	if inspect.NetworkSettings != nil && inspect.NetworkSettings.Ports != nil {
		for port, bindings := range inspect.NetworkSettings.Ports {
			if len(bindings) > 0 {
				info.Ports[bindings[0].HostPort] = port.Port()
			}
		}
	}

	if inspect.NetworkSettings != nil {
		for _, net := range inspect.NetworkSettings.Networks {
			if net.IPAddress != "" {
				info.IP = net.IPAddress
				break
			}
		}
	}

	return info, nil
}

func (d *dockerEngine) ExecCommand(ctx context.Context, params outbound.ExecCommandParams) (*outbound.ExecCommandResult, error) {
	execConfig := container.ExecOptions{
		Cmd:          params.Command,
		WorkingDir:   params.WorkingDir,
		Env:          params.Env,
		AttachStdout: true,
		AttachStderr: true,
	}

	createResp, err := d.client.ContainerExecCreate(ctx, params.ContainerID, execConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create exec: %w", err)
	}

	attachResp, err := d.client.ContainerExecAttach(ctx, createResp.ID, container.ExecAttachOptions{
		Detach: false,
		Tty:    false,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to attach exec: %w", err)
	}
	defer attachResp.Close()

	var stdoutBuf, stderrBuf bytes.Buffer
	if _, err := stdcopy.StdCopy(&stdoutBuf, &stderrBuf, attachResp.Reader); err != nil {
		return nil, fmt.Errorf("failed to read exec output: %w", err)
	}

	inspectResp, err := d.client.ContainerExecInspect(ctx, createResp.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to inspect exec: %w", err)
	}

	return &outbound.ExecCommandResult{
		ExitCode: inspectResp.ExitCode,
		Stdout:   stdoutBuf.String(),
		Stderr:   stderrBuf.String(),
	}, nil
}

func (d *dockerEngine) ContainerLogs(ctx context.Context, containerID string, opts outbound.ContainerLogsOptions) (io.ReadCloser, error) {
	tail := opts.Tail
	if tail == "" {
		tail = "100"
	}
	logsOpts := container.LogsOptions{
		Tail:       tail,
		Since:      opts.Since,
		Timestamps: opts.Timestamps,
		ShowStdout: true,
		ShowStderr: true,
	}
	reader, err := d.client.ContainerLogs(ctx, containerID, logsOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to get container logs: %w", err)
	}
	return reader, nil
}

func (d *dockerEngine) ExecInteractive(ctx context.Context, containerID string, shell string) (net.Conn, io.Reader, error) {
	execCreate, err := d.client.ContainerExecCreate(ctx, containerID, container.ExecOptions{
		AttachStdin:  true,
		AttachStdout: true,
		AttachStderr: true,
		Tty:          true,
		Cmd:          []string{shell},
	})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create interactive exec: %w", err)
	}

	hijacked, err := d.client.ContainerExecAttach(ctx, execCreate.ID, container.ExecStartOptions{Tty: true})
	if err != nil {
		return nil, nil, fmt.Errorf("failed to attach interactive exec: %w", err)
	}

	return hijacked.Conn, hijacked.Reader, nil
}

func (d *dockerEngine) ContainerList(ctx context.Context, labels map[string]string) ([]*outbound.ContainerInfo, error) {
	filterArgs := filters.NewArgs()
	for k, v := range labels {
		filterArgs.Add("label", fmt.Sprintf("%s=%s", k, v))
	}

	containers, err := d.client.ContainerList(ctx, container.ListOptions{
		All:     true,
		Filters: filterArgs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to list containers: %w", err)
	}

	var result []*outbound.ContainerInfo
	for _, c := range containers {
		info := &outbound.ContainerInfo{
			ID:      c.ID,
			Running: c.State == "running",
			Status:  c.Status,
		}
		if len(c.Names) > 0 {
			info.Name = strings.TrimPrefix(c.Names[0], "/")
		}
		info.Ports = make(map[string]string)
		for _, p := range c.Ports {
			if p.PublicPort > 0 {
				info.Ports[fmt.Sprintf("%d", p.PublicPort)] = fmt.Sprintf("%d", p.PrivatePort)
			}
		}
		result = append(result, info)
	}
	return result, nil
}