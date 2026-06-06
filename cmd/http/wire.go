package main

import (
	"context"
	"net/http"

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

	// Restore proxy routes from database at startup
	if cfg.Proxy.Enabled {
		if err := c.Invoke(func(proxyStateRepo portOutbound.ProxyStateRepository, bindingRepo portOutbound.AppBindingRepository, proxyEngine portOutbound.ProxyEngine) {
			restoreProxyRoutes(proxyStateRepo, bindingRepo, proxyEngine)
		}); err != nil {
			logw.Warningf("proxy: failed to restore routes at startup: %v", err)
		}
	}

	return handler, cc.Cleanup, nil
}