package entity

import "time"

type Setting struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Section   string    `gorm:"size:50;not null" json:"section"`
	Key       string    `gorm:"size:100;not null" json:"key"`
	Value     string    `gorm:"type:text;not null" json:"value"`
	Type      string    `gorm:"size:20;not null;default:'string'" json:"type"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

func (s *Setting) ToResponse() SettingResponse {
	return SettingResponse{
		Section: s.Section,
		Key:     s.Key,
		Value:   s.Value,
		Type:    s.Type,
	}
}

type SettingResponse struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
	Type    string `json:"type"`
}

type SettingsGroup struct {
	Section  string            `json:"section"`
	Settings []SettingResponse `json:"settings"`
}

type UpdateSettingInput struct {
	Section string `json:"section"`
	Key     string `json:"key"`
	Value   string `json:"value"`
}

type SettingApplied struct {
	Section     string `json:"section"`
	Key         string `json:"key"`
	Value       string `json:"value"`
	HotReloaded bool   `json:"hot_reloaded"`
}

type UpdateSettingsResult struct {
	Applied         []SettingApplied `json:"applied"`
	RestartRequired []string         `json:"restart_required"`
}