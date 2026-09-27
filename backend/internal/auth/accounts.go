package auth

import (
	"context"
	"errors"
	"strings"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"github.com/uptaris/uptaris/backend/internal/passwordpolicy"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Service) Register(ctx context.Context, email, password string) (*models.User, error) {
	if err := passwordpolicy.Validate(password); err != nil {
		return nil, err
	}

	hash, err := HashPassword(password)
	if err != nil {
		return nil, failure("internal_error", "password processing failed")
	}

	user := models.User{Email: strings.ToLower(email), PasswordHash: &hash, Role: "viewer"}
	err = s.db.WithContext(ctx).Create(&user).Error

	return &user, emailRegistrationError(err)
}

func (s *Service) Login(ctx context.Context, email, password string) (*Session, error) {
	if err := passwordpolicy.Validate(password); err != nil {
		return nil, err
	}

	var session *Session

	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		session = nil

		var user models.User

		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("email = ?", strings.ToLower(email)).First(&user).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err != nil || user.PasswordHash == nil || CheckPassword(*user.PasswordHash, password) != nil {
			return failure("invalid_credentials", "email or password invalid")
		}

		// Password verification and issuance share the lock used by password changes.
		session, err = s.issueLocked(tx, user)

		return err
	})

	return session, err
}

// Caller holds the user lock; a failed removal rolls back with the transaction.
func (s *Service) hasSignInMethod(tx *gorm.DB, user models.User) (bool, error) {
	if user.PasswordHash != nil && *user.PasswordHash != "" {
		return true, nil
	}

	var remaining int64

	if err := tx.Model(&models.Passkey{}).
		Where("user_id = ? AND rp_id = ?", user.ID, s.cfg.EffectiveWebAuthnRPID()).
		Count(&remaining).Error; err != nil {
		return false, err
	}

	if remaining > 0 {
		return true, nil
	}

	available := make([]string, 0, len(s.providers))
	for id := range s.providers {
		available = append(available, id)
	}

	if len(available) == 0 {
		return false, nil
	}

	// Stored identities cannot sign in while their providers are disabled.
	err := tx.Model(&models.OAuthIdentity{}).
		Where("user_id = ? AND provider IN ?", user.ID, available).
		Count(&remaining).Error

	return remaining > 0, err
}

func (s *Service) User(ctx context.Context, id uint) (*models.User, error) {
	var user models.User

	err := s.db.WithContext(ctx).First(&user, id).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, failure("invalid_token", "user unavailable")
	}

	return &user, err
}

func emailRegistrationError(err error) error {
	var conflict *pgconn.PgError

	if errors.As(err, &conflict) && conflict.Code == "23505" {
		return failure("email_exists", "email already registered")
	}

	return err
}
