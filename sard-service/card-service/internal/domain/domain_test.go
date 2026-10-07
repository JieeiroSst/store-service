package domain

import (
	"crypto/rand"
	"strings"
	"testing"
	"time"
)

func TestGeneratePAN(t *testing.T) {
	seen := map[string]bool{}
	for i := 0; i < 2000; i++ {
		pan, err := GeneratePAN("730100", rand.Reader)
		if err != nil {
			t.Fatal(err)
		}
		if len(pan) != PANLength || !strings.HasPrefix(pan, "730100") || !ValidPAN(pan) {
			t.Fatalf("bad pan %q", pan)
		}
		seen[pan] = true
	}
	if len(seen) < 1990 {
		t.Fatalf("too many collisions: %d unique of 2000", len(seen))
	}
	if _, err := GeneratePAN("73", rand.Reader); !IsInvalid(err) {
		t.Fatalf("short bin: got %v", err)
	}
}

func TestLuhn(t *testing.T) {
	for _, pan := range []string{"4111111111111111", "5555555555554444", "4012888888881881"} {
		if !ValidPAN(pan) {
			t.Errorf("%s should be valid", pan)
		}
	}
	for _, pan := range []string{"4111111111111112", "411111111111111", "41111111111111a1", ""} {
		if ValidPAN(pan) {
			t.Errorf("%q should be invalid", pan)
		}
	}
	if got := MaskPAN("730100", "1234"); got != "730100******1234" {
		t.Fatalf("mask = %s", got)
	}
}

func TestExpiry(t *testing.T) {
	issued := time.Date(2026, 2, 28, 10, 0, 0, 0, time.UTC)
	e := NewExpiry(issued, 5)
	if e.String() != "02/31" || e.YYMM() != "3102" {
		t.Fatalf("expiry = %s / %s", e, e.YYMM())
	}
	if e.Expired(time.Date(2031, 2, 28, 23, 59, 0, 0, time.UTC)) {
		t.Fatal("card is valid through the last day of its month")
	}
	if !e.Expired(time.Date(2031, 3, 1, 0, 0, 0, 0, time.UTC)) {
		t.Fatal("card should expire on the first day of the next month")
	}
	p, err := ParseExpiry("12/29")
	if err != nil || p != (Expiry{Month: 12, Year: 2029}) {
		t.Fatalf("parse = %+v, %v", p, err)
	}
	for _, bad := range []string{"13/29", "1/29", "12-29", "aa/bb", ""} {
		if _, err := ParseExpiry(bad); err == nil {
			t.Errorf("%q should not parse", bad)
		}
	}
}

func TestNormalizeCardholderName(t *testing.T) {
	cases := map[string]string{
		"Nguyễn Văn An":        "NGUYEN VAN AN",
		"  trần   thị   đào  ": "TRAN THI DAO",
		"Đặng O'Neil-Lê":       "DANG O'NEIL-LE",
	}
	for in, want := range cases {
		got, err := NormalizeCardholderName(in)
		if err != nil || got != want {
			t.Errorf("%q => %q, %v; want %q", in, got, err, want)
		}
	}
	for _, bad := range []string{"A", "Nguyen Van An Nguyen Van An Nguyen", "Bob 2", "李小龙"} {
		if _, err := NormalizeCardholderName(bad); err == nil {
			t.Errorf("%q should be rejected", bad)
		}
	}
}

