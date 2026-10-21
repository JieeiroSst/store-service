package model

import (
	"database/sql/driver"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	TransactionStatusPending           TransactionStatus = "PENDING"
	TransactionStatusAuthorized        TransactionStatus = "AUTHORIZED"
	TransactionStatusCaptured          TransactionStatus = "CAPTURED"
	TransactionStatusSettled           TransactionStatus = "SETTLED"
	TransactionStatusDeclined          TransactionStatus = "DECLINED"
	TransactionStatusVoided            TransactionStatus = "VOIDED"
	TransactionStatusCompleted         TransactionStatus = "COMPLETED"
	TransactionStatusFailed            TransactionStatus = "FAILED"
	TransactionStatusRefunded          TransactionStatus = "REFUNDED"
	TransactionStatusPartiallyRefunded TransactionStatus = "PARTIALLY_REFUNDED"
)

type TransactionType string

const (
	TransactionTypeAuthorization TransactionType = "AUTHORIZATION"
	TransactionTypeCapture       TransactionType = "CAPTURE"
	TransactionTypeSettlement    TransactionType = "SETTLEMENT"
	TransactionTypeRefund        TransactionType = "REFUND"
	TransactionTypeVoid          TransactionType = "VOID"
	TransactionTypeTopUp         TransactionType = "TOPUP"
	TransactionTypeWithdrawal    TransactionType = "WITHDRAWAL"
	TransactionTypeTransferIn    TransactionType = "TRANSFER_IN"
	TransactionTypeTransferOut   TransactionType = "TRANSFER_OUT"
	TransactionTypePayment       TransactionType = "PAYMENT"
	TransactionTypeAdjustment    TransactionType = "ADJUSTMENT"
)

type Metadata map[string]string

func (m Metadata) Value() (driver.Value, error) {
	if len(m) == 0 {
		return nil, nil
	}
	b, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

func (m *Metadata) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*m = nil
		return nil
	case []byte:
		return json.Unmarshal(v, m)
	case string:
		return json.Unmarshal([]byte(v), m)
	}
	return errors.New("metadata: unsupported column type")
}

type Transaction struct {
	ID                   uuid.UUID         `json:"id"                     db:"id"`
	IdempotencyKey       string            `json:"idempotency_key"        db:"idempotency_key"`
	WalletID             uuid.UUID         `json:"wallet_id"              db:"wallet_id"`
	CounterpartyWalletID uuid.NullUUID     `json:"counterparty_wallet_id" db:"counterparty_wallet_id"`
	MerchantID           uuid.NullUUID     `json:"merchant_id"            db:"merchant_id"`
	AcquirerBankID       uuid.NullUUID     `json:"acquirer_bank_id"       db:"acquirer_bank_id"`
	IssuerBankID         uuid.NullUUID     `json:"issuer_bank_id"         db:"issuer_bank_id"`
	CardID               uuid.NullUUID     `json:"card_id"                db:"card_id"`
	BankAccountID        uuid.NullUUID     `json:"bank_account_id"        db:"bank_account_id"`
	ReferenceID          uuid.NullUUID     `json:"reference_id"           db:"reference_id"`
	CardNetwork          CardNetwork       `json:"card_network"           db:"card_network"`
	Amount               decimal.Decimal   `json:"amount"                 db:"amount"`
	Currency             Currency          `json:"currency"               db:"currency"`
	Fee                  decimal.Decimal   `json:"fee"                    db:"fee"`
	RefundedAmount       decimal.Decimal   `json:"refunded_amount"        db:"refunded_amount"`
	Type                 TransactionType   `json:"type"                   db:"type"`
	Status               TransactionStatus `json:"status"                 db:"status"`
	Description          string            `json:"description"            db:"description"`
	AuthorizationCode    string            `json:"authorization_code"     db:"authorization_code"`
	ExternalRef          string            `json:"external_ref"           db:"external_ref"`
	FailureReason        string            `json:"failure_reason"         db:"failure_reason"`
	BatchID              uuid.NullUUID     `json:"batch_id"               db:"batch_id"`
	Metadata             Metadata          `json:"metadata,omitempty"     db:"metadata"`
	CreatedAt            time.Time         `json:"created_at"             db:"created_at"`
	UpdatedAt            time.Time         `json:"updated_at"             db:"updated_at"`
}

