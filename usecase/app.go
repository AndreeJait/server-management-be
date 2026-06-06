package usecase

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"strconv"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/app"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/google/uuid"
)

type appUseCase struct {
	appRepo       outbound.AppRepository
	projectRepo   outbound.ProjectRepository
	deployRepo    outbound.DeploymentRepository
	stepRepo      outbound.PipelineStepRepository
	bindingRepo   outbound.AppBindingRepository
	appFileRepo   outbound.AppFileRepository
	proxyStateRepo outbound.ProxyStateRepository
	dockerEngine  outbound.DockerEngine
	filesystem    outbound.Filesystem
	cf            outbound.Cloudflare
	proxyEngine   outbound.ProxyEngine
}

func NewAppUseCase(
	appRepo outbound.AppRepository,
	projectRepo outbound.ProjectRepository,
	deployRepo outbound.DeploymentRepository,
	stepRepo outbound.PipelineStepRepository,
	bindingRepo outbound.AppBindingRepository,
	appFileRepo outbound.AppFileRepository,
	proxyStateRepo outbound.ProxyStateRepository,
	dockerEngine outbound.DockerEngine,
	filesystem outbound.Filesystem,
	cf outbound.Cloudflare,
	proxyEngine outbound.ProxyEngine,
) app.UseCase {
	return &appUseCase{
		appRepo:       appRepo,
		projectRepo:   projectRepo,
		deployRepo:    deployRepo,
		stepRepo:      stepRepo,
		bindingRepo:   bindingRepo,
		appFileRepo:   appFileRepo,
		proxyStateRepo: proxyStateRepo,
		dockerEngine:  dockerEngine,
		filesystem:    filesystem,
		cf:            cf,
		proxyEngine:  proxyEngine,
	}
}

func generateDeployToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return uuid.New().String() + "-" + hex.EncodeToString(b)
}

func (u *appUseCase) Create(ctx context.Context, projectID uint, name, frameworkPreset string) (*entity.CreateAppResponse, error) {
	_, err := u.projectRepo.FindByID(ctx, projectID)
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("Project not found")
	}

	preset := entity.FrameworkPreset(frameworkPreset)
	if !preset.IsValid() {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid framework preset")
	}

	deployToken := generateDeployToken()
	appID := uuid.New().String()

	a := entity.NewApp(projectID, name, preset, deployToken, appID)
	if err := u.appRepo.Create(ctx, a); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return a.ToCreateResponse(), nil
}

func (u *appUseCase) List(ctx context.Context, projectID uint) ([]*entity.AppResponse, error) {
	apps, err := u.appRepo.FindByProjectID(ctx, projectID)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	responses := make([]*entity.AppResponse, 0, len(apps))
	for _, a := range apps {
		responses = append(responses, a.ToResponse())
	}
	return responses, nil
}

func (u *appUseCase) Get(ctx context.Context, projectID uint, appID string) (*entity.AppResponse, error) {
	id, err := strconv.ParseUint(appID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid app ID")
	}
	a, err := u.appRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("App not found")
	}
	if a.ProjectID != projectID {
		return nil, domainError.ErrForbidden.WithCustomMessage("App does not belong to this project")
	}
	return a.ToResponse(), nil
}

func (u *appUseCase) Update(ctx context.Context, projectID uint, appID string, name, frameworkPreset string, envVars entity.StringMap, volumeMounts entity.VolumeMountList, postDeployCommands entity.StringList, basePath, defaultImage, containerPort, publishPort, containerName string) (*entity.AppResponse, error) {
	id, err := strconv.ParseUint(appID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid app ID")
	}
	a, err := u.appRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("App not found")
	}
	if a.ProjectID != projectID {
		return nil, domainError.ErrForbidden.WithCustomMessage("App does not belong to this project")
	}
	preset := entity.FrameworkPreset(frameworkPreset)
	if !preset.IsValid() {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid framework preset")
	}
	a.Name = name
	a.FrameworkPreset = preset
	a.EnvVars = envVars
	a.VolumeMounts = volumeMounts
	a.PostDeployCommands = postDeployCommands
	if basePath != "" {
		a.BasePath = basePath
	}
	a.DefaultImage = defaultImage
	a.ContainerPort = containerPort
	a.PublishPort = publishPort
	a.ContainerName = containerName
	if err := u.appRepo.Update(ctx, a); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return a.ToResponse(), nil
}

