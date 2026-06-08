package main

import (
	"context"
	"net/http"
	"strconv"

	"github.com/AndreeJait/server-management-be/config"
	portOutbound "github.com/AndreeJait/server-management-be/port/outbound"
	"github.com/AndreeJait/go-utility/v2/logw"
	"go.uber.org/dig"
)

// CleanupCollector accumulates cleanup functions from infrastructure providers.
// After dig resolves all dependencies, its Cleanup method is passed to gracefulw.
type CleanupCollector struct {
	cleanups []func(ctx context.Context) error
}

// Add appends a cleanup function.
func (cc *CleanupCollector) Add(fn func(ctx context.Context) error) {
	cc.cleanups = append(cc.cleanups, fn)
}

// Cleanup runs all collected cleanup functions, returning the first error.
func (cc *CleanupCollector) Cleanup(ctx context.Context) error {
	var firstErr error
	for _, fn := range cc.cleanups {
		if err := fn(ctx); err != nil && firstErr == nil {
			firstErr = err
		}
	}
	return firstErr
}

// wire builds the dependency graph using dig and returns the HTTP handler + cleanup.
func wire(cfg *config.AppConfig) (http.Handler, func(ctx context.Context) error, error) {
	c := dig.New()

	// Root values
	c.Provide(func() *config.AppConfig { return cfg })
	c.Provide(func() *config.RuntimeConfig { return config.NewRuntimeConfig(cfg) })
	c.Provide(func() *CleanupCollector { return &CleanupCollector{} })

	// Register all providers
	provideInfrastructure(c)
	provideServices(c)
	initAppFileDeployFunc(c)
	provideRouter(c)

	// Invoke to build the handler
	var handler http.Handler
	if err := c.Invoke(func(h http.Handler) {
		handler = h
	}); err != nil {
		return nil, nil, err
	}

	// Retrieve cleanup collector
	var cc *CleanupCollector
	if err := c.Invoke(func(collector *CleanupCollector) {
		cc = collector
	}); err != nil {
		return nil, nil, err
	}

	// Load settings from DB into RuntimeConfig (overrides YAML defaults)
	loadSettingsFromDB(c)

	// Restore proxy routes from database at startup
	if cfg.Proxy.Enabled {
		if err := c.Invoke(func(runtimeCfg *config.RuntimeConfig, proxyStateRepo portOutbound.ProxyStateRepository, bindingRepo portOutbound.AppBindingRepository, proxyEngine portOutbound.ProxyEngine) {
			restoreProxyRoutes(proxyStateRepo, bindingRepo, proxyEngine)
		}); err != nil {
			logw.Warningf("proxy: failed to restore routes at startup: %v", err)
		}
	}

	return handler, cc.Cleanup, nil
}

// loadSettingsFromDB reads all settings from the database and applies them to RuntimeConfig,
// so DB values override YAML defaults without a restart.
func loadSettingsFromDB(c *dig.Container) {
	if err := c.Invoke(func(settingRepo portOutbound.SettingRepository, runtimeCfg *config.RuntimeConfig) {
		settings, err := settingRepo.FindAll(context.Background())
		if err != nil {
			logw.Warningf("config: failed to load settings from DB: %v", err)
			return
		}
		for _, s := range settings {
			applySettingFromDB(runtimeCfg, s.Section, s.Key, s.Value)
		}
		logw.Infof("config: loaded %d settings from DB", len(settings))
	}); err != nil {
		logw.Warningf("config: failed to invoke loadSettingsFromDB: %v", err)
	}
}

func applySettingFromDB(rc *config.RuntimeConfig, section, key, value string) {
	switch section + "." + key {
	case "proxy.enabled":
		rc.SetProxyEnabled(value == "true")
	case "proxy.health_check_path":
		rc.SetHealthCheckPath(value)
	case "proxy.shift_interval_sec":
		if v, err := parseInt(value); err == nil {
			rc.SetShiftIntervalSec(v)
		}
	case "proxy.tunnel_service_url":
		rc.SetTunnelServiceURL(value)
	case "proxy.rate_limit_rps":
		if v, err := parseInt(value); err == nil {
			rc.SetRateLimitRPS(v)
		}
	case "docker.network":
		rc.SetDockerNetwork(value)
	case "docker.host_base":
		rc.SetDockerHostBase(value)
	}
}

func parseInt(s string) (int, error) {
	return strconv.Atoi(s)
}