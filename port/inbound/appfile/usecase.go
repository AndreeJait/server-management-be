package appfile

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type DeployFunc func(ctx context.Context, appID, deployToken, image string) (*entity.DeploymentResponse, error)

type UseCase interface {
	Create(ctx context.Context, projectID uint, appID, path, content string) (*entity.AppFileResponse, error)
	List(ctx context.Context, projectID uint, appID string) ([]*entity.AppFileResponse, error)
	Get(ctx context.Context, projectID uint, appID string, fileID uint) (*entity.AppFileResponse, error)
	Update(ctx context.Context, projectID uint, appID string, fileID uint, path, content string) (*entity.AppFileResponse, error)
	Delete(ctx context.Context, projectID uint, appID string, fileID uint) error
	Upload(ctx context.Context, projectID uint, appID, path string, data []byte, mimeType string, fileSize int64) (*entity.AppFileResponse, error)
	Download(ctx context.Context, projectID uint, appID string, fileID uint) ([]byte, string, error)
	CreateFolder(ctx context.Context, projectID uint, appID, path string) (*entity.AppResponse, error)
	DeleteFolder(ctx context.Context, projectID uint, appID, path string) error
	SetDeployFunc(fn DeployFunc)
}