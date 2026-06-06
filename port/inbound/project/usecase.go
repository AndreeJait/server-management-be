package project

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type UseCase interface {
	Create(ctx context.Context, name, description string, ownerID uint) (*entity.ProjectResponse, error)
	List(ctx context.Context, ownerID uint) ([]*entity.ProjectResponse, error)
	Get(ctx context.Context, projectID string, ownerID uint) (*entity.ProjectResponse, error)
	Update(ctx context.Context, projectID string, ownerID uint, name, description string) (*entity.ProjectResponse, error)
	Delete(ctx context.Context, projectID string, ownerID uint) error
}