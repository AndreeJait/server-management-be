package config

import "sync"

// RuntimeConfig wraps AppConfig with thread-safe access and hot-reload support.
type RuntimeConfig struct {
	mu  sync.RWMutex
	cfg *AppConfig

	onRateLimitChange func(rps int)
}

func NewRuntimeConfig(cfg *AppConfig) *RuntimeConfig {
	return &RuntimeConfig{cfg: cfg}
}

// --- Proxy accessors ---

func (r *RuntimeConfig) GetProxyEnabled() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Proxy.Enabled
}

func (r *RuntimeConfig) GetHealthCheckPath() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Proxy.HealthCheckPath
}

func (r *RuntimeConfig) GetShiftIntervalSec() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Proxy.ShiftIntervalSec
}

func (r *RuntimeConfig) GetTunnelServiceURL() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Proxy.TunnelServiceURL
}

func (r *RuntimeConfig) GetRateLimitRPS() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Proxy.RateLimitRPS
}

// --- Docker accessors ---

func (r *RuntimeConfig) GetDockerNetwork() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Docker.Network
}

func (r *RuntimeConfig) GetDockerHostBase() string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.cfg.Docker.HostBase
}

// --- Hot-reload setters ---

func (r *RuntimeConfig) SetProxyEnabled(v bool) {
	r.mu.Lock()
	r.cfg.Proxy.Enabled = v
	r.mu.Unlock()
}

func (r *RuntimeConfig) SetHealthCheckPath(v string) {
	r.mu.Lock()
	r.cfg.Proxy.HealthCheckPath = v
	r.mu.Unlock()
}

func (r *RuntimeConfig) SetShiftIntervalSec(v int) {
	r.mu.Lock()
	r.cfg.Proxy.ShiftIntervalSec = v
	r.mu.Unlock()
}

func (r *RuntimeConfig) SetTunnelServiceURL(v string) {
	r.mu.Lock()
	r.cfg.Proxy.TunnelServiceURL = v
	r.mu.Unlock()
}

func (r *RuntimeConfig) SetRateLimitRPS(v int) {
	r.mu.Lock()
	r.cfg.Proxy.RateLimitRPS = v
	cb := r.onRateLimitChange
	r.mu.Unlock()
	if cb != nil {
		cb(v)
	}
}

func (r *RuntimeConfig) SetDockerNetwork(v string) {
	r.mu.Lock()
	r.cfg.Docker.Network = v
	r.mu.Unlock()
}

func (r *RuntimeConfig) SetDockerHostBase(v string) {
	r.mu.Lock()
	r.cfg.Docker.HostBase = v
	r.mu.Unlock()
}

// --- Callbacks ---

func (r *RuntimeConfig) OnRateLimitChange(fn func(rps int)) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.onRateLimitChange = fn
}