package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"

	"github.com/AndreeJait/server-management-be/docs"
	"github.com/AndreeJait/server-management-be/config"
	"github.com/AndreeJait/go-utility/v2/gracefulw"
	"github.com/AndreeJait/go-utility/v2/logw"
)

// @title Server Management API
// @version 1.0
// @description A self-hosted control plane for automating server deployments, managing Cloudflare tunnels, and zero-downtime Blue/Green deployments.
// @BasePath /

func main() {
	configFlag := flag.String("config", "files/config/app.yaml", "Path to config file")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load(*configFlag)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Set swagger host dynamically from config
	if cfg.HTTP.SwaggerHost != "" {
		docs.SwaggerInfo.Host = cfg.HTTP.SwaggerHost
	} else {
		docs.SwaggerInfo.Host = fmt.Sprintf("%s:%d", cfg.App.Host, cfg.App.HTTPPort)
	}

	// Set swagger schemes dynamically from config
	if len(cfg.HTTP.SwaggerSchemes) > 0 {
		docs.SwaggerInfo.Schemes = cfg.HTTP.SwaggerSchemes
	} else {
		docs.SwaggerInfo.Schemes = []string{"http"}
	}

	// Initialize logger
	if err := logw.Init(&logw.LogConfig{
		Level:       cfg.Log.Level,
		Format:      cfg.Log.Format,
		WriteToFile: cfg.Log.WriteToFile,
		FilePath:    cfg.Log.FilePath,
	}); err != nil {
		fmt.Fprintf(os.Stderr, "failed to init logger: %v\n", err)
		os.Exit(1)
	}

	logw.Infof("Starting %s", cfg.App.Name)

	// Wire all dependencies
	handler, cleanup, err := wire(cfg)
	if err != nil {
		logw.Errorf("failed to wire dependencies: %v", err)
		os.Exit(1)
	}

	// Start server with graceful shutdown
	addr := fmt.Sprintf(":%d", cfg.App.HTTPPort)
	srv := &http.Server{Addr: addr, Handler: handler}

	gracefulw.Register("http-server", srv.Shutdown)
	gracefulw.Register("dependencies", cleanup)

	logw.Infof("HTTP server listening on %s", addr)
	gracefulw.Start(srv.ListenAndServe, cfg.Graceful.ShutdownTimeout)
}