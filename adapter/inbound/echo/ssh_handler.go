package echo

import (
	"strconv"

	"github.com/AndreeJait/server-management-be/domain/entity"
	sshInbound "github.com/AndreeJait/server-management-be/port/inbound/ssh"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/go-utility/v2/statusw"
	"github.com/labstack/echo/v5"
)

type createSSHHostRequest struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	AuthMethod string `json:"auth_method"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
}

func createSSHHost(sshUC sshInbound.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID, err := getUserIDFromContext(c)
		if err != nil {
			return nil, statusw.InvalidCredential.WithCustomMessage("Authentication required")
		}
		var req createSSHHostRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		if req.Name == "" || req.Host == "" || req.Username == "" || req.AuthMethod == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("name, host, username, and auth_method are required")
		}
		if req.Port == 0 {
			req.Port = 22
		}
		result, err := sshUC.Create(c.Request().Context(), req.Name, req.Host, req.Port, req.Username, req.AuthMethod, req.Password, req.PrivateKey, userID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "SSH host created"), nil
	}
}

func listSSHHosts(sshUC sshInbound.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID, err := getUserIDFromContext(c)
		if err != nil {
			return nil, statusw.InvalidCredential.WithCustomMessage("Authentication required")
		}
		hosts, err := sshUC.List(c.Request().Context(), userID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(hosts, "SSH hosts retrieved"), nil
	}
}

func getSSHHost(sshUC sshInbound.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID, err := getUserIDFromContext(c)
		if err != nil {
			return nil, statusw.InvalidCredential.WithCustomMessage("Authentication required")
		}
		hostID, err := strconv.ParseUint(c.Param("hostId"), 10, 64)
		if err != nil {
			return nil, statusw.InvalidReqParam.WithCustomMessage("Invalid host ID")
		}
		result, err := sshUC.Get(c.Request().Context(), uint(hostID), userID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "SSH host retrieved"), nil
	}
}

type updateSSHHostRequest struct {
	Name       string `json:"name"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Username   string `json:"username"`
	AuthMethod string `json:"auth_method"`
	Password   string `json:"password,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
}

func updateSSHHost(sshUC sshInbound.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID, err := getUserIDFromContext(c)
		if err != nil {
			return nil, statusw.InvalidCredential.WithCustomMessage("Authentication required")
		}
		hostID, err := strconv.ParseUint(c.Param("hostId"), 10, 64)
		if err != nil {
			return nil, statusw.InvalidReqParam.WithCustomMessage("Invalid host ID")
		}
		var req updateSSHHostRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		if req.Name == "" || req.Host == "" || req.Username == "" || req.AuthMethod == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("name, host, username, and auth_method are required")
		}
		if req.Port == 0 {
			req.Port = 22
		}
		result, err := sshUC.Update(c.Request().Context(), uint(hostID), userID, req.Name, req.Host, req.Port, req.Username, req.AuthMethod, req.Password, req.PrivateKey)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "SSH host updated"), nil
	}
}

func deleteSSHHost(sshUC sshInbound.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		userID, err := getUserIDFromContext(c)
		if err != nil {
			return nil, statusw.InvalidCredential.WithCustomMessage("Authentication required")
		}
		hostID, err := strconv.ParseUint(c.Param("hostId"), 10, 64)
		if err != nil {
			return nil, statusw.InvalidReqParam.WithCustomMessage("Invalid host ID")
		}
		if err := sshUC.Delete(c.Request().Context(), uint(hostID), userID); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "SSH host deleted"), nil
	}
}

// Ensure entity types are referenced for swagger annotations
var _ = entity.SSHHost{}