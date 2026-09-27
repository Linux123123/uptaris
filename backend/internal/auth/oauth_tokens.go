package auth

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"errors"
	"io"

	"github.com/uptaris/uptaris/backend/internal/models"
)

func (s *Service) encryptOAuthTokens(providerID, subject string, tokens OAuthTokenSet) ([]byte, []byte, error) {
	access, err := s.encryptOAuthToken(providerID, subject, "access", tokens.AccessToken)
	if err != nil {
		return nil, nil, err
	}

	refresh, err := s.encryptOAuthToken(providerID, subject, "refresh", tokens.RefreshToken)
	if err != nil {
		return nil, nil, err
	}

	return access, refresh, nil
}

func (s *Service) encryptOAuthToken(providerID, subject, kind, token string) ([]byte, error) {
	if token == "" {
		return nil, nil
	}

	gcm, err := s.oauthTokenCipher()
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, errors.New("encrypt OAuth token")
	}

	ciphertext := gcm.Seal(nil, nonce, []byte(token), oauthTokenAAD(providerID, subject, kind))

	return append(nonce, ciphertext...), nil
}

func (s *Service) decryptOAuthToken(providerID, subject, kind string, ciphertext []byte) (string, error) {
	if len(ciphertext) == 0 {
		return "", nil
	}

	gcm, err := s.oauthTokenCipher()
	if err != nil {
		return "", err
	}

	if len(ciphertext) < gcm.NonceSize()+gcm.Overhead() {
		return "", errors.New("stored OAuth token is invalid")
	}

	nonce, encrypted := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, encrypted, oauthTokenAAD(providerID, subject, kind))
	if err != nil {
		return "", errors.New("stored OAuth token could not be decrypted")
	}

	return string(plain), nil
}

func (s *Service) oauthTokenCipher() (cipher.AEAD, error) {
	if len(s.cfg.OAuthTokenKey) != 32 {
		return nil, errors.New("OAuth token encryption requires a 32-byte key")
	}

	block, err := aes.NewCipher(s.cfg.OAuthTokenKey)
	if err != nil {
		return nil, err
	}

	return cipher.NewGCM(block)
}

func oauthTokenAAD(providerID, subject, kind string) []byte {
	// Bind ciphertext to this provider identity and to its access/refresh purpose.
	return []byte("uptaris-oauth-token\x00" + providerID + "\x00" + subject + "\x00" + kind)
}

// OAuthTokens retrieves credentials for trusted server-side provider integrations.
// Expiries are preserved so future integrations can decide when to refresh.
func (s *Service) OAuthTokens(ctx context.Context, userID uint, providerID string) (OAuthTokenSet, error) {
	var linked models.OAuthIdentity

	if err := s.db.WithContext(ctx).Where("user_id = ? AND provider = ?", userID, providerID).First(&linked).Error; err != nil {
		return OAuthTokenSet{}, err
	}

	access, err := s.decryptOAuthToken(providerID, linked.ProviderUserID, "access", linked.AccessTokenCiphertext)
	if err != nil {
		return OAuthTokenSet{}, err
	}

	refresh, err := s.decryptOAuthToken(providerID, linked.ProviderUserID, "refresh", linked.RefreshTokenCiphertext)
	if err != nil {
		return OAuthTokenSet{}, err
	}

	return OAuthTokenSet{
		AccessToken:           access,
		RefreshToken:          refresh,
		AccessTokenExpiresAt:  linked.AccessTokenExpiresAt,
		RefreshTokenExpiresAt: linked.RefreshTokenExpiresAt,
	}, nil
}
