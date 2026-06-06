package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type ProjectRepository interface {
	Create(ctx context.Context, project *entity.Project) error
	FindByID(ctx context.Context, id uint) (*entity.Project, error)
	FindByOwnerID(ctx context.Context, ownerID uint) ([]*entity.Project, error)
	Update(ctx context.Context, project *entity.Project) error
	Delete(ctx context.Context, id uint) error
}