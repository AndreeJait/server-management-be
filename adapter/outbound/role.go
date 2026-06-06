package outbound

import (
	"context"
	"fmt"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type roleRepository struct {
	db *gorm.DB
}

func NewRoleRepository(db *DB) outbound.RoleRepository {
	return &roleRepository{db: db.GormDB}
}

func (r *roleRepository) FindPermissionsByRole(ctx context.Context, role string) ([]string, error) {
	var perms []entity.RolePermission
	if err := r.db.WithContext(ctx).Where("role = ?", role).Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("failed to find permissions for role %s: %w", role, err)
	}
	result := make([]string, len(perms))
	for i, p := range perms {
		result[i] = p.Permission
	}
	return result, nil
}

func (r *roleRepository) ListRoles(ctx context.Context) ([]*entity.RoleResponse, error) {
	var perms []entity.RolePermission
	if err := r.db.WithContext(ctx).Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("failed to list role permissions: %w", err)
	}

	roleMap := make(map[string][]string)
	for _, p := range perms {
		roleMap[p.Role] = append(roleMap[p.Role], p.Permission)
	}

	var responses []*entity.RoleResponse
	for role, permissions := range roleMap {
		responses = append(responses, &entity.RoleResponse{Role: role, Permissions: permissions})
	}
	return responses, nil
}

func (r *roleRepository) UpdateRolePermissions(ctx context.Context, role string, permissions []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("role = ?", role).Delete(&entity.RolePermission{}).Error; err != nil {
			return fmt.Errorf("failed to clear permissions for role %s: %w", role, err)
		}
		for _, perm := range permissions {
			rp := entity.RolePermission{Role: role, Permission: perm}
			if err := tx.Create(&rp).Error; err != nil {
				return fmt.Errorf("failed to assign permission %s to role %s: %w", perm, role, err)
			}
		}
		return nil
	})
}

func (r *roleRepository) FindAllPermissions(ctx context.Context) (map[string][]string, error) {
	var perms []entity.RolePermission
	if err := r.db.WithContext(ctx).Find(&perms).Error; err != nil {
		return nil, fmt.Errorf("failed to find all permissions: %w", err)
	}

	result := make(map[string][]string)
	for _, p := range perms {
		result[p.Role] = append(result[p.Role], p.Permission)
	}
	return result, nil
}