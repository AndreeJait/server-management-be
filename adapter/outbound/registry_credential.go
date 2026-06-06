package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type registryCredentialRepository struct {
	db *gorm.DB
}

func NewRegistryCredentialRepository(db *DB) outbound.RegistryCredentialRepository {
	return &registryCredentialRepository{db: db.GormDB}
}

func (r *registryCredentialRepository) Create(ctx context.Context, cred *entity.RegistryCredential) error {
	return r.db.WithContext(ctx).Create(cred).Error
}

func (r *registryCredentialRepository) FindByID(ctx context.Context, id uint) (*entity.RegistryCredential, error) {
	var cred entity.RegistryCredential
	if err := r.db.WithContext(ctx).First(&cred, id).Error; err != nil {
		return nil, err
	}
	return &cred, nil
}

func (r *registryCredentialRepository) FindGlobal(ctx context.Context) ([]*entity.RegistryCredential, error) {
	var creds []*entity.RegistryCredential
	if err := r.db.WithContext(ctx).Where("scope = ?", "global").Find(&creds).Error; err != nil {
		return nil, err
	}
	return creds, nil
}

func (r *registryCredentialRepository) FindByProjectID(ctx context.Context, projectID uint) ([]*entity.RegistryCredential, error) {
	var creds []*entity.RegistryCredential
	if err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&creds).Error; err != nil {
		return nil, err
	}
	return creds, nil
}

func (r *registryCredentialRepository) Update(ctx context.Context, cred *entity.RegistryCredential) error {
	return r.db.WithContext(ctx).Save(cred).Error
}

func (r *registryCredentialRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.RegistryCredential{}, id).Error
}