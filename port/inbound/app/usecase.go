package app

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type UseCase interface {
	Create(ctx context.Context, projectID uint, name, frameworkPreset string) (*entity.CreateAppResponse, error)
	List(ctx context.Context, projectID uint) ([]*entity.AppResponse, error)
	Get(ctx context.Context, projectID uint, appID string) (*entity.AppResponse, error)
	Update(ctx context.Context, projectID uint, appID string, name, frameworkPreset string, envVars entity.StringMap, volumeMounts entity.VolumeMountList, postDeployCommands entity.StringList, basePath, defaultImage, containerPort, publishPort, containerName, filesMountPath string) (*entity.AppResponse, error)
	Delete(ctx context.Context, projectID uint, appID string) error
	RegenerateDeployToken(ctx context.Context, projectID uint, appID string) (*entity.RegenerateTokenResponse, error)
}