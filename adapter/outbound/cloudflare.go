package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/outbound"
)

type cloudflareClient struct {
	apiToken  string
	accountID string
	baseURL   string
	client    *http.Client
}

func NewCloudflareClient(apiToken, accountID string) outbound.Cloudflare {
	return &cloudflareClient{
		apiToken:  apiToken,
		accountID: accountID,
		baseURL:   "https://api.cloudflare.com/client/v4",
		client:    &http.Client{},
	}
}

func (c *cloudflareClient) doRequest(ctx context.Context, method, path string, body any) ([]byte, error) {
	var reqBody io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("marshal request body: %w", err)
		}
		reqBody = bytes.NewReader(data)
	}

	req, err := http.NewRequestWithContext(ctx, method, c.baseURL+path, reqBody)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiToken)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("execute request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("read response: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("cloudflare API error: %s: %s", resp.Status, string(respBody))
	}

	return respBody, nil
}

type cfResponse struct {
	Success bool            `json:"success"`
	Errors  []cfError       `json:"errors"`
	Result  json.RawMessage `json:"result"`
}

type cfError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

type cfPaginatedResponse struct {
	Success bool            `json:"success"`
	Errors  []cfError       `json:"errors"`
	Result  json.RawMessage `json:"result"`
	ResultInfo *struct {
		Page       int `json:"page"`
		PerPage    int `json:"per_page"`
		TotalPages int `json:"total_pages"`
		Count      int `json:"count"`
		TotalCount int `json:"total_count"`
	} `json:"result_info"`
}

func unmarshalResult(data []byte, target any) error {
	var resp cfResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return fmt.Errorf("unmarshal cloudflare response: %w", err)
	}
	if !resp.Success {
		if len(resp.Errors) > 0 {
			return fmt.Errorf("cloudflare API error: %s", resp.Errors[0].Message)
		}
		return fmt.Errorf("cloudflare API error: unknown")
	}
	if err := json.Unmarshal(resp.Result, target); err != nil {
		return fmt.Errorf("unmarshal cloudflare result: %w", err)
	}
	return nil
}

func unmarshalResultRaw(data []byte) (json.RawMessage, error) {
	var resp cfResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("unmarshal cloudflare response: %w", err)
	}
	if !resp.Success {
		if len(resp.Errors) > 0 {
			return nil, fmt.Errorf("cloudflare API error: %s", resp.Errors[0].Message)
		}
		return nil, fmt.Errorf("cloudflare API error: unknown")
	}
	return resp.Result, nil
}

// --- Account operations ---

type cfAccount struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

func (c *cloudflareClient) ListAccounts(ctx context.Context) ([]*entity.CloudflareAccount, error) {
	data, err := c.doRequest(ctx, http.MethodGet, "/accounts", nil)
	if err != nil {
		return nil, err
	}
	var accounts []cfAccount
	if err := unmarshalResult(data, &accounts); err != nil {
		return nil, err
	}
	result := make([]*entity.CloudflareAccount, len(accounts))
	for i, a := range accounts {
		result[i] = &entity.CloudflareAccount{
			ID:   a.ID,
			Name: a.Name,
			Type: a.Type,
		}
	}
	return result, nil
}

// --- Zone operations ---

type cfZone struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

func (c *cloudflareClient) ListZones(ctx context.Context) ([]*entity.CloudflareZone, error) {
	data, err := c.doRequest(ctx, http.MethodGet, "/zones", nil)
	if err != nil {
		return nil, err
	}
	var zones []cfZone
	if err := unmarshalResult(data, &zones); err != nil {
		return nil, err
	}
	result := make([]*entity.CloudflareZone, len(zones))
	for i, z := range zones {
		result[i] = &entity.CloudflareZone{
			ID:     z.ID,
			Name:   z.Name,
			Status: z.Status,
		}
	}
	return result, nil
}

// --- DNS Record operations ---

type cfDNSRecord struct {
	ID      string `json:"id"`
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	Proxied bool   `json:"proxied"`
	TTL     int    `json:"ttl"`
}

func (c *cloudflareClient) ListDNSRecords(ctx context.Context, zoneID string) ([]*entity.CloudflareDNSRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records", zoneID)
	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var records []cfDNSRecord
	if err := unmarshalResult(data, &records); err != nil {
		return nil, err
	}
	result := make([]*entity.CloudflareDNSRecord, len(records))
	for i, r := range records {
		result[i] = &entity.CloudflareDNSRecord{
			ID:      r.ID,
			Type:    r.Type,
			Name:    r.Name,
			Content: r.Content,
			Proxied: r.Proxied,
			TTL:     r.TTL,
		}
	}
	return result, nil
}