func NullID(id uuid.UUID) uuid.NullUUID {
	return uuid.NullUUID{UUID: id, Valid: id != uuid.Nil}
}

func NewTransaction(w *Wallet, typ TransactionType, status TransactionStatus, amount decimal.Decimal, now time.Time) *Transaction {
	return &Transaction{
		ID:             uuid.New(),
		WalletID:       w.ID,
		Amount:         amount,
		Currency:       w.Currency,
		Fee:            decimal.Zero,
		RefundedAmount: decimal.Zero,
		Type:           typ,
		Status:         status,
		CreatedAt:      now,
		UpdatedAt:      now,
	}
}

func (t *Transaction) TotalDebit() decimal.Decimal {
	return t.Amount.Add(t.Fee)
}

func (t *Transaction) Authorize(authCode string, now time.Time) {
	t.Status = TransactionStatusAuthorized
	t.AuthorizationCode = authCode
	t.UpdatedAt = now
}

func (t *Transaction) Capture(now time.Time) error {
	if t.Status != TransactionStatusAuthorized {
		return ErrTransactionState
	}
	t.Type = TransactionTypeCapture
	t.Status = TransactionStatusCaptured
	t.UpdatedAt = now
	return nil
}

func (t *Transaction) Void(now time.Time) error {
	if t.Status != TransactionStatusAuthorized {
		return ErrTransactionState
	}
	t.Type = TransactionTypeVoid
	t.Status = TransactionStatusVoided
	t.UpdatedAt = now
	return nil
}

func (t *Transaction) Complete(externalRef string, now time.Time) error {
	if t.Status != TransactionStatusPending {
		return ErrTransactionState
	}
	t.Status = TransactionStatusCompleted
	t.ExternalRef = externalRef
	t.UpdatedAt = now
	return nil
}

func (t *Transaction) Fail(reason string, now time.Time) error {
	if t.Status != TransactionStatusPending {
		return ErrTransactionState
	}
	t.Status = TransactionStatusFailed
	t.FailureReason = reason
	t.UpdatedAt = now
	return nil
}

func (t *Transaction) RefundableAmount() decimal.Decimal {
	return t.Amount.Sub(t.RefundedAmount)
}

func (t *Transaction) ApplyRefund(amount decimal.Decimal, now time.Time) (decimal.Decimal, error) {
	if t.Type != TransactionTypePayment {
		return decimal.Zero, ErrNotRefundable
	}
	if t.Status != TransactionStatusCompleted && t.Status != TransactionStatusPartiallyRefunded {
		return decimal.Zero, ErrNotRefundable
	}
	remaining := t.RefundableAmount()
	if amount.IsZero() {
		amount = remaining
	}
	if !amount.IsPositive() {
		return decimal.Zero, ErrInvalidAmount
	}
	if amount.GreaterThan(remaining) {
		return decimal.Zero, ErrRefundExceeds
	}
	t.RefundedAmount = t.RefundedAmount.Add(amount)
	if t.RefundedAmount.Equal(t.Amount) {
		t.Status = TransactionStatusRefunded
	} else {
		t.Status = TransactionStatusPartiallyRefunded
	}
	t.UpdatedAt = now
	return amount, nil
}

type TransactionFilter struct {
	WalletID uuid.UUID
	Types    []TransactionType
	Statuses []TransactionStatus
	From     *time.Time
	To       *time.Time
	Page     int
	PageSize int
}

const (
	DefaultPageSize = 20
	MaxPageSize     = 100
)

func NormalizePage(page, pageSize int) (int, int) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = DefaultPageSize
	}
	if pageSize > MaxPageSize {
		pageSize = MaxPageSize
	}
	return page, pageSize
}
