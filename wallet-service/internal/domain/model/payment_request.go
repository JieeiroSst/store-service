package model

import (
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type PaymentRequestStatus string

const (
	PaymentRequestPending   PaymentRequestStatus = "PENDING"
	PaymentRequestPaid      PaymentRequestStatus = "PAID"
	PaymentRequestDeclined  PaymentRequestStatus = "DECLINED"
	PaymentRequestCancelled PaymentRequestStatus = "CANCELLED"
	PaymentRequestExpired   PaymentRequestStatus = "EXPIRED"
)

type PaymentRequestRole string

const (
	RoleIncoming PaymentRequestRole = "incoming"
	RoleOutgoing PaymentRequestRole = "outgoing"
)

type PaymentRequest struct {
	ID                uuid.UUID            `json:"id"                  db:"id"`
	RequesterUserID   uuid.UUID            `json:"requester_user_id"   db:"requester_user_id"`
	RequesterWalletID uuid.UUID            `json:"requester_wallet_id" db:"requester_wallet_id"`
	PayerUserID       uuid.UUID            `json:"payer_user_id"       db:"payer_user_id"`
	PayerWalletID     uuid.NullUUID        `json:"payer_wallet_id"     db:"payer_wallet_id"`
	TransactionID     uuid.NullUUID        `json:"transaction_id"      db:"transaction_id"`
	Amount            decimal.Decimal      `json:"amount"              db:"amount"`
	Currency          Currency             `json:"currency"            db:"currency"`
	Note              string               `json:"note"                db:"note"`
	Status            PaymentRequestStatus `json:"status"              db:"status"`
	ExpiresAt         time.Time            `json:"expires_at"          db:"expires_at"`
	CreatedAt         time.Time            `json:"created_at"          db:"created_at"`
	UpdatedAt         time.Time            `json:"updated_at"          db:"updated_at"`
}

type PaymentRequestFilter struct {
	UserID   uuid.UUID
	Role     PaymentRequestRole
	Status   PaymentRequestStatus
	Page     int
	PageSize int
}

func NewPaymentRequest(requester *Wallet, payerUserID uuid.UUID, amount decimal.Decimal, note string, ttl time.Duration, now time.Time) (*PaymentRequest, error) {
	if payerUserID == uuid.Nil {
		return nil, Invalid("payer_user_id is required")
	}
	if payerUserID == requester.UserID {
		return nil, ErrPaymentRequestSelf
	}
	if err := ValidateAmount(amount, requester.Currency); err != nil {
		return nil, err
	}
	note = strings.TrimSpace(note)
	if len(note) > 255 {
		return nil, Invalid("note must be at most 255 characters")
	}
	return &PaymentRequest{
		ID:                uuid.New(),
		RequesterUserID:   requester.UserID,
		RequesterWalletID: requester.ID,
		PayerUserID:       payerUserID,
		Amount:            amount,
		Currency:          requester.Currency,
		Note:              note,
		Status:            PaymentRequestPending,
		ExpiresAt:         now.Add(ttl),
		CreatedAt:         now,
		UpdatedAt:         now,
	}, nil
}

func (p *PaymentRequest) Expire(now time.Time) bool {
	if p.Status == PaymentRequestPending && !now.Before(p.ExpiresAt) {
		p.Status = PaymentRequestExpired
		p.UpdatedAt = now
		return true
	}
	return false
}

func (p *PaymentRequest) ensurePending(now time.Time) error {
	if p.Expire(now) {
		return ErrPaymentRequestExpired
	}
	if p.Status != PaymentRequestPending {
		return ErrPaymentRequestState
	}
	return nil
}

func (p *PaymentRequest) EnsurePayable(userID uuid.UUID, now time.Time) error {
	if p.PayerUserID != userID {
		return ErrNotPaymentRequestParty
	}
	return p.ensurePending(now)
}

func (p *PaymentRequest) MarkPaid(payerWalletID, txID uuid.UUID, now time.Time) {
	p.Status = PaymentRequestPaid
	p.PayerWalletID = NullID(payerWalletID)
	p.TransactionID = NullID(txID)
	p.UpdatedAt = now
}

func (p *PaymentRequest) Decline(userID uuid.UUID, now time.Time) error {
	if p.PayerUserID != userID {
		return ErrNotPaymentRequestParty
	}
	if err := p.ensurePending(now); err != nil {
		return err
	}
	p.Status = PaymentRequestDeclined
	p.UpdatedAt = now
	return nil
}

func (p *PaymentRequest) Cancel(userID uuid.UUID, now time.Time) error {
	if p.RequesterUserID != userID {
		return ErrNotPaymentRequestParty
	}
	if err := p.ensurePending(now); err != nil {
		return err
	}
	p.Status = PaymentRequestCancelled
	p.UpdatedAt = now
	return nil
}
