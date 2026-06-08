package echo

import (
	"net/http"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/inbound/app"
	"github.com/AndreeJait/server-management-be/port/inbound/appfile"
	"github.com/AndreeJait/server-management-be/port/inbound/auth"
	bindingUC "github.com/AndreeJait/server-management-be/port/inbound/binding"
	"github.com/AndreeJait/server-management-be/port/inbound/cloudflare"
	configUC "github.com/AndreeJait/server-management-be/port/inbound/config"
	"github.com/AndreeJait/server-management-be/port/inbound/deployment"
	proxyUC "github.com/AndreeJait/server-management-be/port/inbound/proxy"
	"github.com/AndreeJait/server-management-be/port/inbound/health"
	"github.com/AndreeJait/server-management-be/port/inbound/project"
	"github.com/AndreeJait/server-management-be/port/inbound/registry"
	"github.com/AndreeJait/server-management-be/port/inbound/role"
	"github.com/AndreeJait/server-management-be/port/inbound/user"
	"github.com/AndreeJait/go-utility/v2/authw"
	httpw "github.com/AndreeJait/go-utility/v2/httpw/echow"
	"github.com/labstack/echo/v5"
)

// Required for swagger annotations
var _ = entity.Health{}
var _ = entity.Project{}
var _ = entity.App{}
var _ = entity.Deployment{}
var _ = entity.PipelineStep{}

