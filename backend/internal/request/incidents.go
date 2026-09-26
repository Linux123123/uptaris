package request

import (
	"time"

	"github.com/uptaris/uptaris/backend/internal/inventory"
)

type IncidentInput struct {
	Title       string     `json:"title" binding:"required" example:"API maintenance"`
	Description string     `json:"description" example:"Primary customer API"`
	Severity    string     `json:"severity" binding:"required" example:"low"`
	Status      string     `json:"status"`
	StartedAt   *time.Time `json:"startedAt" example:"2026-09-26T10:00:00Z"`
	ResolvedAt  *time.Time `json:"resolvedAt" example:"2026-09-26T11:00:00Z"`
}

type IncidentPatchInput = inventory.IncidentPatch
