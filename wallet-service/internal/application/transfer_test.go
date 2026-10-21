package application

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func TestTransferMovesFundsAndWritesLedger(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	src := h.newWallet(alice, model.CurrencyUSD, "500", true)
	dst := h.newWallet(bob, model.CurrencyUSD, "0", false)

	out, err := h.transfer.Transfer(h.ctx, port.TransferCommand{
		IdempotencyKey: "t-1", UserID: alice, FromWalletID: src.ID, ToWalletID: dst.ID,
		Amount: dec("120.50"), Pin: testPin, Description: "dinner",
	})
	mustNoErr(t, err)

	if out.Type != model.TransactionTypeTransferOut || out.Status != model.TransactionStatusCompleted {
		t.Fatalf("unexpected outbound tx: %s %s", out.Type, out.Status)
	}
	mustEqualDec(t, "sender balance", h.reload(src.ID).Balance, "379.50")
	mustEqualDec(t, "receiver balance", h.reload(dst.ID).Balance, "120.50")

	in, err := txRepo{db: h.db}.GetByID(h.ctx, out.ReferenceID.UUID)
	mustNoErr(t, err)
	if in.Type != model.TransactionTypeTransferIn || in.ReferenceID.UUID != out.ID || in.CounterpartyWalletID.UUID != src.ID {
		t.Fatalf("inbound tx not linked: %+v", in)
	}

	srcLedger := h.ledgerFor(src.ID)
	last := srcLedger[len(srcLedger)-1]
	if last.Direction != model.EntryDebit || !last.BalanceAfter.Equal(dec("379.50")) {
		t.Fatalf("unexpected sender ledger entry: %+v", last)
	}
	dstLedger := h.ledgerFor(dst.ID)
	if len(dstLedger) != 1 || dstLedger[0].Direction != model.EntryCredit {
		t.Fatalf("unexpected receiver ledger: %+v", dstLedger)
	}
}

func TestTransferIsIdempotent(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	src := h.newWallet(alice, model.CurrencyUSD, "100", true)
	dst := h.newWallet(bob, model.CurrencyUSD, "0", false)
	cmd := port.TransferCommand{IdempotencyKey: "same-key", UserID: alice, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec("40"), Pin: testPin}

	first, err := h.transfer.Transfer(h.ctx, cmd)
	mustNoErr(t, err)
	second, err := h.transfer.Transfer(h.ctx, cmd)
	mustNoErr(t, err)
	if first.ID != second.ID {
		t.Fatalf("replay returned a different transaction")
	}
	mustEqualDec(t, "sender balance", h.reload(src.ID).Balance, "60")

	other := h.newWallet(alice, model.CurrencyEUR, "100", true)
	_, err = h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: "same-key", UserID: alice, FromWalletID: other.ID, ToUserID: bob, Amount: dec("1"), Pin: testPin})
	mustErr(t, err, model.ErrIdempotencyKeyReused)

	_, err = h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: "same-key", UserID: bob, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec("40"), Pin: testPin})
	mustErr(t, err, model.ErrNotWalletOwner)
}

func TestTransferResolvesRecipientByUserAndCurrency(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	src := h.newWallet(alice, model.CurrencyVND, "1000000", true)
	h.newWallet(bob, model.CurrencyUSD, "0", false)
	bobVND := h.newWallet(bob, model.CurrencyVND, "0", false)

	_, err := h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: "k", UserID: alice, FromWalletID: src.ID, ToUserID: bob, Amount: dec("250000"), Pin: testPin})
	mustNoErr(t, err)
	mustEqualDec(t, "bob VND", h.reload(bobVND.ID).Balance, "250000")
}

func TestTransferRejectsAndRollsBack(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	src := h.newWallet(alice, model.CurrencyUSD, "50", true)
	dst := h.newWallet(bob, model.CurrencyUSD, "0", false)
	eur := h.newWallet(bob, model.CurrencyEUR, "0", false)

	cases := []struct {
		name string
		cmd  port.TransferCommand
		want error
	}{
		{"insufficient funds", port.TransferCommand{IdempotencyKey: "a", UserID: alice, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec("51"), Pin: testPin}, model.ErrInsufficientFunds},
		{"currency mismatch", port.TransferCommand{IdempotencyKey: "b", UserID: alice, FromWalletID: src.ID, ToWalletID: eur.ID, Amount: dec("1"), Pin: testPin}, model.ErrCurrencyMismatch},
		{"same wallet", port.TransferCommand{IdempotencyKey: "c", UserID: alice, FromWalletID: src.ID, ToWalletID: src.ID, Amount: dec("1"), Pin: testPin}, model.ErrSameWallet},
		{"not owner", port.TransferCommand{IdempotencyKey: "d", UserID: bob, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec("1"), Pin: testPin}, model.ErrNotWalletOwner},
		{"bad precision", port.TransferCommand{IdempotencyKey: "e", UserID: alice, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec("1.001"), Pin: testPin}, model.ErrAmountPrecision},
		{"missing key", port.TransferCommand{UserID: alice, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec("1"), Pin: testPin}, model.ErrIdempotencyKeyMissing},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := h.transfer.Transfer(h.ctx, tc.cmd)
			mustErr(t, err, tc.want)
		})
	}

	_, err := h.wallets.Freeze(h.ctx, dst.ID)
	mustNoErr(t, err)
	_, err = h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: "f", UserID: alice, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec("10"), Pin: testPin})
	mustErr(t, err, model.ErrWalletInactive)

	mustEqualDec(t, "sender balance untouched", h.reload(src.ID).Balance, "50")
	mustEqualDec(t, "receiver balance untouched", h.reload(dst.ID).Balance, "0")
}

