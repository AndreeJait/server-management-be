package echo

import (
	"strconv"

	"github.com/AndreeJait/server-management-be/port/inbound/auth"
	"github.com/AndreeJait/server-management-be/port/inbound/user"
	"github.com/AndreeJait/go-utility/v2/authw"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/go-utility/v2/statusw"
	"github.com/labstack/echo/v5"
)

// --- Auth handlers ---

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// @Summary      Login
// @Description  Authenticate and get a JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  loginRequest  true  "Login request"
// @Success      200  {object}  responsew.BaseResponse
// @Router       /auth/login [post]
func login(authUC auth.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req loginRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		userResp, token, err := authUC.Login(c.Request().Context(), req.Email, req.Password)
		if err != nil {
			return nil, err
		}
		return responsew.Success(map[string]any{"user": userResp, "token": token}, "Login successful"), nil
	}
}

type refreshRequest struct {
	UserID string `json:"user_id"`
}

// @Summary      Refresh token
// @Description  Get a new JWT token
// @Tags         auth
// @Accept       json
// @Produce      json
// @Param        body  body  refreshRequest  true  "Refresh request"
// @Success      200  {object}  responsew.BaseResponse
// @Router       /auth/refresh [post]
func refreshToken(authUC auth.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req refreshRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		token, err := authUC.RefreshToken(c.Request().Context(), req.UserID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(map[string]any{"token": token}, "Token refreshed"), nil
	}
}

// @Summary      Get current user
// @Description  Get the authenticated user's profile
// @Tags         auth
// @Security     BearerAuth
// @Produce      json
// @Success      200  {object}  responsew.BaseResponse
// @Router       /auth/me [get]
func getMe(userUC user.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID, err := getUserIDFromContext(c)
		if err != nil || userID == 0 {
			return nil, statusw.InvalidCredential.WithCustomMessage("Authentication required")
		}
		user, err := userUC.Get(c.Request().Context(), strconv.FormatUint(uint64(userID), 10))
		if err != nil {
			return nil, err
		}
		return responsew.Success(user, "Profile retrieved"), nil
	}
}

// --- Helper functions ---

// getUserIDFromContext extracts the user ID from the auth context.
func getUserIDFromContext(c *echo.Context) (uint, error) {
	result := authw.FromContext(c.Request().Context())
	if result == nil {
		return 0, nil
	}
	idStr := result.GetUserID()
	if idStr == "" {
		return 0, nil
	}
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		return 0, err
	}
	return uint(id), nil
}

// getOwnerID extracts the current user's ID and returns it as ownerID.
func getOwnerID(c *echo.Context) (uint, error) {
	id, err := getUserIDFromContext(c)
	if err != nil || id == 0 {
		return 0, statusw.InvalidCredential.WithCustomMessage("Authentication required")
	}
	return id, nil
}

// parseProjectID parses the :id param from the URL.
func parseProjectID(c *echo.Context) (uint, error) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, statusw.InvalidReqParam.WithCustomMessage("Invalid project ID")
	}
	return uint(id), nil
}