package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type AppRepository interface {
	Create(ctx context.Context, app *entity.App) error
	FindByID(ctx context.Context, id uint) (*entity.App, error)
	FindByProjectID(ctx context.Context, projectID uint) ([]*entity.App, error)
	FindByAppID(ctx context.Context, appID string) (*entity.App, error)
	Update(ctx context.Context, app *entity.App) error
	Delete(ctx context.Context, id uint) error
}