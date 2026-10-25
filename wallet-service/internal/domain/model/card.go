package model

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"math/big"
	"strings"
	"time"

	"github.com/google/uuid"
)

type CardNetwork string

const (
	CardNetworkVisa       CardNetwork = "VISA"
	CardNetworkMastercard CardNetwork = "MASTERCARD"
	CardNetworkAmex       CardNetwork = "AMEX"
)

type cardFormat struct {
	prefix string
	length int
}

var cardFormats = map[CardNetwork]cardFormat{
	CardNetworkVisa:       {prefix: "4", length: 16},
	CardNetworkMastercard: {prefix: "51", length: 16},
	CardNetworkAmex:       {prefix: "37", length: 15},
}

func (n CardNetwork) Valid() bool {
	_, ok := cardFormats[n]
	return ok
}

type CardStatus string

const (
	CardStatusActive    CardStatus = "ACTIVE"
	CardStatusBlocked   CardStatus = "BLOCKED"
	CardStatusExpired   CardStatus = "EXPIRED"
	CardStatusCancelled CardStatus = "CANCELLED"
)

const cardValidityYears = 3

type Card struct {
	ID             uuid.UUID   `json:"id"             db:"id"`
	WalletID       uuid.UUID   `json:"wallet_id"      db:"wallet_id"`
	IssuerBankID   uuid.UUID   `json:"issuer_bank_id" db:"issuer_bank_id"`
	CardNumber     string      `json:"card_number"    db:"card_number"`
	CardNumberHash string      `json:"-"              db:"card_number_hash"`
	HolderName     string      `json:"holder_name"    db:"holder_name"`
	Network        CardNetwork `json:"network"        db:"network"`
	ExpiryMonth    int         `json:"expiry_month"   db:"expiry_month"`
	ExpiryYear     int         `json:"expiry_year"    db:"expiry_year"`
	Status         CardStatus  `json:"status"         db:"status"`
	CreatedAt      time.Time   `json:"created_at"     db:"created_at"`
	UpdatedAt      time.Time   `json:"updated_at"     db:"updated_at"`
}

const MinCardHashKeyLen = 32

func NewCard(walletID, issuerBankID uuid.UUID, network CardNetwork, holderName string, hashKey []byte, now time.Time) (*Card, error) {
	if len(hashKey) < MinCardHashKeyLen {
		return nil, ErrCardIssuing
	}
	if !network.Valid() {
		return nil, ErrCardNetwork
	}
	holderName = strings.TrimSpace(holderName)
	if holderName == "" || len(holderName) > 100 {
		return nil, Invalid("holder_name is required and must be at most 100 characters")
	}
	number, err := GenerateCardNumber(network)
	if err != nil {
		return nil, err
	}
	return &Card{
		ID:             uuid.New(),
		WalletID:       walletID,
		IssuerBankID:   issuerBankID,
		CardNumber:     MaskCardNumber(number),
		CardNumberHash: HashPAN(hashKey, number),
		HolderName:     strings.ToUpper(holderName),
		Network:        network,
		ExpiryMonth:    int(now.Month()),
		ExpiryYear:     now.Year() + cardValidityYears,
		Status:         CardStatusActive,
		CreatedAt:      now,
		UpdatedAt:      now,
	}, nil
}

func (c *Card) IsExpired(now time.Time) bool {
	expiry := time.Date(c.ExpiryYear, time.Month(c.ExpiryMonth+1), 0, 23, 59, 59, 0, time.UTC)
	return now.After(expiry)
}

func (c *Card) EnsureUsable(now time.Time) error {
	if c.IsExpired(now) || c.Status == CardStatusExpired {
		return ErrCardExpired
	}
	if c.Status != CardStatusActive {
		return ErrCardBlocked
	}
	return nil
}

func (c *Card) Block(now time.Time) {
	if c.Status == CardStatusActive {
		c.Status = CardStatusBlocked
		c.UpdatedAt = now
	}
}

func (c *Card) Unblock(now time.Time) error {
	if c.Status != CardStatusBlocked {
		return ErrCardNotActive
	}
	if c.IsExpired(now) {
		return ErrCardExpired
	}
	c.Status = CardStatusActive
	c.UpdatedAt = now
	return nil
}

func GenerateCardNumber(network CardNetwork) (string, error) {
	f, ok := cardFormats[network]
	if !ok {
		return "", ErrCardNetwork
	}
	var b strings.Builder
	b.WriteString(f.prefix)
	for b.Len() < f.length-1 {
		n, err := rand.Int(rand.Reader, big.NewInt(10))
		if err != nil {
			return "", err
		}
		b.WriteByte(byte('0' + n.Int64()))
	}
	partial := b.String()
	return partial + string(byte('0'+luhnCheckDigit(partial))), nil
}

func luhnCheckDigit(partial string) int {
	sum := 0
	double := true
	for i := len(partial) - 1; i >= 0; i-- {
		d := int(partial[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return (10 - sum%10) % 10
}

func LuhnValid(number string) bool {
	if len(number) < 2 {
		return false
	}
	return luhnCheckDigit(number[:len(number)-1]) == int(number[len(number)-1]-'0')
}

func MaskCardNumber(number string) string {
	if len(number) < 4 {
		return "****"
	}
	return "**** **** **** " + number[len(number)-4:]
}

func HashPAN(key []byte, pan string) string {
	m := hmac.New(sha256.New, key)
	m.Write([]byte(pan))
	return hex.EncodeToString(m.Sum(nil))
}

func HashSecret(value string) string {
	h := sha256.Sum256([]byte(value))
	return hex.EncodeToString(h[:])
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
