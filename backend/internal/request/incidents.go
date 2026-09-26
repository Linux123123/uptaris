package request

import (
	"strings"
	"time"

	"github.com/uptaris/uptaris/backend/internal/models"
)

type IncidentInput struct {
	Title       string     `json:"title" binding:"required,max=200" example:"API maintenance"`
	Description string     `json:"description" binding:"max=10000" example:"Primary customer API"`
	Severity    string     `json:"severity" binding:"required,oneof=low medium high critical" example:"low" enums:"low,medium,high,critical"`
	Status      string     `json:"status" binding:"omitempty,oneof=open acknowledged resolved" example:"open" enums:"open,acknowledged,resolved"`
	StartedAt   *time.Time `json:"startedAt" example:"2026-09-26T10:00:00Z"`
	ResolvedAt  *time.Time `json:"resolvedAt" example:"2026-09-26T11:00:00Z"`
}

type IncidentPatchInput struct {
	Title       *string    `json:"title" binding:"omitempty,max=200" example:"API maintenance"`
	Description *string    `json:"description" binding:"omitempty,max=10000" example:"Primary customer API"`
	Severity    *string    `json:"severity" binding:"omitempty,oneof=low medium high critical" example:"low" enums:"low,medium,high,critical"`
	Status      *string    `json:"status" binding:"omitempty,oneof=open acknowledged resolved" example:"open" enums:"open,acknowledged,resolved"`
	StartedAt   *time.Time `json:"startedAt" example:"2026-09-26T10:00:00Z"`
	ResolvedAt  *time.Time `json:"resolvedAt" example:"2026-09-26T11:00:00Z"`
}

func (input IncidentPatchInput) Apply(value *models.Incident) {
	if input.Title != nil {
		value.Title = *input.Title
	}
	if input.Description != nil {
		value.Description = *input.Description
	}
	if input.Severity != nil {
		value.Severity = *input.Severity
	}
	if input.Status != nil {
		value.Status = *input.Status
	}
	if input.StartedAt != nil {
		value.StartedAt = *input.StartedAt
	}
	if input.ResolvedAt != nil {
		value.ResolvedAt = input.ResolvedAt
	}
}

func PrepareIncident(value *models.Incident) error {
	value.Title = strings.TrimSpace(value.Title)
	if value.Status != "resolved" {
		value.ResolvedAt = nil
	} else if value.ResolvedAt == nil {
		now := time.Now()
		value.ResolvedAt = &now
	}
	return Validate(value)
}
