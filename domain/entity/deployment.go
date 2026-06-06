package entity

import "time"

type DeploymentStatus string

const (
	DeploymentPending    DeploymentStatus = "pending"
	DeploymentPulling    DeploymentStatus = "pulling"
	DeploymentCreating   DeploymentStatus = "creating"
	DeploymentStarting   DeploymentStatus = "starting"
	DeploymentExecuting  DeploymentStatus = "executing"
	DeploymentRunning    DeploymentStatus = "running"
	DeploymentFailed     DeploymentStatus = "failed"
	DeploymentStopped    DeploymentStatus = "stopped"
)

type Deployment struct {
	ID            uint             `gorm:"primaryKey" json:"id"`
	AppID         string           `gorm:"size:36;index;not null" json:"app_id"`
	Image         string           `gorm:"size:512;not null" json:"image"`
	Status        DeploymentStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	ContainerID   string           `gorm:"size:64" json:"container_id,omitempty"`
	ContainerName string           `gorm:"size:255" json:"container_name,omitempty"`
	Error         string           `gorm:"type:text" json:"error,omitempty"`
	CreatedAt     time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt     time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

func NewDeployment(appID, image string) *Deployment {
	return &Deployment{
		AppID:  appID,
		Image:  image,
		Status: DeploymentPending,
	}
}

type DeploymentResponse struct {
	ID            uint                   `json:"id"`
	AppID         string                 `json:"app_id"`
	Image         string                 `json:"image"`
	Status        string                 `json:"status"`
	ContainerID   string                 `json:"container_id,omitempty"`
	ContainerName string                 `json:"container_name,omitempty"`
	Error         string                 `json:"error,omitempty"`
	CreatedAt     string                 `json:"created_at"`
	Steps         []*PipelineStepResponse `json:"steps,omitempty"`
}

func (d *Deployment) ToResponse(steps ...[]*PipelineStepResponse) *DeploymentResponse {
	resp := &DeploymentResponse{
		ID:            d.ID,
		AppID:         d.AppID,
		Image:         d.Image,
		Status:        string(d.Status),
		ContainerID:   d.ContainerID,
		ContainerName: d.ContainerName,
		Error:         d.Error,
		CreatedAt:     d.CreatedAt.UTC().Format(time.RFC3339),
	}
	if len(steps) > 0 {
		resp.Steps = steps[0]
	}
	return resp
}