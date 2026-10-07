package postgres

import (
	"context"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-info-service/internal/domain"
)

func TestCustomers(t *testing.T) {
	dsn := os.Getenv("CUSTOMER_INFO_TEST_POSTGRES_DSN")
	if dsn == "" {
		t.Skip("CUSTOMER_INFO_TEST_POSTGRES_DSN is not set")
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
	repo := NewCustomers(db)
	now := time.Now().UTC().Truncate(time.Microsecond)
	base := now.UnixNano()
	mk := func(i int64) *domain.Customer {
		c, _ := domain.NewCustomer("cus-"+strconv.FormatInt(base+i, 10), domain.UserProfile{ID: base + i, Name: "Nguyen Van An", Active: true}, now)
		return c
	}
	a, b := mk(1), mk(2)
	for _, c := range []*domain.Customer{a, b} {
		if err := repo.Create(ctx, c); err != nil {
			t.Fatal(err)
		}
	}
	if err := repo.Create(ctx, a); !domain.IsConflict(err) {
		t.Fatalf("duplicate user: %v", err)
	}
	dob := time.Date(2000, 1, 2, 0, 0, 0, 0, time.UTC)
	doc := domain.IdentityDocument{Source: "card_ocr", Number: "079200012345", FullName: "NGUYEN VAN AN", DateOfBirth: &dob, ChecksumValid: true, Confidence: 0.9}
	a.KYC = domain.KYCPolicy{MinAge: 18, MinConfidence: 0.5, RequireFaceMatch: true}.Evaluate(a, domain.EkycResult{Document: &doc, FaceVerified: true}, now)
	a.KYC.DocumentHash = "hash-" + strconv.FormatInt(base, 10)
	if err := repo.Update(ctx, a); err != nil {
		t.Fatal(err)
	}
	b.KYC = a.KYC
	if err := repo.Update(ctx, b); !domain.IsConflict(err) {
		t.Fatalf("same verified document twice: %v", err)
	}
	got, err := repo.GetByDocumentHash(ctx, a.KYC.DocumentHash)
	if err != nil || got.ID != a.ID || got.KYC.Status != domain.KYCVerified || got.KYC.Document.Number != "2345" || !got.KYC.FaceVerified || !got.KYC.Document.ChecksumValid ||
		!got.KYC.Document.DateOfBirth.Equal(dob) || !got.CreatedAt.Equal(now) {
		t.Fatalf("get by hash: %v %+v", err, got)
	}
	if byUser, err := repo.GetByUserID(ctx, b.UserID); err != nil || byUser.ID != b.ID || byUser.KYC.Status != domain.KYCNone {
		t.Fatalf("by user: %v %+v", err, byUser)
	}
	if _, err := repo.Get(ctx, "missing"); err != domain.ErrNotFound {
		t.Fatalf("missing: %v", err)
	}
}
