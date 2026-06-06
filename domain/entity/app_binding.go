package entity

import "time"

type AppBinding struct {
	ID           uint      `gorm:"primaryKey" json:"id"`
	AppID        string    `gorm:"size:36;index;not null" json:"app_id"`
	ZoneID       string    `gorm:"size:255;not null" json:"zone_id"`
	DNSRecordID  string    `gorm:"size:255;not null" json:"dns_record_id"`
	Domain       string    `gorm:"size:512;not null" json:"domain"`
	TunnelID     string    `gorm:"size:255;not null" json:"tunnel_id"`
	CreatedAt    time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type AppBindingResponse struct {
	ID          uint   `json:"id"`
	AppID       string `json:"app_id"`
	ZoneID      string `json:"zone_id"`
	DNSRecordID string `json:"dns_record_id"`
	Domain      string `json:"domain"`
	TunnelID    string `json:"tunnel_id"`
	CreatedAt   string `json:"created_at"`
}

func NewAppBinding(appID, zoneID, dnsRecordID, domain, tunnelID string) *AppBinding {
	return &AppBinding{
		AppID:       appID,
		ZoneID:      zoneID,
		DNSRecordID: dnsRecordID,
		Domain:      domain,
		TunnelID:    tunnelID,
	}
}

func (b *AppBinding) ToResponse() *AppBindingResponse {
	return &AppBindingResponse{
		ID:          b.ID,
		AppID:       b.AppID,
		ZoneID:      b.ZoneID,
		DNSRecordID: b.DNSRecordID,
		Domain:      b.Domain,
		TunnelID:    b.TunnelID,
		CreatedAt:   b.CreatedAt.UTC().Format(time.RFC3339),
	}
}