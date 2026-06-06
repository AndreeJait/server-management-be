package health

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

// UseCase defines the inbound port for health checking.
type UseCase interface {
	Check(ctx context.Context) *entity.Health
}