package usecase

import (
	"context"
	"strconv"
	"time"

	"github.com/AndreeJait/server-management-be/domain/entity"
	domainError "github.com/AndreeJait/server-management-be/domain/error"
	"github.com/AndreeJait/server-management-be/port/inbound/auth"
	"github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/authw"
	"github.com/AndreeJait/go-utility/v2/jwtw"
	jwtv5 "github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

type authUseCase struct {
	userRepo outbound.UserRepository
	jwt      jwtw.JWT
	jwtTTL   time.Duration
	rbac     *authw.RBAC
}

func NewAuthUseCase(userRepo outbound.UserRepository, jwt jwtw.JWT, jwtTTL time.Duration, rbac *authw.RBAC) auth.UseCase {
	return &authUseCase{userRepo: userRepo, jwt: jwt, jwtTTL: jwtTTL, rbac: rbac}
}

func (u *authUseCase) Login(ctx context.Context, email, password string) (*entity.UserResponse, string, error) {
	user, err := u.userRepo.FindByEmail(ctx, email)
	if err != nil {
		return nil, "", domainError.ErrUnauthorized.WithCustomMessage("Invalid email or password")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password)); err != nil {
		return nil, "", domainError.ErrUnauthorized.WithCustomMessage("Invalid email or password")
	}

	roles, err := u.userRepo.FindRolesByUserID(ctx, user.ID)
	if err != nil {
		return nil, "", domainError.ErrInternalServer.WithError(err)
	}

	token, err := u.generateToken(user.ID, email, roles)
	if err != nil {
		return nil, "", domainError.ErrInternalServer.WithError(err)
	}

	return user.ToResponse(roles), token, nil
}

func (u *authUseCase) RefreshToken(ctx context.Context, userID string) (string, error) {
	id, err := strconv.ParseUint(userID, 10, 64)
	if err != nil {
		return "", domainError.ErrInvalidParam.WithCustomMessage("Invalid user ID")
	}

	user, err := u.userRepo.FindByID(ctx, uint(id))
	if err != nil {
		return "", domainError.ErrNotFound.WithCustomMessage("User not found")
	}

	roles, err := u.userRepo.FindRolesByUserID(ctx, user.ID)
	if err != nil {
		return "", domainError.ErrInternalServer.WithError(err)
	}

	token, err := u.generateToken(user.ID, user.Email, roles)
	if err != nil {
		return "", domainError.ErrInternalServer.WithError(err)
	}

	return token, nil
}

// TokenData holds the data embedded in JWT tokens.
type TokenData struct {
	UserID uint     `json:"user_id"`
	Email  string   `json:"email"`
	Roles  []string `json:"roles"`
}

func (u *authUseCase) generateToken(userID uint, email string, roles []string) (string, error) {
	data := TokenData{
		UserID: userID,
		Email:  email,
		Roles:  roles,
	}
	claims := jwtw.NewClaims(data, u.jwtTTL, "server-management", strconv.FormatUint(uint64(userID), 10))
	return u.jwt.Create(&claims)
}

// NewClaimsFactory returns a function that creates empty JWT claims for the authenticator.
func NewClaimsFactory() func() jwtv5.Claims {
	return func() jwtv5.Claims {
		return &jwtw.MyClaims[TokenData]{}
	}
}

// ExtractJWTClaims maps parsed JWT claims to an authw.Result.
func ExtractJWTClaims(claims jwtv5.Claims) *authw.Result {
	c, ok := claims.(*jwtw.MyClaims[TokenData])
	if !ok {
		return &authw.Result{}
	}
	return &authw.Result{
		UserID:      strconv.FormatUint(uint64(c.Data.UserID), 10),
		Username:    c.Data.Email,
		Roles:       c.Data.Roles,
		Permissions: []string{},
	}
}