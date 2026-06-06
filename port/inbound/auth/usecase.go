package auth

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

// UseCase defines the inbound port for authentication operations.
type UseCase interface {
	Login(ctx context.Context, email, password string) (*entity.UserResponse, string, error)
	RefreshToken(ctx context.Context, userID string) (string, error)
}