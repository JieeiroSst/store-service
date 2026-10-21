package application

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func TestTopUpCreditsAfterBankCollection(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "0", false)
	acct := h.linkAccount(alice, "0123456789")
	cmd := port.TopUpCommand{IdempotencyKey: "top-1", UserID: alice, WalletID: w.ID, BankAccountID: acct.ID, Amount: dec("250")}

	tx, err := h.funding.TopUp(h.ctx, cmd)
	mustNoErr(t, err)
	if tx.Status != model.TransactionStatusCompleted || tx.ExternalRef == "" {
		t.Fatalf("unexpected top-up tx: %+v", tx)
	}
	mustEqualDec(t, "balance", h.reload(w.ID).Balance, "250")

	again, err := h.funding.TopUp(h.ctx, cmd)
	mustNoErr(t, err)
	if again.ID != tx.ID || len(h.bank.collects) != 1 {
		t.Fatalf("replay must not collect twice (collects=%d)", len(h.bank.collects))
	}
	mustEqualDec(t, "balance after replay", h.reload(w.ID).Balance, "250")
}

func TestTopUpBankFailureMarksTransactionFailed(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "0", false)
	acct := h.linkAccount(alice, "0123456789")
	h.bank.collectErr = errors.New("insufficient bank balance")

	_, err := h.funding.TopUp(h.ctx, port.TopUpCommand{IdempotencyKey: "top-fail", UserID: alice, WalletID: w.ID, BankAccountID: acct.ID, Amount: dec("10")})
	mustErr(t, err, model.ErrBankDeclined)

	stored, err := txRepo{db: h.db}.GetByIdempotencyKey(h.ctx, "top-fail")
	mustNoErr(t, err)
	if stored.Status != model.TransactionStatusFailed || stored.FailureReason == "" {
		t.Fatalf("expected FAILED top-up, got %+v", stored)
	}
	mustEqualDec(t, "balance", h.reload(w.ID).Balance, "0")
}

func TestTopUpRejectsSomeoneElsesBankAccount(t *testing.T) {
	h := newHarness(t)
	alice, mallory := uuid.New(), uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "0", false)
	foreign := h.linkAccount(mallory, "99887766")

	_, err := h.funding.TopUp(h.ctx, port.TopUpCommand{IdempotencyKey: "x", UserID: alice, WalletID: w.ID, BankAccountID: foreign.ID, Amount: dec("10")})
	mustErr(t, err, model.ErrBankAccountAbsent)
}

func TestWithdrawDebitsAmountPlusFee(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "1000", true)
	acct := h.linkAccount(alice, "0123456789")

	tx, err := h.funding.Withdraw(h.ctx, port.WithdrawCommand{IdempotencyKey: "wd-1", UserID: alice, WalletID: w.ID, BankAccountID: acct.ID, Amount: dec("400"), Pin: testPin})
	mustNoErr(t, err)
	mustEqualDec(t, "fee", tx.Fee, "2")
	if tx.Status != model.TransactionStatusCompleted {
		t.Fatalf("expected completed withdrawal, got %s", tx.Status)
	}
	mustEqualDec(t, "balance", h.reload(w.ID).Balance, "598")
	mustEqualDec(t, "payout amount excludes fee", h.bank.payouts[0].Amount, "400")
}

func TestWithdrawMinimumFee(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "100", true)
	acct := h.linkAccount(alice, "0123456789")

	tx, err := h.funding.Withdraw(h.ctx, port.WithdrawCommand{IdempotencyKey: "wd-min", UserID: alice, WalletID: w.ID, BankAccountID: acct.ID, Amount: dec("10"), Pin: testPin})
	mustNoErr(t, err)
	mustEqualDec(t, "minimum fee", tx.Fee, "0.5")
}

func TestWithdrawPayoutFailureIsReversed(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "1000", true)
	acct := h.linkAccount(alice, "0123456789")
	h.bank.payoutErr = errors.New("bank offline")

	_, err := h.funding.Withdraw(h.ctx, port.WithdrawCommand{IdempotencyKey: "wd-fail", UserID: alice, WalletID: w.ID, BankAccountID: acct.ID, Amount: dec("400"), Pin: testPin})
	mustErr(t, err, model.ErrBankDeclined)

	mustEqualDec(t, "balance restored", h.reload(w.ID).Balance, "1000")
	stored, err := txRepo{db: h.db}.GetByIdempotencyKey(h.ctx, "wd-fail")
	mustNoErr(t, err)
	if stored.Status != model.TransactionStatusFailed {
		t.Fatalf("expected FAILED withdrawal, got %s", stored.Status)
	}
	entries := h.ledgerFor(w.ID)
	last, prev := entries[len(entries)-1], entries[len(entries)-2]
	if prev.Direction != model.EntryDebit || last.Direction != model.EntryCredit || !last.Amount.Equal(prev.Amount) {
		t.Fatalf("expected debit followed by reversing credit, got %+v / %+v", prev, last)
	}
}
