package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type RegistryCredentialRepository interface {
	Create(ctx context.Context, cred *entity.RegistryCredential) error
	FindByID(ctx context.Context, id uint) (*entity.RegistryCredential, error)
	FindGlobal(ctx context.Context) ([]*entity.RegistryCredential, error)
	FindByProjectID(ctx context.Context, projectID uint) ([]*entity.RegistryCredential, error)
	Update(ctx context.Context, cred *entity.RegistryCredential) error
	Delete(ctx context.Context, id uint) error
}