func (u *appUseCase) Delete(ctx context.Context, projectID uint, appID string) error {
	id, err := strconv.ParseUint(appID, 10, 64)
	if err != nil {
		return domainError.ErrInvalidParam.WithCustomMessage("Invalid app ID")
	}
	a, err := u.appRepo.FindByID(ctx, uint(id))
	if err != nil {
		return domainError.ErrNotFound.WithCustomMessage("App not found")
	}
	if a.ProjectID != projectID {
		return domainError.ErrForbidden.WithCustomMessage("App does not belong to this project")
	}

	appIDStr := a.AppID

	// 1. Stop and remove running containers
	running, _ := u.deployRepo.FindRunningByAppID(ctx, appIDStr)
	for _, d := range running {
		if d.ContainerID != "" {
			u.dockerEngine.StopContainer(ctx, d.ContainerID, 10)
			u.dockerEngine.RemoveContainer(ctx, d.ContainerID, true)
		}
	}

	// 2. Clean up Cloudflare bindings (DNS records + tunnel ingress)
	bindings, _ := u.bindingRepo.FindByAppID(ctx, appIDStr)
	for _, b := range bindings {
		_ = u.cf.DeleteDNSRecord(ctx, b.ZoneID, b.DNSRecordID)

		tunnelConfig, tcErr := u.cf.GetTunnelConfig(ctx, b.TunnelID)
		if tcErr == nil {
			var ingress []entity.TunnelIngressRule
			for _, rule := range tunnelConfig.Ingress {
				if rule.Hostname != b.Domain {
					ingress = append(ingress, rule)
				}
			}
			_ = u.cf.UpdateTunnelConfig(ctx, b.TunnelID, &entity.CloudflareTunnelConfig{Ingress: ingress})
		}
	}
	_ = u.bindingRepo.DeleteByAppID(ctx, appIDStr)

	// 3. Remove proxy route
	if u.proxyEngine != nil {
		u.proxyEngine.RemoveRoute(appIDStr)
	}

	// 4. Delete proxy state
	_ = u.proxyStateRepo.DeleteByAppID(ctx, appIDStr)

	// 5. Delete pipeline steps for each deployment
	deployments, _ := u.deployRepo.FindByAppID(ctx, appIDStr)
	for _, d := range deployments {
		_ = u.stepRepo.DeleteByDeploymentID(ctx, d.ID)
	}

	// 6. Delete deployments
	_ = u.deployRepo.DeleteByAppID(ctx, appIDStr)

	// 7. Delete app files
	_ = u.appFileRepo.DeleteByAppID(ctx, appIDStr)

	// 8. Remove filesystem directories
	basePath := a.BasePath
	if basePath == "" {
		basePath = "/home/user/docker"
	}
	_ = u.filesystem.RemoveAll(fmt.Sprintf("%s/%s/files", basePath, appIDStr))
	_ = u.filesystem.RemoveAll(fmt.Sprintf("%s/%s", basePath, appIDStr))

	// 9. Delete the app record
	if err := u.appRepo.Delete(ctx, uint(id)); err != nil {
		return domainError.ErrInternalServer.WithError(err)
	}
	return nil
}

func (u *appUseCase) RegenerateDeployToken(ctx context.Context, projectID uint, appID string) (*entity.RegenerateTokenResponse, error) {
	id, err := strconv.ParseUint(appID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid app ID")
	}
	a, err := u.appRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("App not found")
	}
	if a.ProjectID != projectID {
		return nil, domainError.ErrForbidden.WithCustomMessage("App does not belong to this project")
	}
	a.DeployToken = generateDeployToken()
	if err := u.appRepo.Update(ctx, a); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	return &entity.RegenerateTokenResponse{
		App:         a.ToResponse(),
		DeployToken: a.DeployToken,
	}, nil
}