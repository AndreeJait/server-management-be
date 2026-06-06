package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type appBindingRepository struct {
	db *gorm.DB
}

func NewAppBindingRepository(db *DB) outbound.AppBindingRepository {
	return &appBindingRepository{db: db.GormDB}
}

func (r *appBindingRepository) Create(ctx context.Context, binding *entity.AppBinding) error {
	return r.db.WithContext(ctx).Create(binding).Error
}

func (r *appBindingRepository) FindByAppID(ctx context.Context, appID string) ([]*entity.AppBinding, error) {
	var bindings []*entity.AppBinding
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).Find(&bindings).Error; err != nil {
		return nil, err
	}
	return bindings, nil
}

func (r *appBindingRepository) FindByID(ctx context.Context, id uint) (*entity.AppBinding, error) {
	var binding entity.AppBinding
	if err := r.db.WithContext(ctx).First(&binding, id).Error; err != nil {
		return nil, err
	}
	return &binding, nil
}

func (r *appBindingRepository) Update(ctx context.Context, binding *entity.AppBinding) error {
	return r.db.WithContext(ctx).Save(binding).Error
}

func (r *appBindingRepository) Delete(ctx context.Context, id uint) error {
	return r.db.WithContext(ctx).Delete(&entity.AppBinding{}, id).Error
}

func (r *appBindingRepository) DeleteByAppID(ctx context.Context, appID string) error {
	return r.db.WithContext(ctx).Where("app_id = ?", appID).Delete(&entity.AppBinding{}).Error
}