func TestValidatePIN(t *testing.T) {
	for _, ok := range []string{"2580", "135790", "9071"} {
		if err := ValidatePIN(ok); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	for _, bad := range []string{"1111", "1234", "9876", "123", "1234567", "12a4"} {
		if ValidatePIN(bad) == nil {
			t.Errorf("%s should be rejected", bad)
		}
	}
}

func TestCardLifecycle(t *testing.T) {
	now := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	c := &Card{Status: StatusPending, Expiry: NewExpiry(now, 5)}
	if err := c.Block("", now); !IsConflict(err) {
		t.Fatalf("block pending: %v", err)
	}
	if err := c.Activate(now); err != nil || c.Status != StatusNormal {
		t.Fatal(err)
	}
	if err := c.Activate(now); !IsConflict(err) {
		t.Fatalf("double activate: %v", err)
	}
	if err := c.Block("travel", now); err != nil || c.Status != StatusBlocked || c.StatusReason != "travel" {
		t.Fatalf("block: %v %s", err, c.Status)
	}
	if err := c.Unblock(now); err != nil || c.Status != StatusNormal {
		t.Fatalf("unblock: %v %s", err, c.Status)
	}
	if err := c.Report(StatusRobbed, "pickpocket", now); err != nil || c.Status != StatusRobbed {
		t.Fatalf("robbed: %v %s", err, c.Status)
	}
	if err := c.Cancel("", now); !IsConflict(err) {
		t.Fatalf("cancel terminal: %v", err)
	}
	if _, err := ParseReportReason("STOLEN"); !IsInvalid(err) {
		t.Fatalf("pismo uses ROBBED: %v", err)
	}
	until := now.Add(24 * time.Hour)
	temp := &Card{Status: StatusNormal, Expiry: NewExpiry(now, 1), ValidUntil: &until}
	if temp.Expired(now) || !temp.Expired(until) {
		t.Fatal("temporary card expires at valid_until")
	}
}

func TestAccount(t *testing.T) {
	now := time.Date(2026, 10, 7, 0, 0, 0, 0, time.UTC)
	catalog := NewCatalog(Programs)
	gold, _ := catalog.Get("CREDIT_GOLD")
	debit, _ := catalog.Get("DEBIT_CLASSIC")
	eligible := Customer{ID: "cus", UserID: 7, FullName: "An", Eligible: true}

	if _, err := OpenAccount("a", Customer{ID: "cus", Reasons: []string{"kyc is NONE"}}, gold, 0, now); !IsConflict(err) {
		t.Fatalf("ineligible customer: %v", err)
	}
	if _, err := OpenAccount("a", eligible, debit, 1000, now); !IsInvalid(err) {
		t.Fatalf("credit limit on debit: %v", err)
	}
	if _, err := OpenAccount("a", eligible, gold, gold.MaxCreditLimit+1, now); !IsInvalid(err) {
		t.Fatalf("over max credit: %v", err)
	}
	a, err := OpenAccount("a", eligible, gold, 0, now)
	if err != nil || a.CreditLimit != gold.DefaultCreditLimit || a.Available() != gold.DefaultCreditLimit || a.UserID != 7 {
		t.Fatalf("open: %v %+v", err, a)
	}
	a.Hold(10_000_000, now)
	a.Release(10_000_000, now)
	a.Post(-30_000_000, now)
	a.Hold(5_000_000, now)
	if a.Available() != 15_000_000 || a.Outstanding() != 30_000_000 {
		t.Fatalf("available %d outstanding %d", a.Available(), a.Outstanding())
	}
	if err := a.SetCreditLimit(20_000_000, gold, now); !IsConflict(err) {
		t.Fatalf("limit below usage: %v", err)
	}
	if err := a.Cancel("", now); !IsConflict(err) {
		t.Fatalf("cancel with holds and debt: %v", err)
	}
	a.Release(5_000_000, now)
	if err := a.ReceivePayment(30_000_000, now); err != nil || a.Balance != 0 {
		t.Fatalf("payment: %v %d", err, a.Balance)
	}
	if err := a.Block("fraud check", now); err != nil {
		t.Fatal(err)
	}
	if err := a.Unblock(now); err != nil {
		t.Fatal(err)
	}
	if err := a.Cancel("closed by customer", now); err != nil || a.Status != AccountCancelled {
		t.Fatalf("cancel: %v", err)
	}
	if err := a.ReceivePayment(1, now); !IsConflict(err) {
		t.Fatalf("payment on cancelled: %v", err)
	}
}

func TestLimitsAndControls(t *testing.T) {
	max := Limits{PerTransaction: 100, Daily: 200}
	for _, bad := range []Limits{{0, 10}, {50, 40}, {150, 200}, {100, 300}} {
		if bad.Validate(max) == nil {
			t.Errorf("%+v should be rejected", bad)
		}
	}
	if err := (Limits{100, 200}).Validate(max); err != nil {
		t.Fatal(err)
	}
	p, _ := NewCatalog(Programs).Get("DEBIT_CLASSIC")
	if c := p.ControlsFor(CardVirtual); c != (Controls{Ecommerce: true}) {
		t.Fatalf("virtual cards are ecommerce only: %+v", c)
	}
	if err := (Controls{ATM: true}).Within(p.ControlsFor(CardVirtual)); err == nil {
		t.Fatal("ATM must not be enabled on a virtual card")
	}
}

func TestDecide(t *testing.T) {
	now := time.Date(2026, 10, 7, 12, 0, 0, 0, time.UTC)
	expiry := NewExpiry(now, 3)
	program := Program{StepUpThreshold: 1000}
	base := func() (*Card, *Account) {
		return &Card{
				Status:   StatusNormal,
				Expiry:   expiry,
				Limits:   Limits{PerTransaction: 1500, Daily: 2000},
				Controls: Controls{POS: true, Ecommerce: true, ATM: true},
				PINHash:  "h",
			}, &Account{
				Status:   AccountNormal,
				Currency: "VND",
				Balance:  1800,
			}
	}
	req := AuthorizationRequest{Expiry: expiry, Amount: 500, Currency: "VND", Channel: ChannelPOS, ProcessingCode: ProcessingPurchase}
	match := Credentials{CVV: CheckMatch, PIN: CheckMatch, StepUp: CheckMatch}

	cases := []struct {
		name  string
		card  func(*Card, *Account)
		req   func(*AuthorizationRequest)
		cred  Credentials
		spent int64
		want  ResponseCode
	}{
		{"approved", nil, nil, match, 0, CodeApproved},
		{"account blocked", func(c *Card, a *Account) { a.Status = AccountBlocked }, nil, match, 0, CodeRestrictedCard},
		{"account cancelled", func(c *Card, a *Account) { a.Status = AccountCancelled }, nil, match, 0, CodeClosedCard},
		{"pending", func(c *Card, a *Account) { c.Status = StatusPending }, nil, match, 0, CodeCardNotActivated},
		{"blocked", func(c *Card, a *Account) { c.Status = StatusBlocked }, nil, match, 0, CodeRestrictedCard},
		{"lost", func(c *Card, a *Account) { c.Status = StatusLost }, nil, match, 0, CodeLostCard},
		{"robbed", func(c *Card, a *Account) { c.Status = StatusRobbed }, nil, match, 0, CodeStolenCard},
		{"damaged", func(c *Card, a *Account) { c.Status = StatusDamaged }, nil, match, 0, CodeClosedCard},
		{"wrong expiry", nil, func(r *AuthorizationRequest) { r.Expiry = Expiry{Month: 1, Year: 2040} }, match, 0, CodeExpiredCard},
		{"bad cvv", nil, nil, Credentials{CVV: CheckMismatch}, 0, CodeCVVMismatch},
		{"ecom without cvv", nil, func(r *AuthorizationRequest) { r.Channel = ChannelEcommerce }, Credentials{}, 0, CodeCVVMismatch},
		{"atm wrong pin", nil, func(r *AuthorizationRequest) { r.Channel = ChannelATM }, Credentials{PIN: CheckMismatch}, 0, CodeIncorrectPIN},
		{"pin locked", func(c *Card, a *Account) { c.PINFailures = MaxPINAttempts }, func(r *AuthorizationRequest) { r.Channel = ChannelATM }, match, 0, CodePINTriesExceeded},
		{"withdrawal at pos", nil, func(r *AuthorizationRequest) { r.ProcessingCode = ProcessingWithdrawal }, match, 0, CodeNotPermitted},
		{"withdrawal at atm", nil, func(r *AuthorizationRequest) {
			r.ProcessingCode, r.Channel = ProcessingWithdrawal, ChannelATM
		}, match, 0, CodeApproved},
		{"channel disabled", nil, func(r *AuthorizationRequest) { r.Channel = ChannelContactless }, match, 0, CodeNotPermitted},
		{"currency", nil, func(r *AuthorizationRequest) { r.Currency = "USD" }, match, 0, CodeInvalidTxn},
		{"per transaction", nil, func(r *AuthorizationRequest) { r.Amount = 1501 }, match, 0, CodeExceedsLimit},
		{"daily", nil, nil, match, 1501, CodeExceedsLimit},
		{"step-up missing", nil, func(r *AuthorizationRequest) { r.Channel, r.Amount = ChannelEcommerce, 1000 }, Credentials{CVV: CheckMatch}, 0, CodeAuthenticationReqd},
		{"step-up rejected", nil, func(r *AuthorizationRequest) { r.Channel, r.Amount = ChannelEcommerce, 1000 }, Credentials{CVV: CheckMatch, StepUp: CheckMismatch}, 0, CodeAuthenticationReqd},
		{"below step-up", nil, func(r *AuthorizationRequest) { r.Channel, r.Amount = ChannelEcommerce, 999 }, Credentials{CVV: CheckMatch}, 0, CodeApproved},
		{"insufficient funds", func(c *Card, a *Account) { a.Held = 1500 }, nil, match, 0, CodeInsufficientFunds},
		{"credit line counts", func(c *Card, a *Account) { a.Balance, a.CreditLimit = -100, 700 }, nil, match, 0, CodeApproved},
		{"refund skips card checks", func(c *Card, a *Account) { c.Status = StatusLost; a.Balance = 0 }, func(r *AuthorizationRequest) {
			r.ProcessingCode, r.Amount = ProcessingRefund, 1_000_000
		}, Credentials{}, 0, CodeApproved},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			c, a := base()
			if tc.card != nil {
				tc.card(c, a)
			}
			r := req
			if tc.req != nil {
				tc.req(&r)
			}
			if got := Decide(c, a, program, r, tc.cred, tc.spent, now); got.Code != tc.want {
				t.Fatalf("got %s (%s), want %s", got.Code, got.Reason, tc.want)
			}
		})
	}
}