// RegisterRoutes registers all HTTP routes on the Echo engine.
func RegisterRoutes(
	e *echo.Echo,
	healthUC health.UseCase,
	authUC auth.UseCase,
	userUC user.UseCase,
	roleUC role.UseCase,
	projectUC project.UseCase,
	appUC app.UseCase,
	registryUC registry.UseCase,
	deployUC deployment.UseCase,
	appFileUC appfile.UseCase,
	cfUC cloudflare.UseCase,
	bindingUC bindingUC.UseCase,
	proxyUC proxyUC.UseCase,
	configUC configUC.UseCase,
	authenticator authw.Authenticator,
	rbac *authw.RBAC,
) {
	// Public routes
	e.GET("/health", httpw.Bind(checkHealth(healthUC)))
	e.POST("/auth/login", httpw.Bind(login(authUC)))
	e.POST("/auth/refresh", httpw.Bind(refreshToken(authUC)))

	// Webhook routes (no auth middleware — uses deploy_token)
	e.POST("/webhook/deploy", httpw.Bind(webhookDeploy(deployUC)))

	// Terminal (WebSocket) — authenticates via query param token, not header
	e.GET("/projects/:id/apps/:appId/terminal", echo.WrapHandler(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		terminalHandler(deployUC, appUC, authenticator, rbac)(w, r)
	})))

	// Protected routes — auth middleware applied at group level
	protected := e.Group("", authMiddleware(authenticator))
	protected.GET("/auth/me", httpw.Bind(getMe(userUC)))

	// Project routes
	protected.POST("/projects", httpw.Bind(createProject(projectUC)))
	protected.GET("/projects", httpw.Bind(listProjects(projectUC)))
	protected.GET("/projects/:id", httpw.Bind(getProject(projectUC)))
	protected.PUT("/projects/:id", httpw.Bind(updateProject(projectUC)))
	protected.DELETE("/projects/:id", httpw.Bind(deleteProject(projectUC)))

	// App routes (nested under projects)
	protected.POST("/projects/:id/apps", httpw.Bind(createApp(appUC, projectUC)))
	protected.GET("/projects/:id/apps", httpw.Bind(listApps(appUC, projectUC)))
	protected.GET("/projects/:id/apps/:appId", httpw.Bind(getApp(appUC, projectUC)))
	protected.PUT("/projects/:id/apps/:appId", httpw.Bind(updateApp(appUC, projectUC)))
	protected.DELETE("/projects/:id/apps/:appId", httpw.Bind(deleteApp(appUC, projectUC)))
	protected.POST("/projects/:id/apps/:appId/regenerate-token", httpw.Bind(regenerateDeployToken(appUC, projectUC)))

	// Deployment routes
	protected.GET("/projects/:id/apps/:appId/deployments", httpw.Bind(listDeployments(deployUC, projectUC, appUC)))
	protected.GET("/projects/:id/apps/:appId/deployments/:deployId", httpw.Bind(getDeployment(deployUC, projectUC)))
	protected.GET("/projects/:id/apps/:appId/logs", httpw.Bind(getContainerLogs(deployUC, projectUC, appUC)))
	protected.POST("/projects/:id/apps/:appId/deploy", httpw.Bind(deployApp(deployUC, projectUC, appUC)))
	protected.POST("/projects/:id/apps/:appId/stop", httpw.Bind(stopContainer(deployUC, projectUC, appUC)))


	// App file routes
	protected.GET("/projects/:id/apps/:appId/files", httpw.Bind(listAppFiles(appFileUC, projectUC)))
	protected.POST("/projects/:id/apps/:appId/files", httpw.Bind(createAppFile(appFileUC, projectUC)))
	protected.PUT("/projects/:id/apps/:appId/files/:fileId", httpw.Bind(updateAppFile(appFileUC, projectUC)))
	protected.DELETE("/projects/:id/apps/:appId/files/:fileId", httpw.Bind(deleteAppFile(appFileUC, projectUC)))
	protected.POST("/projects/:id/apps/:appId/files/upload", httpw.Bind(uploadAppFile(appFileUC, projectUC)))
	protected.GET("/projects/:id/apps/:appId/files/:fileId/download", downloadAppFile(appFileUC, projectUC))

	// Folder routes
	protected.POST("/projects/:id/apps/:appId/folders", httpw.Bind(createFolder(appFileUC, projectUC)))
	protected.DELETE("/projects/:id/apps/:appId/folders", httpw.Bind(deleteFolder(appFileUC, projectUC)))

	// Project-scoped registry credentials
	protected.POST("/projects/:id/registry-credentials", httpw.Bind(createProjectRegistryCredential(registryUC, projectUC)))
	protected.GET("/projects/:id/registry-credentials", httpw.Bind(listProjectRegistryCredentials(registryUC, projectUC)))
	protected.PUT("/projects/:id/registry-credentials/:credId", httpw.Bind(updateRegistryCredential(registryUC)))
	protected.DELETE("/projects/:id/registry-credentials/:credId", httpw.Bind(deleteRegistryCredential(registryUC)))

	// Cloudflare routes — read (cloudflare:read)
	cfRead := protected.Group("/cloudflare", rbacMiddleware(rbac, "cloudflare:read"))
	cfRead.GET("/accounts", httpw.Bind(listAccounts(cfUC)))
	cfRead.GET("/zones", httpw.Bind(listZones(cfUC)))
	cfRead.GET("/zones/:zoneId/dns", httpw.Bind(listDNSRecords(cfUC)))
	cfRead.GET("/tunnels", httpw.Bind(listTunnels(cfUC)))
	cfRead.GET("/tunnels/:tunnelId/config", httpw.Bind(getTunnelConfig(cfUC)))
	cfRead.GET("/access-apps", httpw.Bind(listAccessApps(cfUC)))

	// Cloudflare routes — write (cloudflare:write)
	cfWrite := protected.Group("/cloudflare", rbacMiddleware(rbac, "cloudflare:write"))
	cfWrite.POST("/zones/:zoneId/dns", httpw.Bind(createDNSRecord(cfUC)))
	cfWrite.PUT("/zones/:zoneId/dns/:recordId", httpw.Bind(updateDNSRecord(cfUC)))
	cfWrite.DELETE("/zones/:zoneId/dns/:recordId", httpw.Bind(deleteDNSRecord(cfUC)))
	cfWrite.POST("/access-apps", httpw.Bind(createAccessApp(cfUC)))
	cfWrite.PUT("/access-apps/:appId", httpw.Bind(updateAccessApp(cfUC)))
	cfWrite.DELETE("/access-apps/:appId", httpw.Bind(deleteAccessApp(cfUC)))

	// App binding routes
	protected.POST("/projects/:id/apps/:appId/bindings", httpw.Bind(createBinding(bindingUC, projectUC)))
	protected.GET("/projects/:id/apps/:appId/bindings", httpw.Bind(listBindings(bindingUC, projectUC)))
	protected.GET("/bindings/:bindingId", httpw.Bind(getBinding(bindingUC)))
	protected.DELETE("/bindings/:bindingId", httpw.Bind(deleteBinding(bindingUC)))

	// Proxy routes — read (proxy:read)
	proxyRead := protected.Group("/proxy", rbacMiddleware(rbac, "proxy:read"))
	proxyRead.GET("/state", httpw.Bind(listProxyStates(proxyUC)))
	proxyRead.GET("/state/:appId", httpw.Bind(getProxyState(proxyUC)))

	// Proxy routes — write (proxy:write)
	proxyWrite := protected.Group("/proxy", rbacMiddleware(rbac, "proxy:write"))
	proxyWrite.PUT("/state/:appId/traffic", httpw.Bind(setTraffic(proxyUC)))
	proxyWrite.POST("/state/:appId/rollback", httpw.Bind(rollbackProxy(proxyUC)))

	// Config routes — read (configs:read)
	configRead := protected.Group("/config", rbacMiddleware(rbac, "configs:read"))
	configRead.GET("/settings", httpw.Bind(getSettings(configUC)))
	configRead.GET("/settings/:section", httpw.Bind(getSettingsBySection(configUC)))
	configRead.GET("/proxy/access-stats", httpw.Bind(getDomainRequestCounts(configUC)))

	// Config routes — write (configs:write)
	configWrite := protected.Group("/config", rbacMiddleware(rbac, "configs:write"))
	configWrite.PUT("/settings", httpw.Bind(updateSettings(configUC)))

	// Admin routes — requires users:write permission
	admin := protected.Group("/admin", rbacMiddleware(rbac, "users:write"))
	admin.POST("/users", httpw.Bind(createUser(userUC)))
	admin.GET("/users", httpw.Bind(listUsers(userUC)))
	admin.GET("/users/:id", httpw.Bind(getUser(userUC)))
	admin.PUT("/users/:id", httpw.Bind(updateUser(userUC)))
	admin.PUT("/users/:id/roles", httpw.Bind(updateUserRoles(userUC)))
	admin.GET("/roles", httpw.Bind(listRoles(roleUC)))
	admin.PUT("/roles/:name/permissions", httpw.Bind(updateRolePermissions(roleUC)))

	// Global registry credentials — admin only
	admin.POST("/registry-credentials", httpw.Bind(createGlobalRegistryCredential(registryUC)))
	admin.GET("/registry-credentials", httpw.Bind(listGlobalRegistryCredentials(registryUC)))
	admin.PUT("/registry-credentials/:id", httpw.Bind(updateAdminRegistryCredential(registryUC)))
	admin.DELETE("/registry-credentials/:id", httpw.Bind(deleteAdminRegistryCredential(registryUC)))
}