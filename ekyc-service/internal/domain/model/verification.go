package model

import "time"

type VerificationStatus string

const (
	VerificationPending  VerificationStatus = "pending"
	VerificationVerified VerificationStatus = "verified"
	VerificationFailed   VerificationStatus = "failed"
)

type EkycVerification struct {
	ID         string             `json:"id" gorm:"type:uuid;primaryKey"`
	UserID     string             `json:"user_id" gorm:"uniqueIndex;not null"`
	Status     VerificationStatus `json:"status" gorm:"not null;default:pending"`
	MatchScore float64            `json:"match_score"`
	VerifiedAt *time.Time         `json:"verified_at,omitempty"`
	CreatedAt  time.Time          `json:"created_at"`
	UpdatedAt  time.Time          `json:"updated_at"`
}

func (EkycVerification) TableName() string { return "ekyc_verifications" }

type EkycStatus struct {
	UserID       string            `json:"user_id"`
	Identity     *CitizenIdentity  `json:"identity,omitempty"`
	Face         *FaceBiometric    `json:"face,omitempty"`
	Verification *EkycVerification `json:"verification,omitempty"`
}
