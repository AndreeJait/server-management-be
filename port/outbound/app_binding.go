package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type AppBindingRepository interface {
	Create(ctx context.Context, binding *entity.AppBinding) error
	FindByAppID(ctx context.Context, appID string) ([]*entity.AppBinding, error)
	FindByID(ctx context.Context, id uint) (*entity.AppBinding, error)
	Update(ctx context.Context, binding *entity.AppBinding) error
	Delete(ctx context.Context, id uint) error
	DeleteByAppID(ctx context.Context, appID string) error
}