package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type OAuthIdentity struct {
	Subject       string
	Email         string
	EmailVerified bool
	Tokens        OAuthTokenSet
}

type OAuthTokenSet struct {
	AccessToken           string `json:"-"`
	RefreshToken          string `json:"-"`
	AccessTokenExpiresAt  *time.Time
	RefreshTokenExpiresAt *time.Time
}

type OAuthProvider interface {
	ID() string
	DisplayName() string
	AuthorizationURL(state string) string
	Exchange(ctx context.Context, code string) (OAuthIdentity, error)
}

type OAuthProviderInfo struct {
	ID       string `json:"id"`
	Name     string `json:"name"`
	LoginURL string `json:"loginUrl"`
}

func oauthStateDigest(value string) string {
	sum := sha256.Sum256([]byte(value))

	return hex.EncodeToString(sum[:])
}

func (s *Service) OAuthProviders() []OAuthProviderInfo {
	ids := make([]string, 0, len(s.providers))
	for id := range s.providers {
		ids = append(ids, id)
	}

	sort.Strings(ids)
	result := make([]OAuthProviderInfo, 0, len(ids))
	for _, id := range ids {
		provider := s.providers[id]
		result = append(result, OAuthProviderInfo{ID: id, Name: provider.DisplayName(), LoginURL: "/auth/oauth/" + id})
	}

	return result
}

func (s *Service) StartOAuth(ctx context.Context, providerID, intent string, userID uint) (string, string, error) {
	provider, ok := s.providers[providerID]
	if !ok {
		return "", "", failure("oauth_provider_unavailable", "OAuth provider is not configured")
	}

	if (intent != "login" && intent != "link") || (intent == "link" && userID == 0) {
		return "", "", failure("invalid_oauth_state", "OAuth request invalid")
	}

	var linked *uint

	if intent == "link" {
		linked = &userID
	}

	state, err := NewRefresh()
	if err != nil {
		return "", "", err
	}

	if err := s.db.WithContext(ctx).
		Unscoped().
		Where("expires_at <= ? OR consumed_at IS NOT NULL", time.Now()).
		Delete(&models.OAuthState{}).Error; err != nil {
		return "", "", err
	}

	row := models.OAuthState{
		StateHash: oauthStateDigest(state),
		Provider:  providerID,
		Intent:    intent,
		UserID:    linked,
		ExpiresAt: time.Now().Add(10 * time.Minute),
	}

	if err := s.db.WithContext(ctx).Create(&row).Error; err != nil {
		return "", "", err
	}

	return provider.AuthorizationURL(state), state, nil
}

