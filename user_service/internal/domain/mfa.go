package domain

import "time"

const MFAMethodTOTP = "totp"

type MFAChallenge struct {
	UserID   int    `json:"user_id"`
	Username string `json:"username"`
	Provider string `json:"provider"`
}

type MFAStatus struct {
	Enabled                bool
	EnabledAt              *time.Time
	RecoveryCodesRemaining int
}

type TOTPSetup struct {
	Secret     string
	OtpauthURI string
	QRCodePNG  []byte
	Issuer     string
	Account    string
	Digits     int
	Period     int
	Algorithm  string
}
