package application

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

func TestWalletPerCurrencyUniqueness(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	h.newWallet(alice, model.CurrencyUSD, "0", false)
	_, err := h.wallets.Create(h.ctx, alice, "usd")
	mustErr(t, err, model.ErrWalletExists)
	_, err = h.wallets.Create(h.ctx, alice, "VND")
	mustNoErr(t, err)
	_, err = h.wallets.Create(h.ctx, alice, "JPY")
	mustErr(t, err, model.ErrUnsupportedCurrency)

	list, err := h.wallets.ListByUser(h.ctx, alice)
	mustNoErr(t, err)
	if len(list) != 2 {
		t.Fatalf("expected two wallets, got %d", len(list))
	}
}

func TestCreditInvalidatesCachedWallet(t *testing.T) {
	h := newHarness(t)
	w := h.newWallet(uuid.New(), model.CurrencyUSD, "0", false)

	cached, err := h.wallets.Get(h.ctx, w.ID)
	mustNoErr(t, err)
	mustEqualDec(t, "initial", cached.Balance, "0")

	_, err = h.wallets.Credit(h.ctx, port.CreditCommand{WalletID: w.ID, Amount: dec("15"), Description: "promo"})
	mustNoErr(t, err)
	fresh, err := h.wallets.Get(h.ctx, w.ID)
	mustNoErr(t, err)
	mustEqualDec(t, "after credit", fresh.Balance, "15")

	_, err = h.wallets.Credit(h.ctx, port.CreditCommand{WalletID: w.ID, Amount: dec("0")})
	mustErr(t, err, model.ErrInvalidAmount)
}

func TestCloseWalletRequiresEmptyBalance(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "5", false)

	_, err := h.wallets.Close(h.ctx, w.ID, uuid.New())
	mustErr(t, err, model.ErrNotWalletOwner)
	_, err = h.wallets.Close(h.ctx, w.ID, alice)
	mustErr(t, err, model.ErrWalletNotEmpty)

	empty := h.newWallet(alice, model.CurrencyEUR, "0", false)
	closed, err := h.wallets.Close(h.ctx, empty.ID, alice)
	mustNoErr(t, err)
	if closed.Status != model.WalletStatusClosed {
		t.Fatalf("expected CLOSED, got %s", closed.Status)
	}
	_, err = h.wallets.Unfreeze(h.ctx, empty.ID)
	mustErr(t, err, model.ErrWalletClosed)
	_, err = h.wallets.Create(h.ctx, alice, model.CurrencyEUR)
	mustNoErr(t, err)
}

func TestStatementSummarizesLedger(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	w := h.newWallet(alice, model.CurrencyUSD, "100", true)
	h.newWallet(bob, model.CurrencyUSD, "0", false)

	h.advance(24 * time.Hour)
	from := h.now
	_, err := h.transfer.Transfer(h.ctx, port.TransferCommand{IdempotencyKey: "s1", UserID: alice, FromWalletID: w.ID, ToUserID: bob, Amount: dec("30"), Pin: testPin})
	mustNoErr(t, err)
	_, err = h.wallets.Credit(h.ctx, port.CreditCommand{WalletID: w.ID, Amount: dec("5")})
	mustNoErr(t, err)
	h.advance(time.Hour)

	st, err := h.history.Statement(h.ctx, w.ID, from, h.now)
	mustNoErr(t, err)
	mustEqualDec(t, "opening", st.OpeningBalance, "100")
	mustEqualDec(t, "closing", st.ClosingBalance, "75")
	mustEqualDec(t, "debits", st.TotalDebit, "30")
	mustEqualDec(t, "credits", st.TotalCredit, "5")
	if st.DebitCount != 1 || st.CreditCount != 1 {
		t.Fatalf("unexpected counts: %+v", st)
	}

	_, err = h.history.Statement(h.ctx, w.ID, h.now, from)
	if model.AsError(err).Kind != model.KindInvalid {
		t.Fatalf("expected invalid range error, got %v", err)
	}
}

