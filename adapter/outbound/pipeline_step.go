package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type pipelineStepRepository struct {
	db *gorm.DB
}

func NewPipelineStepRepository(db *DB) outbound.PipelineStepRepository {
	return &pipelineStepRepository{db: db.GormDB}
}

func (r *pipelineStepRepository) CreateBatch(ctx context.Context, steps []*entity.PipelineStep) error {
	return r.db.WithContext(ctx).Create(steps).Error
}

func (r *pipelineStepRepository) FindByDeploymentID(ctx context.Context, deploymentID uint) ([]*entity.PipelineStep, error) {
	var steps []*entity.PipelineStep
	if err := r.db.WithContext(ctx).Where("deployment_id = ?", deploymentID).Order("step_order ASC").Find(&steps).Error; err != nil {
		return nil, err
	}
	return steps, nil
}

func (r *pipelineStepRepository) Update(ctx context.Context, step *entity.PipelineStep) error {
	return r.db.WithContext(ctx).Save(step).Error
}

func (r *pipelineStepRepository) DeleteByDeploymentID(ctx context.Context, deploymentID uint) error {
	return r.db.WithContext(ctx).Where("deployment_id = ?", deploymentID).Delete(&entity.PipelineStep{}).Error
}