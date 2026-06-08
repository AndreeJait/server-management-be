package usecase

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"io"
	"net"
	"strconv"
	"strings"

	"github.com/AndreeJait/server-management-be/config"
	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/deployment"
	proxyInbound "github.com/AndreeJait/server-management-be/port/inbound/proxy"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/logw"
)

type deploymentUseCase struct {
	appRepo      outbound.AppRepository
	projectRepo  outbound.ProjectRepository
	deployRepo   outbound.DeploymentRepository
	stepRepo     outbound.PipelineStepRepository
	credRepo     outbound.RegistryCredentialRepository
	dockerEngine outbound.DockerEngine
	appFileRepo  outbound.AppFileRepository
	filesystem   outbound.Filesystem
	bindingRepo  outbound.AppBindingRepository
	proxyUC      proxyInbound.UseCase
	executor     PipelineExecutor
	runtimeCfg   *config.RuntimeConfig
}

func NewDeploymentUseCase(
	appRepo outbound.AppRepository,
	projectRepo outbound.ProjectRepository,
	deployRepo outbound.DeploymentRepository,
	stepRepo outbound.PipelineStepRepository,
	credRepo outbound.RegistryCredentialRepository,
	dockerEngine outbound.DockerEngine,
	appFileRepo outbound.AppFileRepository,
	filesystem outbound.Filesystem,
	bindingRepo outbound.AppBindingRepository,
	proxyUC proxyInbound.UseCase,
	runtimeCfg *config.RuntimeConfig,
) deployment.UseCase {
	uc := &deploymentUseCase{
		appRepo:      appRepo,
		projectRepo:  projectRepo,
		deployRepo:   deployRepo,
		stepRepo:     stepRepo,
		credRepo:     credRepo,
		dockerEngine: dockerEngine,
		appFileRepo:  appFileRepo,
		filesystem:   filesystem,
		bindingRepo:  bindingRepo,
		proxyUC:      proxyUC,
		runtimeCfg:   runtimeCfg,
	}
	uc.executor = PipelineExecutor{
		DockerEngine:  dockerEngine,
		AppFileRepo:   appFileRepo,
		Filesystem:    filesystem,
		StepRepo:      stepRepo,
		DeployRepo:    deployRepo,
		DockerNetwork: runtimeCfg.GetDockerNetwork(),
		HostBase:      runtimeCfg.GetDockerHostBase(),
	}
	return uc
}

func (u *deploymentUseCase) Deploy(ctx context.Context, appID, deployToken, image string) (*entity.DeploymentResponse, error) {
	app, err := u.appRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("App not found")
	}

	if subtle.ConstantTimeCompare([]byte(app.DeployToken), []byte(deployToken)) != 1 {
		return nil, domainError.ErrUnauthorized.WithCustomMessage("Invalid deploy token")
	}

	// Use app's default image if not provided in the request
	if image == "" {
		image = app.DefaultImage
	}
	if image == "" {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("image is required: provide it in the request or set a default image on the app")
	}

	// If app has bindings and proxy is enabled, use blue/green deployment
	if u.proxyUC != nil {
		bindings, _ := u.bindingRepo.FindByAppID(ctx, appID)
		if len(bindings) > 0 {
			authRegistry, authUser, authPass := u.resolveRegistryAuth(ctx, app.ProjectID, image)
			return u.proxyUC.DeployBlueGreen(ctx, appID, image, authRegistry, authUser, authPass)
		}
	}

	d := entity.NewDeployment(appID, image)
	if err := u.deployRepo.Create(ctx, d); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	template := entity.DefaultPipelineTemplate(app.FrameworkPreset)
	template = MergeAppConfig(template, app)
	template = MergeAppFiles(template, app, u.runtimeCfg.GetDockerHostBase())
	steps := make([]*entity.PipelineStep, 0, len(template.Steps))
	for _, def := range template.Steps {
		steps = append(steps, entity.NewPipelineStep(d.ID, def.Name, def.Order, def.Config))
	}
	if err := u.stepRepo.CreateBatch(ctx, steps); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	authRegistry, authUser, authPass := u.resolveRegistryAuth(ctx, app.ProjectID, image)

	go u.executePipeline(d.ID, appID, image, authRegistry, authUser, authPass, steps, app.ContainerName)

	stepResponses := make([]*entity.PipelineStepResponse, 0, len(steps))
	for _, s := range steps {
		stepResponses = append(stepResponses, s.ToResponse())
	}
	return d.ToResponse(stepResponses), nil
}

