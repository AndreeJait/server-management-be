package cloudflare

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
)

type UseCase interface {
	ListAccounts(ctx context.Context) ([]*entity.CloudflareAccount, error)
	ListZones(ctx context.Context) ([]*entity.CloudflareZone, error)
	ListDNSRecords(ctx context.Context, zoneID string) ([]*entity.CloudflareDNSRecord, error)
	CreateDNSRecord(ctx context.Context, zoneID, recordType, name, content string, ttl int, proxied bool) (*entity.CloudflareDNSRecord, error)
	UpdateDNSRecord(ctx context.Context, zoneID, recordID, recordType, name, content string, ttl int, proxied bool) (*entity.CloudflareDNSRecord, error)
	DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error
	ListTunnels(ctx context.Context) ([]*entity.CloudflareTunnel, error)
	GetTunnelConfig(ctx context.Context, tunnelID string) (*entity.CloudflareTunnelConfig, error)
	ListAccessApps(ctx context.Context) ([]*entity.CloudflareAccessApp, error)
	CreateAccessApp(ctx context.Context, name, domain, sessionDuration string) (*entity.CloudflareAccessApp, error)
	UpdateAccessApp(ctx context.Context, appID, name, domain, sessionDuration string) (*entity.CloudflareAccessApp, error)
	DeleteAccessApp(ctx context.Context, appID string) error
}