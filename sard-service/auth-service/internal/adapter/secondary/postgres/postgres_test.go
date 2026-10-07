package postgres

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/JIeeiroSst/auth-service/internal/domain"
)

func TestChallenges(t *testing.T) {
	dsn := os.Getenv("AUTH_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("AUTH_TEST_POSTGRES_DSN is not set")
	}
	ctx := context.Background()
	db, err := Connect(ctx, dsn, 4)
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	if err := db.Migrate(ctx); err != nil {
		t.Fatal(err)
	}
	repo := NewChallenges(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	user := now.UnixNano()
	policy := domain.Policy{OTPTTL: 5 * time.Minute, MaxAttempts: 3, MaxResends: 2, ValidityTime: 10 * time.Minute}
	card := domain.CardRef{CardID: "card", UserID: user, MaskedPAN: "m", Status: "NORMAL"}
	req := domain.InitiateRequest{Amount: 10, Currency: "VND", Merchant: "m"}
	c, _ := domain.NewChallenge("ch-"+now.Format("150405.000000"), card, "anguyen", req, now.Add(time.Minute), policy, now)
	c.Resent(now.Add(90*time.Second), now)
	if err := repo.Create(ctx, c); err != nil {
		t.Fatal(err)
	}
	pending, err := repo.ListPendingByUser(ctx, user, now)
	if err != nil || len(pending) != 1 {
		t.Fatalf("pending: %v %d", err, len(pending))
	}
	if err := c.Verify(user, true, now, policy); err != nil {
		t.Fatal(err)
	}
	err = db.WithLock(ctx, "challenge:"+c.ID, func(ctx context.Context) error { return repo.Update(ctx, c) })
	if err != nil {
		t.Fatal(err)
	}
	got, err := repo.Get(ctx, c.ID)
	if err != nil || got.Status != domain.StatusAuthenticated || got.ValidUntil == nil || !got.ValidUntil.Equal(now.Add(10*time.Minute)) ||
		got.Username != "anguyen" || got.Resends != 1 || !got.OTPExpiresAt.Equal(now.Add(90*time.Second)) {
		t.Fatalf("get: %v %+v", err, got)
	}
	if pending, _ := repo.ListPendingByUser(ctx, user, now); len(pending) != 0 {
		t.Fatal("authenticated challenge is no longer pending")
	}
	if _, err := repo.Get(ctx, "missing"); err != domain.ErrNotFound {
		t.Fatalf("missing: %v", err)
	}
}
