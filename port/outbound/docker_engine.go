package outbound

import (
	"context"
	"io"
	"net"
)

type DockerEngine interface {
	Ping(ctx context.Context) error
	PullImage(ctx context.Context, image, authRegistry, authUsername, authPassword string, progressSink io.Writer) error
	CreateContainer(ctx context.Context, params CreateContainerParams) (string, error)
	StartContainer(ctx context.Context, containerID string) error
	StopContainer(ctx context.Context, containerID string, timeoutSeconds int) error
	RemoveContainer(ctx context.Context, containerID string, force bool) error
	InspectContainer(ctx context.Context, containerID string) (*ContainerInfo, error)
	ExecCommand(ctx context.Context, params ExecCommandParams) (*ExecCommandResult, error)
	ContainerLogs(ctx context.Context, containerID string, opts ContainerLogsOptions) (io.ReadCloser, error)
	ContainerList(ctx context.Context, labels map[string]string) ([]*ContainerInfo, error)
	ExecInteractive(ctx context.Context, containerID string, shell string) (conn net.Conn, reader io.Reader, err error)
}

type VolumeMount struct {
	HostPath      string
	ContainerPath string
	Mode          string // "rw" or "ro"
}

type CreateContainerParams struct {
	Image        string
	Name         string
	Ports        map[string]string
	Env          map[string]string
	VolumeMounts []VolumeMount
	AutoRemove   bool
	HealthCheck  *HealthCheckConfig
	Labels       map[string]string
	Network      string
}

type ExecCommandParams struct {
	ContainerID string
	Command     []string
	WorkingDir  string
	Env         []string
}

type ExecCommandResult struct {
	ExitCode int
	Stdout   string
	Stderr   string
}

type HealthCheckConfig struct {
	Test           []string
	IntervalSec    int
	TimeoutSec     int
	Retries        int
	StartPeriodSec int
}

type ContainerInfo struct {
	ID      string
	Name    string
	Running bool
	Status  string
	Health  string
	IP      string
	Ports   map[string]string
}

type ContainerLogsOptions struct {
	Tail       string
	Since      string
	Timestamps bool
}