func TestAuthorizationRequestValidate(t *testing.T) {
	pan, _ := GeneratePAN("730100", rand.Reader)
	ok := AuthorizationRequest{PAN: pan[:4] + " " + pan[4:], Amount: 10, Currency: "vnd", Channel: "ecom"}
	if err := ok.Validate(); err != nil {
		t.Fatal(err)
	}
	if ok.PAN != pan || ok.Currency != "VND" || ok.Channel != ChannelEcommerce || ok.ProcessingCode != ProcessingPurchase {
		t.Fatalf("not normalized: %+v", ok)
	}
	bad := []AuthorizationRequest{
		{PAN: "4111111111111112", Amount: 10, Currency: "VND", Channel: "POS"},
		{PAN: pan, Amount: 0, Currency: "VND", Channel: "POS"},
		{PAN: pan, Amount: 10, Currency: "VND", Channel: "CASH"},
		{PAN: pan, Amount: 10, Currency: "VND", Channel: "POS", CVV: "12"},
		{PAN: pan, Amount: 10, Currency: "VND", Channel: "POS", ProcessingCode: "28"},
		{PAN: pan, Amount: 10, Currency: "VND", Channel: "POS", MCC: "59"},
	}
	for i, r := range bad {
		if err := r.Validate(); !IsInvalid(err) {
			t.Errorf("case %d: got %v", i, err)
		}
	}
}

