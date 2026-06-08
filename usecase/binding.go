package usecase

import (
	"context"
	"fmt"

	"github.com/AndreeJait/server-management-be/config"
	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/inbound/binding"
	"github.com/AndreeJait/server-management-be/port/outbound"
)

type bindingUseCase struct {
	bindingRepo  outbound.AppBindingRepository
	appRepo      outbound.AppRepository
	cf           outbound.Cloudflare
	proxyEngine  outbound.ProxyEngine
	runtimeCfg   *config.RuntimeConfig
	deployRepo   outbound.DeploymentRepository
	dockerEngine outbound.DockerEngine
}

func NewBindingUseCase(bindingRepo outbound.AppBindingRepository, appRepo outbound.AppRepository, cf outbound.Cloudflare, proxyEngine outbound.ProxyEngine, runtimeCfg *config.RuntimeConfig, deployRepo outbound.DeploymentRepository, dockerEngine outbound.DockerEngine) binding.UseCase {
	return &bindingUseCase{
		bindingRepo:  bindingRepo,
		appRepo:      appRepo,
		cf:           cf,
		proxyEngine:  proxyEngine,
		runtimeCfg:   runtimeCfg,
		deployRepo:   deployRepo,
		dockerEngine: dockerEngine,
	}
}

func (u *bindingUseCase) Create(ctx context.Context, appID, zoneID, domain, tunnelID, service string) (*entity.AppBindingResponse, error) {
	app, err := u.appRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, fmt.Errorf("app not found: %w", err)
	}

	// Use the BE's tunnel service URL as default (tunnel routes to BE, BE proxies to containers)
	if service == "" {
		service = u.runtimeCfg.GetTunnelServiceURL()
	}

	dnsRecord, err := u.cf.CreateDNSRecord(ctx, zoneID, "CNAME", domain, tunnelID+".cfargotunnel.com", 1, true)
	if err != nil {
		return nil, fmt.Errorf("failed to create DNS record: %w", err)
	}

	tunnelConfig, err := u.cf.GetTunnelConfig(ctx, tunnelID)
	if err != nil {
		_ = u.cf.DeleteDNSRecord(ctx, zoneID, dnsRecord.ID)
		return nil, fmt.Errorf("failed to get tunnel config: %w", err)
	}

	newRule := entity.TunnelIngressRule{
		Hostname: domain,
		Service:  service,
	}

	var ingress []entity.TunnelIngressRule
	if len(tunnelConfig.Ingress) > 0 {
		ingress = append(ingress, tunnelConfig.Ingress[:len(tunnelConfig.Ingress)-1]...)
		ingress = append(ingress, newRule)
		ingress = append(ingress, tunnelConfig.Ingress[len(tunnelConfig.Ingress)-1])
	} else {
		ingress = append(ingress, newRule)
	}

	updatedConfig := &entity.CloudflareTunnelConfig{Ingress: ingress}
	if err := u.cf.UpdateTunnelConfig(ctx, tunnelID, updatedConfig); err != nil {
		_ = u.cf.DeleteDNSRecord(ctx, zoneID, dnsRecord.ID)
		return nil, fmt.Errorf("failed to update tunnel config: %w", err)
	}

	b := entity.NewAppBinding(appID, zoneID, dnsRecord.ID, domain, tunnelID)
	if err := u.bindingRepo.Create(ctx, b); err != nil {
		_ = u.cf.DeleteDNSRecord(ctx, zoneID, dnsRecord.ID)
		_ = u.cf.UpdateTunnelConfig(ctx, tunnelID, tunnelConfig)
		return nil, fmt.Errorf("failed to save binding: %w", err)
	}

	if u.proxyEngine != nil {
		u.proxyEngine.RegisterDomain(appID, domain)

		// Populate proxy route immediately if app has running containers
		running, _ := u.deployRepo.FindRunningByAppID(ctx, appID)
		if len(running) > 0 && running[0].ContainerID != "" {
			info, inspectErr := u.dockerEngine.InspectContainer(ctx, running[0].ContainerID)
			if inspectErr == nil && info.Running {
				containerPort := app.ContainerPort
				if containerPort == "" {
					containerPort = defaultPortForPreset(app.FrameworkPreset)
				}
				targetAddr := fmt.Sprintf("%s:%s", info.IP, containerPort)
				if info.IP == "" {
					targetAddr = fmt.Sprintf("localhost:%s", containerPort)
				}
				u.proxyEngine.UpdateRoute(appID, targetAddr, "", 0)
			}
		}
	}

	return b.ToResponse(), nil
}

func (u *bindingUseCase) List(ctx context.Context, appID string) ([]*entity.AppBindingResponse, error) {
	bindings, err := u.bindingRepo.FindByAppID(ctx, appID)
	if err != nil {
		return nil, err
	}
	result := make([]*entity.AppBindingResponse, len(bindings))
	for i, b := range bindings {
		result[i] = b.ToResponse()
	}
	return result, nil
}

func (u *bindingUseCase) Get(ctx context.Context, bindingID uint) (*entity.AppBindingResponse, error) {
	b, err := u.bindingRepo.FindByID(ctx, bindingID)
	if err != nil {
		return nil, err
	}
	return b.ToResponse(), nil
}

func (u *bindingUseCase) Delete(ctx context.Context, bindingID uint) error {
	b, err := u.bindingRepo.FindByID(ctx, bindingID)
	if err != nil {
		return err
	}

	if err := u.cf.DeleteDNSRecord(ctx, b.ZoneID, b.DNSRecordID); err != nil {
		return fmt.Errorf("failed to delete DNS record: %w", err)
	}

	tunnelConfig, err := u.cf.GetTunnelConfig(ctx, b.TunnelID)
	if err != nil {
		return fmt.Errorf("failed to get tunnel config: %w", err)
	}

	var ingress []entity.TunnelIngressRule
	for _, rule := range tunnelConfig.Ingress {
		if rule.Hostname != b.Domain {
			ingress = append(ingress, rule)
		}
	}

	if err := u.cf.UpdateTunnelConfig(ctx, b.TunnelID, &entity.CloudflareTunnelConfig{Ingress: ingress}); err != nil {
		return fmt.Errorf("failed to update tunnel config: %w", err)
	}

	if err := u.bindingRepo.Delete(ctx, bindingID); err != nil {
		return err
	}

	if u.proxyEngine != nil {
		u.proxyEngine.RemoveRoute(b.AppID)
	}

	return nil
}

func defaultPortForPreset(preset entity.FrameworkPreset) string {
	switch preset {
	case entity.FrameworkNextjs, entity.FrameworkNuxt, entity.FrameworkAstro,
		entity.FrameworkReact, entity.FrameworkVue, entity.FrameworkSvelte, entity.FrameworkRemix:
		return "3000"
	case entity.FrameworkPython, entity.FrameworkNodejs,
		entity.FrameworkLaravel, entity.FrameworkRails:
		return "8000"
	default:
		return "8080"
	}
}