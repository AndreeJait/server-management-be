package usecase

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/proxy"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/logw"
)

type proxyUseCase struct {
	proxyStateRepo outbound.ProxyStateRepository
	appRepo        outbound.AppRepository
	bindingRepo    outbound.AppBindingRepository
	deployRepo     outbound.DeploymentRepository
	stepRepo       outbound.PipelineStepRepository
	credRepo       outbound.RegistryCredentialRepository
	dockerEngine   outbound.DockerEngine
	appFileRepo    outbound.AppFileRepository
	filesystem     outbound.Filesystem
	proxyEngine    outbound.ProxyEngine
	executor       PipelineExecutor
	shiftInterval  time.Duration
}

func NewProxyUseCase(
	proxyStateRepo outbound.ProxyStateRepository,
	appRepo outbound.AppRepository,
	bindingRepo outbound.AppBindingRepository,
	deployRepo outbound.DeploymentRepository,
	stepRepo outbound.PipelineStepRepository,
	credRepo outbound.RegistryCredentialRepository,
	dockerEngine outbound.DockerEngine,
	appFileRepo outbound.AppFileRepository,
	filesystem outbound.Filesystem,
	proxyEngine outbound.ProxyEngine,
	shiftIntervalSec int,
	dockerNetwork string,
) proxy.UseCase {
	if shiftIntervalSec <= 0 {
		shiftIntervalSec = 30
	}
	uc := &proxyUseCase{
		proxyStateRepo: proxyStateRepo,
		appRepo:        appRepo,
		bindingRepo:    bindingRepo,
		deployRepo:     deployRepo,
		stepRepo:       stepRepo,
		credRepo:       credRepo,
		dockerEngine:   dockerEngine,
		appFileRepo:    appFileRepo,
		filesystem:     filesystem,
		proxyEngine:    proxyEngine,
		shiftInterval:  time.Duration(shiftIntervalSec) * time.Second,
	}
	uc.executor = PipelineExecutor{
		DockerEngine:  dockerEngine,
		AppFileRepo:   appFileRepo,
		Filesystem:    filesystem,
		StepRepo:      stepRepo,
		DeployRepo:    deployRepo,
		DockerNetwork: dockerNetwork,
	}
	return uc
}

func (u *proxyUseCase) DeployBlueGreen(ctx context.Context, appID, image, authRegistry, authUser, authPass string) (*entity.DeploymentResponse, error) {
	app, err := u.appRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("App not found")
	}

	// Use app's default image if not provided
	if image == "" {
		image = app.DefaultImage
	}
	if image == "" {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("image is required: provide it in the request or set a default image on the app")
	}

	ps, err := u.proxyStateRepo.FindByAppID(ctx, appID)
	if err != nil {
		ps = entity.NewProxyState(appID)
		if err := u.proxyStateRepo.Create(ctx, ps); err != nil {
			return nil, domainError.ErrInternalServer.WithError(err)
		}
	}

	targetSlot := ps.InactiveSlot()

	d := entity.NewDeployment(appID, image)
	if err := u.deployRepo.Create(ctx, d); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	ps.Status = entity.ProxyStateDeploying
	ps.SetSlotContainerID(targetSlot, "")
	u.proxyStateRepo.Update(ctx, ps)

	template := entity.DefaultPipelineTemplate(app.FrameworkPreset)
	template = MergeAppConfig(template, app)
	template = MergeAppFiles(template, app)

	steps := make([]*entity.PipelineStep, 0, len(template.Steps))
	for _, def := range template.Steps {
		steps = append(steps, entity.NewPipelineStep(d.ID, def.Name, def.Order, def.Config))
	}
	if err := u.stepRepo.CreateBatch(ctx, steps); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	go u.executeBlueGreenPipeline(d.ID, appID, image, authRegistry, authUser, authPass, steps, ps, targetSlot, app.ContainerName, app.ContainerPort)

	stepResponses := make([]*entity.PipelineStepResponse, 0, len(steps))
	for _, s := range steps {
		stepResponses = append(stepResponses, s.ToResponse())
	}
	return d.ToResponse(stepResponses), nil
}

