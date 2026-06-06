package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type ProxyStateRepository interface {
	Create(ctx context.Context, ps *entity.ProxyState) error
	FindByAppID(ctx context.Context, appID string) (*entity.ProxyState, error)
	FindAll(ctx context.Context) ([]*entity.ProxyState, error)
	Update(ctx context.Context, ps *entity.ProxyState) error
	DeleteByAppID(ctx context.Context, appID string) error
}