func (u *deploymentUseCase) DeployApp(ctx context.Context, appID, image string) (*entity.DeploymentResponse, error) {
	app, err := u.appRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("App not found")
	}

	if image == "" {
		image = app.DefaultImage
	}
	if image == "" {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("image is required: provide it in the request or set a default image on the app")
	}

	if u.proxyUC != nil {
		bindings, _ := u.bindingRepo.FindByAppID(ctx, appID)
		if len(bindings) > 0 {
			authRegistry, authUser, authPass := u.resolveRegistryAuth(ctx, app.ProjectID, image)
			return u.proxyUC.DeployBlueGreen(ctx, appID, image, authRegistry, authUser, authPass)
		}
	}

	d := entity.NewDeployment(appID, image)
	if err := u.deployRepo.Create(ctx, d); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	template := entity.DefaultPipelineTemplate(app.FrameworkPreset)
	template = MergeAppConfig(template, app)
	template = MergeAppFiles(template, app, u.runtimeCfg.GetDockerHostBase())
	steps := make([]*entity.PipelineStep, 0, len(template.Steps))
	for _, def := range template.Steps {
		steps = append(steps, entity.NewPipelineStep(d.ID, def.Name, def.Order, def.Config))
	}
	if err := u.stepRepo.CreateBatch(ctx, steps); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	authRegistry, authUser, authPass := u.resolveRegistryAuth(ctx, app.ProjectID, image)

	go u.executePipeline(d.ID, appID, image, authRegistry, authUser, authPass, steps, app.ContainerName)

	stepResponses := make([]*entity.PipelineStepResponse, 0, len(steps))
	for _, s := range steps {
		stepResponses = append(stepResponses, s.ToResponse())
	}
	return d.ToResponse(stepResponses), nil
}

func (u *deploymentUseCase) ListByAppID(ctx context.Context, appID string) ([]*entity.DeploymentResponse, error) {
	deployments, err := u.deployRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	responses := make([]*entity.DeploymentResponse, 0, len(deployments))
	for _, d := range deployments {
		responses = append(responses, d.ToResponse())
	}
	return responses, nil
}

func (u *deploymentUseCase) Get(ctx context.Context, deploymentID string) (*entity.DeploymentResponse, error) {
	id, err := strconv.ParseUint(deploymentID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid deployment ID")
	}
	d, err := u.deployRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("Deployment not found")
	}
	steps, _ := u.stepRepo.FindByDeploymentID(ctx, d.ID)
	stepResponses := make([]*entity.PipelineStepResponse, 0, len(steps))
	for _, s := range steps {
		stepResponses = append(stepResponses, s.ToResponse())
	}
	return d.ToResponse(stepResponses), nil
}

func (u *deploymentUseCase) resolveRegistryAuth(ctx context.Context, projectID uint, image string) (string, string, string) {
	registryHost := extractRegistryHost(image)

	projectCreds, err := u.credRepo.FindByProjectID(ctx, projectID)
	if err == nil {
		for _, cred := range projectCreds {
			if matchesRegistry(cred.RegistryURL, registryHost) {
				return cred.RegistryURL, cred.Username, cred.Password
			}
		}
	}

	globalCreds, err := u.credRepo.FindGlobal(ctx)
	if err == nil {
		for _, cred := range globalCreds {
			if matchesRegistry(cred.RegistryURL, registryHost) {
				return cred.RegistryURL, cred.Username, cred.Password
			}
		}
	}

	return "", "", ""
}

func extractRegistryHost(image string) string {
	parts := strings.SplitN(image, "/", 2)
	if len(parts) > 1 && strings.Contains(parts[0], ".") {
		return parts[0]
	}
	return ""
}

func matchesRegistry(credentialURL, imageRegistry string) bool {
	if imageRegistry == "" {
		lower := strings.ToLower(credentialURL)
		return lower == "" || lower == "docker.io" || strings.HasPrefix(lower, "index.docker.io") || strings.HasPrefix(lower, "registry-1.docker.io")
	}
	return strings.EqualFold(credentialURL, imageRegistry) || strings.HasPrefix(strings.ToLower(credentialURL), strings.ToLower(imageRegistry))
}

func (u *deploymentUseCase) executePipeline(deploymentID uint, appID, image, authRegistry, authUser, authPass string, steps []*entity.PipelineStep, containerNameOverride string) {
	ctx := context.Background()

	d, err := u.deployRepo.FindByID(ctx, deploymentID)
	if err != nil {
		logw.Errorf("deployment %d: failed to load: %v", deploymentID, err)
		return
	}

	containerID, containerName, failed := u.executor.RunPipelineSteps(ctx, steps, d, appID, image, authRegistry, authUser, authPass, nil, containerNameOverride)

	if failed {
		if containerID != "" {
			u.dockerEngine.StopContainer(ctx, containerID, 10)
			u.dockerEngine.RemoveContainer(ctx, containerID, true)
		}
		return
	}

	d.Status = entity.DeploymentRunning
	d.ContainerID = containerID
	d.ContainerName = containerName
	u.deployRepo.Update(ctx, d)

	running, _ := u.deployRepo.FindRunningByAppID(ctx, appID)
	if app, err := u.appRepo.FindByAppID(ctx, appID); err == nil {
		app.ContainerCount = len(running)
		u.appRepo.Update(ctx, app)
	}
}

