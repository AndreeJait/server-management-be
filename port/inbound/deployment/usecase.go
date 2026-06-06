package deployment

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"io"
	"net"
)

type UseCase interface {
	Deploy(ctx context.Context, appID, deployToken, image string) (*entity.DeploymentResponse, error)
	DeployApp(ctx context.Context, appID, image string) (*entity.DeploymentResponse, error)
	StopContainer(ctx context.Context, appID string) error
	ListByAppID(ctx context.Context, appID string) ([]*entity.DeploymentResponse, error)
	Get(ctx context.Context, deploymentID string) (*entity.DeploymentResponse, error)
	GetContainerLogs(ctx context.Context, appID string, tail int) (string, error)
	GetRunningContainerID(ctx context.Context, appID string) (string, error)
	ExecInteractive(ctx context.Context, appID string, containerID string, shell string) (conn net.Conn, reader io.Reader, err error)
}