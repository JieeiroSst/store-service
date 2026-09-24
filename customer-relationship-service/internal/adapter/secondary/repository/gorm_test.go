package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
)

func newDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:"), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	sqlDB, _ := db.DB()
	sqlDB.SetMaxOpenConns(1) // every connection to :memory: is its own database
	if err := db.AutoMigrate(model.Models()...); err != nil {
		t.Fatal(err)
	}
	return db
}

func TestList_FilterSearchAndTotal(t *testing.T) {
	ctx := context.Background()
	repo := New[model.Contact](newDB(t))
	for _, c := range []model.Contact{
		{AccountID: 1, Name: "Ada Lovelace", Email: "ada@x.io"},
		{AccountID: 1, Name: "Grace Hopper", Email: "grace@x.io"},
		{AccountID: 2, Name: "Alan Turing", Email: "alan@x.io"},
	} {
		c := c
		if err := repo.Create(ctx, &c); err != nil {
			t.Fatal(err)
		}
	}

	got, total, err := repo.List(ctx, port.ListQuery{Limit: 1, Equals: map[string]any{"account_id": uint64(1)}})
	if err != nil || total != 2 || len(got) != 1 {
		t.Fatalf("filter+page: got %d rows, total %d, err %v; want 1 row, total 2", len(got), total, err)
	}

	got, total, err = repo.List(ctx, port.ListQuery{Limit: 10, Search: "hopper"})
	if err != nil || total != 1 || got[0].Name != "Grace Hopper" {
		t.Fatalf("search: got %+v, total %d, err %v", got, total, err)
	}
}

func TestUpdate_ReplacesZeroValuesAndKeepsCreatedAt(t *testing.T) {
	ctx := context.Background()
	repo := New[model.Account](newDB(t))

	a := &model.Account{Name: "acme", Description: "old", Phone: "1"}
	if err := repo.Create(ctx, a); err != nil {
		t.Fatal(err)
	}
	before, _ := repo.GetByID(ctx, a.ID)
	time.Sleep(1100 * time.Millisecond) // sqlite stores timestamps to the second

	if err := repo.Update(ctx, a.ID, &model.Account{Name: "acme 2"}); err != nil {
		t.Fatal(err)
	}
	after, _ := repo.GetByID(ctx, a.ID)

	if after.Name != "acme 2" || after.Description != "" || after.Phone != "" {
		t.Fatalf("PUT did not replace all columns: %+v", after)
	}
	if !after.CreatedAt.Equal(before.CreatedAt) {
		t.Fatalf("created_at changed: %v -> %v", before.CreatedAt, after.CreatedAt)
	}
	if !after.UpdatedAt.After(before.UpdatedAt) {
		t.Fatalf("updated_at not bumped: %v -> %v", before.UpdatedAt, after.UpdatedAt)
	}
}

func TestInTx_RollsBackOnError(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	repo := New[model.Account](db)

	boom := errors.New("boom")
	err := NewTxRunner(db).InTx(ctx, func(ctx context.Context) error {
		if err := repo.Create(ctx, &model.Account{Name: "ghost"}); err != nil {
			return err
		}
		return boom
	})
	if !errors.Is(err, boom) {
		t.Fatalf("err = %v", err)
	}
	if _, total, _ := repo.List(ctx, port.ListQuery{Limit: 10}); total != 0 {
		t.Fatalf("rolled-back row survived: total = %d", total)
	}
}

func TestExpireDue_ClosesOnlyLapsedActiveContracts(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	repo := New[model.Contract](db)
	now := time.Now()
	past, future := now.Add(-time.Hour), now.Add(time.Hour)

	rows := []model.Contract{
		{AccountID: 1, Status: model.ContractApproved, EndDate: &past},   // 1: expires
		{AccountID: 1, Status: model.ContractPending, EndDate: &past},    // 2: expires
		{AccountID: 1, Status: model.ContractApproved, EndDate: &future}, // 3: still running
		{AccountID: 1, Status: model.ContractApproved},                   // 4: no end date
		{AccountID: 1, Status: model.ContractRejected, EndDate: &past},   // 5: already final
	}
	for i := range rows {
		if err := repo.Create(ctx, &rows[i]); err != nil {
			t.Fatal(err)
		}
	}

	n, err := NewContractExpirer(db).ExpireDue(ctx, now)
	if err != nil || n != 2 {
		t.Fatalf("expired %d, err %v; want 2", n, err)
	}

	want := map[uint]string{
		1: model.ContractExpired, 2: model.ContractExpired, 3: model.ContractApproved,
		4: model.ContractApproved, 5: model.ContractRejected,
	}
	for id, status := range want {
		got, _ := repo.GetByID(ctx, id)
		if got.Status != status {
			t.Errorf("contract %d: status %q, want %q", id, got.Status, status)
		}
	}

	if n, _ := NewContractExpirer(db).ExpireDue(ctx, now); n != 0 {
		t.Fatalf("second run expired %d, want 0", n)
	}
}

func TestExpireOne(t *testing.T) {
	ctx := context.Background()
	db := newDB(t)
	repo := New[model.Contract](db)
	exp := NewContractExpirer(db)
	now := time.Now().Truncate(time.Second)
	past, future := now.Add(-time.Hour), now.Add(time.Hour)

	rows := []model.Contract{
		{AccountID: 1, Status: model.ContractApproved, EndDate: &past},   // 1: due
		{AccountID: 1, Status: model.ContractApproved, EndDate: &now},    // 2: due exactly now (<=)
		{AccountID: 1, Status: model.ContractApproved, EndDate: &future}, // 3: not yet
		{AccountID: 1, Status: model.ContractExpired, EndDate: &past},    // 4: already expired
		{AccountID: 1, Status: model.ContractApproved},                   // 5: no end date
	}
	for i := range rows {
		if err := repo.Create(ctx, &rows[i]); err != nil {
			t.Fatal(err)
		}
	}

	for id, want := range map[uint]bool{1: true, 2: true, 3: false, 4: false, 5: false, 99: false} {
		got, err := exp.ExpireOne(ctx, id, now)
		if err != nil || got != want {
			t.Errorf("ExpireOne(%d) = %v, %v; want %v", id, got, err, want)
		}
	}
	if again, _ := exp.ExpireOne(ctx, 1, now); again {
		t.Fatal("expiring twice reported success twice")
	}
	if c, _ := repo.GetByID(ctx, 3); c.Status != model.ContractApproved {
		t.Fatalf("contract 3 = %s", c.Status)
	}
}
