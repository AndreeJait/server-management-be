package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type DeploymentRepository interface {
	Create(ctx context.Context, deployment *entity.Deployment) error
	FindByID(ctx context.Context, id uint) (*entity.Deployment, error)
	FindByAppID(ctx context.Context, appID string) ([]*entity.Deployment, error)
	FindRunningByAppID(ctx context.Context, appID string) ([]*entity.Deployment, error)
	FindLatestByAppID(ctx context.Context, appID string) (*entity.Deployment, error)
	Update(ctx context.Context, deployment *entity.Deployment) error
	DeleteByAppID(ctx context.Context, appID string) error
}