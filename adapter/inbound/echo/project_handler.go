package echo

import (
	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/inbound/app"
	"github.com/AndreeJait/server-management-be/port/inbound/project"
	"github.com/AndreeJait/go-utility/v2/responsew"
	"github.com/labstack/echo/v5"
)

// --- Project handlers ---

type createProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func createProject(projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		ownerID, err := getOwnerID(c)
		if err != nil {
			return nil, err
		}
		var req createProjectRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		p, err := projectUC.Create(c.Request().Context(), req.Name, req.Description, ownerID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(p, "Project created"), nil
	}
}

func listProjects(projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		ownerID, err := getOwnerID(c)
		if err != nil {
			return nil, err
		}
		projects, err := projectUC.List(c.Request().Context(), ownerID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(projects, "Projects retrieved"), nil
	}
}

func getProject(projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		ownerID, err := getOwnerID(c)
		if err != nil {
			return nil, err
		}
		p, err := projectUC.Get(c.Request().Context(), c.Param("id"), ownerID)
		if err != nil {
			return nil, err
		}
		return responsew.Success(p, "Project retrieved"), nil
	}
}

type updateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

func updateProject(projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		ownerID, err := getOwnerID(c)
		if err != nil {
			return nil, err
		}
		var req updateProjectRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		p, err := projectUC.Update(c.Request().Context(), c.Param("id"), ownerID, req.Name, req.Description)
		if err != nil {
			return nil, err
		}
		return responsew.Success(p, "Project updated"), nil
	}
}

func deleteProject(projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		ownerID, err := getOwnerID(c)
		if err != nil {
			return nil, err
		}
		if err := projectUC.Delete(c.Request().Context(), c.Param("id"), ownerID); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "Project deleted"), nil
	}
}

// validateProjectOwnership checks the current user owns the project referenced by :id.
func validateProjectOwnership(c *echo.Context, projectUC project.UseCase) (uint, error) {
	ownerID, err := getOwnerID(c)
	if err != nil {
		return 0, err
	}
	pid, err := parseProjectID(c)
	if err != nil {
		return 0, err
	}
	_, err = projectUC.Get(c.Request().Context(), c.Param("id"), ownerID)
	if err != nil {
		return 0, err
	}
	return pid, nil
}

// --- App handlers ---

type createAppRequest struct {
	Name            string `json:"name"`
	FrameworkPreset string `json:"framework_preset"`
}

func createApp(appUC app.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		var req createAppRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		a, err := appUC.Create(c.Request().Context(), pid, req.Name, req.FrameworkPreset)
		if err != nil {
			return nil, err
		}
		return responsew.Success(a, "App created"), nil
	}
}

func listApps(appUC app.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		apps, err := appUC.List(c.Request().Context(), pid)
		if err != nil {
			return nil, err
		}
		return responsew.Success(apps, "Apps retrieved"), nil
	}
}

func getApp(appUC app.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		a, err := appUC.Get(c.Request().Context(), pid, c.Param("appId"))
		if err != nil {
			return nil, err
		}
		return responsew.Success(a, "App retrieved"), nil
	}
}

type updateAppRequest struct {
	Name               string                  `json:"name"`
	FrameworkPreset    string                  `json:"framework_preset"`
	EnvVars            entity.StringMap         `json:"env_vars"`
	VolumeMounts       entity.VolumeMountList   `json:"volume_mounts"`
	PostDeployCommands entity.StringList        `json:"post_deploy_commands"`
	BasePath           string                  `json:"base_path"`
	DefaultImage       string                  `json:"default_image"`
	ContainerPort      string                  `json:"container_port"`
	PublishPort        string                  `json:"publish_port"`
	ContainerName      string                  `json:"container_name"`
	FilesMountPath     string                  `json:"files_mount_path"`
}

func updateApp(appUC app.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		var req updateAppRequest
		if err := c.Bind(&req); err != nil {
			return nil, err
		}
		a, err := appUC.Update(c.Request().Context(), pid, c.Param("appId"), req.Name, req.FrameworkPreset, req.EnvVars, req.VolumeMounts, req.PostDeployCommands, req.BasePath, req.DefaultImage, req.ContainerPort, req.PublishPort, req.ContainerName, req.FilesMountPath)
		if err != nil {
			return nil, err
		}
		return responsew.Success(a, "App updated"), nil
	}
}

func deleteApp(appUC app.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		if err := appUC.Delete(c.Request().Context(), pid, c.Param("appId")); err != nil {
			return nil, err
		}
		return responsew.Success(nil, "App deleted"), nil
	}
}

func regenerateDeployToken(appUC app.UseCase, projectUC project.UseCase) func(c *echo.Context) (any, error) {
	return func(c *echo.Context) (any, error) {
		pid, err := validateProjectOwnership(c, projectUC)
		if err != nil {
			return nil, err
		}
		result, err := appUC.RegenerateDeployToken(c.Request().Context(), pid, c.Param("appId"))
		if err != nil {
			return nil, err
		}
		return responsew.Success(result, "Deploy token regenerated"), nil
	}
}