func (s *Service) CompleteOAuth(ctx context.Context, providerID, state, code string) (*Session, error) {
	// Consume state before contacting the provider; failed callbacks cannot reuse it.
	pending, err := s.consumeOAuthState(ctx, providerID, state)
	if err != nil {
		return nil, err
	}

	provider, ok := s.providers[providerID]
	if !ok || code == "" {
		return nil, failure("oauth_auth_failed", "OAuth sign-in failed")
	}

	identity, err := provider.Exchange(ctx, code)
	if err != nil || identity.Subject == "" || identity.Tokens.AccessToken == "" {
		return nil, failure("oauth_auth_failed", "OAuth provider could not authenticate account")
	}

	accessCiphertext, refreshCiphertext, err := s.encryptOAuthTokens(providerID, identity.Subject, identity.Tokens)
	if err != nil {
		return nil, err
	}

	identity.Email = strings.ToLower(strings.TrimSpace(identity.Email))

	var session *Session

	err = database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		session = nil

		var user models.User
		var linked models.OAuthIdentity

		lookup := tx.Where("provider = ? AND provider_user_id = ?", providerID, identity.Subject).First(&linked)
		if lookup.Error == nil {
			if pending.Intent == "link" && (pending.UserID == nil || linked.UserID != *pending.UserID) {
				return failure("oauth_identity_in_use", "OAuth identity is linked to another account")
			}

			locked, err := lockUser(tx, linked.UserID)
			if err != nil {
				return err
			}

			user = locked
			// Unlink uses the same user lock. Recheck the identity after acquiring it.
			if err := tx.First(&linked, linked.ID).Error; err != nil {
				if errors.Is(err, gorm.ErrRecordNotFound) {
					return failure("oauth_auth_failed", "OAuth identity was disconnected; sign in again")
				}

				return err
			}

			updates := map[string]any{
				"access_token_ciphertext": accessCiphertext,
				"access_token_expires_at": identity.Tokens.AccessTokenExpiresAt,
			}

			// Providers may omit a refresh token on later grants; keep the stored one.
			if identity.Tokens.RefreshToken != "" {
				updates["refresh_token_ciphertext"] = refreshCiphertext
				updates["refresh_token_expires_at"] = identity.Tokens.RefreshTokenExpiresAt
			}

			if err := tx.Model(&linked).Updates(updates).Error; err != nil {
				return err
			}

			if pending.Intent == "link" {
				return nil
			}

			// Keep identity verification and issuance under the lock shared with unlink.
			session, err = s.issueLocked(tx, user)

			return err
		}

		if !errors.Is(lookup.Error, gorm.ErrRecordNotFound) {
			return lookup.Error
		}

		if pending.Intent == "link" {
			if pending.UserID == nil {
				return failure("invalid_oauth_state", "OAuth request invalid")
			}

			if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, *pending.UserID).Error; err != nil {
				return err
			}

			var existing models.OAuthIdentity

			if err := tx.Where("user_id = ? AND provider = ?", user.ID, providerID).First(&existing).Error; err == nil {
				return failure("oauth_provider_already_linked", "Account already has an identity from this provider")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}
		} else {
			// A verified email permits signup, never automatic linking to an existing account.
			if identity.Email == "" || !identity.EmailVerified {
				return failure("oauth_email_unavailable", "Provider did not provide a verified primary email")
			}

			if err := tx.Where("LOWER(email) = ?", identity.Email).First(&user).Error; err == nil {
				return failure("email_exists", "Sign in with existing account, then connect provider from account settings")
			} else if !errors.Is(err, gorm.ErrRecordNotFound) {
				return err
			}

			user = models.User{Email: identity.Email, Role: "viewer"}
			if err := tx.Create(&user).Error; err != nil {
				return emailRegistrationError(err)
			}
		}

		if err := tx.Create(&models.OAuthIdentity{
			UserID:                 user.ID,
			Provider:               providerID,
			ProviderUserID:         identity.Subject,
			AccessTokenCiphertext:  accessCiphertext,
			RefreshTokenCiphertext: refreshCiphertext,
			AccessTokenExpiresAt:   identity.Tokens.AccessTokenExpiresAt,
			RefreshTokenExpiresAt:  identity.Tokens.RefreshTokenExpiresAt,
		}).Error; err != nil {
			return err
		}

		if pending.Intent == "link" {
			return nil
		}

		// Newly created users stay private to this transaction until issuance commits.
		session, err = s.issueLocked(tx, user)

		return err
	})
	if err != nil {
		// Concurrent callbacks may collide after the identity lookup.
		var conflict *pgconn.PgError
		if errors.As(err, &conflict) && conflict.Code == "23505" {
			switch conflict.ConstraintName {
			case "oauth_identities_provider_identity_unique":
				return nil, failure("oauth_identity_in_use", "OAuth identity is already linked; sign in again")
			case "oauth_identities_user_provider_unique":
				return nil, failure("oauth_provider_already_linked", "Account already has an identity from this provider")
			}
		}

		return nil, err
	}

	return session, nil
}

func (s *Service) CancelOAuth(ctx context.Context, providerID, state string) error {
	_, err := s.consumeOAuthState(ctx, providerID, state)

	return err
}

func (s *Service) consumeOAuthState(ctx context.Context, providerID, state string) (models.OAuthState, error) {
	var pending models.OAuthState

	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("state_hash = ? AND provider = ? AND consumed_at IS NULL AND expires_at > ?", oauthStateDigest(state), providerID, time.Now()).
			First(&pending).Error
		if err != nil {
			return err
		}

		return tx.Model(&pending).Update("consumed_at", time.Now()).Error
	})
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.OAuthState{}, failure("invalid_oauth_state", "OAuth request expired or already used")
	}

	if err != nil {
		return models.OAuthState{}, err
	}

	return pending, nil
}

func (s *Service) UnlinkOAuth(ctx context.Context, providerID string, userID uint) error {
	return database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		user, err := lockUser(tx, userID)
		if err != nil {
			return err
		}

		result := tx.Unscoped().Where("user_id = ? AND provider = ?", userID, providerID).Delete(&models.OAuthIdentity{})
		if result.Error != nil {
			return result.Error
		}

		if result.RowsAffected == 0 {
			return gorm.ErrRecordNotFound
		}

		canSignIn, err := s.hasSignInMethod(tx, user)
		if err != nil {
			return err
		}

		if !canSignIn {
			return failure("last_signin_method", "Add another sign-in method before disconnecting this provider")
		}

		return nil
	})
}
