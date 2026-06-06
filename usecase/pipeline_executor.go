package usecase

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/logw"
)

type PipelineExecutor struct {
	DockerEngine  outbound.DockerEngine
	AppFileRepo   outbound.AppFileRepository
	Filesystem    outbound.Filesystem
	StepRepo      outbound.PipelineStepRepository
	DeployRepo    outbound.DeploymentRepository
	DockerNetwork string
}

func (e *PipelineExecutor) ExecuteStep(ctx context.Context, step *entity.PipelineStep, d *entity.Deployment, appID, image, authRegistry, authUser, authPass string, containerID *string, containerName *string, labels map[string]string, containerNameOverride string) error {
	switch step.Name {
	case "pull_image":
		return e.DockerEngine.PullImage(ctx, image, authRegistry, authUser, authPass, os.Stderr)

	case "write_files":
		config := parseWriteFilesConfig(step.Config)
		files, err := e.AppFileRepo.FindByAppID(ctx, appID)
		if err != nil {
			return fmt.Errorf("failed to load app files: %w", err)
		}
		for _, f := range files {
			hostPath := filepath.Join(config.HostBase, strings.TrimPrefix(f.Path, "/"))
			dir := filepath.Dir(hostPath)
			if err := e.Filesystem.MkdirAll(dir, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", dir, err)
			}
			if err := e.Filesystem.WriteFile(hostPath, []byte(f.Content), 0644); err != nil {
				return fmt.Errorf("failed to write file %s: %w", hostPath, err)
			}
		}
		return nil

	case "create_container":
		config := parseContainerConfig(step.Config)
		name := containerNameOverride
		if name == "" {
			name = fmt.Sprintf("%s-deploy-%d", appID, d.ID)
		}
		// Remove any existing container with the same name (e.g., fixed container names like "cafe-fe")
		e.DockerEngine.RemoveContainer(ctx, name, true)
		params := outbound.CreateContainerParams{
			Image:        image,
			Name:         name,
			Ports:        config.Ports,
			Env:          config.Env,
			AutoRemove:   config.AutoRemove,
			VolumeMounts: make([]outbound.VolumeMount, 0, len(config.VolumeMounts)),
			Labels:       labels,
			Network:      e.DockerNetwork,
		}
		for _, vm := range config.VolumeMounts {
			params.VolumeMounts = append(params.VolumeMounts, outbound.VolumeMount{
				HostPath:      vm.HostPath,
				ContainerPath: vm.ContainerPath,
				Mode:          vm.Mode,
			})
		}
		if config.HealthCheckPath != "" {
			params.HealthCheck = &outbound.HealthCheckConfig{
				Test:           []string{"CMD-SHELL", fmt.Sprintf("(curl -sf http://localhost:%s%s || wget -qO- http://localhost:%s%s) > /dev/null 2>&1 || exit 1", config.Port, config.HealthCheckPath, config.Port, config.HealthCheckPath)},
				IntervalSec:    10,
				TimeoutSec:     5,
				Retries:        3,
				StartPeriodSec: 10,
			}
		}
		id, err := e.DockerEngine.CreateContainer(ctx, params)
		if err != nil {
			return err
		}
		*containerID = id
		*containerName = name
		return nil

	case "start_container":
		if *containerID == "" {
			return fmt.Errorf("no container ID from previous step")
		}
		return e.DockerEngine.StartContainer(ctx, *containerID)

	case "verify_health":
		if *containerID == "" {
			return fmt.Errorf("no container ID from previous step")
		}
		config := parseVerifyConfig(step.Config)
		return e.waitForHealthy(ctx, *containerID, config.TimeoutSeconds, config.IntervalSeconds)

	case "exec_command":
		if *containerID == "" {
			return fmt.Errorf("no container ID from previous step")
		}
		config := parseExecCommandConfig(step.Config)
		result, err := e.DockerEngine.ExecCommand(ctx, outbound.ExecCommandParams{
			ContainerID: *containerID,
			Command:      config.Command,
			WorkingDir:   config.WorkingDir,
		})
		if err != nil {
			return err
		}
		var output strings.Builder
		if result.Stdout != "" {
			output.WriteString(result.Stdout)
		}
		if result.Stderr != "" {
			if output.Len() > 0 {
				output.WriteString("\n")
			}
			output.WriteString(result.Stderr)
		}
		step.Output = output.String()
		if result.ExitCode != 0 {
			return fmt.Errorf("exec command exited with code %d: %s", result.ExitCode, result.Stderr)
		}
		return nil

	default:
		return fmt.Errorf("unknown pipeline step: %s", step.Name)
	}
}

