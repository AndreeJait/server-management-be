package config

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/AndreeJait/go-utility/v2/logw"
	"github.com/spf13/viper"
)

// AppConfig holds all application configuration.
type AppConfig struct {
	App struct {
		Name     string `mapstructure:"name"`
		Host     string `mapstructure:"host"`
		Env      string `mapstructure:"env"`
		HTTPPort int    `mapstructure:"http_port"`
	} `mapstructure:"app"`

	HTTP struct {
		SwaggerHost    string   `mapstructure:"swagger_host"`
		SwaggerSchemes []string `mapstructure:"swagger_schemes"`
		EnableSwagger  bool     `mapstructure:"enable_swagger"`
		DebugMode      bool     `mapstructure:"debug_mode"`
		CORSOrigins    []string `mapstructure:"cors_origins"`
	} `mapstructure:"http"`

	Log struct {
		Level       string         `mapstructure:"level"`
		Format      logw.LogFormat `mapstructure:"format"`
		WriteToFile bool           `mapstructure:"write_to_file"`
		FilePath    string         `mapstructure:"file_path"`
	} `mapstructure:"log"`

	DB struct {
		Driver          string        `mapstructure:"driver"`
		Dialect         string        `mapstructure:"dialect"`
		DSN             string        `mapstructure:"dsn"`
		MaxOpenConns    int           `mapstructure:"max_open_conns"`
		MaxIdleConns    int           `mapstructure:"max_idle_conns"`
		ConnMaxLifetime time.Duration `mapstructure:"conn_max_lifetime"`
		DebugMode      bool          `mapstructure:"debug_mode"`
	} `mapstructure:"db"`

	Redis struct {
		Address  string `mapstructure:"address"`
		Password string `mapstructure:"password"`
		DB       int    `mapstructure:"db"`
		PoolSize int    `mapstructure:"pool_size"`
		DebugMode bool  `mapstructure:"debug_mode"`
	} `mapstructure:"redis"`

	Graceful struct {
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`
	} `mapstructure:"graceful"`

	Auth struct {
		JWTSecret     string        `mapstructure:"jwt_secret"`
		JWTTTL       time.Duration `mapstructure:"jwt_ttl"`
		AdminEmail    string        `mapstructure:"admin_email"`
		AdminPassword string        `mapstructure:"admin_password"`
	} `mapstructure:"auth"`

	Docker struct {
		Host     string `mapstructure:"host"`
		Network  string `mapstructure:"network"`
		HostBase string `mapstructure:"host_base"`
	} `mapstructure:"docker"`

	Cloudflare struct {
		APIToken  string `mapstructure:"api_token"`
		AccountID string `mapstructure:"account_id"`
	} `mapstructure:"cloudflare"`

	Proxy struct {
		Enabled          bool   `mapstructure:"enabled"`
		HealthCheckPath  string `mapstructure:"health_check_path"`
		ShiftIntervalSec int    `mapstructure:"shift_interval_sec"`
		TunnelServiceURL string `mapstructure:"tunnel_service_url"`
		RateLimitRPS     int    `mapstructure:"rate_limit_rps"`
	} `mapstructure:"proxy"`
}

// Load reads the base config file at configPath, then merges app.local.yaml
// from the same directory if it exists. Environment variables override both.
func Load(configPath string) (*AppConfig, error) {
	v := viper.New()

	v.SetConfigFile(configPath)
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to load config: %w", err)
	}

	// Merge local overrides if app.local.yaml exists alongside the base config
	localPath := strings.Replace(configPath, "app.yaml", "app.local.yaml", 1)
	if _, err := os.Stat(localPath); err == nil {
		v.SetConfigFile(localPath)
		if err := v.MergeInConfig(); err != nil {
			return nil, fmt.Errorf("failed to load local config: %w", err)
		}
	}

	cfg := &AppConfig{}
	if err := v.Unmarshal(cfg); err != nil {
		return nil, fmt.Errorf("failed to unmarshal config: %w", err)
	}

	// Apply defaults
	if cfg.App.Host == "" {
		cfg.App.Host = "localhost"
	}
	if cfg.App.HTTPPort == 0 {
		cfg.App.HTTPPort = 8080
	}
	if cfg.DB.Driver == "" {
		cfg.DB.Driver = "gorm"
	}
	if cfg.DB.Dialect == "" {
		cfg.DB.Dialect = "postgres"
	}
	if cfg.Graceful.ShutdownTimeout == 0 {
		cfg.Graceful.ShutdownTimeout = 10 * time.Second
	}
	if cfg.Docker.Host == "" {
		cfg.Docker.Host = "unix:///var/run/docker.sock"
	}
	if cfg.Docker.HostBase == "" {
		cfg.Docker.HostBase = "/home/andree/docker/app-server"
	}
	if cfg.Auth.JWTSecret == "" {
		cfg.Auth.JWTSecret = "change-me-in-production"
	}
	if cfg.Auth.JWTTTL == 0 {
		cfg.Auth.JWTTTL = 24 * time.Hour
	}
	if cfg.Proxy.HealthCheckPath == "" {
		cfg.Proxy.HealthCheckPath = "/health"
	}
	if cfg.Proxy.ShiftIntervalSec == 0 {
		cfg.Proxy.ShiftIntervalSec = 30
	}
	if cfg.Proxy.TunnelServiceURL == "" {
		cfg.Proxy.TunnelServiceURL = fmt.Sprintf("http://localhost:%d", cfg.App.HTTPPort)
	}

	return cfg, nil
}