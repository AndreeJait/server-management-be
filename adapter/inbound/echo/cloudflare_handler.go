package echo

import (
	"strconv"

	bindingUC "github.com/AndreeJait/server-management-be/port/inbound/binding"
	"github.com/AndreeJait/server-management-be/port/inbound/cloudflare"
	"github.com/AndreeJait/server-management-be/port/inbound/project"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/go-utility/v2/statusw"
	"github.com/labstack/echo/v5"
)

// --- Cloudflare handlers ---

func listAccounts(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		accounts, err := cfUC.ListAccounts(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(accounts, "Accounts retrieved"), nil
	}
}

func listZones(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		zones, err := cfUC.ListZones(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(zones, "Zones retrieved"), nil
	}
}

func listDNSRecords(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		zoneID := c.Param("zoneId")
		if zoneID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("zone ID is required")
		}
		records, err := cfUC.ListDNSRecords(c.Request().Context(), zoneID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(records, "DNS records retrieved"), nil
	}
}

type createDNSRecordRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

func createDNSRecord(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		zoneID := c.Param("zoneId")
		if zoneID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("zone ID is required")
		}
		var req createDNSRecordRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		record, err := cfUC.CreateDNSRecord(c.Request().Context(), zoneID, req.Type, req.Name, req.Content, req.TTL, req.Proxied)
		if err != nil {
			return nil, err
		}
		return responsew.Success(record, "DNS record created"), nil
	}
}

type updateDNSRecordRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

func updateDNSRecord(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		zoneID := c.Param("zoneId")
		recordID := c.Param("recordId")
		if zoneID == "" || recordID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("zone ID and record ID are required")
		}
		var req updateDNSRecordRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		record, err := cfUC.UpdateDNSRecord(c.Request().Context(), zoneID, recordID, req.Type, req.Name, req.Content, req.TTL, req.Proxied)
		if err != nil {
			return nil, err
		}
		return responsew.Success(record, "DNS record updated"), nil
	}
}

func deleteDNSRecord(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		zoneID := c.Param("zoneId")
		recordID := c.Param("recordId")
		if zoneID == "" || recordID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("zone ID and record ID are required")
		}
		if err := cfUC.DeleteDNSRecord(c.Request().Context(), zoneID, recordID); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "DNS record deleted"), nil
	}
}

func listTunnels(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		tunnels, err := cfUC.ListTunnels(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(tunnels, "Tunnels retrieved"), nil
	}
}

func getTunnelConfig(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		tunnelID := c.Param("tunnelId")
		if tunnelID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("tunnel ID is required")
		}
		config, err := cfUC.GetTunnelConfig(c.Request().Context(), tunnelID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(config, "Tunnel config retrieved"), nil
	}
}

func listAccessApps(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		apps, err := cfUC.ListAccessApps(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(apps, "Access apps retrieved"), nil
	}
}

type createAccessAppRequest struct {
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	SessionDuration string `json:"session_duration"`
}

func createAccessApp(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req createAccessAppRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		app, err := cfUC.CreateAccessApp(c.Request().Context(), req.Name, req.Domain, req.SessionDuration)
		if err != nil {
			return nil, err
		}
		return responsew.Success(app, "Access app created"), nil
	}
}

type updateAccessAppRequest struct {
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	SessionDuration string `json:"session_duration"`
}

func updateAccessApp(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		appID := c.Param("appId")
		if appID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("app ID is required")
		}
		var req updateAccessAppRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		app, err := cfUC.UpdateAccessApp(c.Request().Context(), appID, req.Name, req.Domain, req.SessionDuration)
		if err != nil {
			return nil, err
		}
		return responsew.Success(app, "Access app updated"), nil
	}
}

func deleteAccessApp(cfUC cloudflare.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		appID := c.Param("appId")
		if appID == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("app ID is required")
		}
		if err := cfUC.DeleteAccessApp(c.Request().Context(), appID); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "Access app deleted"), nil
	}
}

// --- Binding handlers ---

type createBindingRequest struct {
	AppID    string `json:"app_id"`
	ZoneID   string `json:"zone_id"`
	Domain   string `json:"domain"`
	TunnelID string `json:"tunnel_id"`
	Service  string `json:"service"`
}

func createBinding(bUC bindingUC.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		if _, err := validateProjectOwnership(c, projectUC); err != nil {
			return nil, err
		}
		var req createBindingRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		result, err := bUC.Create(c.Request().Context(), req.AppID, req.ZoneID, req.Domain, req.TunnelID, req.Service)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Binding created"), nil
	}
}

func listBindings(bUC bindingUC.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		if _, err := validateProjectOwnership(c, projectUC); err != nil {
			return nil, err
		}
		appID := c.Param("appId")
		result, err := bUC.List(c.Request().Context(), appID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Bindings retrieved"), nil
	}
}

func getBinding(bUC bindingUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		bindingID, err := strconv.ParseUint(c.Param("bindingId"), 10, 64)
		if err != nil {
			return nil, statusw.InvalidReqParam.WithCustomMessage("Invalid binding ID")
		}
		result, err := bUC.Get(c.Request().Context(), uint(bindingID))
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Binding retrieved"), nil
	}
}

func deleteBinding(bUC bindingUC.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		bindingID, err := strconv.ParseUint(c.Param("bindingId"), 10, 64)
		if err != nil {
			return nil, statusw.InvalidReqParam.WithCustomMessage("Invalid binding ID")
		}
		if err := bUC.Delete(c.Request().Context(), uint(bindingID)); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "Binding deleted"), nil
	}
}