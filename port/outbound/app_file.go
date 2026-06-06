package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type AppFileRepository interface {
	Create(ctx context.Context, file *entity.AppFile) error
	FindByID(ctx context.Context, id uint) (*entity.AppFile, error)
	FindByAppID(ctx context.Context, appID string) ([]*entity.AppFile, error)
	Update(ctx context.Context, file *entity.AppFile) error
	Delete(ctx context.Context, id uint) error
	DeleteByAppID(ctx context.Context, appID string) error
}