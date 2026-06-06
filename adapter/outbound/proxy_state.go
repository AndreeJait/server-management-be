package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type proxyStateRepository struct {
	db *gorm.DB
}

func NewProxyStateRepository(db *DB) outbound.ProxyStateRepository {
	return &proxyStateRepository{db: db.GormDB}
}

func (r *proxyStateRepository) Create(ctx context.Context, ps *entity.ProxyState) error {
	return r.db.WithContext(ctx).Create(ps).Error
}

func (r *proxyStateRepository) FindByAppID(ctx context.Context, appID string) (*entity.ProxyState, error) {
	var ps entity.ProxyState
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).First(&ps).Error; err != nil {
		return nil, err
	}
	return &ps, nil
}

func (r *proxyStateRepository) FindAll(ctx context.Context) ([]*entity.ProxyState, error) {
	var states []*entity.ProxyState
	if err := r.db.WithContext(ctx).Find(&states).Error; err != nil {
		return nil, err
	}
	return states, nil
}

func (r *proxyStateRepository) Update(ctx context.Context, ps *entity.ProxyState) error {
	return r.db.WithContext(ctx).Save(ps).Error
}

func (r *proxyStateRepository) DeleteByAppID(ctx context.Context, appID string) error {
	return r.db.WithContext(ctx).Where("app_id = ?", appID).Delete(&entity.ProxyState{}).Error
}