func (u *proxyUseCase) executeBlueGreenPipeline(
	deploymentID uint, appID, image, authRegistry, authUser, authPass string,
	steps []*entity.PipelineStep, ps *entity.ProxyState, targetSlot entity.SlotName,
	containerNameOverride, appContainerPort string,
) {
	ctx := context.Background()

	d, err := u.deployRepo.FindByID(ctx, deploymentID)
	if err != nil {
		logw.Errorf("blue-green deployment %d: failed to load: %v", deploymentID, err)
		return
	}

	labels := map[string]string{
		"com.server-management.app-id":        appID,
		"com.server-management.slot":          string(targetSlot),
		"com.server-management.deployment-id":  fmt.Sprintf("%d", deploymentID),
	}

	containerID, containerName, failed := u.executor.RunPipelineSteps(ctx, steps, d, appID, image, authRegistry, authUser, authPass, labels, containerNameOverride)

	if failed {
		if containerID != "" {
			u.dockerEngine.StopContainer(ctx, containerID, 10)
			u.dockerEngine.RemoveContainer(ctx, containerID, true)
		}
		u.rollback(ctx, ps)
		return
	}

	d.Status = entity.DeploymentRunning
	d.ContainerID = containerID
	d.ContainerName = containerName
	u.deployRepo.Update(ctx, d)

	info, err := u.dockerEngine.InspectContainer(ctx, containerID)
	if err != nil {
		logw.Errorf("blue-green: failed to inspect new container %s: %v", containerID, err)
		u.dockerEngine.StopContainer(ctx, containerID, 10)
		u.dockerEngine.RemoveContainer(ctx, containerID, true)
		u.rollback(ctx, ps)
		return
	}

	containerPort := "8080"
	if appContainerPort != "" {
		containerPort = appContainerPort
	} else {
		for _, p := range info.Ports {
			containerPort = p
			break
		}
	}

	targetAddr := fmt.Sprintf("%s:%s", info.IP, containerPort)
	if info.IP == "" {
		targetAddr = fmt.Sprintf("localhost:%s", containerPort)
	}

	ps.SetSlotContainerID(targetSlot, containerID)
	ps.SetSlotTarget(targetSlot, targetAddr)
	ps.Status = entity.ProxyStateShifting
	ps.TrafficPercent = 0
	u.proxyStateRepo.Update(ctx, ps)
	u.proxyEngine.UpdateRoute(appID, ps.BlueTarget, ps.GreenTarget, ps.TrafficPercent)

	u.executeTrafficShift(ctx, ps, targetSlot)
}

func (u *proxyUseCase) executeTrafficShift(ctx context.Context, ps *entity.ProxyState, targetSlot entity.SlotName) {
	shiftSteps := []int{25, 50, 75, 100}

	for _, percent := range shiftSteps {
		time.Sleep(u.shiftInterval)

		containerID := ps.GetSlotContainerID(targetSlot)
		if containerID == "" {
			logw.Errorf("blue-green: lost container for slot %s during shift", targetSlot)
			u.rollback(ctx, ps)
			return
		}

		info, err := u.dockerEngine.InspectContainer(ctx, containerID)
		if err != nil || !info.Running {
			logw.Errorf("blue-green: container %s unhealthy during shift: %v", containerID, err)
			u.rollback(ctx, ps)
			return
		}

		targetAddr := ps.GetSlotTarget(targetSlot)
		if !u.httpHealthCheck(ctx, targetAddr, ps.HealthCheckPath) {
			logw.Errorf("blue-green: HTTP health check failed for %s%s before shifting to %d%%", targetAddr, ps.HealthCheckPath, percent)
			u.rollback(ctx, ps)
			return
		}

		ps.TrafficPercent = percent
		u.proxyStateRepo.Update(ctx, ps)
		u.proxyEngine.UpdateRoute(ps.AppID, ps.BlueTarget, ps.GreenTarget, ps.TrafficPercent)
		logw.Infof("blue-green: app=%s traffic shifted to %d%% on slot %s", ps.AppID, percent, targetSlot)
	}

	oldContainerID := ps.GetSlotContainerID(ps.ActiveSlot)
	ps.ActiveSlot = targetSlot
	ps.TrafficPercent = 100
	ps.Status = entity.ProxyStateActive
	u.proxyStateRepo.Update(ctx, ps)
	u.proxyEngine.UpdateRoute(ps.AppID, ps.BlueTarget, ps.GreenTarget, ps.TrafficPercent)

	if oldContainerID != "" {
		logw.Infof("blue-green: stopping old container %s for app=%s", oldContainerID, ps.AppID)
		u.dockerEngine.StopContainer(ctx, oldContainerID, 10)
		u.dockerEngine.RemoveContainer(ctx, oldContainerID, false)

		oldSlot := ps.InactiveSlot()
		ps.SetSlotContainerID(oldSlot, "")
		ps.SetSlotTarget(oldSlot, "")
		u.proxyStateRepo.Update(ctx, ps)
		u.proxyEngine.UpdateRoute(ps.AppID, ps.BlueTarget, ps.GreenTarget, ps.TrafficPercent)
	}

	running, _ := u.deployRepo.FindRunningByAppID(ctx, ps.AppID)
	if app, err := u.appRepo.FindByAppID(ctx, ps.AppID); err == nil {
		app.ContainerCount = len(running)
		u.appRepo.Update(ctx, app)
	}

	logw.Infof("blue-green: deployment complete for app=%s, active slot=%s", ps.AppID, targetSlot)
}

