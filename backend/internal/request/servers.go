package request

import (
	"github.com/uptaris/uptaris/backend/internal/inventory"
)

type ServerInput struct {
	Name            string `json:"name" binding:"required" example:"Public API"`
	Address         string `json:"address" binding:"required" example:"api.example.com"`
	OperatingSystem string `json:"operatingSystem" binding:"required" example:"Ubuntu 24.04"`
	Description     string `json:"description" example:"Primary customer API"`
	Status          string `json:"status"`
}

type ServerPatchInput = inventory.ServerPatch
