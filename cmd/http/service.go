package main

import (
	"context"

	"github.com/AndreeJait/server-management-be/adapter/outbound"
	"github.com/AndreeJait/server-management-be/config"
	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/inbound/app"
	"github.com/AndreeJait/server-management-be/port/inbound/appfile"
	"github.com/AndreeJait/server-management-be/port/inbound/auth"
	bindingInbound "github.com/AndreeJait/server-management-be/port/inbound/binding"
	cloudflareInbound "github.com/AndreeJait/server-management-be/port/inbound/cloudflare"
	"github.com/AndreeJait/server-management-be/port/inbound/deployment"
	"github.com/AndreeJait/server-management-be/port/inbound/health"
	proxyInbound "github.com/AndreeJait/server-management-be/port/inbound/proxy"
	"github.com/AndreeJait/server-management-be/port/inbound/project"
	"github.com/AndreeJait/server-management-be/port/inbound/registry"
	"github.com/AndreeJait/server-management-be/port/inbound/role"
	"github.com/AndreeJait/server-management-be/port/inbound/user"
	portOutbound "github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/server-management-be/usecase"
	"github.com/AndreeJait/go-utility/v2/authw"
	"github.com/AndreeJait/go-utility/v2/jwtw"
	"github.com/AndreeJait/go-utility/v2/logw"
	"go.uber.org/dig"
)

// provideServices registers repository and use case providers into the dig container.
func provideServices(c *dig.Container) {
	// kyan:provider:start
	c.Provide(newHealthRepository)
	c.Provide(newHealthUseCase)
	c.Provide(newUserRepository)
	c.Provide(newRoleRepository)
	c.Provide(newAuthUseCase)
	c.Provide(newUserUseCase)
	c.Provide(newRoleUseCase)
	c.Provide(newProjectRepository)
	c.Provide(newAppRepository)
	c.Provide(newRegistryCredentialRepository)
	c.Provide(newProjectUseCase)
	c.Provide(newAppUseCase)
	c.Provide(newRegistryUseCase)
	c.Provide(newDeploymentRepository)
	c.Provide(newPipelineStepRepository)
	c.Provide(newAppFileRepository)
	c.Provide(newFilesystem)
	c.Provide(newAppFileUseCase)
	c.Provide(newCloudflareClient)
	c.Provide(newAppBindingRepository)
	c.Provide(newCloudflareUseCase)
	c.Provide(newProxyStateRepository)
	c.Provide(newProxyEngine)
	c.Provide(newProxyUseCase)
	c.Provide(newDeploymentUseCase)
	c.Provide(newBindingUseCase)
	// kyan:provider:end
}

func newHealthRepository(db *outbound.DB, redisConn *outbound.RedisConn) portOutbound.HealthRepository {
	return outbound.NewHealthRepository(db, redisConn)
}

func newHealthUseCase(cfg *config.AppConfig, repo portOutbound.HealthRepository) health.UseCase {
	return usecase.NewHealthUseCase(cfg.App.Name, repo)
}

func newUserRepository(db *outbound.DB) portOutbound.UserRepository {
	return outbound.NewUserRepository(db)
}

func newRoleRepository(db *outbound.DB) portOutbound.RoleRepository {
	return outbound.NewRoleRepository(db)
}

func newAuthUseCase(userRepo portOutbound.UserRepository, j jwtw.JWT, rbac *authw.RBAC, cfg *config.AppConfig) auth.UseCase {
	return usecase.NewAuthUseCase(userRepo, j, cfg.Auth.JWTTTL, rbac)
}

func newUserUseCase(userRepo portOutbound.UserRepository, rbac *authw.RBAC) user.UseCase {
	return usecase.NewUserUseCase(userRepo, rbac)
}

func newRoleUseCase(roleRepo portOutbound.RoleRepository, userRepo portOutbound.UserRepository, rbac *authw.RBAC) role.UseCase {
	return usecase.NewRoleUseCase(roleRepo, userRepo, rbac)
}

func newProjectRepository(db *outbound.DB) portOutbound.ProjectRepository {
	return outbound.NewProjectRepository(db)
}

func newAppRepository(db *outbound.DB) portOutbound.AppRepository {
	return outbound.NewAppRepository(db)
}

func newRegistryCredentialRepository(db *outbound.DB) portOutbound.RegistryCredentialRepository {
	return outbound.NewRegistryCredentialRepository(db)
}

func newProjectUseCase(projectRepo portOutbound.ProjectRepository) project.UseCase {
	return usecase.NewProjectUseCase(projectRepo)
}

func newAppUseCase(
	appRepo portOutbound.AppRepository,
	projectRepo portOutbound.ProjectRepository,
	deployRepo portOutbound.DeploymentRepository,
	stepRepo portOutbound.PipelineStepRepository,
	bindingRepo portOutbound.AppBindingRepository,
	appFileRepo portOutbound.AppFileRepository,
	proxyStateRepo portOutbound.ProxyStateRepository,
	dockerEngine portOutbound.DockerEngine,
	filesystem portOutbound.Filesystem,
	cf portOutbound.Cloudflare,
	proxyEngine portOutbound.ProxyEngine,
) app.UseCase {
	return usecase.NewAppUseCase(appRepo, projectRepo, deployRepo, stepRepo, bindingRepo, appFileRepo, proxyStateRepo, dockerEngine, filesystem, cf, proxyEngine)
}

func newRegistryUseCase(credRepo portOutbound.RegistryCredentialRepository) registry.UseCase {
	return usecase.NewRegistryUseCase(credRepo)
}

