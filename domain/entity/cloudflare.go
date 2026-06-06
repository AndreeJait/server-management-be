package entity

type CloudflareAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type,omitempty"`
}

type CloudflareZone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status,omitempty"`
}

type CloudflareDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

type CloudflareTunnel struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	TunType   string `json:"type,omitempty"`
	ConnsCount int   `json:"conns_count,omitempty"`
}

type TunnelIngressRule struct {
	Hostname      string          `json:"hostname,omitempty"`
	Service       string          `json:"service"`
	Path          string          `json:"path,omitempty"`
	OriginRequest *OriginRequest  `json:"originRequest,omitempty"`
}

type OriginRequest struct {
	NoTLSVerify    *bool `json:"noTLSVerify,omitempty"`
	ConnectTimeout *int  `json:"connectTimeout,omitempty"`
	HTTP2Origin    *bool `json:"http2Origin,omitempty"`
}

type CloudflareTunnelConfig struct {
	Ingress []TunnelIngressRule `json:"ingress"`
}

type CloudflareAccessApp struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	SessionDuration string `json:"session_duration,omitempty"`
}