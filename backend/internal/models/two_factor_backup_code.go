package models

// TwoFactorBackupCode stores a single-use recovery credential hash.
type TwoFactorBackupCode struct {
	Model
	UserID   uint   `json:"-"`
	CodeHash string `json:"-"`
}

func (TwoFactorBackupCode) TableName() string { return "two_factor_backup_codes" }
