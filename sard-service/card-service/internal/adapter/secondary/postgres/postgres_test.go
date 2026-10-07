package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JIeeiroSst/card-service/internal/domain"
)

func connect(t *testing.T) *DB {
	t.Helper()
	dsn := os.Getenv("CARD_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("CARD_TEST_POSTGRES_DSN is not set")
	}
	ctx := context.Background()
	db, err := Connect(ctx, dsn, 4)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(db.Close)
	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatalf("migrate must be idempotent: %v", err)
	}
	return db
}

func fixture(t *testing.T, ctx context.Context, db *DB, suffix string, now time.Time) (*domain.Account, *domain.Card) {
	t.Helper()
	gold, _ := domain.NewCatalog(domain.Programs).Get("CREDIT_GOLD")
	a, err := domain.OpenAccount("acc-"+suffix, domain.Customer{ID: "cus-" + suffix, UserID: 7, FullName: "An", Eligible: true}, gold, 0, now)
	if err != nil {
		t.Fatal(err)
	}
	if err := NewAccounts(db).Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	until := now.Add(24 * time.Hour)
	c := &domain.Card{
		ID: "card-" + suffix, AccountID: a.ID, CustomerID: a.CustomerID, ProgramCode: gold.Code, Type: domain.CardTemporary,
		CardholderName: "AN", PANHash: "hash-" + suffix, PANCipher: []byte{1, 2, 3}, BIN: gold.BIN, Last4: "1234",
		Expiry: domain.Expiry{Month: 10, Year: 2031}, ValidUntil: &until, ServiceCode: "101", Status: domain.StatusNormal,
		Limits: gold.DefaultLimits, Controls: domain.Controls{Ecommerce: true}, CreatedAt: now, UpdatedAt: now, ActivatedAt: &now,
	}
	if err := NewCards(db).Create(ctx, c); err != nil {
		t.Fatal(err)
	}
	return a, c
}

func TestRepositories(t *testing.T) {
	db := connect(t)
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Microsecond)
	suffix := now.Format("150405.000000")
	accounts, cards, auths, txns := NewAccounts(db), NewCards(db), NewAuthorizations(db), NewTransactions(db)
	a, c := fixture(t, ctx, db, suffix, now)

	dupAccount := *a
	dupAccount.ID = "other-" + suffix
	if err := accounts.Create(ctx, &dupAccount); !domain.IsConflict(err) {
		t.Fatalf("second open account in the program: %v", err)
	}
	dupCard := *c
	dupCard.ID = "other-" + suffix
	if err := cards.Create(ctx, &dupCard); err != domain.ErrDuplicatePAN {
		t.Fatalf("duplicate pan: %v", err)
	}

	a.Hold(500, now)
	a.Post(-200, now)
	if err := accounts.Update(ctx, a); err != nil {
		t.Fatal(err)
	}
	gotA, err := accounts.Get(ctx, a.ID)
	if err != nil || gotA.Held != 500 || gotA.Balance != -200 || gotA.Mode != domain.ModeCredit || gotA.UserID != 7 || !gotA.CreatedAt.Equal(now) {
		t.Fatalf("account: %v %+v", err, gotA)
	}
	gotC, err := cards.GetByPANHash(ctx, c.PANHash)
	if err != nil || gotC.Type != domain.CardTemporary || gotC.ValidUntil == nil || !gotC.ValidUntil.Equal(*c.ValidUntil) || gotC.Controls != c.Controls {
		t.Fatalf("card: %v %+v", err, gotC)
	}
	if list, err := cards.ListByAccount(ctx, a.ID); err != nil || len(list) != 1 {
		t.Fatalf("cards by account: %v %d", err, len(list))
	}

	mk := func(id string, pc domain.ProcessingCode, amount int64, d domain.Decision) *domain.Authorization {
		au := domain.NewAuthorization(id+suffix, c, domain.AuthorizationRequest{Amount: amount, Currency: "VND", Channel: domain.ChannelEcommerce, ProcessingCode: pc}, d, now)
		if err := auths.Create(ctx, au); err != nil {
			t.Fatal(err)
		}
		return au
	}
	ok := domain.Decision{Code: domain.CodeApproved, Reason: "approved"}
	held := mk("a1-", domain.ProcessingPurchase, 300, ok)
	cleared := mk("a2-", domain.ProcessingPurchase, 400, ok)
	mk("a3-", domain.ProcessingPurchase, 999, domain.Decision{Code: domain.CodeExceedsLimit})
	mk("a4-", domain.ProcessingRefund, 50, ok)
	if _, err := cleared.Confirm(250, now); err != nil {
		t.Fatal(err)
	}
	if err := auths.Update(ctx, cleared); err != nil {
		t.Fatal(err)
	}
	if sum, err := auths.SumDebits(ctx, c.ID, now.Add(-time.Minute)); err != nil || sum != 550 {
		t.Fatalf("sum debits: %d %v", sum, err)
	}
	if _, err := held.Cancel(now); err != nil {
		t.Fatal(err)
	}
	if err := auths.Update(ctx, held); err != nil {
		t.Fatal(err)
	}
	if sum, _ := auths.SumDebits(ctx, c.ID, now.Add(-time.Minute)); sum != 250 {
		t.Fatalf("sum after cancel: %d", sum)
	}
	if list, err := auths.ListByCard(ctx, c.ID, 2); err != nil || len(list) != 2 {
		t.Fatalf("list auths: %v %d", err, len(list))
	}

	for i, amt := range []int64{1000, -250} {
		tx := &domain.Transaction{ID: suffix + "-t" + string(rune('0'+i)), AccountID: a.ID, Type: domain.TxnPurchase,
			ProcessingCode: domain.ProcessingPurchase, Amount: amt, BalanceAfter: amt, CreatedAt: now.Add(time.Duration(i) * time.Second)}
		if err := txns.Create(ctx, tx); err != nil {
			t.Fatal(err)
		}
	}
	if list, err := txns.ListByAccount(ctx, a.ID, 10); err != nil || len(list) != 2 || list[0].Amount != -250 {
		t.Fatalf("transactions: %v %+v", err, list)
	}
}

func TestWithLockNestsAndRollsBack(t *testing.T) {
	db := connect(t)
	ctx := context.Background()
	now := time.Now().UTC()
	suffix := "lock-" + now.Format("150405.000000")
	err := db.WithLock(ctx, "account:"+suffix, func(ctx context.Context) error {
		return db.WithLock(ctx, "card:"+suffix, func(ctx context.Context) error {
			fixture(t, ctx, db, suffix, now)
			return domain.Conflict("abort")
		})
	})
	if !domain.IsConflict(err) {
		t.Fatalf("got %v", err)
	}
	if _, err := NewAccounts(db).Get(ctx, "acc-"+suffix); err != domain.ErrNotFound {
		t.Fatalf("nested writes must roll back: %v", err)
	}
}
