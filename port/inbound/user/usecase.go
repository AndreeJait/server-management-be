package user

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

// UseCase defines the inbound port for user management operations.
type UseCase interface {
	Create(ctx context.Context, email, password, name string, roles []string) (*entity.UserResponse, error)
	List(ctx context.Context) ([]*entity.UserResponse, error)
	Get(ctx context.Context, userID string) (*entity.UserResponse, error)
	Update(ctx context.Context, userID string, name string) (*entity.UserResponse, error)
	UpdateRoles(ctx context.Context, userID string, roles []string) (*entity.UserResponse, error)
}