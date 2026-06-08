package entity

import "time"

type DomainRequestCount struct {
	ID            uint      `gorm:"primaryKey" json:"id"`
	Domain        string    `gorm:"size:512;uniqueIndex;not null" json:"domain"`
	Count         int64     `gorm:"not null;default:0" json:"count"`
	LastRequestAt time.Time `json:"last_request_at"`
	CreatedAt     time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (d *DomainRequestCount) ToResponse() DomainRequestCountResponse {
	return DomainRequestCountResponse{
		Domain:        d.Domain,
		Count:         d.Count,
		LastRequestAt: d.LastRequestAt.Format(time.RFC3339),
	}
}

type DomainRequestCountResponse struct {
	Domain        string `json:"domain"`
	Count         int64  `json:"count"`
	LastRequestAt string `json:"last_request_at"`
}