func (e *PipelineExecutor) waitForHealthy(ctx context.Context, containerID string, timeoutSec, intervalSec int) error {
	deadline := time.Now().Add(time.Duration(timeoutSec) * time.Second)
	for {
		info, err := e.DockerEngine.InspectContainer(ctx, containerID)
		if err != nil {
			return fmt.Errorf("health check: inspect failed: %w", err)
		}

		if !info.Running {
			return fmt.Errorf("health check: container is not running (status: %s)", info.Status)
		}

		if info.Health == "healthy" {
			return nil
		}
		if info.Health == "unhealthy" {
			return fmt.Errorf("health check: container is unhealthy")
		}

		if time.Now().After(deadline) {
			if info.Health == "" {
				return nil
			}
			return fmt.Errorf("health check: timed out after %d seconds (status: %s)", timeoutSec, info.Health)
		}

		time.Sleep(time.Duration(intervalSec) * time.Second)
	}
}

func (e *PipelineExecutor) RunPipelineSteps(ctx context.Context, steps []*entity.PipelineStep, d *entity.Deployment, appID, image, authRegistry, authUser, authPass string, labels map[string]string, containerNameOverride string) (containerID string, containerName string, failed bool) {
	for _, step := range steps {
		now := time.Now()
		step.Status = entity.StepRunning
		step.StartedAt = &now
		e.StepRepo.Update(ctx, step)

		switch step.Name {
		case "pull_image":
			d.Status = entity.DeploymentPulling
		case "write_files":
			d.Status = entity.DeploymentPulling
		case "create_container":
			d.Status = entity.DeploymentCreating
		case "start_container":
			d.Status = entity.DeploymentStarting
		case "verify_health":
			d.Status = entity.DeploymentStarting
		case "exec_command":
			d.Status = entity.DeploymentExecuting
		}
		e.DeployRepo.Update(ctx, d)

		stepErr := e.ExecuteStep(ctx, step, d, appID, image, authRegistry, authUser, authPass, &containerID, &containerName, labels, containerNameOverride)

		finishedNow := time.Now()
		step.FinishedAt = &finishedNow

		if stepErr != nil {
			step.Status = entity.StepFailed
			step.Error = stepErr.Error()
			e.StepRepo.Update(ctx, step)

			d.Status = entity.DeploymentFailed
			d.Error = fmt.Sprintf("Step %s failed: %s", step.Name, stepErr.Error())
			e.DeployRepo.Update(ctx, d)
			return containerID, containerName, true
		}

		step.Status = entity.StepDone
		e.StepRepo.Update(ctx, step)
	}
	return containerID, containerName, false
}

// --- Config types and parsers (shared) ---

type containerConfig struct {
	Port            json.Number         `json:"port"`
	ContainerPort   string             `json:"container_port"`
	PublishPort     string             `json:"publish_port"`
	Env             map[string]string   `json:"env"`
	AutoRemove      bool                `json:"auto_remove"`
	HealthCheckPath string              `json:"healthcheck_path"`
	VolumeMounts    []volumeMountConfig `json:"volume_mounts"`
	Ports           map[string]string
}

type volumeMountConfig struct {
	HostPath      string `json:"host_path"`
	ContainerPath string `json:"container_path"`
	Mode          string `json:"mode"`
}

func parseContainerConfig(configJSON string) containerConfig {
	var c containerConfig
	if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
		logw.Warningf("failed to parse container config: %v, using defaults", err)
		return containerConfig{Env: map[string]string{}, Ports: map[string]string{"8080": "8080"}}
	}

	// Resolve port string from json.Number or string
	portStr := c.Port.String()
	if portStr == "" {
		portStr = "8080"
	}
	if c.Env == nil {
		c.Env = map[string]string{}
	}
	if c.VolumeMounts == nil {
		c.VolumeMounts = []volumeMountConfig{}
	}

	// Determine effective container port
	effectivePort := portStr
	if c.ContainerPort != "" {
		effectivePort = c.ContainerPort
	}

	// Determine port publishing
	if c.PublishPort != "" {
		c.Ports = map[string]string{c.PublishPort: effectivePort}
	} else if c.ContainerPort != "" {
		c.Ports = map[string]string{}
	} else {
		c.Ports = map[string]string{portStr: portStr}
	}

	// Store effective port back for health check reference
	c.Port = json.Number(effectivePort)
	return c
}

