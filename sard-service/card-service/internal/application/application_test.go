package application

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/crypto"
	"github.com/JIeeiroSst/card-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/card-service/internal/domain"
)

type fakeClock struct {
	mu  sync.Mutex
	now time.Time
}

func (c *fakeClock) Now() time.Time {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.now
}

func (c *fakeClock) Advance(d time.Duration) {
	c.mu.Lock()
	c.now = c.now.Add(d)
	c.mu.Unlock()
}

type nopMetrics struct{}

func (nopMetrics) ObserveIssued(string, domain.CardType)            {}
func (nopMetrics) ObserveLifecycle(string)                          {}
func (nopMetrics) ObserveAuthorization(string, domain.ResponseCode) {}

type customers map[string]domain.Customer

func (c customers) GetCustomer(_ context.Context, id string) (domain.Customer, error) {
	cu, ok := c[id]
	if !ok {
		return domain.Customer{}, domain.Invalid("customer %s does not exist", id)
	}
	return cu, nil
}

type authenticator struct {
	mu   sync.Mutex
	ok   map[string]bool
	seen []string
}

func (a *authenticator) Enabled() bool { return true }

func (a *authenticator) Consume(_ context.Context, id, _ string, _ int64, _ string) error {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.seen = append(a.seen, id)
	if !a.ok[id] {
		return domain.Conflict("payment authentication rejected")
	}
	delete(a.ok, id)
	return nil
}

type env struct {
	accounts *AccountService
	cards    *CardService
	auths    *AuthorizationService
	authn    *authenticator
	clock    *fakeClock
}

func newEnv(t *testing.T) env {
	t.Helper()
	keys := crypto.DeriveKeys([]byte("application-test-master-key-0123456789"))
	vault, err := crypto.NewVault(keys)
	if err != nil {
		t.Fatal(err)
	}
	sec := crypto.NewSecurity(keys)
	store := memory.NewStore()
	accounts, cards, auths, txns := memory.NewAccounts(store), memory.NewCards(store), memory.NewAuthorizations(store), memory.NewTransactions(store)
	locker := memory.NewLocker()
	loc, _ := time.LoadLocation("Asia/Ho_Chi_Minh")
	clock := &fakeClock{now: time.Date(2026, 10, 7, 3, 0, 0, 0, time.UTC)}
	catalog := domain.NewCatalog(domain.Programs)
	dir := customers{
		"cus-1": {ID: "cus-1", UserID: 7, FullName: "Nguyễn Văn An", Eligible: true},
		"cus-2": {ID: "cus-2", UserID: 8, FullName: "Tran Thi Binh", Eligible: true},
		"cus-x": {ID: "cus-x", UserID: 9, FullName: "No Kyc", Reasons: []string{"kyc is NONE"}},
	}
	authn := &authenticator{ok: map[string]bool{}}
	return env{
		accounts: NewAccountService(catalog, accounts, cards, txns, dir, locker, clock, nopMetrics{}),
		cards:    NewCardService(catalog, accounts, cards, locker, vault, sec, clock, nopMetrics{}),
		auths:    NewAuthorizationService(Settings{LimitsLocation: loc}, catalog, accounts, cards, auths, txns, authn, locker, vault, sec, clock, nopMetrics{}),
		authn:    authn,
		clock:    clock,
	}
}

func (e env) open(t *testing.T, customer, program string, deposit int64) *domain.Account {
	t.Helper()
	ctx := context.Background()
	a, err := e.accounts.Open(ctx, OpenAccountCommand{CustomerID: customer, ProgramCode: program})
	if err != nil {
		t.Fatal(err)
	}
	if deposit > 0 {
		if a, _, err = e.accounts.ReceivePayment(ctx, a.ID, deposit, "top up"); err != nil {
			t.Fatal(err)
		}
	}
	return a
}

func (e env) card(t *testing.T, accountID string, typ domain.CardType) *IssuedCard {
	t.Helper()
	ctx := context.Background()
	issued, err := e.cards.Issue(ctx, IssueCommand{AccountID: accountID, Type: typ})
	if err != nil {
		t.Fatal(err)
	}
	if issued.Card.Status == domain.StatusPending {
		if _, err := e.cards.Activate(ctx, issued.Card.ID, issued.CVV); err != nil {
			t.Fatal(err)
		}
	}
	return issued
}

