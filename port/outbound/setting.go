package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type SettingRepository interface {
	FindAll(ctx context.Context) ([]*entity.Setting, error)
	FindBySection(ctx context.Context, section string) ([]*entity.Setting, error)
	FindByKey(ctx context.Context, section, key string) (*entity.Setting, error)
	Upsert(ctx context.Context, setting *entity.Setting) error
	UpsertBatch(ctx context.Context, settings []*entity.Setting) error
}