func newDeploymentRepository(db *outbound.DB) portOutbound.DeploymentRepository {
	return outbound.NewDeploymentRepository(db)
}

func newPipelineStepRepository(db *outbound.DB) portOutbound.PipelineStepRepository {
	return outbound.NewPipelineStepRepository(db)
}

func newDeploymentUseCase(
	appRepo portOutbound.AppRepository,
	projectRepo portOutbound.ProjectRepository,
	deployRepo portOutbound.DeploymentRepository,
	stepRepo portOutbound.PipelineStepRepository,
	credRepo portOutbound.RegistryCredentialRepository,
	dockerEngine portOutbound.DockerEngine,
	appFileRepo portOutbound.AppFileRepository,
	filesystem portOutbound.Filesystem,
	bindingRepo portOutbound.AppBindingRepository,
	proxyUC proxyInbound.UseCase,
	cfg *config.AppConfig,
) deployment.UseCase {
	return usecase.NewDeploymentUseCase(appRepo, projectRepo, deployRepo, stepRepo, credRepo, dockerEngine, appFileRepo, filesystem, bindingRepo, proxyUC, cfg.Docker.Network)
}

func newAppFileRepository(db *outbound.DB) portOutbound.AppFileRepository {
	return outbound.NewAppFileRepository(db)
}

func newFilesystem() portOutbound.Filesystem {
	return outbound.NewFilesystem()
}

func newAppFileUseCase(appFileRepo portOutbound.AppFileRepository, appRepo portOutbound.AppRepository, deployRepo portOutbound.DeploymentRepository, filesystem portOutbound.Filesystem) appfile.UseCase {
	uc := usecase.NewAppFileUseCase(appFileRepo, appRepo, deployRepo, filesystem)
	return uc
}

func initAppFileDeployFunc(c *dig.Container) {
	var appFileUC appfile.UseCase
	var deployUC deployment.UseCase
	if err := c.Invoke(func(afUC appfile.UseCase, dUC deployment.UseCase) {
		appFileUC = afUC
		deployUC = dUC
	}); err != nil {
		panic(err)
	}
	appFileUC.SetDeployFunc(func(ctx context.Context, appID, deployToken, image string) (*entity.DeploymentResponse, error) {
		return deployUC.Deploy(ctx, appID, deployToken, image)
	})
}

func newCloudflareClient(cfg *config.AppConfig) portOutbound.Cloudflare {
	return outbound.NewCloudflareClient(cfg.Cloudflare.APIToken, cfg.Cloudflare.AccountID)
}

func newAppBindingRepository(db *outbound.DB) portOutbound.AppBindingRepository {
	return outbound.NewAppBindingRepository(db)
}

func newCloudflareUseCase(cf portOutbound.Cloudflare) cloudflareInbound.UseCase {
	return usecase.NewCloudflareUseCase(cf)
}

func newProxyStateRepository(db *outbound.DB) portOutbound.ProxyStateRepository {
	return outbound.NewProxyStateRepository(db)
}

func newProxyEngine() portOutbound.ProxyEngine {
	return outbound.NewProxyEngine()
}

func newProxyUseCase(
	proxyStateRepo portOutbound.ProxyStateRepository,
	appRepo portOutbound.AppRepository,
	bindingRepo portOutbound.AppBindingRepository,
	deployRepo portOutbound.DeploymentRepository,
	stepRepo portOutbound.PipelineStepRepository,
	credRepo portOutbound.RegistryCredentialRepository,
	dockerEngine portOutbound.DockerEngine,
	appFileRepo portOutbound.AppFileRepository,
	filesystem portOutbound.Filesystem,
	proxyEngine portOutbound.ProxyEngine,
	cfg *config.AppConfig,
) proxyInbound.UseCase {
	return usecase.NewProxyUseCase(proxyStateRepo, appRepo, bindingRepo, deployRepo, stepRepo, credRepo, dockerEngine, appFileRepo, filesystem, proxyEngine, cfg.Proxy.ShiftIntervalSec, cfg.Docker.Network)
}

func newBindingUseCase(bindingRepo portOutbound.AppBindingRepository, appRepo portOutbound.AppRepository, cf portOutbound.Cloudflare, proxyEngine portOutbound.ProxyEngine, cfg *config.AppConfig, deployRepo portOutbound.DeploymentRepository, dockerEngine portOutbound.DockerEngine) bindingInbound.UseCase {
	return usecase.NewBindingUseCase(bindingRepo, appRepo, cf, proxyEngine, cfg.Proxy.TunnelServiceURL, deployRepo, dockerEngine)
}

func restoreProxyRoutes(proxyStateRepo portOutbound.ProxyStateRepository, bindingRepo portOutbound.AppBindingRepository, proxyEngine portOutbound.ProxyEngine) {
	ctx := context.Background()
	states, err := proxyStateRepo.FindAll(ctx)
	if err != nil {
		logw.Warningf("proxy: failed to restore routes: %v", err)
		return
	}
	for _, ps := range states {
		bindings, err := bindingRepo.FindByAppID(ctx, ps.AppID)
		if err != nil || len(bindings) == 0 {
			continue
		}
		domain := bindings[0].Domain
		proxyEngine.RegisterDomain(ps.AppID, domain)
		proxyEngine.UpdateRoute(ps.AppID, ps.BlueTarget, ps.GreenTarget, ps.TrafficPercent)
		logw.Infof("proxy: restored route for app=%s domain=%s", ps.AppID, domain)
	}
}

// kyan:service:start
// kyan:service:end