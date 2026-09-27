package request

type TwoFactorSetup struct {
	CurrentPassword string `json:"currentPassword" binding:"max=128"`
}

type TwoFactorVerification struct {
	Code       string `json:"code" binding:"required,max=32"`
	BackupCode bool   `json:"backupCode"`
}

type TwoFactorManagement struct {
	CurrentPassword string `json:"currentPassword" binding:"max=128"`
	Code            string `json:"code" binding:"required,max=32"`
	BackupCode      bool   `json:"backupCode"`
}
