package binding

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type UseCase interface {
	Create(ctx context.Context, appID, zoneID, domain, tunnelID, service string) (*entity.AppBindingResponse, error)
	List(ctx context.Context, appID string) ([]*entity.AppBindingResponse, error)
	Get(ctx context.Context, bindingID uint) (*entity.AppBindingResponse, error)
	Delete(ctx context.Context, bindingID uint) error
}