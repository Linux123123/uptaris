package oauth

import (
	"strconv"
	"time"

	"github.com/uptaris/uptaris/backend/internal/auth"
	"golang.org/x/oauth2"
)

// Tokens preserves standard expiry and provider-specific refresh expiry.
func Tokens(token *oauth2.Token) auth.OAuthTokenSet {
	tokens := auth.OAuthTokenSet{
		AccessToken:  token.AccessToken,
		RefreshToken: token.RefreshToken,
	}

	if !token.Expiry.IsZero() {
		tokens.AccessTokenExpiresAt = &token.Expiry
	}

	// OAuth2 decodes JSON numbers as float64 and form-encoded numbers as int64.
	var seconds int64

	switch value := token.Extra("refresh_token_expires_in").(type) {
	case float64:
		seconds = int64(value)
	case int64:
		seconds = value
	case string:
		seconds, _ = strconv.ParseInt(value, 10, 64)
	}

	if seconds > 0 {
		expires := time.Now().Add(time.Duration(seconds) * time.Second)
		tokens.RefreshTokenExpiresAt = &expires
	}

	return tokens
}
