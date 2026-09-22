package model

import "time"

type IdentitySource string

const (
	IdentitySourceCardOCR IdentitySource = "card_ocr"
	IdentitySourceNFCChip IdentitySource = "nfc_chip"
)

type CitizenIdentity struct {
	ID               string         `json:"id" gorm:"type:uuid;primaryKey"`
	UserID           string         `json:"user_id" gorm:"uniqueIndex;not null"`
	Source           IdentitySource `json:"source" gorm:"not null;default:card_ocr"`
	DocumentNumber   string         `json:"document_number" gorm:"index"`
	Surname          string         `json:"surname"`
	GivenNames       string         `json:"given_names"`
	Nationality      string         `json:"nationality"`
	DateOfBirth      string         `json:"date_of_birth"`
	Sex              string         `json:"sex"`
	DateOfExpiry     string         `json:"date_of_expiry"`
	MRZLine1         string         `json:"mrz_line1"`
	MRZLine2         string         `json:"mrz_line2"`
	MRZLine3         string         `json:"mrz_line3"`
	ChecksumValid    bool           `json:"checksum_valid"`
	Confidence       float64        `json:"confidence"`
	FrontImageKey    string         `json:"-"`
	BackImageKey     string         `json:"-"`
	NFCVerified      bool           `json:"nfc_verified"`
	ChipFaceImageKey string         `json:"-"`
	CreatedAt        time.Time      `json:"created_at"`
	UpdatedAt        time.Time      `json:"updated_at"`
}

func (CitizenIdentity) TableName() string { return "citizen_identities" }

type MRZResult struct {
	Line1          string
	Line2          string
	Line3          string
	DocumentNumber string
	Surname        string
	GivenNames     string
	Nationality    string
	DateOfBirth    string
	Sex            string
	DateOfExpiry   string
	ChecksumValid  bool
	Confidence     float64
}