func TestBankAccountDefaults(t *testing.T) {
	h := newHarness(t)
	alice := uuid.New()
	first := h.linkAccount(alice, "11112222")
	h.advance(time.Second)
	second := h.linkAccount(alice, "33334444")
	if !first.IsDefault || second.IsDefault {
		t.Fatalf("first linked account must be default")
	}
	if second.AccountNumber != "****4444" {
		t.Fatalf("account number must be masked, got %s", second.AccountNumber)
	}

	_, err := h.accounts.Link(h.ctx, port.LinkBankAccountCommand{UserID: alice, BankID: h.issuer.ID, AccountNumber: "1111 2222", AccountName: "dup"})
	mustErr(t, err, model.ErrBankAccountFound)

	_, err = h.accounts.SetDefault(h.ctx, uuid.New(), second.ID)
	mustErr(t, err, model.ErrBankAccountAbsent)
	_, err = h.accounts.SetDefault(h.ctx, alice, second.ID)
	mustNoErr(t, err)

	mustNoErr(t, h.accounts.Remove(h.ctx, alice, second.ID))
	list, err := h.accounts.List(h.ctx, alice)
	mustNoErr(t, err)
	if len(list) != 1 || list[0].ID != first.ID || !list[0].IsDefault {
		t.Fatalf("removing the default must promote the remaining account: %+v", list)
	}
}

func TestPaymentRequestLifecycle(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	aliceW := h.newWallet(alice, model.CurrencyUSD, "0", false)
	bobW := h.newWallet(bob, model.CurrencyUSD, "100", true)

	pr, err := h.requests.Create(h.ctx, port.CreatePaymentRequestCommand{RequesterUserID: alice, RequesterWalletID: aliceW.ID, PayerUserID: bob, Amount: dec("25"), Note: "movie"})
	mustNoErr(t, err)

	_, err = h.requests.Decline(h.ctx, pr.ID, alice)
	mustErr(t, err, model.ErrNotPaymentRequestParty)

	incoming, total, err := h.requests.List(h.ctx, model.PaymentRequestFilter{UserID: bob, Role: model.RoleIncoming})
	mustNoErr(t, err)
	if total != 1 || incoming[0].ID != pr.ID {
		t.Fatalf("bob should see the incoming request")
	}

	paid, err := h.requests.Pay(h.ctx, port.PayPaymentRequestCommand{RequestID: pr.ID, UserID: bob, WalletID: bobW.ID, Pin: testPin})
	mustNoErr(t, err)
	if paid.Status != model.PaymentRequestPaid || !paid.TransactionID.Valid {
		t.Fatalf("expected PAID with transaction, got %+v", paid)
	}
	mustEqualDec(t, "alice", h.reload(aliceW.ID).Balance, "25")
	mustEqualDec(t, "bob", h.reload(bobW.ID).Balance, "75")

	again, err := h.requests.Pay(h.ctx, port.PayPaymentRequestCommand{RequestID: pr.ID, UserID: bob, WalletID: bobW.ID, Pin: testPin})
	mustNoErr(t, err)
	if again.TransactionID != paid.TransactionID {
		t.Fatalf("paying twice must not create a second transfer")
	}
	mustEqualDec(t, "bob after replay", h.reload(bobW.ID).Balance, "75")

	_, err = h.requests.Cancel(h.ctx, pr.ID, alice)
	mustErr(t, err, model.ErrPaymentRequestState)

	_, err = h.requests.Create(h.ctx, port.CreatePaymentRequestCommand{RequesterUserID: alice, RequesterWalletID: aliceW.ID, PayerUserID: alice, Amount: dec("1")})
	mustErr(t, err, model.ErrPaymentRequestSelf)
}

func TestPaymentRequestExpires(t *testing.T) {
	h := newHarness(t)
	alice, bob := uuid.New(), uuid.New()
	aliceW := h.newWallet(alice, model.CurrencyUSD, "0", false)
	bobW := h.newWallet(bob, model.CurrencyUSD, "100", true)
	pr, err := h.requests.Create(h.ctx, port.CreatePaymentRequestCommand{RequesterUserID: alice, RequesterWalletID: aliceW.ID, PayerUserID: bob, Amount: dec("10")})
	mustNoErr(t, err)

	h.advance(25 * time.Hour)
	got, err := h.requests.Get(h.ctx, pr.ID)
	mustNoErr(t, err)
	if got.Status != model.PaymentRequestExpired {
		t.Fatalf("expected EXPIRED, got %s", got.Status)
	}
	_, err = h.requests.Pay(h.ctx, port.PayPaymentRequestCommand{RequestID: pr.ID, UserID: bob, WalletID: bobW.ID, Pin: testPin})
	mustErr(t, err, model.ErrPaymentRequestExpired)
	mustEqualDec(t, "bob untouched", h.reload(bobW.ID).Balance, "100")
}
