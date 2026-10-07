package http

import (
	"time"

	"github.com/JIeeiroSst/card-service/internal/application"
	"github.com/JIeeiroSst/card-service/internal/domain"
)

type limitsDTO struct {
	PerTransaction int64 `json:"per_transaction"`
	Daily          int64 `json:"daily"`
}

type controlsDTO struct {
	POS         bool `json:"pos"`
	Contactless bool `json:"contactless"`
	Ecommerce   bool `json:"ecommerce"`
	ATM         bool `json:"atm"`
}

func (c controlsDTO) domain() domain.Controls {
	return domain.Controls{POS: c.POS, Contactless: c.Contactless, Ecommerce: c.Ecommerce, ATM: c.ATM}
}

type programDTO struct {
	Code               string      `json:"code"`
	Name               string      `json:"name"`
	Mode               string      `json:"mode"`
	BIN                string      `json:"bin"`
	Currency           string      `json:"currency"`
	ValidityYears      int         `json:"validity_years"`
	CardTypes          []string    `json:"card_types"`
	MaxActiveVirtual   int         `json:"max_active_virtual"`
	DefaultCreditLimit int64       `json:"default_credit_limit,omitempty"`
	MaxCreditLimit     int64       `json:"max_credit_limit,omitempty"`
	DefaultLimits      limitsDTO   `json:"default_limits"`
	MaxLimits          limitsDTO   `json:"max_limits"`
	SupportedChannels  controlsDTO `json:"supported_channels"`
	StepUpThreshold    int64       `json:"step_up_threshold"`
}

