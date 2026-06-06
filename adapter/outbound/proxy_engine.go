package outbound

import (
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"sync"

	"github.com/AndreeJait/server-management-be/port/outbound"
)

type routeEntry struct {
	BlueTarget     string
	GreenTarget    string
	TrafficPercent int
}

type proxyEngine struct {
	mu     sync.RWMutex
	routes map[string]*routeEntry // domain -> route entry
	appMap map[string]string      // appID -> domain
}

func NewProxyEngine() outbound.ProxyEngine {
	return &proxyEngine{
		routes: make(map[string]*routeEntry),
		appMap: make(map[string]string),
	}
}

func (p *proxyEngine) UpdateRoute(appID, blueTarget, greenTarget string, trafficPercent int) {
	p.mu.Lock()
	defer p.mu.Unlock()

	domain, exists := p.appMap[appID]
	if !exists {
		domain = appID
	}

	p.routes[domain] = &routeEntry{
		BlueTarget:     blueTarget,
		GreenTarget:    greenTarget,
		TrafficPercent: trafficPercent,
	}
	log.Printf("proxy: updated route for domain=%s blue=%s green=%s traffic=%d%%",
		domain, blueTarget, greenTarget, trafficPercent)
}

func (p *proxyEngine) RegisterDomain(appID, domain string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.appMap[appID] = domain

	if entry, exists := p.routes[appID]; exists {
		delete(p.routes, appID)
		p.routes[domain] = entry
	}
}

func (p *proxyEngine) RemoveRoute(appID string) {
	p.mu.Lock()
	defer p.mu.Unlock()
	domain, exists := p.appMap[appID]
	if exists {
		delete(p.routes, domain)
		delete(p.appMap, appID)
	}
}

func (p *proxyEngine) Handler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Host
		if idx := strings.LastIndex(host, ":"); idx != -1 {
			host = host[:idx]
		}

		target, ok := p.selectTarget(host)
		if !ok {
			http.Error(w, "no upstream for host: "+host, http.StatusBadGateway)
			return
		}

		targetURL, err := url.Parse(fmt.Sprintf("http://%s", target))
		if err != nil {
			http.Error(w, "invalid upstream target", http.StatusInternalServerError)
			return
		}

		originalHost := r.Host
		proxy := &httputil.ReverseProxy{
			Director: func(req *http.Request) {
				req.URL.Scheme = targetURL.Scheme
				req.URL.Host = targetURL.Host
				req.Host = targetURL.Host
				req.Header.Set("X-Forwarded-Host", originalHost)
			},
			ErrorHandler: func(w http.ResponseWriter, r *http.Request, err error) {
				log.Printf("proxy: error forwarding to host=%s path=%s: %v", host, r.URL.Path, err)
				http.Error(w, "upstream error", http.StatusBadGateway)
			},
		}
		proxy.ServeHTTP(w, r)
	})
}

func (p *proxyEngine) Lookup(host string) (string, bool) {
	target, ok := p.selectTarget(host)
	return target, ok
}

func (p *proxyEngine) selectTarget(host string) (string, bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	entry, exists := p.routes[host]
	if !exists {
		return "", false
	}

	if entry.BlueTarget == "" && entry.GreenTarget == "" {
		return "", false
	}
	if entry.BlueTarget == "" {
		return entry.GreenTarget, true
	}
	if entry.GreenTarget == "" {
		return entry.BlueTarget, true
	}

	if rand.Intn(100) < entry.TrafficPercent {
		return entry.GreenTarget, true
	}
	return entry.BlueTarget, true
}