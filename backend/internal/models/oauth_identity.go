package models

import "time"

type OAuthIdentity struct {
	Model
	UserID                 uint       `gorm:"uniqueIndex:oauth_identities_user_provider_unique;not null"`
	Provider               string     `gorm:"uniqueIndex:oauth_identities_user_provider_unique;uniqueIndex:oauth_identities_provider_identity_unique;not null"`
	ProviderUserID         string     `gorm:"uniqueIndex:oauth_identities_provider_identity_unique;not null"`
	AccessTokenCiphertext  []byte     `gorm:"type:bytea;not null" json:"-"`
	RefreshTokenCiphertext []byte     `gorm:"type:bytea" json:"-"`
	AccessTokenExpiresAt   *time.Time `json:"-"`
	RefreshTokenExpiresAt  *time.Time `json:"-"`
}

func (OAuthIdentity) TableName() string { return "oauth_identities" }
