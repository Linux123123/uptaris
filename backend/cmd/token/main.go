package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/config"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func main() {
	email := flag.String("email", "", "Existing user's email (required)")
	ttl := flag.Duration("ttl", 0, "Token lifetime, for example 720h for 30 days (required)")
	flag.Parse()

	if err := run(strings.ToLower(strings.TrimSpace(*email)), *ttl); err != nil {
		fmt.Fprintln(os.Stderr, "token:", err)
		os.Exit(1)
	}
}

func run(email string, ttl time.Duration) error {
	if email == "" || ttl < time.Second || flag.NArg() != 0 {
		return errors.New("usage: token -email <existing-user-email> -ttl <duration of at least 1s>")
	}

	cfg := config.Load()
	db, err := database.Open(cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	defer sqlDB.Close()

	var token string
	err = db.Transaction(func(tx *gorm.DB) error {
		var user models.User
		// Serialize issuance with user deletion and role changes, as login does.
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("email = ?", email).First(&user).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return fmt.Errorf("no user with email %q", email)
			}
			return err
		}

		refresh, err := auth.NewRefresh()
		if err != nil {
			return err
		}
		session := models.AuthSession{
			UserID:           user.ID,
			RefreshTokenHash: auth.HashRefresh(refresh, cfg.RefreshTokenPepper),
			ExpiresAt:        time.Now().Add(ttl),
		}
		if err := tx.Create(&session).Error; err != nil {
			return err
		}

		token, _, err = auth.Issue(uint64(user.ID), session.ID, user.Role, cfg.JWTAccessSecret, ttl)
		return err
	})
	if err != nil {
		return fmt.Errorf("issue token: %w", err)
	}

	_, err = fmt.Fprintln(os.Stdout, token)
	return err
}