type accountDTO struct {
	ID           string    `json:"id"`
	CustomerID   string    `json:"customer_id"`
	ProgramCode  string    `json:"program_code"`
	Mode         string    `json:"mode"`
	Currency     string    `json:"currency"`
	HolderName   string    `json:"holder_name"`
	Status       string    `json:"status"`
	StatusReason string    `json:"status_reason,omitempty"`
	CreditLimit  int64     `json:"credit_limit"`
	Balance      int64     `json:"balance"`
	Held         int64     `json:"held"`
	Available    int64     `json:"available"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type cardDTO struct {
	ID               string      `json:"id"`
	AccountID        string      `json:"account_id"`
	CustomerID       string      `json:"customer_id"`
	ProgramCode      string      `json:"program_code"`
	Type             string      `json:"type"`
	CardholderName   string      `json:"cardholder_name"`
	MaskedPAN        string      `json:"masked_pan"`
	Last4            string      `json:"last4"`
	Expiry           string      `json:"expiry"`
	ValidUntil       *time.Time  `json:"valid_until,omitempty"`
	Status           string      `json:"status"`
	StatusReason     string      `json:"status_reason,omitempty"`
	Limits           limitsDTO   `json:"limits"`
	Controls         controlsDTO `json:"controls"`
	SpentToday       *int64      `json:"spent_today,omitempty"`
	PINSet           bool        `json:"pin_set"`
	PINLocked        bool        `json:"pin_locked"`
	ReplacesCardID   string      `json:"replaces_card_id,omitempty"`
	ReplacedByCardID string      `json:"replaced_by_card_id,omitempty"`
	CreatedAt        time.Time   `json:"created_at"`
	UpdatedAt        time.Time   `json:"updated_at"`
	ActivatedAt      *time.Time  `json:"activated_at,omitempty"`
}

type issuedCardDTO struct {
	Card   cardDTO `json:"card"`
	PAN    string  `json:"pan"`
	CVV    string  `json:"cvv"`
	Notice string  `json:"notice"`
}

type authorizationDTO struct {
	ID               string    `json:"id,omitempty"`
	AccountID        string    `json:"account_id,omitempty"`
	CardID           string    `json:"card_id,omitempty"`
	ProcessingCode   string    `json:"processing_code"`
	Status           string    `json:"status"`
	Approved         bool      `json:"approved"`
	ResponseCode     string    `json:"response_code"`
	Reason           string    `json:"reason"`
	AuthCode         string    `json:"auth_code,omitempty"`
	Amount           int64     `json:"amount"`
	ConfirmedAmount  int64     `json:"confirmed_amount,omitempty"`
	Currency         string    `json:"currency"`
	Channel          string    `json:"channel"`
	Merchant         string    `json:"merchant,omitempty"`
	MCC              string    `json:"mcc,omitempty"`
	AuthenticationID string    `json:"authentication_id,omitempty"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

type transactionDTO struct {
	ID              string    `json:"id"`
	CardID          string    `json:"card_id,omitempty"`
	AuthorizationID string    `json:"authorization_id,omitempty"`
	Type            string    `json:"type"`
	ProcessingCode  string    `json:"processing_code"`
	Amount          int64     `json:"amount"`
	BalanceAfter    int64     `json:"balance_after"`
	Description     string    `json:"description,omitempty"`
	CreatedAt       time.Time `json:"created_at"`
}

type resolvedCardDTO struct {
	CardID     string `json:"card_id"`
	AccountID  string `json:"account_id"`
	CustomerID string `json:"customer_id"`
	UserID     int64  `json:"user_id"`
	MaskedPAN  string `json:"masked_pan"`
	Status     string `json:"status"`
}

const issueNotice = "pan and cvv are returned only once; they are never stored in clear and cannot be retrieved again"

func toLimits(l domain.Limits) limitsDTO {
	return limitsDTO{PerTransaction: l.PerTransaction, Daily: l.Daily}
}

func toControls(c domain.Controls) controlsDTO {
	return controlsDTO{POS: c.POS, Contactless: c.Contactless, Ecommerce: c.Ecommerce, ATM: c.ATM}
}

func toProgram(p domain.Program) programDTO {
	types := make([]string, 0, len(p.CardTypes))
	for _, t := range p.CardTypes {
		types = append(types, string(t))
	}
	return programDTO{
		Code: p.Code, Name: p.Name, Mode: string(p.Mode), BIN: p.BIN, Currency: p.Currency,
		ValidityYears: p.ValidityYears, CardTypes: types, MaxActiveVirtual: p.MaxActiveVirtual,
		DefaultCreditLimit: p.DefaultCreditLimit, MaxCreditLimit: p.MaxCreditLimit,
		DefaultLimits: toLimits(p.DefaultLimits), MaxLimits: toLimits(p.MaxLimits),
		SupportedChannels: toControls(p.Supported), StepUpThreshold: p.StepUpThreshold,
	}
}

func toAccount(a *domain.Account) accountDTO {
	return accountDTO{
		ID: a.ID, CustomerID: a.CustomerID, ProgramCode: a.ProgramCode, Mode: string(a.Mode), Currency: a.Currency,
		HolderName: a.HolderName, Status: string(a.Status), StatusReason: a.StatusReason, CreditLimit: a.CreditLimit,
		Balance: a.Balance, Held: a.Held, Available: a.Available(), CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}

func toCard(c *domain.Card) cardDTO {
	return cardDTO{
		ID: c.ID, AccountID: c.AccountID, CustomerID: c.CustomerID, ProgramCode: c.ProgramCode, Type: string(c.Type),
		CardholderName: c.CardholderName, MaskedPAN: c.MaskedPAN(), Last4: c.Last4, Expiry: c.Expiry.String(),
		ValidUntil: c.ValidUntil, Status: string(c.Status), StatusReason: c.StatusReason, Limits: toLimits(c.Limits),
		Controls: toControls(c.Controls), PINSet: c.PINSet(), PINLocked: c.PINLocked(), ReplacesCardID: c.ReplacesCardID,
		ReplacedByCardID: c.ReplacedByCardID, CreatedAt: c.CreatedAt, UpdatedAt: c.UpdatedAt, ActivatedAt: c.ActivatedAt,
	}
}

func toIssued(i *application.IssuedCard) issuedCardDTO {
	return issuedCardDTO{Card: toCard(i.Card), PAN: i.PAN, CVV: i.CVV, Notice: issueNotice}
}

func toAuthorization(a *domain.Authorization) authorizationDTO {
	return authorizationDTO{
		ID: a.ID, AccountID: a.AccountID, CardID: a.CardID, ProcessingCode: string(a.ProcessingCode),
		Status: string(a.Status), Approved: a.Approved(), ResponseCode: string(a.Code), Reason: a.Reason,
		AuthCode: a.AuthCode, Amount: a.Amount, ConfirmedAmount: a.ConfirmedAmount, Currency: a.Currency,
		Channel: string(a.Channel), Merchant: a.Merchant, MCC: a.MCC, AuthenticationID: a.AuthenticationID,
		CreatedAt: a.CreatedAt, UpdatedAt: a.UpdatedAt,
	}
}

func toTransaction(t *domain.Transaction) transactionDTO {
	return transactionDTO{
		ID: t.ID, CardID: t.CardID, AuthorizationID: t.AuthorizationID, Type: string(t.Type),
		ProcessingCode: string(t.ProcessingCode), Amount: t.Amount, BalanceAfter: t.BalanceAfter,
		Description: t.Description, CreatedAt: t.CreatedAt,
	}
}

func mapList[T, D any](in []*T, f func(*T) D) []D {
	out := make([]D, 0, len(in))
	for _, v := range in {
		out = append(out, f(v))
	}
	return out
}