type cfCreateDNSRecordRequest struct {
	Type    string `json:"type"`
	Name    string `json:"name"`
	Content string `json:"content"`
	TTL     int    `json:"ttl"`
	Proxied bool   `json:"proxied"`
}

func (c *cloudflareClient) CreateDNSRecord(ctx context.Context, zoneID, recordType, name, content string, ttl int, proxied bool) (*entity.CloudflareDNSRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records", zoneID)
	body := cfCreateDNSRecordRequest{
		Type:    recordType,
		Name:    name,
		Content: content,
		TTL:     ttl,
		Proxied: proxied,
	}
	data, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	var record cfDNSRecord
	if err := unmarshalResult(data, &record); err != nil {
		return nil, err
	}
	return &entity.CloudflareDNSRecord{
		ID:      record.ID,
		Type:    record.Type,
		Name:    record.Name,
		Content: record.Content,
		Proxied: record.Proxied,
		TTL:     record.TTL,
	}, nil
}

func (c *cloudflareClient) UpdateDNSRecord(ctx context.Context, zoneID, recordID, recordType, name, content string, ttl int, proxied bool) (*entity.CloudflareDNSRecord, error) {
	path := fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, recordID)
	body := cfCreateDNSRecordRequest{
		Type:    recordType,
		Name:    name,
		Content: content,
		TTL:     ttl,
		Proxied: proxied,
	}
	data, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}
	var record cfDNSRecord
	if err := unmarshalResult(data, &record); err != nil {
		return nil, err
	}
	return &entity.CloudflareDNSRecord{
		ID:      record.ID,
		Type:    record.Type,
		Name:    record.Name,
		Content: record.Content,
		Proxied: record.Proxied,
		TTL:     record.TTL,
	}, nil
}

func (c *cloudflareClient) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	path := fmt.Sprintf("/zones/%s/dns_records/%s", zoneID, recordID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}

// --- Tunnel operations ---

type cfTunnel struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Status     string `json:"status"`
	ConnsCount int    `json:"conns_count"`
}

func (c *cloudflareClient) ListTunnels(ctx context.Context) ([]*entity.CloudflareTunnel, error) {
	path := fmt.Sprintf("/accounts/%s/cfd_tunnel", c.accountID)
	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var tunnels []cfTunnel
	if err := unmarshalResult(data, &tunnels); err != nil {
		return nil, err
	}
	result := make([]*entity.CloudflareTunnel, len(tunnels))
	for i, t := range tunnels {
		result[i] = &entity.CloudflareTunnel{
			ID:         t.ID,
			Name:       t.Name,
			Status:     t.Status,
			ConnsCount: t.ConnsCount,
		}
	}
	return result, nil
}

type cfTunnelConfig struct {
	Ingress []cfIngressRule `json:"ingress"`
}

type cfIngressRule struct {
	Hostname      string          `json:"hostname,omitempty"`
	Service       string          `json:"service"`
	Path          string          `json:"path,omitempty"`
	OriginRequest *cfOriginRequest `json:"originRequest,omitempty"`
}

type cfOriginRequest struct {
	NoTLSVerify    *bool `json:"noTLSVerify,omitempty"`
	ConnectTimeout *int  `json:"connectTimeout,omitempty"`
	HTTP2Origin    *bool `json:"http2Origin,omitempty"`
}

func (c *cloudflareClient) GetTunnelConfig(ctx context.Context, tunnelID string) (*entity.CloudflareTunnelConfig, error) {
	path := fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations", c.accountID, tunnelID)
	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}

	resultRaw, err := unmarshalResultRaw(data)
	if err != nil {
		return nil, err
	}

	// The API wraps config in a "config" object
	var wrapper struct {
		Config cfTunnelConfig `json:"config"`
	}
	if err := json.Unmarshal(resultRaw, &wrapper); err != nil {
		return nil, fmt.Errorf("unmarshal tunnel config: %w", err)
	}

	ingress := make([]entity.TunnelIngressRule, len(wrapper.Config.Ingress))
	for i, r := range wrapper.Config.Ingress {
		rule := entity.TunnelIngressRule{
			Hostname: r.Hostname,
			Service:  r.Service,
			Path:     r.Path,
		}
		if r.OriginRequest != nil {
			rule.OriginRequest = &entity.OriginRequest{
				NoTLSVerify:    r.OriginRequest.NoTLSVerify,
				ConnectTimeout: r.OriginRequest.ConnectTimeout,
				HTTP2Origin:    r.OriginRequest.HTTP2Origin,
			}
		}
		ingress[i] = rule
	}
	return &entity.CloudflareTunnelConfig{Ingress: ingress}, nil
}

