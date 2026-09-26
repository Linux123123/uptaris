package inventory

import (
	"slices"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/uptaris/uptaris/backend/internal/models"
)

func textValid(value string, max int) bool {
	return strings.TrimSpace(value) != "" &&
		utf8.RuneCountInString(value) <= max
}

func ValidateServer(server *models.Server) bool {
	server.Name = strings.TrimSpace(server.Name)
	server.Address = strings.TrimSpace(server.Address)
	server.OperatingSystem = strings.TrimSpace(server.OperatingSystem)
	return textValid(server.Name, 200) &&
		textValid(server.Address, 2048) &&
		textValid(server.OperatingSystem, 200) &&
		len(server.Description) <= 10000 &&
		valid(server.Status, "up", "down", "paused")
}

func ValidateMonitor(monitor *models.Monitor) bool {
	monitor.Name = strings.TrimSpace(monitor.Name)
	monitor.Target = strings.TrimSpace(monitor.Target)
	monitor.ExpectedHealth = strings.TrimSpace(monitor.ExpectedHealth)
	return textValid(monitor.Name, 200) &&
		textValid(monitor.Target, 2048) &&
		textValid(monitor.ExpectedHealth, 200) &&
		valid(monitor.Type, "http", "tcp", "icmp") &&
		valid(monitor.Status, "up", "down", "paused") &&
		monitor.IntervalSeconds >= 1 &&
		monitor.IntervalSeconds <= 86400
}

func ValidateIncident(incident *models.Incident) bool {
	incident.Title = strings.TrimSpace(incident.Title)
	if incident.Status != "resolved" {
		incident.ResolvedAt = nil
	} else if incident.ResolvedAt == nil {
		now := time.Now()
		incident.ResolvedAt = &now
	}
	return textValid(incident.Title, 200) &&
		len(incident.Description) <= 10000 &&
		valid(incident.Severity, "low", "medium", "high", "critical") &&
		valid(incident.Status, "open", "acknowledged", "resolved") &&
		!incident.StartedAt.IsZero() &&
		(incident.ResolvedAt == nil || !incident.ResolvedAt.Before(incident.StartedAt))
}

func valid(v string, allowed ...string) bool {
	return slices.Contains(allowed, v)
}
