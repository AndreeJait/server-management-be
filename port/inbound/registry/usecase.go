package registry

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type UseCase interface {
	Create(ctx context.Context, scope string, projectID *uint, registryURL, username, password string) (*entity.RegistryCredentialResponse, error)
	ListGlobal(ctx context.Context) ([]*entity.RegistryCredentialResponse, error)
	ListByProject(ctx context.Context, projectID uint) ([]*entity.RegistryCredentialResponse, error)
	Update(ctx context.Context, credID string, registryURL, username, password string) (*entity.RegistryCredentialResponse, error)
	Delete(ctx context.Context, credID string) error
}