func (c *cloudflareClient) UpdateTunnelConfig(ctx context.Context, tunnelID string, config *entity.CloudflareTunnelConfig) error {
	path := fmt.Sprintf("/accounts/%s/cfd_tunnel/%s/configurations", c.accountID, tunnelID)
	ingress := make([]cfIngressRule, len(config.Ingress))
	for i, r := range config.Ingress {
		rule := cfIngressRule{
			Hostname: r.Hostname,
			Service:  r.Service,
			Path:     r.Path,
		}
		if r.OriginRequest != nil {
			rule.OriginRequest = &cfOriginRequest{
				NoTLSVerify:    r.OriginRequest.NoTLSVerify,
				ConnectTimeout: r.OriginRequest.ConnectTimeout,
				HTTP2Origin:    r.OriginRequest.HTTP2Origin,
			}
		}
		ingress[i] = rule
	}
	body := struct {
		Config cfTunnelConfig `json:"config"`
	}{
		Config: cfTunnelConfig{Ingress: ingress},
	}
	_, err := c.doRequest(ctx, http.MethodPut, path, body)
	return err
}

// --- Access Application operations ---

type cfAccessApp struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	SessionDuration string `json:"session_duration"`
}

func (c *cloudflareClient) ListAccessApps(ctx context.Context) ([]*entity.CloudflareAccessApp, error) {
	path := fmt.Sprintf("/accounts/%s/access/apps", c.accountID)
	data, err := c.doRequest(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	var apps []cfAccessApp
	if err := unmarshalResult(data, &apps); err != nil {
		return nil, err
	}
	result := make([]*entity.CloudflareAccessApp, len(apps))
	for i, a := range apps {
		result[i] = &entity.CloudflareAccessApp{
			ID:              a.ID,
			Name:            a.Name,
			Domain:          a.Domain,
			SessionDuration: a.SessionDuration,
		}
	}
	return result, nil
}

type cfCreateAccessAppRequest struct {
	Name            string `json:"name"`
	Domain          string `json:"domain"`
	SessionDuration string `json:"session_duration,omitempty"`
}

func (c *cloudflareClient) CreateAccessApp(ctx context.Context, name, domain, sessionDuration string) (*entity.CloudflareAccessApp, error) {
	path := fmt.Sprintf("/accounts/%s/access/apps", c.accountID)
	body := cfCreateAccessAppRequest{
		Name:            name,
		Domain:          domain,
		SessionDuration: sessionDuration,
	}
	data, err := c.doRequest(ctx, http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	var app cfAccessApp
	if err := unmarshalResult(data, &app); err != nil {
		return nil, err
	}
	return &entity.CloudflareAccessApp{
		ID:              app.ID,
		Name:            app.Name,
		Domain:          app.Domain,
		SessionDuration: app.SessionDuration,
	}, nil
}

func (c *cloudflareClient) UpdateAccessApp(ctx context.Context, appID, name, domain, sessionDuration string) (*entity.CloudflareAccessApp, error) {
	path := fmt.Sprintf("/accounts/%s/access/apps/%s", c.accountID, appID)
	body := cfCreateAccessAppRequest{
		Name:            name,
		Domain:          domain,
		SessionDuration: sessionDuration,
	}
	data, err := c.doRequest(ctx, http.MethodPut, path, body)
	if err != nil {
		return nil, err
	}
	var app cfAccessApp
	if err := unmarshalResult(data, &app); err != nil {
		return nil, err
	}
	return &entity.CloudflareAccessApp{
		ID:              app.ID,
		Name:            app.Name,
		Domain:          app.Domain,
		SessionDuration: app.SessionDuration,
	}, nil
}

func (c *cloudflareClient) DeleteAccessApp(ctx context.Context, appID string) error {
	path := fmt.Sprintf("/accounts/%s/access/apps/%s", c.accountID, appID)
	_, err := c.doRequest(ctx, http.MethodDelete, path, nil)
	return err
}