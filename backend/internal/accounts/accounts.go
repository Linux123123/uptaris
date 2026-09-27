package accounts

import (
	"context"
	"time"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/passwordpolicy"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Service struct {
	db *gorm.DB
}

type ProviderConnection struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Available bool   `json:"available"`
	Connected bool   `json:"connected"`
}

type Settings struct {
	Email                string               `json:"email"`
	HasPassword          bool                 `json:"hasPassword"`
	Providers            []ProviderConnection `json:"providers"`
	TwoFactorEnabled     bool                 `json:"twoFactorEnabled"`
	BackupCodesRemaining int64                `json:"backupCodesRemaining"`
}

type Error struct {
	Code    string
	Message string
}

func (e *Error) Error() string { return e.Message }

var ErrCurrentPassword = &Error{Code: "invalid_current_password", Message: "current password is incorrect"}

func New(db *gorm.DB) *Service { return &Service{db: db} }

func (s *Service) Settings(ctx context.Context, userID uint, available []auth.OAuthProviderInfo) (Settings, error) {
	var user models.User

	if err := s.db.WithContext(ctx).First(&user, userID).Error; err != nil {
		return Settings{}, err
	}

	var identities []models.OAuthIdentity

	if err := s.db.WithContext(ctx).Select("provider").
		Where("user_id = ?", userID).
		Order("provider").
		Find(&identities).Error; err != nil {
		return Settings{}, err
	}

	connected := make(map[string]bool, len(identities))
	for _, identity := range identities {
		connected[identity.Provider] = true
	}

	providers := make([]ProviderConnection, 0, len(available)+len(identities))
	known := make(map[string]bool, len(available))
	for _, provider := range available {
		known[provider.ID] = true
		providers = append(providers, ProviderConnection{
			ID:        provider.ID,
			Name:      provider.Name,
			Available: true,
			Connected: connected[provider.ID],
		})
	}

	// Keep connections to disabled providers visible so users can remove them.
	for _, identity := range identities {
		if !known[identity.Provider] {
			providers = append(providers, ProviderConnection{ID: identity.Provider, Name: identity.Provider, Connected: true})
		}
	}

	var enabled int64

	if err := s.db.WithContext(ctx).
		Model(&models.TwoFactor{}).
		Where("user_id = ? AND enabled_at IS NOT NULL", userID).
		Count(&enabled).Error; err != nil {
		return Settings{}, err
	}

	var remaining int64

	if err := s.db.WithContext(ctx).
		Model(&models.TwoFactorBackupCode{}).
		Where("user_id = ?", userID).
		Count(&remaining).Error; err != nil {
		return Settings{}, err
	}

	return Settings{
		Email:                user.Email,
		HasPassword:          user.PasswordHash != nil && *user.PasswordHash != "",
		Providers:            providers,
		TwoFactorEnabled:     enabled > 0,
		BackupCodesRemaining: remaining,
	}, nil
}

func (s *Service) SetPassword(ctx context.Context, userID, sessionID uint, current, next string) error {
	if err := passwordpolicy.Validate(next); err != nil {
		return err
	}

	hash, err := auth.HashPassword(next)
	if err != nil {
		return &Error{Code: "password_processing_failed", Message: "password processing failed"}
	}

	return database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		var user models.User

		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error; err != nil {
			return err
		}

		if user.PasswordHash != nil && *user.PasswordHash != "" && (current == "" || auth.CheckPassword(*user.PasswordHash, current) != nil) {
			return ErrCurrentPassword
		}

		if err := tx.Model(&user).Update("password_hash", hash).Error; err != nil {
			return err
		}

		if err := tx.Unscoped().Where("user_id = ?", user.ID).Delete(&models.TwoFactorChallenge{}).Error; err != nil {
			return err
		}

		return tx.
			Model(&models.AuthSession{}).
			Where("user_id = ? AND id <> ? AND revoked_at IS NULL", user.ID, sessionID).
			Update("revoked_at", time.Now()).Error
	})
}