func (e env) authorize(t *testing.T, c *IssuedCard, mut func(*domain.AuthorizationRequest)) *domain.Authorization {
	t.Helper()
	req := domain.AuthorizationRequest{
		PAN: c.PAN, Expiry: c.Card.Expiry, CVV: c.CVV, Amount: 1_000_000, Currency: "VND",
		Channel: domain.ChannelEcommerce, Merchant: "Tiki",
	}
	if mut != nil {
		mut(&req)
	}
	a, err := e.auths.Authorize(context.Background(), req)
	if err != nil {
		t.Fatal(err)
	}
	return a
}

func TestOpenAccount(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	a := e.open(t, "cus-1", "CREDIT_GOLD", 0)
	if a.UserID != 7 || a.HolderName != "Nguyễn Văn An" || a.Available() != 50_000_000 {
		t.Fatalf("%+v", a)
	}
	if _, err := e.accounts.Open(ctx, OpenAccountCommand{CustomerID: "cus-1", ProgramCode: "CREDIT_GOLD"}); !domain.IsConflict(err) {
		t.Fatalf("second account in the same program: %v", err)
	}
	if _, err := e.accounts.Open(ctx, OpenAccountCommand{CustomerID: "cus-x", ProgramCode: "DEBIT_CLASSIC"}); !domain.IsConflict(err) {
		t.Fatalf("customer without kyc: %v", err)
	}
	if _, err := e.accounts.Open(ctx, OpenAccountCommand{CustomerID: "nobody", ProgramCode: "DEBIT_CLASSIC"}); !domain.IsInvalid(err) {
		t.Fatalf("unknown customer: %v", err)
	}
	if _, err := e.accounts.Open(ctx, OpenAccountCommand{CustomerID: "cus-1", ProgramCode: "GOLDEN"}); !domain.IsInvalid(err) {
		t.Fatalf("unknown program: %v", err)
	}
	if list, _ := e.accounts.ListByCustomer(ctx, "cus-1"); len(list) != 1 {
		t.Fatalf("accounts: %d", len(list))
	}
}

func TestIssueCards(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	debit := e.open(t, "cus-1", "DEBIT_CLASSIC", 0)
	plastic, err := e.cards.Issue(ctx, IssueCommand{AccountID: debit.ID, Type: domain.CardPlastic})
	if err != nil || plastic.Card.Status != domain.StatusPending || plastic.Card.CardholderName != "NGUYEN VAN AN" || plastic.Card.ServiceCode != "201" {
		t.Fatalf("plastic: %v %+v", err, plastic.Card)
	}
	if !domain.ValidPAN(plastic.PAN) || plastic.Card.BIN != "730100" || plastic.PAN[12:] != plastic.Card.Last4 {
		t.Fatalf("pan %s", plastic.PAN)
	}
	if _, err := e.cards.Issue(ctx, IssueCommand{AccountID: debit.ID, Type: domain.CardPlastic}); !domain.IsConflict(err) {
		t.Fatalf("second plastic: %v", err)
	}
	virtual, err := e.cards.Issue(ctx, IssueCommand{AccountID: debit.ID, Type: domain.CardVirtual, PrintedName: "An Shopping"})
	if err != nil || virtual.Card.Status != domain.StatusNormal || virtual.Card.Controls != (domain.Controls{Ecommerce: true}) || virtual.Card.CardholderName != "AN SHOPPING" {
		t.Fatalf("virtual: %v %+v", err, virtual.Card)
	}
	temp, err := e.cards.Issue(ctx, IssueCommand{AccountID: debit.ID, Type: domain.CardTemporary})
	if err != nil || temp.Card.ValidUntil == nil || !temp.Card.ValidUntil.Equal(e.clock.Now().Add(24*time.Hour)) {
		t.Fatalf("temporary: %v %+v", err, temp.Card)
	}
	if _, err := e.cards.Issue(ctx, IssueCommand{AccountID: debit.ID, Type: domain.CardTemporary}); !domain.IsConflict(err) {
		t.Fatalf("second live temporary: %v", err)
	}
	if a := e.authorize(t, temp, func(r *domain.AuthorizationRequest) { r.Amount = 1 }); a.Code != domain.CodeInsufficientFunds {
		t.Fatalf("temporary card works until it expires: %+v", a)
	}
	e.clock.Advance(25 * time.Hour)
	if a := e.authorize(t, temp, nil); a.Code != domain.CodeExpiredCard {
		t.Fatalf("expired temporary: %+v", a)
	}
	if _, err := e.cards.Issue(ctx, IssueCommand{AccountID: debit.ID, Type: domain.CardTemporary}); err != nil {
		t.Fatalf("new temporary after expiry: %v", err)
	}

	prepaid := e.open(t, "cus-2", "PREPAID_ONLINE", 0)
	if _, err := e.cards.Issue(ctx, IssueCommand{AccountID: prepaid.ID, Type: domain.CardPlastic}); !domain.IsInvalid(err) {
		t.Fatalf("prepaid online has no plastic: %v", err)
	}
	if _, err := e.cards.SetPIN(ctx, virtual.Card.ID, "2580"); !domain.IsInvalid(err) {
		t.Fatalf("pin on a virtual card: %v", err)
	}
}

