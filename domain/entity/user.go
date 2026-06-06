package entity

import "time"

// User represents a registered user in the system.
type User struct {
	ID        uint      `gorm:"primaryKey" json:"id"`
	Email     string    `gorm:"uniqueIndex;size:255;not null" json:"email"`
	Password  string    `gorm:"size:255;not null" json:"-"`
	Name      string    `gorm:"size:255;not null" json:"name"`
	CreatedAt time.Time `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt time.Time `gorm:"autoUpdateTime" json:"updated_at"`
}

// UserRole represents the many-to-many relationship between users and roles.
type UserRole struct {
	ID     uint   `gorm:"primaryKey" json:"id"`
	UserID uint   `gorm:"index;not null" json:"user_id"`
	Role   string `gorm:"size:50;not null" json:"role"`
}

// NewUser creates a User with the default "viewer" role.
func NewUser(email, hashedPassword, name string) *User {
	return &User{
		Email:    email,
		Password: hashedPassword,
		Name:     name,
	}
}

// UserResponse is the API-safe representation of a User (excludes Password).
type UserResponse struct {
	ID        uint     `json:"id"`
	Email     string   `json:"email"`
	Name      string   `json:"name"`
	Roles     []string `json:"roles"`
	CreatedAt string   `json:"created_at"`
}

// ToResponse converts a User (with loaded roles) to a UserResponse.
func (u *User) ToResponse(roles []string) *UserResponse {
	return &UserResponse{
		ID:        u.ID,
		Email:     u.Email,
		Name:      u.Name,
		Roles:     roles,
		CreatedAt: u.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// RolePermission represents a role-to-permission mapping.
type RolePermission struct {
	ID         uint   `gorm:"primaryKey" json:"id"`
	Role       string `gorm:"size:50;not null;uniqueIndex:idx_role_permission" json:"role"`
	Permission string `gorm:"size:100;not null;uniqueIndex:idx_role_permission" json:"permission"`
}

// RoleResponse is the API representation of a role and its permissions.
type RoleResponse struct {
	Role        string   `json:"role"`
	Permissions []string `json:"permissions"`
}