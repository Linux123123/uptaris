package request

import (
	"github.com/uptaris/uptaris/backend/internal/inventory"
)

type MonitorInput struct {
	Name            string `json:"name" binding:"required" example:"Public API"`
	Type            string `json:"type" binding:"required" example:"http"`
	Target          string `json:"target" binding:"required" example:"https://api.example.com/health"`
	IntervalSeconds int    `json:"intervalSeconds" example:"60"`
	ExpectedHealth  string `json:"expectedHealth" binding:"required" example:"200"`
	Status          string `json:"status"`
}

type MonitorPatchInput = inventory.MonitorPatch
