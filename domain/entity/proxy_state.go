package entity

import "time"

type SlotName string

const (
	SlotBlue  SlotName = "blue"
	SlotGreen SlotName = "green"
)

type ProxyStateStatus string

const (
	ProxyStateIdle        ProxyStateStatus = "idle"
	ProxyStateDeploying   ProxyStateStatus = "deploying"
	ProxyStateShifting    ProxyStateStatus = "shifting"
	ProxyStateActive      ProxyStateStatus = "active"
	ProxyStateRollingBack ProxyStateStatus = "rolling_back"
	ProxyStateFailed      ProxyStateStatus = "failed"
)

type ProxyState struct {
	ID                  uint             `gorm:"primaryKey" json:"id"`
	AppID               string           `gorm:"size:36;uniqueIndex;not null" json:"app_id"`
	ActiveSlot          SlotName         `gorm:"size:10;not null;default:'blue'" json:"active_slot"`
	BlueContainerID     string           `gorm:"size:64" json:"blue_container_id,omitempty"`
	GreenContainerID    string           `gorm:"size:64" json:"green_container_id,omitempty"`
	BlueTarget          string           `gorm:"size:255" json:"blue_target,omitempty"`
	GreenTarget         string           `gorm:"size:255" json:"green_target,omitempty"`
	TrafficPercent       int              `gorm:"not null;default:100" json:"traffic_percent"`
	HealthCheckPath     string           `gorm:"size:512;not null;default:'/health'" json:"health_check_path"`
	HealthCheckInterval int              `gorm:"not null;default:10" json:"health_check_interval"`
	Status              ProxyStateStatus `gorm:"size:20;not null;default:'idle'" json:"status"`
	CreatedAt           time.Time        `gorm:"autoCreateTime" json:"created_at"`
	UpdatedAt           time.Time        `gorm:"autoUpdateTime" json:"updated_at"`
}

func NewProxyState(appID string) *ProxyState {
	return &ProxyState{
		AppID:               appID,
		ActiveSlot:          SlotBlue,
		TrafficPercent:      100,
		HealthCheckPath:     "/health",
		HealthCheckInterval: 10,
		Status:              ProxyStateIdle,
	}
}

type ProxyStateResponse struct {
	ID                  uint   `json:"id"`
	AppID               string `json:"app_id"`
	ActiveSlot          string `json:"active_slot"`
	BlueContainerID     string `json:"blue_container_id,omitempty"`
	GreenContainerID    string `json:"green_container_id,omitempty"`
	BlueTarget          string `json:"blue_target,omitempty"`
	GreenTarget         string `json:"green_target,omitempty"`
	TrafficPercent       int    `json:"traffic_percent"`
	HealthCheckPath      string `json:"health_check_path"`
	HealthCheckInterval  int    `json:"health_check_interval"`
	Status              string `json:"status"`
	CreatedAt           string `json:"created_at"`
}

func (ps *ProxyState) ToResponse() *ProxyStateResponse {
	return &ProxyStateResponse{
		ID:                  ps.ID,
		AppID:               ps.AppID,
		ActiveSlot:          string(ps.ActiveSlot),
		BlueContainerID:     ps.BlueContainerID,
		GreenContainerID:    ps.GreenContainerID,
		BlueTarget:          ps.BlueTarget,
		GreenTarget:         ps.GreenTarget,
		TrafficPercent:      ps.TrafficPercent,
		HealthCheckPath:     ps.HealthCheckPath,
		HealthCheckInterval: ps.HealthCheckInterval,
		Status:              string(ps.Status),
		CreatedAt:           ps.CreatedAt.UTC().Format(time.RFC3339),
	}
}

func (ps *ProxyState) InactiveSlot() SlotName {
	if ps.ActiveSlot == SlotBlue {
		return SlotGreen
	}
	return SlotBlue
}

func (ps *ProxyState) SetSlotContainerID(slot SlotName, containerID string) {
	if slot == SlotBlue {
		ps.BlueContainerID = containerID
	} else {
		ps.GreenContainerID = containerID
	}
}

func (ps *ProxyState) GetSlotContainerID(slot SlotName) string {
	if slot == SlotBlue {
		return ps.BlueContainerID
	}
	return ps.GreenContainerID
}

func (ps *ProxyState) SetSlotTarget(slot SlotName, target string) {
	if slot == SlotBlue {
		ps.BlueTarget = target
	} else {
		ps.GreenTarget = target
	}
}

func (ps *ProxyState) GetSlotTarget(slot SlotName) string {
	if slot == SlotBlue {
		return ps.BlueTarget
	}
	return ps.GreenTarget
}