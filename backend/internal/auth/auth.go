package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"fmt"
	"strconv"
	"time"

	"errors"

	"github.com/alexedwards/argon2id"
	"github.com/golang-jwt/jwt/v5"
)

var passwordParams = &argon2id.Params{Memory: 19 * 1024, Iterations: 2, Parallelism: 1, SaltLength: 16, KeyLength: 32}

type Claims struct {
	Role      string `json:"role"`
	SessionID uint   `json:"sid"`
	jwt.RegisteredClaims
}

func HashPassword(password string) (string, error) {
	return argon2id.CreateHash(password, passwordParams)
}

func CheckPassword(encoded, password string) error {
	match, err := argon2id.ComparePasswordAndHash(password, encoded)
	if err != nil {
		return err
	}
	if !match {
		return errors.New("password mismatch")
	}
	return nil
}

func NewRefresh() (string, error) {
	b := make([]byte, 32)
	if _, e := rand.Read(b); e != nil {
		return "", e
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func HashRefresh(token, pepper string) string {
	sum := sha256.Sum256([]byte(token + pepper))
	return fmt.Sprintf("%x", sum)
}

func NewJTI() (string, error) {
	return NewRefresh()
}

func Issue(userID uint64, sessionID uint, role, secret string, ttl time.Duration) (string, string, error) {
	jti, e := NewJTI()
	if e != nil {
		return "", "", e
	}
	now := time.Now()
	claims := Claims{
		Role:      role,
		SessionID: sessionID,
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   strconv.FormatUint(userID, 10),
			ID:        jti,
			Issuer:    "uptaris",
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signed, e := token.SignedString([]byte(secret))
	return signed, jti, e
}

func Parse(token, secret string) (*Claims, error) {
	claims := new(Claims)
	_, err := jwt.ParseWithClaims(token, claims, func(t *jwt.Token) (interface{}, error) {
		if t.Method.Alg() != jwt.SigningMethodHS256.Alg() {
			return nil, errors.New("unexpected signing method")
		}
		return []byte(secret), nil
	}, jwt.WithValidMethods([]string{"HS256"}), jwt.WithIssuer("uptaris"), jwt.WithExpirationRequired(), jwt.WithIssuedAt())
	if err == nil && (claims.SessionID == 0 || claims.ID == "" || claims.Subject == "" || claims.ExpiresAt == nil) {
		err = errors.New("required claims missing")
	}
	return claims, err
}
