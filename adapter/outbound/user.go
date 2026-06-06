package outbound

import (
	"context"
	"fmt"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"gorm.io/gorm"
)

type userRepository struct {
	db *gorm.DB
}

func NewUserRepository(db *DB) outbound.UserRepository {
	return &userRepository{db: db.GormDB}
}

func (r *userRepository) Create(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Create(user).Error
}

func (r *userRepository) FindByID(ctx context.Context, id uint) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).First(&user, id).Error; err != nil {
		return nil, fmt.Errorf("user not found: %w", err)
	}
	return &user, nil
}

func (r *userRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	var user entity.User
	if err := r.db.WithContext(ctx).Where("email = ?", email).First(&user).Error; err != nil {
		return nil, fmt.Errorf("user not found by email: %w", err)
	}
	return &user, nil
}

func (r *userRepository) List(ctx context.Context) ([]*entity.User, error) {
	var users []*entity.User
	if err := r.db.WithContext(ctx).Find(&users).Error; err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}
	return users, nil
}

func (r *userRepository) UpdateRoles(ctx context.Context, userID uint, roles []string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&entity.UserRole{}).Error; err != nil {
			return fmt.Errorf("failed to clear user roles: %w", err)
		}
		for _, role := range roles {
			userRole := entity.UserRole{UserID: userID, Role: role}
			if err := tx.Create(&userRole).Error; err != nil {
				return fmt.Errorf("failed to assign role %s: %w", role, err)
			}
		}
		return nil
	})
}

func (r *userRepository) FindRolesByUserID(ctx context.Context, userID uint) ([]string, error) {
	var userRoles []entity.UserRole
	if err := r.db.WithContext(ctx).Where("user_id = ?", userID).Find(&userRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to find roles for user %d: %w", userID, err)
	}
	roles := make([]string, len(userRoles))
	for i, ur := range userRoles {
		roles[i] = ur.Role
	}
	return roles, nil
}

func (r *userRepository) Update(ctx context.Context, user *entity.User) error {
	return r.db.WithContext(ctx).Save(user).Error
}

func (r *userRepository) FindUserIDsByRole(ctx context.Context, role string) ([]string, error) {
	var userRoles []entity.UserRole
	if err := r.db.WithContext(ctx).Where("role = ?", role).Find(&userRoles).Error; err != nil {
		return nil, fmt.Errorf("failed to find user IDs for role %s: %w", role, err)
	}
	ids := make([]string, len(userRoles))
	for i, ur := range userRoles {
		ids[i] = fmt.Sprintf("%d", ur.UserID)
	}
	return ids, nil
}