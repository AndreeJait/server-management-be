package entity

import "time"

type StepStatus string

const (
	StepPending StepStatus = "pending"
	StepRunning StepStatus = "running"
	StepDone    StepStatus = "done"
	StepFailed  StepStatus = "failed"
	StepSkipped StepStatus = "skipped"
)

type PipelineStep struct {
	ID           uint       `gorm:"primaryKey" json:"id"`
	DeploymentID uint       `gorm:"index;not null" json:"deployment_id"`
	Name         string     `gorm:"size:100;not null" json:"name"`
	StepOrder    int        `gorm:"not null" json:"step_order"`
	Status       StepStatus `gorm:"size:20;not null;default:'pending'" json:"status"`
	Config       string     `gorm:"type:jsonb" json:"config,omitempty"`
	Output       string     `gorm:"type:text" json:"output,omitempty"`
	Error        string     `gorm:"type:text" json:"error,omitempty"`
	StartedAt    *time.Time `json:"started_at,omitempty"`
	FinishedAt   *time.Time `json:"finished_at,omitempty"`
	CreatedAt    time.Time  `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt    time.Time  `gorm:"autoUpdateTime" json:"updated_at"`
}

func NewPipelineStep(deploymentID uint, name string, order int, config string) *PipelineStep {
	return &PipelineStep{
		DeploymentID: deploymentID,
		Name:         name,
		StepOrder:    order,
		Status:       StepPending,
		Config:       config,
	}
}

type PipelineStepResponse struct {
	ID           uint   `json:"id"`
	DeploymentID uint   `json:"deployment_id"`
	Name         string `json:"name"`
	StepOrder    int    `json:"step_order"`
	Status       string `json:"status"`
	Config       string `json:"config,omitempty"`
	Output       string `json:"output,omitempty"`
	Error        string `json:"error,omitempty"`
	StartedAt    string `json:"started_at,omitempty"`
	FinishedAt   string `json:"finished_at,omitempty"`
}

func (s *PipelineStep) ToResponse() *PipelineStepResponse {
	resp := &PipelineStepResponse{
		ID:           s.ID,
		DeploymentID: s.DeploymentID,
		Name:         s.Name,
		StepOrder:    s.StepOrder,
		Status:       string(s.Status),
		Config:       s.Config,
		Output:       s.Output,
		Error:        s.Error,
	}
	if s.StartedAt != nil {
		resp.StartedAt = s.StartedAt.UTC().Format(time.RFC3339)
	}
	if s.FinishedAt != nil {
		resp.FinishedAt = s.FinishedAt.UTC().Format(time.RFC3339)
	}
	return resp
}