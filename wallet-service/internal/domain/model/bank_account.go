package model

import (
	"regexp"
	"strings"
	"time"

	"github.com/google/uuid"
)

type BankAccountStatus string

const (
	BankAccountActive  BankAccountStatus = "ACTIVE"
	BankAccountRemoved BankAccountStatus = "REMOVED"
)

var accountNumberPattern = regexp.MustCompile(`^[0-9]{6,20}$`)

type BankAccount struct {
	ID                uuid.UUID         `json:"id"             db:"id"`
	UserID            uuid.UUID         `json:"user_id"        db:"user_id"`
	BankID            uuid.UUID         `json:"bank_id"        db:"bank_id"`
	AccountNumber     string            `json:"account_number" db:"account_number"`
	AccountNumberHash string            `json:"-"              db:"account_number_hash"`
	AccountName       string            `json:"account_name"   db:"account_name"`
	IsDefault         bool              `json:"is_default"     db:"is_default"`
	Status            BankAccountStatus `json:"status"         db:"status"`
	CreatedAt         time.Time         `json:"created_at"     db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"     db:"updated_at"`
}

func NewBankAccount(userID, bankID uuid.UUID, accountNumber, accountName string, now time.Time) (*BankAccount, error) {
	accountNumber = strings.ReplaceAll(strings.TrimSpace(accountNumber), " ", "")
	if !accountNumberPattern.MatchString(accountNumber) {
		return nil, ErrBankAccountNumber
	}
	accountName = strings.TrimSpace(accountName)
	if accountName == "" || len(accountName) > 100 {
		return nil, Invalid("account_name is required and must be at most 100 characters")
	}
	return &BankAccount{
		ID:                uuid.New(),
		UserID:            userID,
		BankID:            bankID,
		AccountNumber:     "****" + accountNumber[len(accountNumber)-4:],
		AccountNumberHash: HashSecret(bankID.String() + ":" + accountNumber),
		AccountName:       strings.ToUpper(accountName),
		Status:            BankAccountActive,
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (a *BankAccount) EnsureUsableBy(userID uuid.UUID) error {
	if a.UserID != userID || a.Status != BankAccountActive {
		return ErrBankAccountAbsent
	}
	return nil
}

func (a *BankAccount) Remove(now time.Time) {
	a.Status = BankAccountRemoved
	a.IsDefault = false
	a.UpdatedAt = now
}
