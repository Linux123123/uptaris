package main

import (
	"fmt"
	"time"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/config"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
)

func main() {
	cfg := config.Load()
	if cfg.Environment == "production" {
		panic("demo seeding is disabled in production")
	}

	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		panic(err)
	}

	err = database.Transaction(db, func(db *gorm.DB) error {
		var userCount int64

		if err := db.Model(&models.User{}).Count(&userCount).Error; err != nil {
			return err
		}

		if userCount > 0 {
			fmt.Println("database already seeded")

			return nil
		}

		password, err := auth.HashPassword("UptarisDemo!2026")
		if err != nil {
			return err
		}

		users := []models.User{
			{Email: "viewer@uptaris.local", PasswordHash: &password, Role: "viewer"},
			{Email: "operator@uptaris.local", PasswordHash: &password, Role: "operator"},
			{Email: "admin@uptaris.local", PasswordHash: &password, Role: "admin"},
		}

		for i := range users {
			if err := db.Create(&users[i]).Error; err != nil {
				return err
			}
		}

		names := []string{"Vilnius API", "Kaunas Web", "Klaipeda DB", "Riga Gateway", "Helsinki Cache"}
		for i, name := range names {
			server := models.Server{
				OwnerID:         users[1].ID,
				Name:            name,
				Address:         fmt.Sprintf("10.20.0.%d", i+10),
				OperatingSystem: "Ubuntu 24.04",
				Description:     "Demo infrastructure service",
				Status:          "up",
			}

			if err := db.Create(&server).Error; err != nil {
				return err
			}

			monitor := models.Monitor{
				ServerID:        server.ID,
				Name:            name + " HTTP",
				Type:            "http",
				Target:          "https://example.com/health",
				IntervalSeconds: 60,
				ExpectedHealth:  "200",
				Status:          "up",
			}

			if err := db.Create(&monitor).Error; err != nil {
				return err
			}

			incident := models.Incident{
				MonitorID:   monitor.ID,
				Title:       name + " maintenance",
				Description: "Scheduled demo maintenance",
				Severity:    "low",
				Status:      "resolved",
				StartedAt:   time.Now().Add(-48 * time.Hour),
				ResolvedAt:  ptr(time.Now().Add(-47 * time.Hour)),
			}

			if err := db.Create(&incident).Error; err != nil {
				return err
			}
		}

		return nil
	})
	if err != nil {
		panic(err)
	}

	fmt.Println("seeded users: viewer@uptaris.local, operator@uptaris.local, admin@uptaris.local; password: UptarisDemo!2026")
}

func ptr(t time.Time) *time.Time { return &t }
