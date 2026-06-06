package echo

import (
	"github.com/AndreeJait/server-management-be/port/inbound/registry"
	"github.com/AndreeJait/server-management-be/port/inbound/project"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/labstack/echo/v5"
)

// --- Registry credential handlers ---

type createRegistryCredentialRequest struct {
	RegistryURL string `json:"registry_url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

func createProjectRegistryCredential(registryUC registry.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		var req createRegistryCredentialRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		cred, err := registryUC.Create(c.Request().Context(), "project", &pid, req.RegistryURL, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return responsew.Success(cred, "Registry credential created"), nil
	}
}

func listProjectRegistryCredentials(registryUC registry.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		creds, err := registryUC.ListByProject(c.Request().Context(), pid)
		if err != nil {
			return nil, err
		}
		return responsew.Success(creds, "Registry credentials retrieved"), nil
	}
}

type updateRegistryCredentialRequest struct {
	RegistryURL string `json:"registry_url"`
	Username    string `json:"username"`
	Password    string `json:"password"`
}

func updateRegistryCredential(registryUC registry.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req updateRegistryCredentialRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		cred, err := registryUC.Update(c.Request().Context(), c.Param("credId"), req.RegistryURL, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return responsew.Success(cred, "Registry credential updated"), nil
	}
}

func deleteRegistryCredential(registryUC registry.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		if err := registryUC.Delete(c.Request().Context(), c.Param("credId")); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "Registry credential deleted"), nil
	}
}

func createGlobalRegistryCredential(registryUC registry.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req createRegistryCredentialRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		cred, err := registryUC.Create(c.Request().Context(), "global", nil, req.RegistryURL, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return responsew.Success(cred, "Global registry credential created"), nil
	}
}

func listGlobalRegistryCredentials(registryUC registry.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		creds, err := registryUC.ListGlobal(c.Request().Context())
		if err != nil {
			return nil, err
		}
		return responsew.Success(creds, "Global registry credentials retrieved"), nil
	}
}

func updateAdminRegistryCredential(registryUC registry.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req updateRegistryCredentialRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		cred, err := registryUC.Update(c.Request().Context(), c.Param("id"), req.RegistryURL, req.Username, req.Password)
		if err != nil {
			return nil, err
		}
		return responsew.Success(cred, "Registry credential updated"), nil
	}
}

func deleteAdminRegistryCredential(registryUC registry.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		if err := registryUC.Delete(c.Request().Context(), c.Param("id")); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "Registry credential deleted"), nil
	}
}