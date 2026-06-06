package entity

import "time"

type AppFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AppID     string    `gorm:"index;not null" json:"app_id"`
	Path      string    `gorm:"size:1024;not null" json:"path"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type AppFileResponse struct {
	ID        uint   `json:"id"`
	AppID     string `json:"app_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewAppFile(appID, path, content string) *AppFile {
	return &AppFile{
		AppID:   appID,
		Path:    path,
		Content: content,
	}
}

func (f *AppFile) ToResponse() *AppFileResponse {
	return &AppFileResponse{
		ID:        f.ID,
		AppID:     f.AppID,
		Path:      f.Path,
		Content:   f.Content,
		CreatedAt: f.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: f.UpdatedAt.UTC().Format(time.RFC3339),
	}
}