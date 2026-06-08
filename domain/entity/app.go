package entity

import "time"

type FrameworkPreset string

const (
	FrameworkNextjs  FrameworkPreset = "nextjs"
	FrameworkNuxt    FrameworkPreset = "nuxt"
	FrameworkVue     FrameworkPreset = "vue"
	FrameworkReact   FrameworkPreset = "react"
	FrameworkSvelte   FrameworkPreset = "svelte"
	FrameworkAstro   FrameworkPreset = "astro"
	FrameworkRemix   FrameworkPreset = "remix"
	FrameworkGo      FrameworkPreset = "go"
	FrameworkPython  FrameworkPreset = "python"
	FrameworkJava    FrameworkPreset = "java"
	FrameworkNodejs  FrameworkPreset = "nodejs"
	FrameworkRust    FrameworkPreset = "rust"
	FrameworkDotnet  FrameworkPreset = "dotnet"
	FrameworkLaravel FrameworkPreset = "laravel"
	FrameworkRails   FrameworkPreset = "rails"
	FrameworkCustom  FrameworkPreset = "custom"
)

var ValidFrameworkPresets = map[FrameworkPreset]bool{
	FrameworkNextjs:  true,
	FrameworkNuxt:    true,
	FrameworkVue:     true,
	FrameworkReact:   true,
	FrameworkSvelte:  true,
	FrameworkAstro:   true,
	FrameworkRemix:   true,
	FrameworkGo:      true,
	FrameworkPython:  true,
	FrameworkJava:    true,
	FrameworkNodejs:  true,
	FrameworkRust:    true,
	FrameworkDotnet:  true,
	FrameworkLaravel: true,
	FrameworkRails:   true,
	FrameworkCustom:  true,
}

func (f FrameworkPreset) IsValid() bool {
	return ValidFrameworkPresets[f]
}

type App struct {
	ID                  uint            `gorm:"primaryKey" json:"id"`
	ProjectID           uint            `gorm:"index;not null" json:"project_id"`
	Name                string          `gorm:"size:255;not null" json:"name"`
	FrameworkPreset     FrameworkPreset `gorm:"size:50;not null;default:'custom'" json:"framework_preset"`
	ContainerCount      int             `gorm:"not null;default:0" json:"container_count"`
	ContainerPort     string          `gorm:"size:10;not null;default:''" json:"container_port,omitempty"`
	PublishPort       string          `gorm:"size:10;not null;default:''" json:"publish_port,omitempty"`
	ContainerName     string          `gorm:"size:255;not null;default:''" json:"container_name,omitempty"`
	DefaultImage      string          `gorm:"size:512" json:"default_image,omitempty"`
	DeployToken       string          `gorm:"size:255;not null" json:"-"`
	AppID             string          `gorm:"size:36;uniqueIndex;not null" json:"app_id"`
	EnvVars             StringMap       `gorm:"type:jsonb;not null;default:'{}'" json:"env_vars,omitempty"`
	VolumeMounts        VolumeMountList `gorm:"type:jsonb;not null;default:'[]'" json:"volume_mounts,omitempty"`
	PostDeployCommands  StringList      `gorm:"type:jsonb;not null;default:'[]'" json:"post_deploy_commands,omitempty"`
	BasePath            string          `gorm:"size:512;not null;default:''" json:"base_path,omitempty"`
	FilesMountPath      string          `gorm:"size:512;not null;default:'/app/files'" json:"files_mount_path,omitempty"`
	CreatedAt           time.Time       `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time       `gorm:"autoUpdateTime" json:"updated_at"`
}

func NewApp(projectID uint, name string, preset FrameworkPreset, deployToken, appID string) *App {
	return &App{
		ProjectID:          projectID,
		Name:               name,
		FrameworkPreset:    preset,
		DeployToken:        deployToken,
		AppID:              appID,
		EnvVars:            StringMap{},
		VolumeMounts:       VolumeMountList{},
		PostDeployCommands: StringList{},
	}
}

type AppResponse struct {
	ID                  uint            `json:"id"`
	ProjectID           uint            `json:"project_id"`
	Name                string          `json:"name"`
	FrameworkPreset     string          `json:"framework_preset"`
	ContainerPort       string          `json:"container_port,omitempty"`
	PublishPort         string          `json:"publish_port,omitempty"`
	ContainerName       string          `json:"container_name,omitempty"`
	DefaultImage        string          `json:"default_image,omitempty"`
	ContainerCount      int             `json:"container_count"`
	AppID               string          `json:"app_id"`
	EnvVars             StringMap       `json:"env_vars,omitempty"`
	VolumeMounts        VolumeMountList `json:"volume_mounts,omitempty"`
	PostDeployCommands  StringList      `json:"post_deploy_commands,omitempty"`
	BasePath            string          `json:"base_path,omitempty"`
	FilesMountPath      string          `json:"files_mount_path,omitempty"`
	DockerHostBase      string          `json:"docker_host_base"`
	CreatedAt           string          `json:"created_at"`
}

func (a *App) ToResponse() *AppResponse {
	return &AppResponse{
		ID:                 a.ID,
		ProjectID:          a.ProjectID,
		Name:               a.Name,
		FrameworkPreset:    string(a.FrameworkPreset),
		ContainerPort:      a.ContainerPort,
		PublishPort:        a.PublishPort,
		ContainerName:      a.ContainerName,
		DefaultImage:       a.DefaultImage,
		ContainerCount:     a.ContainerCount,
		AppID:              a.AppID,
		EnvVars:            a.EnvVars,
		VolumeMounts:       a.VolumeMounts,
		PostDeployCommands: a.PostDeployCommands,
		BasePath:           a.BasePath,
		FilesMountPath:     a.FilesMountPath,
		CreatedAt:          a.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (a *App) ToCreateResponse() *CreateAppResponse {
	return &CreateAppResponse{
		App:         a.ToResponse(),
		DeployToken: a.DeployToken,
	}
}

type CreateAppResponse struct {
	App         *AppResponse `json:"app"`
	DeployToken string       `json:"deploy_token"`
}

type RegenerateTokenResponse struct {
	App         *AppResponse `json:"app"`
	DeployToken string       `json:"deploy_token"`
}