package domain

import (
	"time"

	"github.com/google/uuid"
)

type CardNetwork string

const (
	CardNetworkVisa       CardNetwork = "VISA"
	CardNetworkMastercard CardNetwork = "MASTERCARD"
	CardNetworkAmex       CardNetwork = "AMEX"
)

type CardStatus string

const (
	CardStatusActive    CardStatus = "ACTIVE"
	CardStatusBlocked   CardStatus = "BLOCKED"
	CardStatusExpired   CardStatus = "EXPIRED"
	CardStatusCancelled CardStatus = "CANCELLED"
)

type Card struct {
	ID             uuid.UUID  `json:"id"               db:"id"`
	WalletID       uuid.UUID  `json:"wallet_id"        db:"wallet_id"`
	IssuerBankID   uuid.UUID  `json:"issuer_bank_id"   db:"issuer_bank_id"`
	CardNumber     string     `json:"card_number"      db:"card_number"`      // masked: **** **** **** 1234
	CardNumberHash string     `json:"-"                db:"card_number_hash"` // SHA-256 for lookup
	HolderName     string     `json:"holder_name"      db:"holder_name"`
	Network        CardNetwork `json:"network"         db:"network"`
	ExpiryMonth    int        `json:"expiry_month"     db:"expiry_month"`
	ExpiryYear     int        `json:"expiry_year"      db:"expiry_year"`
	Status         CardStatus `json:"status"           db:"status"`
	CreatedAt      time.Time  `json:"created_at"       db:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"       db:"updated_at"`
}

func (c *Card) IsExpired() bool {
	now := time.Now()
	expiry := time.Date(c.ExpiryYear, time.Month(c.ExpiryMonth+1), 0, 23, 59, 59, 0, time.UTC)
	return now.After(expiry)
}

func (c *Card) IsUsable() bool {
	return c.Status == CardStatusActive && !c.IsExpired()
}

type Bank struct {
	ID        uuid.UUID `json:"id"         db:"id"`
	Name      string    `json:"name"       db:"name"`
	Code      string    `json:"code"       db:"code"`
	Country   string    `json:"country"    db:"country"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}

type Merchant struct {
	ID             uuid.UUID `json:"id"               db:"id"`
	Name           string    `json:"name"             db:"name"`
	MCC            string    `json:"mcc"              db:"mcc"` 
	AcquirerBankID uuid.UUID `json:"acquirer_bank_id" db:"acquirer_bank_id"`
	Country        string    `json:"country"          db:"country"`
	CreatedAt      time.Time `json:"created_at"       db:"created_at"`
}
