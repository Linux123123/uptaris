package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/pquerna/otp/totp"
	"github.com/uptaris/uptaris/backend/internal/database"
	"github.com/uptaris/uptaris/backend/internal/models"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const challengeTTL = 5 * time.Minute
const enrollmentTTL = 10 * time.Minute
const maxChallengeAttempts = 5

// Enrollment is returned only to the authenticated account owner. Never cache it.
type Enrollment struct {
	Secret    string    `json:"secret"`
	URI       string    `json:"uri"`
	ExpiresAt time.Time `json:"expiresAt"`
}

func (s *Service) factorCipher() (cipher.AEAD, error) {
	if len(s.cfg.TwoFactorKey) != 32 {
		return nil, failure("two_factor_unavailable", "two-factor encryption is not configured")
	}

	block, err := aes.NewCipher(s.cfg.TwoFactorKey)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

// Associated data prevents moving an encrypted secret to another account.
func factorAAD(userID uint) []byte { return []byte(fmt.Sprintf("uptaris-totp\x00%d", userID)) }

func (s *Service) encryptFactor(userID uint, secret string) ([]byte, error) {
	gcm, err := s.factorCipher()
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, []byte(secret), factorAAD(userID)), nil
}

func (s *Service) decryptFactor(factor models.TwoFactor) (string, error) {
	gcm, err := s.factorCipher()
	if err != nil {
		return "", err
	}

	if len(factor.SecretCiphertext) < gcm.NonceSize()+gcm.Overhead() {
		return "", failure("internal_error", "authenticator secret is invalid")
	}

	plain, err := gcm.Open(nil, factor.SecretCiphertext[:gcm.NonceSize()], factor.SecretCiphertext[gcm.NonceSize():], factorAAD(factor.UserID))
	if err != nil {
		return "", failure("internal_error", "authenticator secret could not be decrypted")
	}

	return string(plain), nil
}

func (s *Service) factorHash(kind, value string) string {
	mac := hmac.New(sha256.New, []byte(s.cfg.RefreshTokenPepper))
	mac.Write([]byte("uptaris-two-factor\x00" + kind + "\x00" + value))

	return hex.EncodeToString(mac.Sum(nil))
}

func normalizeBackup(code string) string {
	return strings.ToLower(strings.ReplaceAll(strings.TrimSpace(code), "-", ""))
}

func (s *Service) newChallenge(tx *gorm.DB, userID uint) (string, error) {
	token, err := NewRefresh()
	if err != nil {
		return "", err
	}

	if err := tx.Unscoped().Where("expires_at <= ?", time.Now()).Delete(&models.TwoFactorChallenge{}).Error; err != nil {
		return "", err
	}

	// Only the most recent first-factor sign-in may proceed.
	if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.TwoFactorChallenge{}).Error; err != nil {
		return "", err
	}

	challenge := models.TwoFactorChallenge{
		UserID:    userID,
		TokenHash: s.factorHash("challenge", token),
		ExpiresAt: time.Now().Add(challengeTTL),
	}

	return token, tx.Create(&challenge).Error
}

func lockUser(tx *gorm.DB, userID uint) (models.User, error) {
	var user models.User

	err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).First(&user, userID).Error

	return user, err
}

func checkCurrentPassword(user models.User, password string) error {
	if user.PasswordHash != nil && *user.PasswordHash != "" && CheckPassword(*user.PasswordHash, password) != nil {
		return failure("invalid_current_password", "current password is incorrect")
	}

	return nil
}

func (s *Service) SetupTwoFactor(ctx context.Context, userID uint, current string) (Enrollment, error) {
	var result Enrollment

	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		user, err := lockUser(tx, userID)
		if err != nil {
			return err
		}

		if err := checkCurrentPassword(user, current); err != nil {
			return err
		}

		var factor models.TwoFactor

		err = tx.Where("user_id = ?", userID).First(&factor).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if factor.EnabledAt != nil {
			return failure("two_factor_enabled", "two-factor authentication is already enabled")
		}

		key, err := totp.Generate(totp.GenerateOpts{Issuer: "Uptaris", AccountName: user.Email, SecretSize: 20})
		if err != nil {
			return err
		}

		encrypted, err := s.encryptFactor(userID, key.Secret())
		if err != nil {
			return err
		}

		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.TwoFactor{}).Error; err != nil {
			return err
		}

		expires := time.Now().Add(enrollmentTTL)
		pending := models.TwoFactor{UserID: userID, SecretCiphertext: encrypted, ExpiresAt: expires, LastUsedStep: -1}
		if err := tx.Create(&pending).Error; err != nil {
			return err
		}

		result = Enrollment{Secret: key.Secret(), URI: key.URL(), ExpiresAt: expires}

		return nil
	})

	return result, err
}