type execCommandConfig struct {
	Command   []string `json:"command"`
	WorkingDir string  `json:"working_dir"`
}

func parseExecCommandConfig(configJSON string) execCommandConfig {
	var c struct {
		Command   string `json:"command"`
		WorkingDir string `json:"working_dir"`
	}
	if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
		return execCommandConfig{
			Command: []string{"sh", "-c", "echo error"},
		}
	}
	return execCommandConfig{
		Command:    []string{"sh", "-c", c.Command},
		WorkingDir: c.WorkingDir,
	}
}

type verifyConfig struct {
	TimeoutSeconds  int `json:"timeout_seconds"`
	IntervalSeconds int `json:"interval_seconds"`
}

func parseVerifyConfig(configJSON string) verifyConfig {
	var c verifyConfig
	if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
		return verifyConfig{TimeoutSeconds: 30, IntervalSeconds: 2}
	}
	if c.TimeoutSeconds == 0 {
		c.TimeoutSeconds = 30
	}
	if c.IntervalSeconds == 0 {
		c.IntervalSeconds = 2
	}
	return c
}

type writeFilesConfig struct {
	AppID    string `json:"app_id"`
	HostBase string `json:"host_base"`
}

func parseWriteFilesConfig(configJSON string) writeFilesConfig {
	var c writeFilesConfig
	if err := json.Unmarshal([]byte(configJSON), &c); err != nil {
		return writeFilesConfig{}
	}
	return c
}

func MergeAppConfig(template *entity.PipelineTemplateDefinition, app *entity.App) *entity.PipelineTemplateDefinition {
	merged := *template

	for i, step := range merged.Steps {
		if step.Name == "create_container" {
			merged.Steps[i].Config = mergeEnvVarsIntoConfig(step.Config, app.EnvVars)
			merged.Steps[i].Config = mergeVolumeMountsIntoConfig(merged.Steps[i].Config, app.VolumeMounts)
			if app.ContainerPort != "" || app.PublishPort != "" {
				merged.Steps[i].Config = mergePortIntoConfig(merged.Steps[i].Config, app.ContainerPort, app.PublishPort)
			}
		}
	}

	nextOrder := len(merged.Steps) + 1
	for _, cmd := range app.PostDeployCommands {
		config := fmt.Sprintf(`{"command":%q}`, cmd)
		merged.Steps = append(merged.Steps, entity.PipelineStepDefinition{
			Name:   "exec_command",
			Order:  nextOrder,
			Config: config,
		})
		nextOrder++
	}

	return &merged
}

func MergeAppFiles(template *entity.PipelineTemplateDefinition, app *entity.App) *entity.PipelineTemplateDefinition {
	merged := *template

	hostBase := "/home/user/docker/" + app.AppID + "/files"
	writeFilesConfig := fmt.Sprintf(`{"app_id":%q,"host_base":%q}`, app.AppID, hostBase)

	createIdx := -1
	for i, step := range merged.Steps {
		if step.Name == "create_container" {
			createIdx = i
			break
		}
	}

	if createIdx == -1 {
		return &merged
	}

	writeFilesStep := entity.PipelineStepDefinition{
		Name:   "write_files",
		Order:  merged.Steps[createIdx].Order,
		Config: writeFilesConfig,
	}

	newSteps := make([]entity.PipelineStepDefinition, 0, len(merged.Steps)+1)
	newSteps = append(newSteps, merged.Steps[:createIdx]...)
	newSteps = append(newSteps, writeFilesStep)
	for _, step := range merged.Steps[createIdx:] {
		newSteps = append(newSteps, entity.PipelineStepDefinition{
			Name:   step.Name,
			Order:  step.Order + 1,
			Config: step.Config,
		})
	}

	return &entity.PipelineTemplateDefinition{Steps: newSteps}
}