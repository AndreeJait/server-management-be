package echo

import (
	proxyUC "github.com/AndreeJait/server-management-be/port/inbound/proxy"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/go-utility/v2/statusw"
	"github.com/labstack/echo/v5"
)

func listProxyStates(proxyUC proxyUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		states, err := proxyUC.ListProxyStates(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(states, "Proxy states retrieved"), nil
	}
}

func getProxyState(proxyUC proxyUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		appID := c.Param("appId")
		if appID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("app_id is required")
		}
		state, err := proxyUC.GetProxyState(c.Request().Context(), appID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(state, "Proxy state retrieved"), nil
	}
}

type setTrafficRequest struct {
	Percent int `json:"percent"`
}

func setTraffic(proxyUC proxyUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		appID := c.Param("appId")
		if appID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("app_id is required")
		}
		var req setTrafficRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		result, err := proxyUC.SetTraffic(c.Request().Context(), appID, req.Percent)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Traffic percentage updated"), nil
	}
}

func rollbackProxy(proxyUC proxyUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		appID := c.Param("appId")
		if appID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("app_id is required")
		}
		result, err := proxyUC.Rollback(c.Request().Context(), appID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Rollback initiated"), nil
	}
}