func TestPinIsRequiredAndLocksAfterRepeatedFailures(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	noPin := h.newWallet(alice, model.CurrencyEUR, "100", false)
	src := h.newWallet(alice, model.CurrencyUSD, "100", true)
	dst := h.newWallet(bob, model.CurrencyUSD, "0", false)
	h.newWallet(bob, model.CurrencyEUR, "0", false)

	_, err := h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: "np", UserID: alice, FromWalletID: noPin.ID, ToUserID: bob, Amount: dec("1"), Pin: testPin})
	mustErr(t, err, model.ErrPinNotSet)

	attempt := func(key, pin string) error {
		_, err := h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: key, UserID: alice, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec("1"), Pin: pin})
		return err
	}
	for i := 0; i < model.PinMaxAttempts-1; i++ {
		mustErr(t, attempt("bad", "000000"), model.ErrPinInvalid)
	}
	if got := h.reload(src.ID).PinFailedAttempts; got != model.PinMaxAttempts-1 {
		t.Fatalf("failed attempts must persist even though the operation failed, got %d", got)
	}
	mustErr(t, attempt("bad", "000000"), model.ErrPinLocked)
	mustErr(t, attempt("good", testPin), model.ErrPinLocked)

	h.advance(model.PinLockDuration + time.Minute)
	mustNoErr(t, attempt("good", testPin))
	if w := h.reload(src.ID); w.PinFailedAttempts != 0 || w.PinLockedUntil != nil {
		t.Fatalf("successful PIN must reset the lock state: %+v", w)
	}
}

func TestChangingPinRequiresCurrentPin(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "0", true)

	err := h.wallets.SetPin(h.ctx, port.SetPinCommand{WalletID: w.ID, UserID: alice, CurrentPin: "999999", NewPin: "654321"})
	mustErr(t, err, model.ErrPinInvalid)
	err = h.wallets.SetPin(h.ctx, port.SetPinCommand{WalletID: w.ID, UserID: alice, CurrentPin: testPin, NewPin: "12ab56"})
	mustErr(t, err, model.ErrPinFormat)
	mustNoErr(t, h.wallets.SetPin(h.ctx, port.SetPinCommand{WalletID: w.ID, UserID: alice, CurrentPin: testPin, NewPin: "654321"}))
	if !plainPins.Compare(plainPins{}, h.reload(w.ID).PinHash, "654321") {
		t.Fatalf("pin was not changed")
	}
}

func TestTierLimitsAreEnforcedPerTransactionAndDaily(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	src := h.newWallet(alice, model.CurrencyUSD, "20000", true)
	dst := h.newWallet(bob, model.CurrencyUSD, "0", false)
	send := func(key, amount string) error {
		_, err := h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: key, UserID: alice, FromWalletID: src.ID, ToWalletID: dst.ID, Amount: dec(amount), Pin: testPin})
		return err
	}

	mustErr(t, send("big", "2000.01"), model.ErrPerTransactionLimit)
	mustNoErr(t, send("d1", "2000"))
	mustNoErr(t, send("d2", "2000"))
	mustErr(t, send("d3", "1000.01"), model.ErrDailyLimit)

	usage, err := h.wallets.Limits(h.ctx, src.ID)
	mustNoErr(t, err)
	mustEqualDec(t, "used today", usage.UsedToday, "4000")
	mustEqualDec(t, "remaining", usage.RemainingToday, "1000")

	_, err = h.wallets.ChangeTier(h.ctx, src.ID, model.TierVerified)
	mustNoErr(t, err)
	mustNoErr(t, send("d3", "5000"))

	h.advance(24 * time.Hour)
	_, err = h.wallets.ChangeTier(h.ctx, src.ID, model.TierBasic)
	mustNoErr(t, err)
	mustNoErr(t, send("next-day", "2000"))
}
