package echo

import (
	"github.com/AndreeJait/server-management-be/port/inbound/health"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/labstack/echo/v5"
)

// --- Health handlers ---

// @Summary      Health check
// @Description  Check if the service is healthy including DB and Redis connectivity
// @Tags         health
// @Success      200  {object}  responsew.BaseResponse{data=entity.Health}
// @Router       /health [get]
func checkHealth(healthUC health.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		health := healthUC.Check(c.Request().Context())
		return responsew.Success(health, "Service is healthy"), nil
	}
}