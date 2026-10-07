package domain

import (
	"strings"
	"time"
)

const (
	OTPLength      = 6
	maxMerchantLen = 100
	maxAmount      = 10_000_000_000
)

type Status string

const (
	StatusPending       Status = "PENDING"
	StatusAuthenticated Status = "AUTHENTICATED"
	StatusFailed        Status = "FAILED"
	StatusDeclined      Status = "DECLINED"
	StatusExpired       Status = "EXPIRED"
	StatusUsed          Status = "USED"
)

type Policy struct {
	OTPTTL       time.Duration
	MaxAttempts  int
	MaxResends   int
	ValidityTime time.Duration
}

type CardRef struct {
	CardID     string
	AccountID  string
	CustomerID string
	UserID     int64
	MaskedPAN  string
	Status     string
}

type InitiateRequest struct {
	PAN      string
	Expiry   string
	Amount   int64
	Currency string
	Merchant string
	MCC      string
}

func (r *InitiateRequest) Validate() error {
	r.PAN = strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(r.PAN))
	if len(r.PAN) < 12 || len(r.PAN) > 19 || !isDigits(r.PAN) {
		return Invalid("pan must be 12-19 digits")
	}
	r.Expiry = strings.TrimSpace(r.Expiry)
	if len(r.Expiry) != 5 || r.Expiry[2] != '/' || !isDigits(r.Expiry[:2]+r.Expiry[3:]) {
		return Invalid("expiry must be MM/YY")
	}
	if r.Amount <= 0 || r.Amount > maxAmount {
		return Invalid("amount must be between 1 and %d", int64(maxAmount))
	}
	r.Currency = strings.ToUpper(strings.TrimSpace(r.Currency))
	if len(r.Currency) != 3 {
		return Invalid("currency must be an ISO 4217 code")
	}
	r.Merchant = strings.TrimSpace(r.Merchant)
	if r.Merchant == "" || len(r.Merchant) > maxMerchantLen {
		return Invalid("merchant must be 1-%d characters", maxMerchantLen)
	}
	if r.MCC != "" && (len(r.MCC) != 4 || !isDigits(r.MCC)) {
		return Invalid("mcc must be 4 digits")
	}
	return nil
}