func (u *proxyUseCase) httpHealthCheck(ctx context.Context, targetAddr, path string) bool {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	url := fmt.Sprintf("http://%s%s", targetAddr, path)

	client := &http.Client{Timeout: 5 * time.Second}
	resp, err := client.Get(url)
	if err != nil {
		return false
	}
	defer resp.Body.Close()
	return resp.StatusCode >= 200 && resp.StatusCode < 400
}

func (u *proxyUseCase) rollback(ctx context.Context, ps *entity.ProxyState) {
	logw.Warningf("blue-green: rolling back app=%s", ps.AppID)

	ps.TrafficPercent = 0
	ps.Status = entity.ProxyStateRollingBack
	u.proxyStateRepo.Update(ctx, ps)
	u.proxyEngine.UpdateRoute(ps.AppID, ps.BlueTarget, ps.GreenTarget, ps.TrafficPercent)

	time.Sleep(5 * time.Second)

	targetSlot := ps.InactiveSlot()
	newContainerID := ps.GetSlotContainerID(targetSlot)
	if newContainerID != "" {
		u.dockerEngine.StopContainer(ctx, newContainerID, 10)
		u.dockerEngine.RemoveContainer(ctx, newContainerID, true)
		ps.SetSlotContainerID(targetSlot, "")
		ps.SetSlotTarget(targetSlot, "")
	}

	ps.Status = entity.ProxyStateActive
	ps.TrafficPercent = 100
	u.proxyStateRepo.Update(ctx, ps)
	u.proxyEngine.UpdateRoute(ps.AppID, ps.BlueTarget, ps.GreenTarget, ps.TrafficPercent)
}

func (u *proxyUseCase) GetProxyState(ctx context.Context, appID string) (*entity.ProxyStateResponse, error) {
	ps, err := u.proxyStateRepo.FindByAppID(ctx, appID)
	if err != nil {
		// Auto-create idle proxy state for apps that haven't been deployed yet
		if _, appErr := u.appRepo.FindByAppID(ctx, appID); appErr != nil {
			return nil, domainError.ErrNotFound.WithCustomMessage("App not found")
		}
		ps = entity.NewProxyState(appID)
		if createErr := u.proxyStateRepo.Create(ctx, ps); createErr != nil {
			return nil, domainError.ErrInternalServer.WithError(createErr)
		}
	}
	return ps.ToResponse(), nil
}

func (u *proxyUseCase) ListProxyStates(ctx context.Context) ([]*entity.ProxyStateResponse, error) {
	states, err := u.proxyStateRepo.FindAll(ctx)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	responses := make([]*entity.ProxyStateResponse, 0, len(states))
	for _, ps := range states {
		responses = append(responses, ps.ToResponse())
	}
	return responses, nil
}

func (u *proxyUseCase) SetTraffic(ctx context.Context, appID string, percent int) (*entity.ProxyStateResponse, error) {
	if percent < 0 || percent > 100 {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Traffic percent must be 0-100")
	}
	ps, err := u.proxyStateRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("Proxy state not found for app")
	}
	ps.TrafficPercent = percent
	u.proxyStateRepo.Update(ctx, ps)
	u.proxyEngine.UpdateRoute(appID, ps.BlueTarget, ps.GreenTarget, ps.TrafficPercent)
	return ps.ToResponse(), nil
}

func (u *proxyUseCase) Rollback(ctx context.Context, appID string) (*entity.ProxyStateResponse, error) {
	ps, err := u.proxyStateRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("Proxy state not found for app")
	}
	if ps.Status != entity.ProxyStateShifting && ps.Status != entity.ProxyStateDeploying {
		return nil, domainError.ErrConflict.WithCustomMessage("Can only rollback during deploying or shifting states")
	}
	u.rollback(ctx, ps)
	ps, _ = u.proxyStateRepo.FindByAppID(ctx, appID)
	return ps.ToResponse(), nil
}