// matchingStep tracks time steps, rejecting replay even within the accepted clock skew.
func matchingStep(secret, code string, now time.Time, last int64) (int64, bool) {
	if len(code) != 6 {
		return 0, false
	}

	for _, c := range code {
		if c < '0' || c > '9' {
			return 0, false
		}
	}

	step := now.Unix() / 30
	for _, candidate := range []int64{step, step - 1, step + 1} {
		if candidate <= last {
			continue
		}

		expected, err := totp.GenerateCode(secret, time.Unix(candidate*30, 0))
		if err == nil && subtle.ConstantTimeCompare([]byte(code), []byte(expected)) == 1 {
			return candidate, true
		}
	}

	return 0, false
}

func (s *Service) verifyFactor(tx *gorm.DB, factor *models.TwoFactor, code string, backup bool) (bool, error) {
	now := time.Now()
	if factor.LockedUntil != nil && factor.LockedUntil.After(now) {
		return false, failure("two_factor_locked", "too many incorrect codes; retry in five minutes")
	}

	attempts := factor.FailedAttempts
	if factor.LockedUntil != nil {
		// The lockout has elapsed; start a fresh attempt window.
		attempts = 0
	}

	var ok bool

	if backup {
		// Deleting the matching hash consumes the backup code in this transaction.
		result := tx.Unscoped().
			Where("user_id = ? AND code_hash = ?", factor.UserID, s.factorHash("backup", normalizeBackup(code))).
			Delete(&models.TwoFactorBackupCode{})
		if result.Error != nil {
			return false, result.Error
		}

		ok = result.RowsAffected == 1
	} else {
		secret, err := s.decryptFactor(*factor)
		if err != nil {
			return false, err
		}

		step, matched := matchingStep(secret, code, now, factor.LastUsedStep)
		ok = matched
		if ok {
			if err := tx.Model(factor).Update("last_used_step", step).Error; err != nil {
				return false, err
			}
		}
	}

	updates := map[string]any{"failed_attempts": 0, "locked_until": nil}
	if !ok {
		attempts++
		updates["failed_attempts"] = attempts
		if attempts >= maxChallengeAttempts {
			updates["locked_until"] = now.Add(challengeTTL)
		}
	}

	return ok, tx.Model(factor).Updates(updates).Error
}

func (s *Service) replaceBackupCodes(tx *gorm.DB, userID uint) ([]string, error) {
	if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.TwoFactorBackupCode{}).Error; err != nil {
		return nil, err
	}

	codes := make([]string, 10)
	for i := range codes {
		bytes := make([]byte, 10)
		if _, err := rand.Read(bytes); err != nil {
			return nil, err
		}

		raw := hex.EncodeToString(bytes)
		codes[i] = raw[:5] + "-" + raw[5:10] + "-" + raw[10:15] + "-" + raw[15:]
		row := models.TwoFactorBackupCode{UserID: userID, CodeHash: s.factorHash("backup", raw)}
		if err := tx.Create(&row).Error; err != nil {
			return nil, err
		}
	}

	return codes, nil
}

func (s *Service) ConfirmTwoFactor(ctx context.Context, userID, sessionID uint, code string) ([]string, error) {
	var codes []string
	var invalid bool

	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		invalid = false
		if _, err := lockUser(tx, userID); err != nil {
			return err
		}

		var factor models.TwoFactor

		if err := tx.
			Where("user_id = ? AND enabled_at IS NULL AND expires_at > ?", userID, time.Now()).
			First(&factor).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return failure("invalid_enrollment", "enrollment expired or already enabled; start again")
			}

			return err
		}

		ok, err := s.verifyFactor(tx, &factor, code, false)
		if err != nil {
			return err
		}

		if !ok {
			invalid = true

			return nil
		}

		if err := tx.Model(&factor).Update("enabled_at", time.Now()).Error; err != nil {
			return err
		}

		codes, err = s.replaceBackupCodes(tx, userID)
		if err != nil {
			return err
		}

		return tx.
			Model(&models.AuthSession{}).
			Where("user_id = ? AND id <> ? AND revoked_at IS NULL", userID, sessionID).
			Update("revoked_at", time.Now()).Error
	})
	if err == nil && invalid {
		return nil, failure("invalid_two_factor_code", "code is incorrect or already used")
	}

	return codes, err
}