type Challenge struct {
	ID              string
	CardID          string
	AccountID       string
	CustomerID      string
	UserID          int64
	Username        string
	MaskedPAN       string
	Amount          int64
	Currency        string
	Merchant        string
	MCC             string
	OTPExpiresAt    time.Time
	Attempts        int
	MaxAttempts     int
	Resends         int
	Status          Status
	FailureReason   string
	ExpiresAt       time.Time
	AuthenticatedAt *time.Time
	ValidUntil      *time.Time
	UsedAt          *time.Time
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

func NewChallenge(id string, card CardRef, username string, req InitiateRequest, otpExpiresAt time.Time, p Policy, now time.Time) (*Challenge, error) {
	if card.Status != "NORMAL" {
		return nil, Conflict("card is %s", card.Status)
	}
	if card.UserID <= 0 || username == "" {
		return nil, Conflict("card has no cardholder enrolled for payment authentication")
	}
	expiresAt := now.Add(p.OTPTTL)
	return &Challenge{
		ID:           id,
		CardID:       card.CardID,
		AccountID:    card.AccountID,
		CustomerID:   card.CustomerID,
		UserID:       card.UserID,
		Username:     username,
		MaskedPAN:    card.MaskedPAN,
		Amount:       req.Amount,
		Currency:     req.Currency,
		Merchant:     req.Merchant,
		MCC:          req.MCC,
		OTPExpiresAt: earliest(otpExpiresAt, expiresAt),
		MaxAttempts:  p.MaxAttempts,
		Status:       StatusPending,
		ExpiresAt:    expiresAt,
		CreatedAt:    now,
		UpdatedAt:    now,
	}, nil
}

func (c *Challenge) Clone() *Challenge {
	cp := *c
	for _, p := range []**time.Time{&cp.AuthenticatedAt, &cp.ValidUntil, &cp.UsedAt} {
		if *p != nil {
			t := **p
			*p = &t
		}
	}
	return &cp
}

func (c *Challenge) Refresh(now time.Time) bool {
	switch {
	case c.Status == StatusPending && !now.Before(c.ExpiresAt):
		c.expire("otp expired", now)
		return true
	case c.Status == StatusAuthenticated && c.ValidUntil != nil && !now.Before(*c.ValidUntil):
		c.expire("authentication was not used in time", now)
		return true
	}
	return false
}

func (c *Challenge) Precheck(userID int64, now time.Time) error {
	if userID != c.UserID {
		return ErrForbidden
	}
	if c.Refresh(now) {
		return ErrExpired
	}
	if c.Status != StatusPending {
		return Conflict("payment authentication is %s", c.Status)
	}
	if !now.Before(c.OTPExpiresAt) {
		return ErrOTPExpired
	}
	return nil
}

func (c *Challenge) Verify(userID int64, otpMatches bool, now time.Time, p Policy) error {
	if err := c.Precheck(userID, now); err != nil {
		return err
	}
	if !otpMatches {
		c.Attempts++
		c.UpdatedAt = now
		if c.Attempts >= c.MaxAttempts {
			c.Status = StatusFailed
			c.FailureReason = "too many incorrect otp attempts"
		}
		return &OTPMismatchError{Remaining: c.MaxAttempts - c.Attempts}
	}
	valid := now.Add(p.ValidityTime)
	c.Status = StatusAuthenticated
	c.AuthenticatedAt = &now
	c.ValidUntil = &valid
	c.UpdatedAt = now
	return nil
}

func (c *Challenge) CanResend(userID int64, now time.Time, p Policy) error {
	if userID != c.UserID {
		return ErrForbidden
	}
	if c.Refresh(now) {
		return ErrExpired
	}
	if c.Status != StatusPending {
		return Conflict("payment authentication is %s", c.Status)
	}
	if c.Resends >= p.MaxResends {
		return Conflict("otp was already resent %d times", c.Resends)
	}
	return nil
}

func (c *Challenge) Resent(otpExpiresAt, now time.Time) {
	c.Resends++
	c.OTPExpiresAt = earliest(otpExpiresAt, c.ExpiresAt)
	c.UpdatedAt = now
}

func earliest(a, b time.Time) time.Time {
	if a.Before(b) {
		return a
	}
	return b
}

func (c *Challenge) Decline(userID int64, now time.Time) error {
	if userID != c.UserID {
		return ErrForbidden
	}
	if c.Refresh(now) {
		return ErrExpired
	}
	if c.Status != StatusPending {
		return Conflict("payment authentication is %s", c.Status)
	}
	c.Status = StatusDeclined
	c.FailureReason = "declined by the cardholder"
	c.UpdatedAt = now
	return nil
}

func (c *Challenge) Consume(cardID string, amount int64, currency string, now time.Time) error {
	if c.Refresh(now) {
		return ErrExpired
	}
	if c.Status != StatusAuthenticated {
		return Conflict("payment authentication is %s", c.Status)
	}
	if cardID != c.CardID {
		return Conflict("payment authentication was issued for another card")
	}
	if !strings.EqualFold(currency, c.Currency) {
		return Conflict("currency %s does not match the authenticated currency %s", currency, c.Currency)
	}
	if amount <= 0 || amount > c.Amount {
		return Conflict("amount %d exceeds the authenticated amount %d", amount, c.Amount)
	}
	c.Status = StatusUsed
	c.UsedAt = &now
	c.UpdatedAt = now
	return nil
}

func (c *Challenge) expire(reason string, now time.Time) {
	c.Status = StatusExpired
	c.FailureReason = reason
	c.UpdatedAt = now
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
