package usecase

import (
	"context"
	"fmt"
	"strconv"

	"github.com/AndreeJait/server-management-be/config"
	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
)

type configUseCase struct {
	settingRepo      outbound.SettingRepository
	domainCountRepo  outbound.DomainRequestCountRepository
	runtimeCfg       *config.RuntimeConfig
}

func NewConfigUseCase(
	settingRepo outbound.SettingRepository,
	domainCountRepo outbound.DomainRequestCountRepository,
	runtimeCfg *config.RuntimeConfig,
) *configUseCase {
	return &configUseCase{
		settingRepo:     settingRepo,
		domainCountRepo: domainCountRepo,
		runtimeCfg:     runtimeCfg,
	}
}

var hotReloadable = map[string]bool{
	"proxy.health_check_path":  true,
	"proxy.shift_interval_sec": true,
	"proxy.rate_limit_rps":     true,
}

func (u *configUseCase) GetSettings(ctx context.Context) ([]*entity.SettingsGroup, error) {
	settings, err := u.settingRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	return groupSettings(settings), nil
}

func (u *configUseCase) GetSettingsBySection(ctx context.Context, section string) (*entity.SettingsGroup, error) {
	settings, err := u.settingRepo.FindBySection(ctx, section)
	if err != nil {
		return nil, fmt.Errorf("failed to load settings: %w", err)
	}
	groups := groupSettings(settings)
	if len(groups) == 0 {
		return &entity.SettingsGroup{Section: section}, nil
	}
	return groups[0], nil
}

func (u *configUseCase) UpdateSettings(ctx context.Context, updates []entity.UpdateSettingInput) (*entity.UpdateSettingsResult, error) {
	var toSave []*entity.Setting
	var applied []entity.SettingApplied
	var restartRequired []string

	for _, input := range updates {
		if err := validateSetting(input); err != nil {
			return nil, fmt.Errorf("invalid setting %s.%s: %w", input.Section, input.Key, err)
		}

		toSave = append(toSave, &entity.Setting{
			Section: input.Section,
			Key:     input.Key,
			Value:   input.Value,
			Type:    typeForSetting(input.Section, input.Key),
		})

		key := input.Section + "." + input.Key
		isHotReload := hotReloadable[key]

		applied = append(applied, entity.SettingApplied{
			Section:     input.Section,
			Key:         input.Key,
			Value:       input.Value,
			HotReloaded: isHotReload,
		})

		if !isHotReload {
			restartRequired = append(restartRequired, key)
		}
	}

	if err := u.settingRepo.UpsertBatch(ctx, toSave); err != nil {
		return nil, fmt.Errorf("failed to save settings: %w", err)
	}

	for _, input := range updates {
		applySetting(u.runtimeCfg, input.Section, input.Key, input.Value)
	}

	return &entity.UpdateSettingsResult{
		Applied:         applied,
		RestartRequired: restartRequired,
	}, nil
}

func (u *configUseCase) GetDomainRequestCounts(ctx context.Context) ([]*entity.DomainRequestCountResponse, error) {
	counts, err := u.domainCountRepo.FindAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to load access stats: %w", err)
	}
	result := make([]*entity.DomainRequestCountResponse, len(counts))
	for i, c := range counts {
		resp := c.ToResponse()
		result[i] = &resp
	}
	return result, nil
}

func groupSettings(settings []*entity.Setting) []*entity.SettingsGroup {
	sectionMap := make(map[string]*entity.SettingsGroup)
	var order []string
	for _, s := range settings {
		g, ok := sectionMap[s.Section]
		if !ok {
			g = &entity.SettingsGroup{Section: s.Section}
			sectionMap[s.Section] = g
			order = append(order, s.Section)
		}
		g.Settings = append(g.Settings, s.ToResponse())
	}
	result := make([]*entity.SettingsGroup, 0, len(order))
	for _, sec := range order {
		result = append(result, sectionMap[sec])
	}
	return result
}

func validateSetting(input entity.UpdateSettingInput) error {
	typ := typeForSetting(input.Section, input.Key)
	switch typ {
	case "bool":
		if input.Value != "true" && input.Value != "false" {
			return fmt.Errorf("must be 'true' or 'false'")
		}
	case "int":
		if _, err := strconv.Atoi(input.Value); err != nil {
			return fmt.Errorf("must be an integer")
		}
	}
	return nil
}

func typeForSetting(section, key string) string {
	switch section + "." + key {
	case "proxy.enabled":
		return "bool"
	case "proxy.shift_interval_sec", "proxy.rate_limit_rps":
		return "int"
	default:
		return "string"
	}
}

func applySetting(rc *config.RuntimeConfig, section, key, value string) {
	switch section + "." + key {
	case "proxy.enabled":
		rc.SetProxyEnabled(value == "true")
	case "proxy.health_check_path":
		rc.SetHealthCheckPath(value)
	case "proxy.shift_interval_sec":
		if v, err := strconv.Atoi(value); err == nil {
			rc.SetShiftIntervalSec(v)
		}
	case "proxy.tunnel_service_url":
		rc.SetTunnelServiceURL(value)
	case "proxy.rate_limit_rps":
		if v, err := strconv.Atoi(value); err == nil {
			rc.SetRateLimitRPS(v)
		}
	case "docker.network":
		rc.SetDockerNetwork(value)
	case "docker.host_base":
		rc.SetDockerHostBase(value)
	}
}