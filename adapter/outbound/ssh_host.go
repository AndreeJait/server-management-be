package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type sshHostRepository struct {
	db *gorm.DB
}

func NewSSHHostRepository(db *DB) outbound.SSHHostRepository {
	return &sshHostRepository{db: db.GormDB}
}

func (r *sshHostRepository) Create(ctx context.Context, host *entity.SSHHost) error {
	return r.db.WithContext(ctx).Create(host).Error
}

func (r *sshHostRepository) FindByID(ctx context.Context, id uint) (*entity.SSHHost, error) {
	var host entity.SSHHost
	if err := r.db.WithContext(ctx).First(&host, id).Error; err != nil {
		return nil, err
	}
	return &host, nil
}

func (r *sshHostRepository) FindByOwnerID(ctx context.Context, ownerID uint) ([]*entity.SSHHost, error) {
	var hosts []*entity.SSHHost
	if err := r.db.WithContext(ctx).Where("owner_id = ?", ownerID).Find(&hosts).Error; err != nil {
		return nil, err
	}
	return hosts, nil
}

func (r *sshHostRepository) Update(ctx context.Context, host *entity.SSHHost) error {
	return r.db.WithContext(ctx).Save(host).Error
}

func (r *sshHostRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.SSHHost{}, id).Error
}