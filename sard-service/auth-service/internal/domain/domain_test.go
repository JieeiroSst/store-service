package domain

import (
	"errors"
	"testing"
	"time"
)

var (
	now    = time.Date(2026, 10, 7, 9, 0, 0, 0, time.UTC)
	policy = Policy{OTPTTL: 5 * time.Minute, MaxAttempts: 3, MaxResends: 2, ValidityTime: 10 * time.Minute}
	card   = CardRef{CardID: "card-1", AccountID: "acc-1", CustomerID: "cus-1", UserID: 7, MaskedPAN: "730100******1234", Status: "NORMAL"}
	req    = InitiateRequest{PAN: "7301001234561234", Expiry: "10/31", Amount: 2_000_000, Currency: "VND", Merchant: "Tiki"}
)

func newChallenge(t *testing.T) *Challenge {
	t.Helper()
	c, err := NewChallenge("ch-1", card, "anguyen", req, now.Add(time.Minute), policy, now)
	if err != nil {
		t.Fatal(err)
	}
	return c
}

func TestNewChallengeRules(t *testing.T) {
	blocked := card
	blocked.Status = "BLOCKED"
	if _, err := NewChallenge("x", blocked, "u", req, now.Add(time.Minute), policy, now); !IsConflict(err) {
		t.Fatalf("blocked card: %v", err)
	}
	orphan := card
	orphan.UserID = 0
	if _, err := NewChallenge("x", orphan, "u", req, now.Add(time.Minute), policy, now); !IsConflict(err) {
		t.Fatalf("no cardholder: %v", err)
	}
	c := newChallenge(t)
	if c.Status != StatusPending || !c.ExpiresAt.Equal(now.Add(5*time.Minute)) || !c.OTPExpiresAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("%+v", c)
	}
	if _, err := NewChallenge("x", card, "", req, now.Add(time.Minute), policy, now); !IsConflict(err) {
		t.Fatalf("no username: %v", err)
	}
	long, _ := NewChallenge("x", card, "u", req, now.Add(time.Hour), policy, now)
	if !long.OTPExpiresAt.Equal(long.ExpiresAt) {
		t.Fatal("otp expiry is capped by the challenge expiry")
	}
}

func TestVerifyAndConsume(t *testing.T) {
	c := newChallenge(t)
	if err := c.Verify(8, true, now, policy); !errors.Is(err, ErrForbidden) {
		t.Fatalf("other user: %v", err)
	}
	err := c.Verify(7, false, now, policy)
	var mm *OTPMismatchError
	if !errors.As(err, &mm) || mm.Remaining != 2 || c.Status != StatusPending {
		t.Fatalf("wrong otp: %v %+v", err, c)
	}
	if err := c.Verify(7, true, now.Add(30*time.Second), policy); err != nil || c.Status != StatusAuthenticated {
		t.Fatalf("verify: %v %s", err, c.Status)
	}
	if err := c.Verify(7, true, now, policy); !IsConflict(err) {
		t.Fatalf("verify twice: %v", err)
	}
	later := now.Add(2 * time.Minute)
	for name, tc := range map[string]struct {
		card     string
		amount   int64
		currency string
	}{
		"other card":  {"card-2", 1, "VND"},
		"over amount": {"card-1", 2_000_001, "VND"},
		"currency":    {"card-1", 1, "USD"},
		"zero amount": {"card-1", 0, "VND"},
	} {
		if err := c.Consume(tc.card, tc.amount, tc.currency, later); !IsConflict(err) {
			t.Errorf("%s: %v", name, err)
		}
	}
	if err := c.Consume("card-1", 1_500_000, "vnd", later); err != nil || c.Status != StatusUsed {
		t.Fatalf("consume: %v %s", err, c.Status)
	}
	if err := c.Consume("card-1", 1, "VND", later); !IsConflict(err) {
		t.Fatalf("consume twice: %v", err)
	}
}

func TestLockoutAndExpiry(t *testing.T) {
	c := newChallenge(t)
	for i := 0; i < 3; i++ {
		_ = c.Verify(7, false, now, policy)
	}
	if c.Status != StatusFailed {
		t.Fatalf("after 3 misses: %s", c.Status)
	}
	if err := c.Verify(7, true, now, policy); !IsConflict(err) {
		t.Fatalf("verify failed challenge: %v", err)
	}

	c = newChallenge(t)
	if err := c.Verify(7, true, now.Add(time.Minute), policy); !errors.Is(err, ErrOTPExpired) || c.Status != StatusPending || c.Attempts != 0 {
		t.Fatalf("expired otp keeps the challenge pending: %v %+v", err, c)
	}
	if err := c.Verify(7, true, now.Add(5*time.Minute), policy); !errors.Is(err, ErrExpired) || c.Status != StatusExpired {
		t.Fatalf("expired otp: %v %s", err, c.Status)
	}

	c = newChallenge(t)
	_ = c.Verify(7, true, now, policy)
	if err := c.Consume("card-1", 1, "VND", now.Add(10*time.Minute)); !errors.Is(err, ErrExpired) || c.Status != StatusExpired {
		t.Fatalf("authentication not used in time: %v %s", err, c.Status)
	}

	c = newChallenge(t)
	if err := c.Decline(7, now); err != nil || c.Status != StatusDeclined {
		t.Fatalf("decline: %v %s", err, c.Status)
	}
}

func TestInitiateValidate(t *testing.T) {
	ok := InitiateRequest{PAN: "7301 0012 3456 1234", Expiry: "10/31", Amount: 1, Currency: "vnd", Merchant: " Tiki "}
	if err := ok.Validate(); err != nil || ok.PAN != "7301001234561234" || ok.Currency != "VND" || ok.Merchant != "Tiki" {
		t.Fatalf("%v %+v", err, ok)
	}
	for i, bad := range []InitiateRequest{
		{PAN: "12", Expiry: "10/31", Amount: 1, Currency: "VND", Merchant: "m"},
		{PAN: "7301001234561234", Expiry: "1031", Amount: 1, Currency: "VND", Merchant: "m"},
		{PAN: "7301001234561234", Expiry: "10/31", Amount: 0, Currency: "VND", Merchant: "m"},
		{PAN: "7301001234561234", Expiry: "10/31", Amount: 1, Currency: "VND", Merchant: ""},
		{PAN: "7301001234561234", Expiry: "10/31", Amount: 1, Currency: "VND", Merchant: "m", MCC: "12"},
	} {
		if err := bad.Validate(); !IsInvalid(err) {
			t.Errorf("case %d: %v", i, err)
		}
	}
}

func TestResend(t *testing.T) {
	c := newChallenge(t)
	if err := c.CanResend(8, now, policy); !errors.Is(err, ErrForbidden) {
		t.Fatalf("other user: %v", err)
	}
	for i := 0; i < policy.MaxResends; i++ {
		if err := c.CanResend(7, now, policy); err != nil {
			t.Fatal(err)
		}
		c.Resent(now.Add(2*time.Minute), now)
	}
	if c.Resends != 2 || !c.OTPExpiresAt.Equal(now.Add(2*time.Minute)) {
		t.Fatalf("%+v", c)
	}
	if err := c.CanResend(7, now, policy); !IsConflict(err) {
		t.Fatalf("too many resends: %v", err)
	}
	if err := c.Verify(7, true, now.Add(90*time.Second), policy); err != nil {
		t.Fatalf("the resent otp is valid: %v", err)
	}
}
