package request

import (
	"strings"

	"github.com/uptaris/uptaris/backend/internal/models"
)

type MonitorInput struct {
	Name            string `json:"name" binding:"required,max=200" example:"Public API"`
	Type            string `json:"type" binding:"required,oneof=http tcp icmp" example:"http" enums:"http,tcp,icmp"`
	Target          string `json:"target" binding:"required,max=2048" example:"https://api.example.com/health"`
	IntervalSeconds int    `json:"intervalSeconds" binding:"required,min=1,max=86400" example:"60"`
	ExpectedHealth  string `json:"expectedHealth" binding:"required,max=200" example:"200"`
	Status          string `json:"status" binding:"omitempty,oneof=up down paused" example:"up" enums:"up,down,paused"`
}

type MonitorPatchInput struct {
	Name            *string `json:"name" binding:"omitempty,max=200" example:"Public API"`
	Type            *string `json:"type" binding:"omitempty,oneof=http tcp icmp" example:"http" enums:"http,tcp,icmp"`
	Target          *string `json:"target" binding:"omitempty,max=2048" example:"https://api.example.com/health"`
	IntervalSeconds *int    `json:"intervalSeconds" binding:"omitempty,min=1,max=86400" example:"60"`
	ExpectedHealth  *string `json:"expectedHealth" binding:"omitempty,max=200" example:"200"`
	Status          *string `json:"status" binding:"omitempty,oneof=up down paused" example:"up" enums:"up,down,paused"`
}

func (input MonitorPatchInput) Apply(value *models.Monitor) {
	if input.Name != nil {
		value.Name = *input.Name
	}

	if input.Type != nil {
		value.Type = *input.Type
	}

	if input.Target != nil {
		value.Target = *input.Target
	}

	if input.IntervalSeconds != nil {
		value.IntervalSeconds = *input.IntervalSeconds
	}

	if input.ExpectedHealth != nil {
		value.ExpectedHealth = *input.ExpectedHealth
	}

	if input.Status != nil {
		value.Status = *input.Status
	}
}

func PrepareMonitor(value *models.Monitor) error {
	value.Name = strings.TrimSpace(value.Name)
	value.Target = strings.TrimSpace(value.Target)
	value.ExpectedHealth = strings.TrimSpace(value.ExpectedHealth)

	return Validate(value)
}
