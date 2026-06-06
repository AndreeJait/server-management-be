package usecase

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/role"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/authw"
)

type roleUseCase struct {
	roleRepo outbound.RoleRepository
	userRepo outbound.UserRepository
	rbac     *authw.RBAC
}

func NewRoleUseCase(roleRepo outbound.RoleRepository, userRepo outbound.UserRepository, rbac *authw.RBAC) role.UseCase {
	return &roleUseCase{roleRepo: roleRepo, userRepo: userRepo, rbac: rbac}
}

func (u *roleUseCase) List(ctx context.Context) ([]*entity.RoleResponse, error) {
	roles, err := u.roleRepo.ListRoles(ctx)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}
	if roles == nil {
		roles = []*entity.RoleResponse{}
	}
	return roles, nil
}

func (u *roleUseCase) UpdatePermissions(ctx context.Context, roleName string, permissions []string) (*entity.RoleResponse, error) {
	if err := u.roleRepo.UpdateRolePermissions(ctx, roleName, permissions); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	// Re-register the role in the in-memory RBAC so new permission checks use updated data
	u.rbac.RegisterRole(roleName, permissions...)

	// Invalidate RBAC cache for all users with this role
	userIDs, err := u.userRepo.FindUserIDsByRole(ctx, roleName)
	if err != nil {
		_ = err
	}
	for _, uid := range userIDs {
		_ = u.rbac.InvalidateUser(ctx, uid)
	}

	return &entity.RoleResponse{
		Role:        roleName,
		Permissions: permissions,
	}, nil
}