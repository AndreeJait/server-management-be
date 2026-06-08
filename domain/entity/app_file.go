package entity

import "time"

type AppFile struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	AppID     string    `gorm:"index;not null" json:"app_id"`
	Path      string    `gorm:"size:1024;not null" json:"path"`
	Content   string    `gorm:"type:text;not null" json:"content"`
	FileType  string    `gorm:"size:16;not null;default:'text'" json:"file_type"`
	FileSize  int64     `gorm:"not null;default:0" json:"file_size"`
	MimeType  string    `gorm:"size:256;not null;default:''" json:"mime_type"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

type AppFileResponse struct {
	ID        uint   `json:"id"`
	AppID     string `json:"app_id"`
	Path      string `json:"path"`
	Content   string `json:"content"`
	FileType  string `json:"file_type"`
	FileSize  int64  `json:"file_size"`
	MimeType  string `json:"mime_type"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

func NewAppFile(appID, path, content string) *AppFile {
	return &AppFile{
		AppID:    appID,
		Path:     path,
		Content:  content,
		FileType: "text",
	}
}

func (f *AppFile) ToResponse() *AppFileResponse {
	return &AppFileResponse{
		ID:        f.ID,
		AppID:     f.AppID,
		Path:      f.Path,
		Content:   f.Content,
		FileType:  f.FileType,
		FileSize:  f.FileSize,
		MimeType:  f.MimeType,
		CreatedAt: f.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt: f.UpdatedAt.UTC().Format(time.RFC3339),
	}
}