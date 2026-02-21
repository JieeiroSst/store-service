package domain

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type TransactionStatus string

const (
	TransactionStatusPending    TransactionStatus = "PENDING"
	TransactionStatusAuthorized TransactionStatus = "AUTHORIZED"
	TransactionStatusCaptured   TransactionStatus = "CAPTURED"
	TransactionStatusSettled    TransactionStatus = "SETTLED"
	TransactionStatusDeclined   TransactionStatus = "DECLINED"
	TransactionStatusVoided     TransactionStatus = "VOIDED"
)

type TransactionType string

const (
	TransactionTypeAuthorization TransactionType = "AUTHORIZATION" 
	TransactionTypeCapture       TransactionType = "CAPTURE"      
	TransactionTypeSettlement    TransactionType = "SETTLEMENT"    
	TransactionTypeRefund        TransactionType = "REFUND"
	TransactionTypeVoid          TransactionType = "VOID"
)

type Transaction struct {
	ID                uuid.UUID         `json:"id"                 db:"id"`
	IdempotencyKey    string            `json:"idempotency_key"    db:"idempotency_key"`
	WalletID          uuid.UUID         `json:"wallet_id"          db:"wallet_id"`
	MerchantID        uuid.UUID         `json:"merchant_id"        db:"merchant_id"`
	AcquirerBankID    uuid.UUID         `json:"acquirer_bank_id"   db:"acquirer_bank_id"`
	IssuerBankID      uuid.UUID         `json:"issuer_bank_id"     db:"issuer_bank_id"`
	CardID            uuid.UUID         `json:"card_id"            db:"card_id"`
	CardNetwork       CardNetwork       `json:"card_network"       db:"card_network"`
	Amount            decimal.Decimal   `json:"amount"             db:"amount"`
	Currency          Currency          `json:"currency"           db:"currency"`
	Fee               decimal.Decimal   `json:"fee"                db:"fee"`
	Type              TransactionType   `json:"type"               db:"type"`
	Status            TransactionStatus `json:"status"             db:"status"`
	Description       string            `json:"description"        db:"description"`
	AuthorizationCode string            `json:"authorization_code" db:"authorization_code"`
	BatchID           uuid.UUID         `json:"batch_id"           db:"batch_id"`
	Metadata          map[string]string `json:"metadata,omitempty" db:"-"`
	CreatedAt         time.Time         `json:"created_at"         db:"created_at"`
	UpdatedAt         time.Time         `json:"updated_at"         db:"updated_at"`
}

func (t *Transaction) Authorize(authCode string) {
	t.Status = TransactionStatusAuthorized
	t.AuthorizationCode = authCode
	t.UpdatedAt = time.Now()
}

func (t *Transaction) Capture() error {
	if t.Status != TransactionStatusAuthorized {
		return fmt.Errorf("%w: current status=%s", ErrTransactionNotPending, t.Status)
	}
	t.Type = TransactionTypeCapture
	t.Status = TransactionStatusCaptured
	t.UpdatedAt = time.Now()
	return nil
}

func (t *Transaction) Void() error {
	if t.Status != TransactionStatusAuthorized {
		return fmt.Errorf("%w: current status=%s", ErrTransactionNotPending, t.Status)
	}
	t.Type = TransactionTypeVoid
	t.Status = TransactionStatusVoided
	t.UpdatedAt = time.Now()
	return nil
}

type SettlementBatch struct {
	ID           uuid.UUID         `json:"id"            db:"id"`
	AcquirerID   uuid.UUID         `json:"acquirer_id"   db:"acquirer_id"`
	MerchantID   uuid.UUID         `json:"merchant_id"   db:"merchant_id"`
	TotalAmount  decimal.Decimal   `json:"total_amount"  db:"total_amount"`
	TotalFee     decimal.Decimal   `json:"total_fee"     db:"total_fee"`
	TxnCount     int               `json:"txn_count"     db:"txn_count"`
	Status       TransactionStatus `json:"status"        db:"status"`
	Transactions []Transaction     `json:"transactions,omitempty" db:"-"`
	ProcessedAt  *time.Time        `json:"processed_at,omitempty" db:"processed_at"`
	CreatedAt    time.Time         `json:"created_at"    db:"created_at"`
}

type ClearingRecord struct {
	ID          uuid.UUID       `json:"id"           db:"id"`
	BatchID     uuid.UUID       `json:"batch_id"     db:"batch_id"`
	CardNetwork CardNetwork     `json:"card_network" db:"card_network"`
	AcquirerID  uuid.UUID       `json:"acquirer_id"  db:"acquirer_id"`
	IssuerID    uuid.UUID       `json:"issuer_id"    db:"issuer_id"`
	NetAmount   decimal.Decimal `json:"net_amount"   db:"net_amount"`
	ClearedAt   time.Time       `json:"cleared_at"   db:"cleared_at"`
}
