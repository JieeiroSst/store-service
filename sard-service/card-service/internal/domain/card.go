package domain

import (
	"strings"
	"time"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

const (
	MaxPINAttempts        = 3
	MaxCardholderNameLen  = 26
	minCardholderNameLen  = 2
	maxIDLen              = 64
	maxStatusReasonLen    = 200
	cardholderNameAllowed = " .-'"
)

type Status string

const (
	StatusPending   Status = "PENDING"
	StatusNormal    Status = "NORMAL"
	StatusBlocked   Status = "BLOCKED"
	StatusCancelled Status = "CANCELLED"
	StatusLost      Status = "LOST"
	StatusRobbed    Status = "ROBBED"
	StatusDamaged   Status = "DAMAGED"
)

func (s Status) Terminal() bool {
	switch s {
	case StatusCancelled, StatusLost, StatusRobbed, StatusDamaged:
		return true
	}
	return false
}

func ParseReportReason(s string) (Status, error) {
	switch st := Status(s); st {
	case StatusLost, StatusRobbed, StatusDamaged:
		return st, nil
	}
	return "", Invalid("reason must be LOST, ROBBED or DAMAGED, got %q", s)
}

type Controls struct {
	POS         bool
	Contactless bool
	Ecommerce   bool
	ATM         bool
}

func (c Controls) Allows(ch Channel) bool {
	switch ch {
	case ChannelPOS:
		return c.POS
	case ChannelContactless:
		return c.Contactless
	case ChannelEcommerce:
		return c.Ecommerce
	case ChannelATM:
		return c.ATM
	}
	return false
}

func (c Controls) Within(supported Controls) error {
	for _, ch := range Channels {
		if c.Allows(ch) && !supported.Allows(ch) {
			return Invalid("channel %s is not supported by this card", ch)
		}
	}
	return nil
}

type Card struct {
	ID               string
	AccountID        string
	CustomerID       string
	ProgramCode      string
	Type             CardType
	CardholderName   string
	PANHash          string
	PANCipher        []byte
	BIN              string
	Last4            string
	Expiry           Expiry
	ValidUntil       *time.Time
	ServiceCode      string
	Status           Status
	StatusReason     string
	Limits           Limits
	Controls         Controls
	PINHash          string
	PINFailures      int
	ReplacesCardID   string
	ReplacedByCardID string
	CreatedAt        time.Time
	UpdatedAt        time.Time
	ActivatedAt      *time.Time
}

func (c *Card) Expired(now time.Time) bool {
	if c.ValidUntil != nil && !now.Before(*c.ValidUntil) {
		return true
	}
	return c.Expiry.Expired(now)
}

func (c *Card) MaskedPAN() string { return MaskPAN(c.BIN, c.Last4) }

func (c *Card) PINSet() bool { return c.PINHash != "" }

func (c *Card) PINLocked() bool { return c.PINFailures >= MaxPINAttempts }

func (c *Card) Clone() *Card {
	cp := *c
	cp.PANCipher = append([]byte(nil), c.PANCipher...)
	for _, p := range []**time.Time{&cp.ActivatedAt, &cp.ValidUntil} {
		if *p != nil {
			t := **p
			*p = &t
		}
	}
	return &cp
}

func (c *Card) Activate(now time.Time) error {
	if c.Status != StatusPending {
		return Conflict("card is %s, only a PENDING card can be activated", c.Status)
	}
	if c.Expired(now) {
		return Conflict("card expired on %s", c.Expiry)
	}
	c.Status = StatusNormal
	c.StatusReason = ""
	c.ActivatedAt = &now
	c.UpdatedAt = now
	return nil
}

func (c *Card) Block(reason string, now time.Time) error {
	if c.Status != StatusNormal {
		return Conflict("card is %s, only a NORMAL card can be blocked", c.Status)
	}
	return c.transition(StatusBlocked, reason, now)
}

func (c *Card) Unblock(now time.Time) error {
	if c.Status != StatusBlocked {
		return Conflict("card is %s, only a BLOCKED card can be unblocked", c.Status)
	}
	return c.transition(StatusNormal, "", now)
}

func (c *Card) Report(reason Status, note string, now time.Time) error {
	if c.Status.Terminal() {
		return Conflict("card is already %s", c.Status)
	}
	return c.transition(reason, note, now)
}

func (c *Card) Cancel(reason string, now time.Time) error {
	if c.Status.Terminal() {
		return Conflict("card is already %s", c.Status)
	}
	return c.transition(StatusCancelled, reason, now)
}

func (c *Card) SetLimits(l Limits, max Limits, now time.Time) error {
	if c.Status.Terminal() {
		return Conflict("card is %s", c.Status)
	}
	if err := l.Validate(max); err != nil {
		return err
	}
	c.Limits = l
	c.UpdatedAt = now
	return nil
}

func (c *Card) SetControls(ctl Controls, supported Controls, now time.Time) error {
	if c.Status.Terminal() {
		return Conflict("card is %s", c.Status)
	}
	if err := ctl.Within(supported); err != nil {
		return err
	}
	c.Controls = ctl
	c.UpdatedAt = now
	return nil
}

func (c *Card) SetPIN(hash string, now time.Time) error {
	if c.Status.Terminal() {
		return Conflict("card is %s", c.Status)
	}
	c.PINHash = hash
	c.PINFailures = 0
	c.UpdatedAt = now
	return nil
}

func (c *Card) RecordPINFailure(now time.Time) {
	c.PINFailures++
	c.UpdatedAt = now
}

func (c *Card) RecordPINSuccess(now time.Time) {
	if c.PINFailures != 0 {
		c.PINFailures = 0
		c.UpdatedAt = now
	}
}

func (c *Card) transition(to Status, reason string, now time.Time) error {
	if len(reason) > maxStatusReasonLen {
		return Invalid("reason must be at most %d characters", maxStatusReasonLen)
	}
	c.Status = to
	c.StatusReason = strings.TrimSpace(reason)
	c.UpdatedAt = now
	return nil
}

func ValidateID(name, id string) error {
	if id == "" || len(id) > maxIDLen {
		return Invalid("%s must be 1-%d characters", name, maxIDLen)
	}
	for _, r := range id {
		if !(r == '-' || r == '_' || r < unicode.MaxASCII && (unicode.IsLetter(r) || unicode.IsDigit(r))) {
			return Invalid("%s may contain only letters, digits, '-' and '_'", name)
		}
	}
	return nil
}

func NormalizeCardholderName(name string) (string, error) {
	var b strings.Builder
	for _, r := range norm.NFD.String(name) {
		switch {
		case unicode.Is(unicode.Mn, r):
			continue
		case r == 'đ' || r == 'Đ':
			r = 'D'
		}
		b.WriteRune(unicode.ToUpper(r))
	}
	out := strings.Join(strings.Fields(b.String()), " ")
	if len(out) < minCardholderNameLen || len(out) > MaxCardholderNameLen {
		return "", Invalid("cardholder_name must be %d-%d characters after normalization", minCardholderNameLen, MaxCardholderNameLen)
	}
	for _, r := range out {
		if !(r >= 'A' && r <= 'Z') && !strings.ContainsRune(cardholderNameAllowed, r) {
			return "", Invalid("cardholder_name contains unsupported character %q", r)
		}
	}
	return out, nil
}

func ValidatePIN(pin string) error {
	if len(pin) < 4 || len(pin) > 6 || !isDigits(pin) {
		return Invalid("pin must be 4-6 digits")
	}
	same, asc, desc := true, true, true
	for i := 1; i < len(pin); i++ {
		d := int(pin[i]) - int(pin[i-1])
		same = same && d == 0
		asc = asc && d == 1
		desc = desc && d == -1
	}
	if same || asc || desc {
		return Invalid("pin is too easy to guess")
	}
	return nil
}
