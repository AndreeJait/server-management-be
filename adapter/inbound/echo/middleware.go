package echo

import (
	"github.com/AndreeJait/go-utility/v2/authw"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/go-utility/v2/statusw"
	"github.com/labstack/echo/v5"
)

// authMiddleware returns Echo middleware that authenticates requests via the Authenticator
// and injects the authw.Result into the request context.
func authMiddleware(authenticator authw.Authenticator) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			result, err := authenticator.Authenticate(c.Request())
			if err != nil {
				code, resp := responsew.Error(err)
				return c.JSON(code, resp)
			}
			c.SetRequest(c.Request().WithContext(authw.WithResult(c.Request().Context(), result)))
			return next(c)
		}
	}
}

// rbacMiddleware returns Echo middleware that checks the authenticated user has the required permission.
func rbacMiddleware(rbac *authw.RBAC, permission string) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c *echo.Context) error {
			result := authw.FromContext(c.Request().Context())
			if result == nil || result.GetUserID() == "" {
				err := statusw.InvalidCredential.WithCustomMessage("Authentication required")
				code, resp := responsew.Error(err)
				return c.JSON(code, resp)
			}
			ok, err := rbac.CheckPermission(c.Request().Context(), result.GetUserID(), permission)
			if err != nil {
				code, resp := responsew.Error(err)
				return c.JSON(code, resp)
			}
			if !ok {
				err := statusw.InvalidAccess.WithCustomMessage("Insufficient permissions")
				code, resp := responsew.Error(err)
				return c.JSON(code, resp)
			}
			return next(c)
		}
	}
}