package echo

import (
	"github.com/AndreeJait/server-management-be/port/inbound/user"
	"github.com/AndreeJait/server-management-be/port/inbound/role"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/labstack/echo/v5"
)

// --- Admin user handlers ---

type createUserRequest struct {
	Email    string   `json:"email"`
	Password string   `json:"password"`
	Name     string   `json:"name"`
	Roles    []string `json:"roles"`
}

// @Summary      Create user
// @Description  Create a new user (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  createUserRequest  true  "Create user request"
// @Success      201  {object}  responsew.BaseResponse
// @Router       /admin/users [post]
func createUser(userUC user.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req createUserRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		userResp, err := userUC.Create(c.Request().Context(), req.Email, req.Password, req.Name, req.Roles)
		if err != nil {
			return nil, err
		}
		return responsew.Success(userResp, "User created"), nil
	}
}

// @Summary      List users
// @Description  List all users (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Success      200  {object}  responsew.BaseResponse
// @Router       /admin/users [get]
func listUsers(userUC user.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		users, err := userUC.List(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(users, "Users retrieved"), nil
	}
}

// @Summary      Get user
// @Description  Get a user by ID (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Param        id  path  int  true  "User ID"
// @Success      200  {object}  responsew.BaseResponse
// @Router       /admin/users/{id} [get]
func getUser(userUC user.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID := c.Param("id")
		user, err := userUC.Get(c.Request().Context(), userID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(user, "User retrieved"), nil
	}
}

type updateUserRequest struct {
	Name string `json:"name"`
}

// @Summary      Update user
// @Description  Update a user's details (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Param        id    path  int              true  "User ID"
// @Param        body  body  updateUserRequest  true  "User update"
// @Success      200  {object}  responsew.BaseResponse
// @Router       /admin/users/{id} [put]
func updateUser(userUC user.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID := c.Param("id")
		var req updateUserRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		user, err := userUC.Update(c.Request().Context(), userID, req.Name)
		if err != nil {
			return nil, err
		}
		return responsew.Success(user, "User updated"), nil
	}
}

type updateRolesRequest struct {
	Roles []string `json:"roles"`
}

// @Summary      Update user roles
// @Description  Update a user's roles (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Param        id    path  int                true  "User ID"
// @Param        body  body  updateRolesRequest  true  "Roles"
// @Success      200  {object}  responsew.BaseResponse
// @Router       /admin/users/{id}/roles [put]
func updateUserRoles(userUC user.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID := c.Param("id")
		var req updateRolesRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		user, err := userUC.UpdateRoles(c.Request().Context(), userID, req.Roles)
		if err != nil {
			return nil, err
		}
		return responsew.Success(user, "Roles updated"), nil
	}
}

// --- Role management handlers ---

// @Summary      List roles
// @Description  List all roles and their permissions (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Success      200  {object}  responsew.BaseResponse
// @Router       /admin/roles [get]
func listRoles(roleUC role.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		roles, err := roleUC.List(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(roles, "Roles retrieved"), nil
	}
}

type updateRolePermissionsRequest struct {
	Permissions []string `json:"permissions"`
}

// @Summary      Update role permissions
// @Description  Update permissions for a role (admin only)
// @Tags         admin
// @Security     BearerAuth
// @Param        name  path  string                      true  "Role name"
// @Param        body  body  updateRolePermissionsRequest  true  "Permissions"
// @Success      200  {object}  responsew.BaseResponse
// @Router       /admin/roles/{name}/permissions [put]
func updateRolePermissions(roleUC role.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		roleName := c.Param("name")
		var req updateRolePermissionsRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		r, err := roleUC.UpdatePermissions(c.Request().Context(), roleName, req.Permissions)
		if err != nil {
			return nil, err
		}
		return responsew.Success(r, "Role permissions updated"), nil
	}
}