func TestDebitPurchaseConfirmAndCancel(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	acc := e.open(t, "cus-1", "DEBIT_CLASSIC", 3_000_000)
	c := e.card(t, acc.ID, domain.CardPlastic)

	a := e.authorize(t, c, func(r *domain.AuthorizationRequest) { r.Channel, r.Amount = domain.ChannelPOS, 1_500_000 })
	if !a.Approved() || a.Status != domain.AuthAuthorized || len(a.AuthCode) != 6 {
		t.Fatalf("purchase: %+v", a)
	}
	got, _ := e.accounts.Get(ctx, acc.ID)
	if got.Held != 1_500_000 || got.Available() != 1_500_000 {
		t.Fatalf("hold: %+v", got)
	}
	if over := e.authorize(t, c, func(r *domain.AuthorizationRequest) { r.Channel, r.Amount = domain.ChannelPOS, 1_500_001 }); over.Code != domain.CodeInsufficientFunds {
		t.Fatalf("insufficient: %+v", over)
	}

	confirmed, err := e.auths.Confirm(ctx, a.ID, 1_200_000)
	if err != nil || confirmed.Status != domain.AuthConfirmed {
		t.Fatalf("confirm: %v", err)
	}
	got, _ = e.accounts.Get(ctx, acc.ID)
	if got.Held != 0 || got.Balance != 1_800_000 {
		t.Fatalf("after clearing: %+v", got)
	}

	second := e.authorize(t, c, func(r *domain.AuthorizationRequest) { r.Channel, r.Amount = domain.ChannelPOS, 500_000 })
	if _, err := e.auths.Cancel(ctx, second.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := e.auths.Cancel(ctx, second.ID); !domain.IsConflict(err) {
		t.Fatalf("cancel twice: %v", err)
	}
	refund := e.authorize(t, c, func(r *domain.AuthorizationRequest) {
		r.Channel, r.ProcessingCode, r.Amount, r.CVV = domain.ChannelPOS, domain.ProcessingRefund, 200_000, ""
	})
	if !refund.Approved() || refund.Status != domain.AuthConfirmed {
		t.Fatalf("refund: %+v", refund)
	}
	got, _ = e.accounts.Get(ctx, acc.ID)
	if got.Held != 0 || got.Balance != 2_000_000 {
		t.Fatalf("after cancel and refund: %+v", got)
	}
	txns, err := e.accounts.Transactions(ctx, acc.ID, 0)
	if err != nil || len(txns) != 3 || txns[0].Type != domain.TxnRefund || txns[1].Type != domain.TxnPurchase || txns[1].Amount != -1_200_000 || txns[2].Type != domain.TxnPayment {
		t.Fatalf("ledger: %v %+v", err, txns)
	}
	if spent, _ := e.auths.SpentToday(ctx, c.Card.ID); spent != 1_200_000 {
		t.Fatalf("spent today counts the confirmed amount only: %d", spent)
	}
}

func TestCreditAccountAndPayment(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	acc := e.open(t, "cus-1", "CREDIT_GOLD", 0)
	c := e.card(t, acc.ID, domain.CardPlastic)
	a := e.authorize(t, c, func(r *domain.AuthorizationRequest) { r.Channel, r.Amount = domain.ChannelPOS, 40_000_000 })
	if !a.Approved() {
		t.Fatalf("credit purchase: %+v", a)
	}
	if _, err := e.auths.Confirm(ctx, a.ID, 0); err != nil {
		t.Fatal(err)
	}
	if _, err := e.accounts.SetCreditLimit(ctx, acc.ID, 30_000_000); !domain.IsConflict(err) {
		t.Fatalf("limit below outstanding: %v", err)
	}
	if _, err := e.accounts.Cancel(ctx, acc.ID, ""); !domain.IsConflict(err) {
		t.Fatalf("cancel with debt: %v", err)
	}
	got, _, err := e.accounts.ReceivePayment(ctx, acc.ID, 40_000_000, "statement payment")
	if err != nil || got.Balance != 0 || got.Available() != 50_000_000 {
		t.Fatalf("payment: %v %+v", err, got)
	}
	cancelled, err := e.accounts.Cancel(ctx, acc.ID, "closed")
	if err != nil || cancelled.Status != domain.AccountCancelled {
		t.Fatalf("cancel: %v", err)
	}
	card, _ := e.cards.Get(ctx, c.Card.ID)
	if card.Status != domain.StatusCancelled {
		t.Fatalf("cards follow the account: %s", card.Status)
	}
}

func TestStepUpAuthentication(t *testing.T) {
	e := newEnv(t)
	acc := e.open(t, "cus-1", "DEBIT_CLASSIC", 10_000_000)
	c := e.card(t, acc.ID, domain.CardVirtual)
	high := func(id string) func(*domain.AuthorizationRequest) {
		return func(r *domain.AuthorizationRequest) { r.Amount, r.AuthenticationID = 2_000_000, id }
	}
	if a := e.authorize(t, c, high("")); a.Code != domain.CodeAuthenticationReqd {
		t.Fatalf("no authentication: %+v", a)
	}
	if a := e.authorize(t, c, high("forged")); a.Code != domain.CodeAuthenticationReqd {
		t.Fatalf("forged authentication: %+v", a)
	}
	e.authn.ok["pa-1"] = true
	a := e.authorize(t, c, high("pa-1"))
	if !a.Approved() || a.AuthenticationID != "pa-1" {
		t.Fatalf("authenticated: %+v", a)
	}
	if a := e.authorize(t, c, high("pa-1")); a.Code != domain.CodeAuthenticationReqd {
		t.Fatalf("replayed authentication: %+v", a)
	}
	if a := e.authorize(t, c, func(r *domain.AuthorizationRequest) { r.Amount = 1_999_999 }); !a.Approved() {
		t.Fatalf("below threshold: %+v", a)
	}
	if len(e.authn.seen) != 3 {
		t.Fatalf("auth-service is only called when an id is given above the threshold: %v", e.authn.seen)
	}
}

func TestConcurrentAuthorizationsShareAccountFunds(t *testing.T) {
	e := newEnv(t)
	acc := e.open(t, "cus-1", "DEBIT_PLATINUM", 10_000_000)
	first := e.card(t, acc.ID, domain.CardPlastic)
	second := e.card(t, acc.ID, domain.CardVirtual)
	var wg sync.WaitGroup
	var mu sync.Mutex
	approved := 0
	for i := 0; i < 40; i++ {
		c, ch := first, domain.ChannelPOS
		if i%2 == 1 {
			c, ch = second, domain.ChannelEcommerce
		}
		wg.Add(1)
		go func() {
			defer wg.Done()
			a := e.authorize(t, c, func(r *domain.AuthorizationRequest) { r.Channel, r.Amount = ch, 1_000_000 })
			if a.Approved() {
				mu.Lock()
				approved++
				mu.Unlock()
			}
		}()
	}
	wg.Wait()
	got, _ := e.accounts.Get(context.Background(), acc.ID)
	if approved != 10 || got.Held != 10_000_000 || got.Available() != 0 {
		t.Fatalf("approved %d held %d", approved, got.Held)
	}
}

func TestPINLockoutAndDailyLimit(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	acc := e.open(t, "cus-1", "DEBIT_CLASSIC", 100_000_000)
	c := e.card(t, acc.ID, domain.CardPlastic)
	if _, err := e.cards.SetPIN(ctx, c.Card.ID, "1234"); !domain.IsInvalid(err) {
		t.Fatalf("weak pin: %v", err)
	}
	if _, err := e.cards.SetPIN(ctx, c.Card.ID, "2580"); err != nil {
		t.Fatal(err)
	}
	atm := func(pin string, amount int64) *domain.Authorization {
		return e.authorize(t, c, func(r *domain.AuthorizationRequest) {
			r.Channel, r.ProcessingCode, r.PIN, r.CVV, r.Amount = domain.ChannelATM, domain.ProcessingWithdrawal, pin, "", amount
		})
	}
	if a := atm("2580", 5_000_000); !a.Approved() {
		t.Fatalf("withdrawal: %+v", a)
	}
	for i := 0; i < domain.MaxPINAttempts; i++ {
		atm("0000", 1)
	}
	if a := atm("2580", 1); a.Code != domain.CodePINTriesExceeded {
		t.Fatalf("locked: %+v", a)
	}
	if _, err := e.cards.SetPIN(ctx, c.Card.ID, "3691"); err != nil {
		t.Fatal(err)
	}
	if a := atm("3691", 20_000_000); !a.Approved() {
		t.Fatalf("new pin: %+v", a)
	}
	if a := atm("3691", 20_000_000); !a.Approved() {
		t.Fatalf("second withdrawal: %+v", a)
	}
	if a := atm("3691", 6_000_000); a.Code != domain.CodeExceedsLimit {
		t.Fatalf("daily limit 50M: %+v", a)
	}
	e.clock.Advance(21 * time.Hour)
	if a := atm("3691", 6_000_000); !a.Approved() {
		t.Fatalf("limit resets at local midnight: %+v", a)
	}
}

func TestReportAndResolve(t *testing.T) {
	e := newEnv(t)
	ctx := context.Background()
	acc := e.open(t, "cus-1", "DEBIT_PLATINUM", 0)
	c := e.card(t, acc.ID, domain.CardPlastic)
	if _, err := e.cards.UpdateLimits(ctx, c.Card.ID, domain.Limits{PerTransaction: 10_000_000, Daily: 20_000_000}); err != nil {
		t.Fatal(err)
	}
	res, err := e.cards.Resolve(ctx, c.PAN, c.Card.Expiry)
	if err != nil || res.Card.ID != c.Card.ID || res.Account.UserID != 7 {
		t.Fatalf("resolve: %v %+v", err, res)
	}
	if _, err := e.cards.Resolve(ctx, c.PAN, domain.Expiry{Month: 1, Year: 2040}); err != domain.ErrNotFound {
		t.Fatalf("resolve with wrong expiry: %v", err)
	}

	old, replacement, err := e.cards.Report(ctx, c.Card.ID, ReportCommand{Reason: domain.StatusDamaged, Reissue: true})
	if err != nil || old.Status != domain.StatusDamaged || old.ReplacedByCardID != replacement.Card.ID {
		t.Fatalf("report: %v %+v", err, old)
	}
	nc := replacement.Card
	if nc.ReplacesCardID != old.ID || nc.Status != domain.StatusPending || nc.Type != domain.CardPlastic || nc.Limits.Daily != 20_000_000 {
		t.Fatalf("replacement: %+v", nc)
	}
	if a := e.authorize(t, c, nil); a.Code != domain.CodeClosedCard {
		t.Fatalf("damaged card declined: %+v", a)
	}
	if _, _, err := e.cards.Report(ctx, c.Card.ID, ReportCommand{Reason: domain.StatusLost}); !domain.IsConflict(err) {
		t.Fatalf("report twice: %v", err)
	}
	if _, err := e.accounts.Block(ctx, acc.ID, "review"); err != nil {
		t.Fatal(err)
	}
	if _, err := e.cards.Issue(ctx, IssueCommand{AccountID: acc.ID, Type: domain.CardVirtual}); !domain.IsConflict(err) {
		t.Fatalf("issue on blocked account: %v", err)
	}
}
