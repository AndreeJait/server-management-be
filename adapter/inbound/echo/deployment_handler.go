package echo

import (
	"strconv"

	"github.com/AndreeJait/server-management-be/port/inbound/app"
	"github.com/AndreeJait/server-management-be/port/inbound/appfile"
	"github.com/AndreeJait/server-management-be/port/inbound/deployment"
	"github.com/AndreeJait/server-management-be/port/inbound/project"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/AndreeJait/go-utility/v2/statusw"
	"github.com/labstack/echo/v5"
)

// --- Webhook handlers ---

type webhookDeployRequest struct {
	AppID       string `json:"app_id"`
	DeployToken string `json:"deploy_token"`
	Image       string `json:"image"`
}

func webhookDeploy(deployUC deployment.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		var req webhookDeployRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		if req.AppID == "" || req.DeployToken == "" {
			return nil, statusw.InvalidReqParam.WithCustomMessage("app_id and deploy_token are required")
		}
		result, err := deployUC.Deploy(c.Request().Context(), req.AppID, req.DeployToken, req.Image)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Deployment triggered"), nil
	}
}

// --- Authenticated deploy handlers ---

type deployAppRequest struct {
	Image string `json:"image"`
}

func deployApp(deployUC deployment.UseCase, projectUC project.UseCase, appUC app.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		a, err := appUC.Get(c.Request().Context(), pid, c.Param("appId"))
		if err != nil {
			return nil, err
		}
		var req deployAppRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		result, err := deployUC.DeployApp(c.Request().Context(), a.AppID, req.Image)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Deployment triggered"), nil
	}
}

func stopContainer(deployUC deployment.UseCase, projectUC project.UseCase, appUC app.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		a, err := appUC.Get(c.Request().Context(), pid, c.Param("appId"))
		if err != nil {
			return nil, err
		}
		if err := deployUC.StopContainer(c.Request().Context(), a.AppID); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "Container stopped"), nil
	}
}

// --- Deployment query handlers ---

func listDeployments(deployUC deployment.UseCase, projectUC project.UseCase, appUC app.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		if _, err := validateProjectOwnership(c, projectUC); err != nil {
			return nil, err
		}
		pid, err := parseProjectID(c)
		if err != nil {
			return nil, err
		}
		a, err := appUC.Get(c.Request().Context(), pid, c.Param("appId"))
		if err != nil {
			return nil, err
		}
		deployments, err := deployUC.ListByAppID(c.Request().Context(), a.AppID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(deployments, "Deployments retrieved"), nil
	}
}

func getDeployment(deployUC deployment.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		if _, err := validateProjectOwnership(c, projectUC); err != nil {
			return nil, err
		}
		d, err := deployUC.Get(c.Request().Context(), c.Param("deployId"))
		if err != nil {
			return nil, err
		}
		return responsew.Success(d, "Deployment retrieved"), nil
	}
}

func getContainerLogs(deployUC deployment.UseCase, projectUC project.UseCase, appUC app.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		a, err := appUC.Get(c.Request().Context(), pid, c.Param("appId"))
		if err != nil {
			return nil, err
		}
		tail := 100
		if t := c.QueryParam("tail"); t != "" {
			if n, err := strconv.Atoi(t); err == nil && n > 0 {
				tail = n
			}
		}
		logs, err := deployUC.GetContainerLogs(c.Request().Context(), a.AppID, tail)
		if err != nil {
			return nil, err
		}
		return responsew.Success(map[string]string{"logs": logs}, "Container logs retrieved"), nil
	}
}

// --- App file handlers ---

type createAppFileRequest struct {
	Path    string `json:"path" validate:"required"`
	Content string `json:"content" validate:"required"`
}

func createAppFile(appFileUC appfile.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		var req createAppFileRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		f, err := appFileUC.Create(c.Request().Context(), pid, c.Param("appId"), req.Path, req.Content)
		if err != nil {
			return nil, err
		}
		return responsew.Success(f, "File created"), nil
	}
}

func listAppFiles(appFileUC appfile.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		files, err := appFileUC.List(c.Request().Context(), pid, c.Param("appId"))
		if err != nil {
			return nil, err
		}
		return responsew.Success(files, "Files retrieved"), nil
	}
}

type updateAppFileRequest struct {
	Path    string `json:"path" validate:"required"`
	Content string `json:"content" validate:"required"`
}

func updateAppFile(appFileUC appfile.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		fileID, err := strconv.ParseUint(c.Param("fileId"), 10, 64)
		if err != nil {
			return nil, statusw.InvalidReqParam.WithCustomMessage("Invalid file ID")
		}
		var req updateAppFileRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		f, err := appFileUC.Update(c.Request().Context(), pid, c.Param("appId"), uint(fileID), req.Path, req.Content)
		if err != nil {
			return nil, err
		}
		return responsew.Success(f, "File updated"), nil
	}
}

func deleteAppFile(appFileUC appfile.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		fileID, err := strconv.ParseUint(c.Param("fileId"), 10, 64)
		if err != nil {
			return nil, statusw.InvalidReqParam.WithCustomMessage("Invalid file ID")
		}
		if err := appFileUC.Delete(c.Request().Context(), pid, c.Param("appId"), uint(fileID)); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "File deleted"), nil
	}
}

// --- Folder handlers ---

type createFolderRequest struct {
	Path string `json:"path" validate:"required"`
}

func createFolder(appFileUC appfile.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		var req createFolderRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		result, err := appFileUC.CreateFolder(c.Request().Context(), pid, c.Param("appId"), req.Path)
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Folder created"), nil
	}
}

type deleteFolderRequest struct {
	Path string `json:"path" validate:"required"`
}

func deleteFolder(appFileUC appfile.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		var req deleteFolderRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		if err := appFileUC.DeleteFolder(c.Request().Context(), pid, c.Param("appId"), req.Path); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "Folder deleted"), nil
	}
}