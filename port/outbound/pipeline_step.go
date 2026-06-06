package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type PipelineStepRepository interface {
	CreateBatch(ctx context.Context, steps []*entity.PipelineStep) error
	FindByDeploymentID(ctx context.Context, deploymentID uint) ([]*entity.PipelineStep, error)
	Update(ctx context.Context, step *entity.PipelineStep) error
	DeleteByDeploymentID(ctx context.Context, deploymentID uint) error
}