package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

// RoleRepository defines the outbound port for role-permission persistence.
type RoleRepository interface {
	FindPermissionsByRole(ctx context.Context, role string) ([]string, error)
	ListRoles(ctx context.Context) ([]*entity.RoleResponse, error)
	UpdateRolePermissions(ctx context.Context, role string, permissions []string) error
	FindAllPermissions(ctx context.Context) (map[string][]string, error)
}