package domain

import (
	"strings"
	"time"
)

const (
	maxMerchantLen = 100
	maxAmount      = 10_000_000_000
)

type Channel string

const (
	ChannelPOS         Channel = "POS"
	ChannelContactless Channel = "CONTACTLESS"
	ChannelEcommerce   Channel = "ECOM"
	ChannelATM         Channel = "ATM"
)

var Channels = []Channel{ChannelPOS, ChannelContactless, ChannelEcommerce, ChannelATM}

func ParseChannel(s string) (Channel, error) {
	ch := Channel(strings.ToUpper(strings.TrimSpace(s)))
	for _, c := range Channels {
		if c == ch {
			return c, nil
		}
	}
	return "", Invalid("channel must be one of POS, CONTACTLESS, ECOM, ATM, got %q", s)
}

type ProcessingCode string

const (
	ProcessingPurchase   ProcessingCode = "00"
	ProcessingWithdrawal ProcessingCode = "01"
	ProcessingRefund     ProcessingCode = "20"
	ProcessingPayment    ProcessingCode = "28"
)

func ParseProcessingCode(s string) (ProcessingCode, error) {
	switch pc := ProcessingCode(strings.TrimSpace(s)); pc {
	case "":
		return ProcessingPurchase, nil
	case ProcessingPurchase, ProcessingWithdrawal, ProcessingRefund:
		return pc, nil
	}
	return "", Invalid("processing_code must be 00 (purchase), 01 (withdrawal) or 20 (refund), got %q", s)
}

func (p ProcessingCode) Debit() bool { return p == ProcessingPurchase || p == ProcessingWithdrawal }

func (p ProcessingCode) TransactionType() TransactionType {
	switch p {
	case ProcessingWithdrawal:
		return TxnWithdrawal
	case ProcessingRefund:
		return TxnRefund
	case ProcessingPayment:
		return TxnPayment
	}
	return TxnPurchase
}

type ResponseCode string

const (
	CodeApproved           ResponseCode = "00"
	CodeInvalidTxn         ResponseCode = "12"
	CodeInvalidCard        ResponseCode = "14"
	CodeLostCard           ResponseCode = "41"
	CodeStolenCard         ResponseCode = "43"
	CodeClosedCard         ResponseCode = "46"
	CodeInsufficientFunds  ResponseCode = "51"
	CodeExpiredCard        ResponseCode = "54"
	CodeIncorrectPIN       ResponseCode = "55"
	CodeNotPermitted       ResponseCode = "57"
	CodeExceedsLimit       ResponseCode = "61"
	CodeRestrictedCard     ResponseCode = "62"
	CodePINTriesExceeded   ResponseCode = "75"
	CodeCardNotActivated   ResponseCode = "78"
	CodeCVVMismatch        ResponseCode = "N7"
	CodeAuthenticationReqd ResponseCode = "1A"
)

type Check int

const (
	CheckNotProvided Check = iota
	CheckMatch
	CheckMismatch
)

type AuthorizationRequest struct {
	PAN              string
	Expiry           Expiry
	CVV              string
	PIN              string
	Amount           int64
	Currency         string
	Channel          Channel
	ProcessingCode   ProcessingCode
	Merchant         string
	MCC              string
	AuthenticationID string
}

func (r *AuthorizationRequest) Validate() error {
	ch, err := ParseChannel(string(r.Channel))
	if err != nil {
		return err
	}
	r.Channel = ch
	pc, err := ParseProcessingCode(string(r.ProcessingCode))
	if err != nil {
		return err
	}
	r.ProcessingCode = pc
	r.PAN = NormalizePAN(r.PAN)
	if !ValidPAN(r.PAN) {
		return Invalid("pan must be %d digits with a valid Luhn check digit", PANLength)
	}
	if r.CVV != "" && (len(r.CVV) != 3 || !isDigits(r.CVV)) {
		return Invalid("cvv must be 3 digits")
	}
	if r.PIN != "" && (len(r.PIN) < 4 || len(r.PIN) > 6 || !isDigits(r.PIN)) {
		return Invalid("pin must be 4-6 digits")
	}
	if r.Amount <= 0 || r.Amount > maxAmount {
		return Invalid("amount must be between 1 and %d", int64(maxAmount))
	}
	r.Currency = strings.ToUpper(strings.TrimSpace(r.Currency))
	if len(r.Currency) != 3 {
		return Invalid("currency must be an ISO 4217 code")
	}
	r.Merchant = strings.TrimSpace(r.Merchant)
	if len(r.Merchant) > maxMerchantLen {
		return Invalid("merchant must be at most %d characters", maxMerchantLen)
	}
	if r.MCC != "" && (len(r.MCC) != 4 || !isDigits(r.MCC)) {
		return Invalid("mcc must be 4 digits")
	}
	r.AuthenticationID = strings.TrimSpace(r.AuthenticationID)
	if len(r.AuthenticationID) > maxIDLen {
		return Invalid("authentication_id is too long")
	}
	return nil
}

func (r AuthorizationRequest) NeedsStepUp(p Program) bool {
	return r.Channel == ChannelEcommerce && r.ProcessingCode == ProcessingPurchase && r.Amount >= p.StepUpThreshold
}

type Credentials struct {
	CVV    Check
	PIN    Check
	StepUp Check
}

type Decision struct {
	Code   ResponseCode
	Reason string
}

func (d Decision) Approved() bool { return d.Code == CodeApproved }