func (u *deploymentUseCase) GetContainerLogs(ctx context.Context, appID string, tail int) (string, error) {
	running, err := u.deployRepo.FindRunningByAppID(ctx, appID)
	if err != nil {
		return "", domainError.ErrNotFound.WithCustomMessage("No running container found for this app")
	}
	if len(running) == 0 {
		return "", domainError.ErrNotFound.WithCustomMessage("No running container found for this app")
	}

	containerID := running[0].ContainerID
	if containerID == "" {
		return "", domainError.ErrNotFound.WithCustomMessage("Container not available")
	}

	tailStr := strconv.Itoa(tail)
	reader, err := u.dockerEngine.ContainerLogs(ctx, containerID, outbound.ContainerLogsOptions{
		Tail:       tailStr,
		Timestamps: true,
	})
	if err != nil {
		return "", domainError.ErrInternalServer.WithError(err)
	}
	defer reader.Close()

	logBytes, err := io.ReadAll(reader)
	if err != nil {
		return "", domainError.ErrInternalServer.WithError(err)
	}

	return string(logBytes), nil
}

func (u *deploymentUseCase) GetRunningContainerID(ctx context.Context, appID string) (string, error) {
	running, err := u.deployRepo.FindRunningByAppID(ctx, appID)
	if err != nil {
		return "", domainError.ErrNotFound.WithCustomMessage("No running container found for this app")
	}
	if len(running) == 0 {
		return "", domainError.ErrNotFound.WithCustomMessage("No running container found for this app")
	}
	containerID := running[0].ContainerID
	if containerID == "" {
		return "", domainError.ErrNotFound.WithCustomMessage("Container not available")
	}
	return containerID, nil
}

func (u *deploymentUseCase) ExecInteractive(ctx context.Context, appID string, containerID string, shell string) (net.Conn, io.Reader, error) {
	conn, reader, err := u.dockerEngine.ExecInteractive(ctx, containerID, shell)
	if err != nil {
		return nil, nil, err
	}
	return conn, reader, nil
}

func (u *deploymentUseCase) StopContainer(ctx context.Context, appID string) error {
	running, err := u.deployRepo.FindRunningByAppID(ctx, appID)
	if err != nil {
		return domainError.ErrNotFound.WithCustomMessage("No running container found for this app")
	}
	if len(running) == 0 {
		return nil
	}

	for _, d := range running {
		if d.ContainerID != "" {
			u.dockerEngine.StopContainer(ctx, d.ContainerID, 10)
			u.dockerEngine.RemoveContainer(ctx, d.ContainerID, true)
		}
		d.Status = entity.DeploymentStopped
		d.ContainerID = ""
		d.ContainerName = ""
		u.deployRepo.Update(ctx, d)
	}

	if app, err := u.appRepo.FindByAppID(ctx, appID); err == nil {
		app.ContainerCount = 0
		u.appRepo.Update(ctx, app)
	}
	return nil
}

// --- Merge helper functions (kept for backward compat, delegate to shared) ---

func (u *deploymentUseCase) mergeAppConfig(template *entity.PipelineTemplateDefinition, app *entity.App) *entity.PipelineTemplateDefinition {
	return MergeAppConfig(template, app)
}

func (u *deploymentUseCase) mergeAppFiles(template *entity.PipelineTemplateDefinition, app *entity.App) *entity.PipelineTemplateDefinition {
	return MergeAppFiles(template, app, u.runtimeCfg.GetDockerHostBase())
}

func mergeEnvVarsIntoConfig(configJSON string, envVars entity.StringMap) string {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return configJSON
	}
	existingEnv, _ := cfg["env"].(map[string]interface{})
	if existingEnv == nil {
		existingEnv = make(map[string]interface{})
	}
	for k, v := range envVars {
		existingEnv[k] = v
	}
	cfg["env"] = existingEnv
	result, err := json.Marshal(cfg)
	if err != nil {
		return configJSON
	}
	return string(result)
}

func mergeVolumeMountsIntoConfig(configJSON string, mounts entity.VolumeMountList) string {
	if len(mounts) == 0 {
		return configJSON
	}
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return configJSON
	}
	var existingMounts []interface{}
	if raw, ok := cfg["volume_mounts"]; ok {
		existingMounts, _ = raw.([]interface{})
	}
	for _, m := range mounts {
		existingMounts = append(existingMounts, map[string]interface{}{
			"host_path":      m.HostPath,
			"container_path": m.ContainerPath,
			"mode":           m.Mode,
		})
	}
	cfg["volume_mounts"] = existingMounts
	result, err := json.Marshal(cfg)
	if err != nil {
		return configJSON
	}
	return string(result)
}

func mergePortIntoConfig(configJSON string, containerPort, publishPort string) string {
	var cfg map[string]interface{}
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return configJSON
	}
	if containerPort != "" {
		cfg["port"] = containerPort
	}
	cfg["container_port"] = containerPort
	cfg["publish_port"] = publishPort
	result, err := json.Marshal(cfg)
	if err != nil {
		return configJSON
	}
	return string(result)
}