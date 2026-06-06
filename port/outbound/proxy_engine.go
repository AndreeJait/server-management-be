package outbound

import "net/http"

type ProxyEngine interface {
	UpdateRoute(appID, blueTarget, greenTarget string, trafficPercent int)
	RegisterDomain(appID, domain string)
	RemoveRoute(appID string)
	Handler() http.Handler
	Lookup(host string) (target string, ok bool)
}