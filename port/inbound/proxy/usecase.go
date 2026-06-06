package proxy

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type UseCase interface {
	DeployBlueGreen(ctx context.Context, appID, image, authRegistry, authUser, authPass string) (*entity.DeploymentResponse, error)
	GetProxyState(ctx context.Context, appID string) (*entity.ProxyStateResponse, error)
	ListProxyStates(ctx context.Context) ([]*entity.ProxyStateResponse, error)
	SetTraffic(ctx context.Context, appID string, percent int) (*entity.ProxyStateResponse, error)
	Rollback(ctx context.Context, appID string) (*entity.ProxyStateResponse, error)
}