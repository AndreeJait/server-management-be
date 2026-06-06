package main

import (
	"net/http"
	"strings"

	echoAdapter "github.com/AndreeJait/server-management-be/adapter/inbound/echo"
	"github.com/AndreeJait/server-management-be/adapter/outbound"
	"github.com/AndreeJait/server-management-be/config"
	"github.com/AndreeJait/server-management-be/port/inbound/app"
	"github.com/AndreeJait/server-management-be/port/inbound/appfile"
	"github.com/AndreeJait/server-management-be/port/inbound/auth"
	bindingInbound "github.com/AndreeJait/server-management-be/port/inbound/binding"
	cloudflareInbound "github.com/AndreeJait/server-management-be/port/inbound/cloudflare"
	"github.com/AndreeJait/server-management-be/port/inbound/deployment"
	"github.com/AndreeJait/server-management-be/port/inbound/health"
	proxyInbound "github.com/AndreeJait/server-management-be/port/inbound/proxy"
	"github.com/AndreeJait/server-management-be/port/inbound/project"
	"github.com/AndreeJait/server-management-be/port/inbound/registry"
	"github.com/AndreeJait/server-management-be/port/inbound/role"
	"github.com/AndreeJait/server-management-be/port/inbound/user"
	portOutbound "github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/authw"
	httpwEcho "github.com/AndreeJait/go-utility/v2/httpw/echow"
	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
	"go.uber.org/dig"
)

// provideRouter registers the HTTP router provider into the dig container.
func provideRouter(c *dig.Container) {
	c.Provide(newRouter)
}

// newRouter creates the Echo engine and registers all routes.
func newRouter(
	// kyan:param:start
	cfg *config.AppConfig,
	healthUC health.UseCase,
	authUC auth.UseCase,
	userUC user.UseCase,
	roleUC role.UseCase,
	projectUC project.UseCase,
	appUC app.UseCase,
	registryUC registry.UseCase,
	deployUC deployment.UseCase,
	appFileUC appfile.UseCase,
	cfUC cloudflareInbound.UseCase,
	bindingUC bindingInbound.UseCase,
	proxyUC proxyInbound.UseCase,
	authenticator authw.Authenticator,
	rbac *authw.RBAC,
	proxyEngine portOutbound.ProxyEngine,
	// kyan:param:end
) (http.Handler, error) {
	e := httpwEcho.New(&httpwEcho.Config{
		DebugMode:     cfg.HTTP.DebugMode,
		EnableSwagger: cfg.HTTP.EnableSwagger,
	})

	// Proxy middleware: intercept requests for bound domains BEFORE CORS
	if cfg.Proxy.Enabled {
		var rateLimiter *outbound.RateLimiter
		if cfg.Proxy.RateLimitRPS > 0 {
			rateLimiter = outbound.NewRateLimiter(cfg.Proxy.RateLimitRPS)
		}
		e.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
			return func(c *echo.Context) error {
				host := c.Request().Host
				if idx := strings.LastIndex(host, ":"); idx != -1 {
					host = host[:idx]
				}
				if _, ok := proxyEngine.Lookup(host); ok {
					var handler http.Handler = proxyEngine.Handler()
					if rateLimiter != nil {
						handler = rateLimiter.Middleware(handler)
					}
					handler.ServeHTTP(c.Response(), c.Request())
					return nil
				}
				return next(c)
			}
		})
	}

	corsOrigins := cfg.HTTP.CORSOrigins
	if len(corsOrigins) == 0 {
		corsOrigins = []string{"http://localhost:3000"}
	}

	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins:     corsOrigins,
		AllowMethods:     []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
	}))

	echoAdapter.RegisterRoutes(e, healthUC, authUC, userUC, roleUC, projectUC, appUC, registryUC, deployUC, appFileUC, cfUC, bindingUC, proxyUC, authenticator, rbac)

	return e, nil
}