func TestAuthorizationSettlement(t *testing.T) {
	now := time.Now()
	card := &Card{ID: "c", AccountID: "a"}
	approved := Decision{CodeApproved, "approved"}
	a := NewAuthorization("x", card, AuthorizationRequest{Amount: 100, ProcessingCode: ProcessingPurchase}, approved, now)
	if a.Status != AuthAuthorized {
		t.Fatalf("purchase holds: %s", a.Status)
	}
	if _, err := a.Confirm(101, now); !IsInvalid(err) {
		t.Fatalf("over-confirm: %v", err)
	}
	release, err := a.Confirm(80, now)
	if err != nil || release != 100 || a.ConfirmedAmount != 80 || a.Status != AuthConfirmed {
		t.Fatalf("confirm: %v %d %+v", err, release, a)
	}
	if _, err := a.Cancel(now); !IsConflict(err) {
		t.Fatalf("cancel confirmed: %v", err)
	}
	refund := NewAuthorization("r", card, AuthorizationRequest{Amount: 50, ProcessingCode: ProcessingRefund}, approved, now)
	if refund.Status != AuthConfirmed || refund.ConfirmedAmount != 50 {
		t.Fatalf("refund posts at once: %+v", refund)
	}
	denied := NewAuthorization("d", card, AuthorizationRequest{Amount: 50}, Decision{CodeExceedsLimit, "x"}, now)
	if denied.Status != AuthDenied {
		t.Fatalf("denied: %s", denied.Status)
	}
}
