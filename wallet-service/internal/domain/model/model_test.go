package model

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

func d(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func TestValidateAmountRespectsCurrencyPrecision(t *testing.T) {
	cases := []struct {
		amount   string
		currency Currency
		want     error
	}{
		{"10.25", CurrencyUSD, nil},
		{"10.255", CurrencyUSD, ErrAmountPrecision},
		{"10000", CurrencyVND, nil},
		{"10000.5", CurrencyVND, ErrAmountPrecision},
		{"0", CurrencyUSD, ErrInvalidAmount},
		{"-1", CurrencyUSD, ErrInvalidAmount},
		{"1", Currency("JPY"), ErrUnsupportedCurrency},
	}
	for _, tc := range cases {
		if err := ValidateAmount(d(tc.amount), tc.currency); !errors.Is(err, tc.want) {
			t.Errorf("%s %s: expected %v, got %v", tc.amount, tc.currency, tc.want, err)
		}
	}
}

func TestLimitsScaleWithCurrency(t *testing.T) {
	usd := LimitsFor(TierBasic, CurrencyUSD)
	vnd := LimitsFor(TierBasic, CurrencyVND)
	if !vnd.Daily.Equal(usd.Daily.Mul(decimal.NewFromInt(25000))) {
		t.Fatalf("VND limits must be scaled from USD: %s vs %s", vnd.Daily, usd.Daily)
	}
	if err := usd.Check(d("1"), usd.Daily); !errors.Is(err, ErrDailyLimit) {
		t.Fatalf("expected daily limit, got %v", err)
	}
}

func TestWalletBalanceOperations(t *testing.T) {
	now := time.Now()
	w, err := NewWallet(uuid.New(), CurrencyUSD, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := w.Credit(d("100"), now); err != nil {
		t.Fatal(err)
	}
	if err := w.Hold(d("60"), now); err != nil {
		t.Fatal(err)
	}
	if err := w.Debit(d("41"), now); !errors.Is(err, ErrInsufficientFunds) {
		t.Fatalf("debit must respect held funds, got %v", err)
	}
	if err := w.CaptureHold(d("60"), now); err != nil {
		t.Fatal(err)
	}
	if !w.Balance.Equal(d("40")) || !w.FrozenAmount.IsZero() {
		t.Fatalf("unexpected balances %s / %s", w.Balance, w.FrozenAmount)
	}
	if err := w.Close(now); !errors.Is(err, ErrWalletNotEmpty) {
		t.Fatalf("expected not empty, got %v", err)
	}
}

func TestPinLockout(t *testing.T) {
	now := time.Now()
	w := &Wallet{}
	for i := 0; i < PinMaxAttempts-1; i++ {
		if w.RegisterPinFailure(now) {
			t.Fatalf("locked too early at attempt %d", i+1)
		}
	}
	if !w.RegisterPinFailure(now) || !w.PinLocked(now) {
		t.Fatalf("expected lock after %d failures", PinMaxAttempts)
	}
	if w.PinLocked(now.Add(PinLockDuration)) {
		t.Fatalf("lock must expire")
	}
	if ValidatePinFormat("12345") == nil || ValidatePinFormat("abcdef") == nil || ValidatePinFormat("123456") != nil {
		t.Fatalf("pin format validation is wrong")
	}
}

func TestGeneratedCardNumbersPassLuhn(t *testing.T) {
	for _, network := range []CardNetwork{CardNetworkVisa, CardNetworkMastercard, CardNetworkAmex} {
		for i := 0; i < 50; i++ {
			n, err := GenerateCardNumber(network)
			if err != nil {
				t.Fatal(err)
			}
			if !LuhnValid(n) || len(n) != cardFormats[network].length {
				t.Fatalf("%s: invalid number %s", network, n)
			}
		}
	}
	if LuhnValid("4111111111111112") || !LuhnValid("4111111111111111") {
		t.Fatalf("luhn check is wrong")
	}
}

func TestCardExpiryAndUnblock(t *testing.T) {
	now := time.Date(2026, 10, 21, 0, 0, 0, 0, time.UTC)
	c, err := NewCard(uuid.New(), uuid.New(), CardNetworkVisa, "alice", now)
	if err != nil {
		t.Fatal(err)
	}
	if c.EnsureUsable(now) != nil {
		t.Fatalf("new card must be usable")
	}
	if err := c.EnsureUsable(time.Date(2029, 11, 1, 0, 0, 0, 0, time.UTC)); !errors.Is(err, ErrCardExpired) {
		t.Fatalf("expected expiry, got %v", err)
	}
	if err := c.Unblock(now); !errors.Is(err, ErrCardNotActive) {
		t.Fatalf("unblocking an active card must fail, got %v", err)
	}
}

func TestFees(t *testing.T) {
	f := DefaultFeeSchedule()
	if got := f.Withdrawal(d("1000"), CurrencyUSD); !got.Equal(d("5")) {
		t.Fatalf("withdrawal fee: %s", got)
	}
	if got := f.Withdrawal(d("100000"), CurrencyVND); !got.Equal(d("12500")) {
		t.Fatalf("VND minimum fee: %s", got)
	}
	if got := f.CardInterchange(d("99.99"), CardNetworkAmex, CurrencyUSD); !got.Equal(d("2.5")) {
		t.Fatalf("amex fee: %s", got)
	}
	if got := f.Transfer(d("100"), CurrencyUSD); !got.IsZero() {
		t.Fatalf("transfers are free by default: %s", got)
	}
}

func TestNetClearingGroupsByBankPairAndNetwork(t *testing.T) {
	issuer, acquirer := uuid.New(), uuid.New()
	tx := func(amount, fee string, network CardNetwork) *Transaction {
		return &Transaction{Amount: d(amount), Fee: d(fee), CardNetwork: network, IssuerBankID: NullID(issuer), AcquirerBankID: NullID(acquirer)}
	}
	records := NetClearing(uuid.New(), []*Transaction{
		tx("100", "1.5", CardNetworkVisa),
		tx("50", "0.75", CardNetworkVisa),
		tx("10", "0.16", CardNetworkMastercard),
	}, time.Now())
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}
	if records[0].CardNetwork != CardNetworkVisa || !records[0].NetAmount.Equal(d("147.75")) {
		t.Fatalf("unexpected visa record: %+v", records[0])
	}
}

func TestRefundAccounting(t *testing.T) {
	now := time.Now()
	tx := &Transaction{Type: TransactionTypePayment, Status: TransactionStatusCompleted, Amount: d("10"), RefundedAmount: decimal.Zero}
	if _, err := tx.ApplyRefund(d("4"), now); err != nil {
		t.Fatal(err)
	}
	if tx.Status != TransactionStatusPartiallyRefunded || !tx.RefundableAmount().Equal(d("6")) {
		t.Fatalf("unexpected state %s %s", tx.Status, tx.RefundableAmount())
	}
	got, err := tx.ApplyRefund(decimal.Zero, now)
	if err != nil || !got.Equal(d("6")) || tx.Status != TransactionStatusRefunded {
		t.Fatalf("full remaining refund failed: %v %s %s", err, got, tx.Status)
	}
}

func TestMetadataRoundTrip(t *testing.T) {
	m := Metadata{"order_id": "A1"}
	v, err := m.Value()
	if err != nil {
		t.Fatal(err)
	}
	var back Metadata
	if err := back.Scan([]byte(v.(string))); err != nil || back["order_id"] != "A1" {
		t.Fatalf("round trip failed: %v %v", err, back)
	}
	empty, _ := Metadata(nil).Value()
	if empty != nil {
		t.Fatalf("empty metadata must be NULL")
	}
}
