package model

import "time"

type UserTOTP struct {
	UserID        string     `gorm:"primaryKey;size:64" json:"user_id"`
	Secret        string     `gorm:"size:255;not null" json:"-"`
	Enabled       bool       `gorm:"not null;default:false" json:"enabled"`
	LastUsedStep  int64      `gorm:"not null;default:0" json:"-"`
	RecoveryCodes string     `gorm:"type:text" json:"-"`
	EnabledAt     *time.Time `json:"enabled_at"`
	CreatedAt     time.Time  `json:"created_at"`
	UpdatedAt     time.Time  `json:"updated_at"`
}

func (UserTOTP) TableName() string { return "user_totps" }

type TOTPStatus struct {
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

type TOTPVerification struct {
	UsedRecoveryCode bool
}
