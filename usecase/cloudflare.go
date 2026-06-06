package usecase

import (
	"context"

	"github.com/AndreeJait/server-management-be/domain/entity"
	"github.com/AndreeJait/server-management-be/port/inbound/cloudflare"
	"github.com/AndreeJait/server-management-be/port/outbound"
)

type cloudflareUseCase struct {
	cf outbound.Cloudflare
}

func NewCloudflareUseCase(cf outbound.Cloudflare) cloudflare.UseCase {
	return &cloudflareUseCase{cf: cf}
}

func (u *cloudflareUseCase) ListAccounts(ctx context.Context) ([]*entity.CloudflareAccount, error) {
	return u.cf.ListAccounts(ctx)
}

func (u *cloudflareUseCase) ListZones(ctx context.Context) ([]*entity.CloudflareZone, error) {
	return u.cf.ListZones(ctx)
}

func (u *cloudflareUseCase) ListDNSRecords(ctx context.Context, zoneID string) ([]*entity.CloudflareDNSRecord, error) {
	return u.cf.ListDNSRecords(ctx, zoneID)
}

func (u *cloudflareUseCase) CreateDNSRecord(ctx context.Context, zoneID, recordType, name, content string, ttl int, proxied bool) (*entity.CloudflareDNSRecord, error) {
	return u.cf.CreateDNSRecord(ctx, zoneID, recordType, name, content, ttl, proxied)
}

func (u *cloudflareUseCase) UpdateDNSRecord(ctx context.Context, zoneID, recordID, recordType, name, content string, ttl int, proxied bool) (*entity.CloudflareDNSRecord, error) {
	return u.cf.UpdateDNSRecord(ctx, zoneID, recordID, recordType, name, content, ttl, proxied)
}

func (u *cloudflareUseCase) DeleteDNSRecord(ctx context.Context, zoneID, recordID string) error {
	return u.cf.DeleteDNSRecord(ctx, zoneID, recordID)
}

func (u *cloudflareUseCase) ListTunnels(ctx context.Context) ([]*entity.CloudflareTunnel, error) {
	return u.cf.ListTunnels(ctx)
}

func (u *cloudflareUseCase) GetTunnelConfig(ctx context.Context, tunnelID string) (*entity.CloudflareTunnelConfig, error) {
	return u.cf.GetTunnelConfig(ctx, tunnelID)
}

func (u *cloudflareUseCase) ListAccessApps(ctx context.Context) ([]*entity.CloudflareAccessApp, error) {
	return u.cf.ListAccessApps(ctx)
}

func (u *cloudflareUseCase) CreateAccessApp(ctx context.Context, name, domain, sessionDuration string) (*entity.CloudflareAccessApp, error) {
	return u.cf.CreateAccessApp(ctx, name, domain, sessionDuration)
}

func (u *cloudflareUseCase) UpdateAccessApp(ctx context.Context, appID, name, domain, sessionDuration string) (*entity.CloudflareAccessApp, error) {
	return u.cf.UpdateAccessApp(ctx, appID, name, domain, sessionDuration)
}

func (u *cloudflareUseCase) DeleteAccessApp(ctx context.Context, appID string) error {
	return u.cf.DeleteAccessApp(ctx, appID)
}