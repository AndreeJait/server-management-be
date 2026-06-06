package entity

import "time"

type CredentialScope string

const (
	CredentialScopeGlobal  CredentialScope = "global"
	CredentialScopeProject CredentialScope = "project"
)

type RegistryCredential struct {
	ID          uint            `gorm:"primaryKey" json:"id"`
	ProjectID   *uint           `gorm:"index" json:"project_id"`
	Scope       CredentialScope `gorm:"size:10;not null;default:'project'" json:"scope"`
	RegistryURL string          `gorm:"size:255;not null" json:"registry_url"`
	Username    string          `gorm:"size:255;not null" json:"username"`
	Password    string          `gorm:"size:255;not null" json:"-"` // TODO: encrypt at rest (Phase 3)
	CreatedAt   time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt   time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func NewGlobalRegistryCredential(registryURL, username, password string) *RegistryCredential {
	return &RegistryCredential{
		Scope:       CredentialScopeGlobal,
		RegistryURL: registryURL,
		Username:    username,
		Password:    password,
	}
}

func NewProjectRegistryCredential(projectID uint, registryURL, username, password string) *RegistryCredential {
	return &RegistryCredential{
		ProjectID:   &projectID,
		Scope:       CredentialScopeProject,
		RegistryURL: registryURL,
		Username:    username,
		Password:    password,
	}
}

type RegistryCredentialResponse struct {
	ID          uint   `json:"id"`
	ProjectID   *uint  `json:"project_id,omitempty"`
	Scope       string `json:"scope"`
	RegistryURL string `json:"registry_url"`
	Username    string `json:"username"`
	CreatedAt   string `json:"created_at"`
}

func (r *RegistryCredential) ToResponse() *RegistryCredentialResponse {
	resp := &RegistryCredentialResponse{
		ID:          r.ID,
		ProjectID:   r.ProjectID,
		Scope:       string(r.Scope),
		RegistryURL: r.RegistryURL,
		Username:    r.Username,
		CreatedAt:   r.CreatedAt.UTC().Format(time.RFC3339),
	}
	if r.Scope == CredentialScopeGlobal {
		resp.ProjectID = nil
	}
	return resp
}