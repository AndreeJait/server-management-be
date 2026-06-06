package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type deploymentRepository struct {
	db *gorm.DB
}

func NewDeploymentRepository(db *DB) outbound.DeploymentRepository {
	return &deploymentRepository{db: db.GormDB}
}

func (r *deploymentRepository) Create(ctx context.Context, d *entity.Deployment) error {
	return r.db.WithContext(ctx).Create(d).Error
}

func (r *deploymentRepository) FindByID(ctx context.Context, id uint) (*entity.Deployment, error) {
	var d entity.Deployment
	if err := r.db.WithContext(ctx).First(&d, id).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *deploymentRepository) FindByAppID(ctx context.Context, appID string) ([]*entity.Deployment, error) {
	var deployments []*entity.Deployment
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).Order("created_at DESC").Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

func (r *deploymentRepository) FindRunningByAppID(ctx context.Context, appID string) ([]*entity.Deployment, error) {
	var deployments []*entity.Deployment
	if err := r.db.WithContext(ctx).Where("app_id = ? AND status = ?", appID, entity.DeploymentRunning).Find(&deployments).Error; err != nil {
		return nil, err
	}
	return deployments, nil
}

func (r *deploymentRepository) Update(ctx context.Context, d *entity.Deployment) error {
	return r.db.WithContext(ctx).Save(d).Error
}

func (r *deploymentRepository) FindLatestByAppID(ctx context.Context, appID string) (*entity.Deployment, error) {
	var d entity.Deployment
	if err := r.db.WithContext(ctx).Where("app_id = ?", appID).Order("created_at DESC").First(&d).Error; err != nil {
		return nil, err
	}
	return &d, nil
}

func (r *deploymentRepository) DeleteByAppID(ctx context.Context, appID string) error {
	return r.db.WithContext(ctx).Where("app_id = ?", appID).Delete(&entity.Deployment{}).Error
}