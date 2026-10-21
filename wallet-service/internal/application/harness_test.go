package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const testPin = "123456"

type harness struct {
	t     *testing.T
	ctx   context.Context
	db    *memDB
	bank  *fakeBank
	cache *memCache
	now   time.Time

	wallets  port.WalletService
	funding  port.FundingService
	transfer port.TransferService
	payments port.PaymentService
	cardPay  port.CardPaymentService
	settle   port.SettlementService
	history  port.HistoryService
	cards    port.CardService
	accounts port.BankAccountService
	requests port.PaymentRequestService

	issuer   model.Bank
	acquirer model.Bank
	merchant model.Merchant
}

func newHarness(t *testing.T) *harness {
	t.Helper()
	h := &harness{
		t:     t,
		ctx:   context.Background(),
		db:    newMemDB(),
		bank:  &fakeBank{},
		cache: &memCache{data: map[string]string{}},
		now:   time.Date(2026, 10, 21, 9, 0, 0, 0, time.UTC),
	}
	h.issuer = model.Bank{ID: uuid.New(), Name: "Issuer", Code: "ISS", Country: "VN"}
	h.acquirer = model.Bank{ID: uuid.New(), Name: "Acquirer", Code: "ACQ", Country: "VN"}
	h.merchant = model.Merchant{ID: uuid.New(), Name: "Coffee", MCC: "5814", AcquirerBankID: h.acquirer.ID, Country: "VN"}
	h.db.banks[h.issuer.ID] = h.issuer
	h.db.banks[h.acquirer.ID] = h.acquirer
	h.db.merchants[h.merchant.ID] = h.merchant

	d := Deps{
		Tx:              fakeTx{db: h.db},
		Wallets:         walletRepo{db: h.db},
		Transactions:    txRepo{db: h.db},
		Ledger:          ledgerRepo{db: h.db},
		Cards:           cardRepo{db: h.db},
		Banks:           bankRepo{db: h.db},
		Merchants:       merchantRepo{db: h.db},
		BankAccounts:    accountRepo{db: h.db},
		PaymentRequests: requestRepo{db: h.db},
		Batches:         batchRepo{db: h.db},
		Clearing:        clearingRepo{db: h.db},
		Cache:           h.cache,
		Pins:            plainPins{},
		BankGateway:     h.bank,
		Fees:            model.DefaultFeeSchedule(),
		PaymentRequest:  PaymentRequestPolicy{TTL: 24 * time.Hour},
		Clock:           func() time.Time { return h.now },
	}
	h.wallets = NewWalletService(d)
	h.funding = NewFundingService(d)
	h.transfer = NewTransferService(d)
	h.payments = NewPaymentService(d)
	h.cardPay = NewCardPaymentService(d)
	h.settle = NewSettlementService(d)
	h.history = NewHistoryService(d)
	h.cards = NewCardService(d)
	h.accounts = NewBankAccountService(d)
	h.requests = NewPaymentRequestService(d)
	return h
}

func (h *harness) advance(d time.Duration) { h.now = h.now.Add(d) }

func (h *harness) newWallet(userID uuid.UUID, currency model.Currency, balance string, withPin bool) *model.Wallet {
	h.t.Helper()
	w, err := h.wallets.Create(h.ctx, userID, currency)
	mustNoErr(h.t, err)
	if balance != "" && balance != "0" {
		_, err = h.wallets.Credit(h.ctx, port.CreditCommand{WalletID: w.ID, Amount: dec(balance), IdempotencyKey: "seed-" + w.ID.String()})
		mustNoErr(h.t, err)
	}
	if withPin {
		mustNoErr(h.t, h.wallets.SetPin(h.ctx, port.SetPinCommand{WalletID: w.ID, UserID: userID, NewPin: testPin}))
	}
	return h.reload(w.ID)
}

func (h *harness) reload(id uuid.UUID) *model.Wallet {
	h.t.Helper()
	w, err := walletRepo{db: h.db}.GetByID(h.ctx, id)
	mustNoErr(h.t, err)
	return w
}

func (h *harness) linkAccount(userID uuid.UUID, number string) *model.BankAccount {
	h.t.Helper()
	a, err := h.accounts.Link(h.ctx, port.LinkBankAccountCommand{UserID: userID, BankID: h.issuer.ID, AccountNumber: number, AccountName: "Test User"})
	mustNoErr(h.t, err)
	return a
}

func (h *harness) ledgerFor(walletID uuid.UUID) []model.LedgerEntry {
	out := []model.LedgerEntry{}
	for _, e := range h.db.ledger {
		if e.WalletID == walletID {
			out = append(out, e)
		}
	}
	return out
}

func dec(s string) decimal.Decimal { return decimal.RequireFromString(s) }

func mustNoErr(t *testing.T, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

func mustErr(t *testing.T, err, want error) {
	t.Helper()
	if !errors.Is(err, want) {
		t.Fatalf("expected error %v, got %v", want, err)
	}
}

func mustEqualDec(t *testing.T, label string, got decimal.Decimal, want string) {
	t.Helper()
	if !got.Equal(dec(want)) {
		t.Fatalf("%s: expected %s, got %s", label, want, got)
	}
}
