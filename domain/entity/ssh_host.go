package entity

import "time"

type AuthMethod string

const (
	AuthMethodPassword   AuthMethod = "password"
	AuthMethodPrivateKey AuthMethod = "private_key"
)

var validAuthMethods = map[AuthMethod]bool{
	AuthMethodPassword:   true,
	AuthMethodPrivateKey: true,
}

func (a AuthMethod) IsValid() bool {
	return validAuthMethods[a]
}

type SSHHost struct {
	ID         uint       `gorm:"primaryKey" json:"id"`
	Name       string     `gorm:"size:255;not null" json:"name"`
	Host       string     `gorm:"size:255;not null" json:"host"`
	Port       int        `gorm:"not null;default:22" json:"port"`
	Username   string     `gorm:"size:255;not null" json:"username"`
	AuthMethod AuthMethod `gorm:"size:20;not null;default:'password'" json:"auth_method"`
	Password   string     `gorm:"type:text;not null;default:''" json:"-"`
	PrivateKey string     `gorm:"type:text;not null;default:''" json:"-"`
	OwnerID    uint       `gorm:"index;not null" json:"owner_id"`
	CreatedAt  time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt  time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func NewSSHHost(name, host string, port int, username string, authMethod AuthMethod, password, privateKey string, ownerID uint) *SSHHost {
	return &SSHHost{
		Name:       name,
		Host:       host,
		Port:       port,
		Username:   username,
		AuthMethod: authMethod,
		Password:   password,
		PrivateKey: privateKey,
		OwnerID:    ownerID,
	}
}

type SSHHostResponse struct {
	ID            uint   `json:"id"`
	Name          string `json:"name"`
	Host          string `json:"host"`
	Port          int    `json:"port"`
	Username      string `json:"username"`
	AuthMethod    string `json:"auth_method"`
	HasPassword   bool   `json:"has_password"`
	HasPrivateKey bool   `json:"has_private_key"`
	OwnerID       uint   `json:"owner_id"`
	CreatedAt     string `json:"created_at"`
}

func (h *SSHHost) ToResponse() *SSHHostResponse {
	return &SSHHostResponse{
		ID:            h.ID,
		Name:          h.Name,
		Host:          h.Host,
		Port:          h.Port,
		Username:      h.Username,
		AuthMethod:    string(h.AuthMethod),
		HasPassword:   h.Password != "",
		HasPrivateKey: h.PrivateKey != "",
		OwnerID:       h.OwnerID,
		CreatedAt:     h.CreatedAt.UTC().Format(time.RFC3339),
	}
}