// ManageTwoFactor requires both current password (when present) and a second factor.
func (s *Service) ManageTwoFactor(ctx context.Context, userID, sessionID uint, current, code string, backup, disable bool) ([]string, error) {
	var codes []string
	var invalid bool

	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		invalid = false
		user, err := lockUser(tx, userID)
		if err != nil {
			return err
		}

		if err := checkCurrentPassword(user, current); err != nil {
			return err
		}

		var factor models.TwoFactor

		if err := tx.Where("user_id = ? AND enabled_at IS NOT NULL", userID).First(&factor).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return failure("two_factor_disabled", "two-factor authentication is not enabled")
			}

			return err
		}

		ok, err := s.verifyFactor(tx, &factor, code, backup)
		if err != nil {
			return err
		}

		if !ok {
			invalid = true

			return nil
		}

		if disable {
			if err := tx.Unscoped().Delete(&factor).Error; err != nil {
				return err
			}

			if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.TwoFactorBackupCode{}).Error; err != nil {
				return err
			}
		} else {
			codes, err = s.replaceBackupCodes(tx, userID)
			if err != nil {
				return err
			}
		}

		if err := tx.Unscoped().Where("user_id = ?", userID).Delete(&models.TwoFactorChallenge{}).Error; err != nil {
			return err
		}

		return tx.
			Model(&models.AuthSession{}).
			Where("user_id = ? AND id <> ? AND revoked_at IS NULL", userID, sessionID).
			Update("revoked_at", time.Now()).Error
	})
	if err == nil && invalid {
		return nil, failure("invalid_two_factor_code", "code is incorrect or already used")
	}

	return codes, err
}

func (s *Service) CompleteTwoFactor(ctx context.Context, token, code string, backup bool) (*Session, error) {
	hash := s.factorHash("challenge", token)

	var session *Session
	var invalid bool

	err := database.Transaction(s.db.WithContext(ctx), func(tx *gorm.DB) error {
		session = nil
		invalid = false

		var challenge models.TwoFactorChallenge

		if err := tx.
			Where("token_hash = ? AND expires_at > ? AND attempts < ?", hash, time.Now(), maxChallengeAttempts).
			First(&challenge).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return failure("invalid_two_factor_challenge", "sign-in expired; sign in again")
			}

			return err
		}

		user, err := lockUser(tx, challenge.UserID)
		if err != nil {
			return err
		}

		// Re-read after locking the user: concurrent success, disable, or new login may have consumed it.
		if err := tx.
			Where("token_hash = ? AND expires_at > ? AND attempts < ?", hash, time.Now(), maxChallengeAttempts).
			First(&challenge).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return failure("invalid_two_factor_challenge", "sign-in expired; sign in again")
			}

			return err
		}

		var factor models.TwoFactor

		if err := tx.Where("user_id = ? AND enabled_at IS NOT NULL", user.ID).First(&factor).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return failure("invalid_two_factor_challenge", "sign-in expired; sign in again")
			}

			return err
		}

		ok, err := s.verifyFactor(tx, &factor, code, backup)
		if err != nil {
			return err
		}

		if !ok {
			invalid = true
			// Commit failed attempts rather than rolling them back with the verification error.

			return tx.Model(&challenge).Update("attempts", challenge.Attempts+1).Error
		}

		deleted := tx.Unscoped().Delete(&challenge)
		if deleted.Error != nil {
			return deleted.Error
		}

		if deleted.RowsAffected != 1 {
			return failure("invalid_two_factor_challenge", "sign-in expired; sign in again")
		}

		session, err = s.createSession(tx, user)

		return err
	})
	if err != nil {
		return nil, err
	}

	if invalid {
		return nil, failure("invalid_two_factor_code", "code is incorrect or already used")
	}

	return session, nil
}

func (s *Service) CancelTwoFactor(ctx context.Context, token string) error {
	return s.db.WithContext(ctx).
		Unscoped().
		Where("token_hash = ?", s.factorHash("challenge", token)).
		Delete(&models.TwoFactorChallenge{}).Error
}
