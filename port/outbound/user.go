package outbound

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

// UserRepository defines the outbound port for user persistence.
type UserRepository interface {
	Create(ctx context.Context, user *entity.User) error
	FindByID(ctx context.Context, id uint) (*entity.User, error)
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	List(ctx context.Context) ([]*entity.User, error)
	Update(ctx context.Context, user *entity.User) error
	UpdateRoles(ctx context.Context, userID uint, roles []string) error
	FindRolesByUserID(ctx context.Context, userID uint) ([]string, error)
	FindUserIDsByRole(ctx context.Context, role string) ([]string, error)
}