func Decide(card *Card, account *Account, p Program, req AuthorizationRequest, cred Credentials, spentToday int64, now time.Time) Decision {
	switch account.Status {
	case AccountBlocked:
		return Decision{CodeRestrictedCard, "account is blocked"}
	case AccountCancelled:
		return Decision{CodeClosedCard, "account is cancelled"}
	}
	if req.ProcessingCode == ProcessingRefund {
		if req.Currency != account.Currency {
			return Decision{CodeInvalidTxn, "currency " + req.Currency + " is not supported"}
		}
		return Decision{CodeApproved, "refund accepted"}
	}
	switch card.Status {
	case StatusPending:
		return Decision{CodeCardNotActivated, "card is not activated"}
	case StatusBlocked:
		return Decision{CodeRestrictedCard, "card is blocked"}
	case StatusLost:
		return Decision{CodeLostCard, "card reported lost"}
	case StatusRobbed:
		return Decision{CodeStolenCard, "card reported robbed"}
	case StatusCancelled, StatusDamaged:
		return Decision{CodeClosedCard, "card is " + strings.ToLower(string(card.Status))}
	}
	if card.Expired(now) {
		return Decision{CodeExpiredCard, "card is expired"}
	}
	if req.Expiry != card.Expiry {
		return Decision{CodeExpiredCard, "expiry date does not match"}
	}
	if cred.CVV == CheckMismatch || req.Channel == ChannelEcommerce && cred.CVV == CheckNotProvided {
		return Decision{CodeCVVMismatch, "cvv missing or incorrect"}
	}
	if req.Channel == ChannelATM || req.ProcessingCode == ProcessingWithdrawal || cred.PIN != CheckNotProvided {
		switch {
		case !card.PINSet():
			return Decision{CodeNotPermitted, "pin is not set"}
		case card.PINLocked():
			return Decision{CodePINTriesExceeded, "pin tries exceeded"}
		case cred.PIN != CheckMatch:
			return Decision{CodeIncorrectPIN, "pin missing or incorrect"}
		}
	}
	if req.ProcessingCode == ProcessingWithdrawal && req.Channel != ChannelATM {
		return Decision{CodeNotPermitted, "cash withdrawal is only allowed at ATM"}
	}
	if !card.Controls.Allows(req.Channel) {
		return Decision{CodeNotPermitted, "channel " + string(req.Channel) + " is disabled for this card"}
	}
	if req.Currency != account.Currency {
		return Decision{CodeInvalidTxn, "currency " + req.Currency + " is not supported"}
	}
	if req.Amount > card.Limits.PerTransaction {
		return Decision{CodeExceedsLimit, "amount exceeds per-transaction limit"}
	}
	if spentToday+req.Amount > card.Limits.Daily {
		return Decision{CodeExceedsLimit, "amount exceeds remaining daily limit"}
	}
	if req.NeedsStepUp(p) && cred.StepUp != CheckMatch {
		if cred.StepUp == CheckMismatch {
			return Decision{CodeAuthenticationReqd, "payment authentication is invalid, expired or already used"}
		}
		return Decision{CodeAuthenticationReqd, "payment authentication required"}
	}
	if req.Amount > account.Available() {
		return Decision{CodeInsufficientFunds, "insufficient available balance"}
	}
	return Decision{CodeApproved, "approved"}
}

type AuthorizationStatus string

const (
	AuthAuthorized AuthorizationStatus = "AUTHORIZED"
	AuthConfirmed  AuthorizationStatus = "CONFIRMED"
	AuthCancelled  AuthorizationStatus = "CANCELLED"
	AuthDenied     AuthorizationStatus = "DENIED"
)

type Authorization struct {
	ID               string
	AccountID        string
	CardID           string
	ProcessingCode   ProcessingCode
	Status           AuthorizationStatus
	Amount           int64
	ConfirmedAmount  int64
	Currency         string
	Channel          Channel
	Merchant         string
	MCC              string
	Code             ResponseCode
	Reason           string
	AuthCode         string
	AuthenticationID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

func NewAuthorization(id string, card *Card, req AuthorizationRequest, d Decision, now time.Time) *Authorization {
	a := &Authorization{
		ID:               id,
		AccountID:        card.AccountID,
		CardID:           card.ID,
		ProcessingCode:   req.ProcessingCode,
		Status:           AuthDenied,
		Amount:           req.Amount,
		Currency:         req.Currency,
		Channel:          req.Channel,
		Merchant:         req.Merchant,
		MCC:              req.MCC,
		Code:             d.Code,
		Reason:           d.Reason,
		AuthenticationID: req.AuthenticationID,
		CreatedAt:        now,
		UpdatedAt:        now,
	}
	switch {
	case !d.Approved():
	case req.ProcessingCode.Debit():
		a.Status = AuthAuthorized
	default:
		a.Status = AuthConfirmed
		a.ConfirmedAmount = req.Amount
	}
	return a
}

func (a *Authorization) Approved() bool { return a.Code == CodeApproved }

func (a *Authorization) Confirm(amount int64, now time.Time) (release int64, err error) {
	if a.Status != AuthAuthorized {
		return 0, Conflict("authorization is %s, only an AUTHORIZED one can be confirmed", a.Status)
	}
	if amount == 0 {
		amount = a.Amount
	}
	if amount < 0 || amount > a.Amount {
		return 0, Invalid("confirmed amount must be between 1 and the authorized amount %d", a.Amount)
	}
	a.Status = AuthConfirmed
	a.ConfirmedAmount = amount
	a.UpdatedAt = now
	return a.Amount, nil
}

func (a *Authorization) Cancel(now time.Time) (release int64, err error) {
	if a.Status != AuthAuthorized {
		return 0, Conflict("authorization is %s, only an AUTHORIZED one can be cancelled", a.Status)
	}
	a.Status = AuthCancelled
	a.UpdatedAt = now
	return a.Amount, nil
}
