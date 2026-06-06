package usecase

import (
	"context"
	"strconv"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/user"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/authw"
	"golang.org/x/crypto/bcrypt"
)

type userUseCase struct {
	userRepo outbound.UserRepository
	rbac     *authw.RBAC
}

func NewUserUseCase(userRepo outbound.UserRepository, rbac *authw.RBAC) user.UseCase {
	return &userUseCase{userRepo: userRepo, rbac: rbac}
}

func (u *userUseCase) Create(ctx context.Context, email, password, name string, roles []string) (*entity.UserResponse, error) {
	existing, _ := u.userRepo.FindByEmail(ctx, email)
	if existing != nil {
		return nil, domainError.ErrConflict.WithCustomMessage("Email already registered")
	}

	hashedPwd, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	user := entity.NewUser(email, string(hashedPwd), name)
	if err := u.userRepo.Create(ctx, user); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	if len(roles) == 0 {
		roles = []string{"viewer"}
	}
	if err := u.userRepo.UpdateRoles(ctx, user.ID, roles); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	return user.ToResponse(roles), nil
}

func (u *userUseCase) List(ctx context.Context) ([]*entity.UserResponse, error) {
	users, err := u.userRepo.List(ctx)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	responses := make([]*entity.UserResponse, 0, len(users))
	for _, usr := range users {
		roles, err := u.userRepo.FindRolesByUserID(ctx, usr.ID)
		if err != nil {
			return nil, domainError.ErrInternalServer.WithError(err)
		}
		responses = append(responses, usr.ToResponse(roles))
	}
	return responses, nil
}

func (u *userUseCase) Get(ctx context.Context, userID string) (*entity.UserResponse, error) {
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid user ID")
	}

	usr, err := u.userRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("User not found")
	}

	roles, err := u.userRepo.FindRolesByUserID(ctx, usr.ID)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	return usr.ToResponse(roles), nil
}

func (u *userUseCase) Update(ctx context.Context, userID string, name string) (*entity.UserResponse, error) {
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid user ID")
	}

	usr, err := u.userRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("User not found")
	}

	usr.Name = name
	if err := u.userRepo.Update(ctx, usr); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	roles, err := u.userRepo.FindRolesByUserID(ctx, usr.ID)
	if err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	return usr.ToResponse(roles), nil
}

func (u *userUseCase) UpdateRoles(ctx context.Context, userID string, roles []string) (*entity.UserResponse, error) {
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return nil, domainError.ErrInvalidParam.WithCustomMessage("Invalid user ID")
	}

	usr, err := u.userRepo.FindByID(ctx, uint(id))
	if err != nil {
		return nil, domainError.ErrNotFound.WithCustomMessage("User not found")
	}

	if err := u.userRepo.UpdateRoles(ctx, usr.ID, roles); err != nil {
		return nil, domainError.ErrInternalServer.WithError(err)
	}

	if err := u.rbac.InvalidateUser(ctx, userID); err != nil {
		_ = err
	}

	return usr.ToResponse(roles), nil
}