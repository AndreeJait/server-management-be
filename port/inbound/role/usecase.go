package role

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

// UseCase defines the inbound port for role management operations.
type UseCase interface {
	List(ctx context.Context) ([]*entity.RoleResponse, error)
	UpdatePermissions(ctx context.Context, role string, permissions []string) (*entity.RoleResponse, error)
}