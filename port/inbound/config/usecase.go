package config

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type UseCase interface {
	GetSettings(ctx context.Context) ([]*entity.SettingsGroup, error)
	GetSettingsBySection(ctx context.Context, section string) (*entity.SettingsGroup, error)
	UpdateSettings(ctx context.Context, updates []entity.UpdateSettingInput) (*entity.UpdateSettingsResult, error)
	GetDomainRequestCounts(ctx context.Context) ([]*entity.DomainRequestCountResponse, error)
}