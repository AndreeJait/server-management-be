package echo

import (
	"github.com/AndreeJait/server-management-be/domain/entity"
	configUC "github.com/AndreeJait/server-management-be/port/inbound/config"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/labstack/echo/v5"
)

// --- Config settings handlers ---

// @Summary      Get all settings
// @Description  Get all configuration settings grouped by section
// @Tags         config
// @Security     BearerAuth
// @Success      200  {object}  responsew.BaseResponse
// @Router       /config/settings [get]
func getSettings(configUC configUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		groups, err := configUC.GetSettings(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(groups, "Settings retrieved"), nil
	}
}

// @Summary      Get settings by section
// @Description  Get configuration settings for a specific section
// @Tags         config
// @Security     BearerAuth
// @Param        section  path  string  true  "Section name"
// @Success      200  {object}  responsew.BaseResponse
// @Router       /config/settings/{section} [get]
func getSettingsBySection(configUC configUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		section := c.Param("section")
		group, err := configUC.GetSettingsBySection(c.Request().Context(), section)
		if err != nil {
			return nil, err
		}
		return responsew.Success(group, "Settings retrieved"), nil
	}
}

type updateSettingsRequest struct {
	Settings []entity.UpdateSettingInput `json:"settings"`
}

// @Summary      Update settings
// @Description  Update multiple configuration settings
// @Tags         config
// @Security     BearerAuth
// @Accept       json
// @Produce      json
// @Param        body  body  updateSettingsRequest  true  "Settings to update"
// @Success      200  {object}  responsew.BaseResponse
// @Router       /config/settings [put]
func updateSettings(configUC configUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req updateSettingsRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		result, err := configUC.UpdateSettings(c.Request().Context(), req.Settings)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Settings updated"), nil
	}
}

// --- Domain access stats handler ---

// @Summary      Get domain access stats
// @Description  Get per-domain request counts from proxy traffic
// @Tags         config
// @Security     BearerAuth
// @Success      200  {object}  responsew.BaseResponse
// @Router       /config/proxy/access-stats [get]
func getDomainRequestCounts(configUC configUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		counts, err := configUC.GetDomainRequestCounts(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(counts, "Access stats retrieved"), nil
	}
}