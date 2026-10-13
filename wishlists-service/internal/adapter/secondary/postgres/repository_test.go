package postgres

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sync"
	"testing"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func testRepo(t *testing.T) *Repository {
	host := os.Getenv("WISHLISTS_TEST_DB_HOST")
	if host == "" {
		t.Skip("WISHLISTS_TEST_DB_HOST not set")
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, env("WISHLISTS_TEST_DB_PORT", "5432"), env("WISHLISTS_TEST_DB_USER", "postgres"),
		os.Getenv("WISHLISTS_TEST_DB_PASSWORD"), env("WISHLISTS_TEST_DB_NAME", "postgres"))
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{TranslateError: true, Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := db.Exec("DROP TABLE IF EXISTS wishlists_line, wishlists_wishlist").Error; err != nil {
		t.Fatal(err)
	}
	r := NewRepository(db)
	var wg sync.WaitGroup
	errs := make([]error, 3)
	for i := range errs {
		wg.Add(1)
		go func() {
			defer wg.Done()
			errs[i] = r.Migrate(ctx)
		}()
	}
	wg.Wait()
	for _, err := range errs {
		if err != nil {
			t.Fatalf("concurrent migrate: %v", err)
		}
	}
	t.Cleanup(func() {
		sqlDB, _ := db.DB()
		_ = sqlDB.Close()
	})
	return r
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func TestRepository(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	if err := r.Ping(ctx); err != nil {
		t.Fatal(err)
	}
	w, _ := model.NewWishlist("1", "Birthday", model.VisibilityShared)
	if err := r.Create(ctx, w); err != nil {
		t.Fatal(err)
	}
	if w.ID == 0 || w.DateCreated.IsZero() {
		t.Fatalf("create did not populate generated fields: %+v", w)
	}
	dup, _ := model.NewWishlist("2", "x", model.VisibilityPrivate)
	dup.Key = w.Key
	if err := r.Create(ctx, dup); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("duplicate key: %v", err)
	}
	pub, _ := model.NewWishlist("1", "Public", model.VisibilityPublic)
	if err := r.Create(ctx, pub); err != nil {
		t.Fatal(err)
	}

	l, err := r.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 42, Title: "Book", Quantity: 1})
	if err != nil {
		t.Fatal(err)
	}
	again, err := r.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 42, Title: "Book", Quantity: model.MaxQuantity})
	if err != nil || again.ID != l.ID || again.Quantity != model.MaxQuantity {
		t.Fatalf("upsert line: %+v %v", again, err)
	}
	if _, err := r.AddLine(ctx, &model.Line{WishlistID: 999999, ProductID: 1, Title: "x", Quantity: 1}); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("line for missing wishlist: %v", err)
	}
	if _, err := r.AddLine(ctx, &model.Line{WishlistID: w.ID, ProductID: 7, Title: "Pen", Quantity: 2}); err != nil {
		t.Fatal(err)
	}

	got, err := r.GetByKey(ctx, w.Key)
	if err != nil || got.ID != w.ID || len(got.Lines) != 2 || got.Lines[0].ProductID != 42 {
		t.Fatalf("get by key: %+v %v", got, err)
	}
	if _, err := r.Get(ctx, 999999); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("missing get: %v", err)
	}

	list, total, err := r.List(ctx, port.ListFilter{OwnerID: "1", Limit: 1})
	if err != nil || len(list) != 1 || total != 2 {
		t.Fatalf("list owner: %d/%d %v", len(list), total, err)
	}
	list, total, err = r.List(ctx, port.ListFilter{OwnerID: "1", PublicOnly: true, Limit: 10})
	if err != nil || len(list) != 1 || list[0].ID != pub.ID || total != 1 {
		t.Fatalf("list public: %+v %v", list, err)
	}

	got.Name, got.Visibility = "Renamed", model.VisibilityPrivate
	if err := r.Update(ctx, got); err != nil {
		t.Fatal(err)
	}
	got.Key = pub.Key
	if err := r.Update(ctx, got); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("update to taken key: %v", err)
	}

	upd, err := r.UpdateLineQuantity(ctx, w.ID, l.ID, 3)
	if err != nil || upd.Quantity != 3 {
		t.Fatalf("update line: %+v %v", upd, err)
	}
	if _, err := r.UpdateLineQuantity(ctx, pub.ID, l.ID, 3); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("update line through other wishlist: %v", err)
	}
	if err := r.DeleteLine(ctx, pub.ID, l.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("delete line through other wishlist: %v", err)
	}
	if err := r.DeleteLine(ctx, w.ID, l.ID); err != nil {
		t.Fatal(err)
	}

	if err := r.Delete(ctx, w.ID); err != nil {
		t.Fatal(err)
	}
	if err := r.Delete(ctx, w.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("double delete: %v", err)
	}
	var lines int64
	r.db.Model(&lineRecord{}).Where("wishlist_id = ?", w.ID).Count(&lines)
	if lines != 0 {
		t.Fatalf("lines should cascade on delete, %d left", lines)
	}
}

func TestRepositoryFeatures(t *testing.T) {
	r := testRepo(t)
	ctx := context.Background()

	var wg sync.WaitGroup
	ids := make([]int64, 8)
	errs := make([]error, 8)
	for i := range ids {
		wg.Add(1)
		go func() {
			defer wg.Done()
			c, _ := model.NewWishlist("10", model.DefaultName, model.VisibilityPrivate)
			w, _, err := r.FirstOrCreate(ctx, c)
			errs[i] = err
			if err == nil {
				ids[i] = w.ID
			}
		}()
	}
	wg.Wait()
	for i := range ids {
		if errs[i] != nil || ids[i] != ids[0] {
			t.Fatalf("concurrent default creation must yield one wishlist: %v %v", ids, errs)
		}
	}
	_, total, _ := r.List(ctx, port.ListFilter{OwnerID: "10", Limit: 10})
	if total != 1 {
		t.Fatalf("expected exactly one default wishlist, got %d", total)
	}
	c, _ := model.NewWishlist("10", "ignored", model.VisibilityPrivate)
	again, created, err := r.FirstOrCreate(ctx, c)
	if err != nil || created || again.ID != ids[0] {
		t.Fatalf("existing default: %+v %v %v", again, created, err)
	}

	src, _ := model.NewWishlist("10", "Source", model.VisibilityShared)
	src.Lines = []model.Line{{ProductID: 1, Title: "A", Quantity: 2}, {ProductID: 2, Title: "B", Quantity: 1}}
	if err := r.Create(ctx, src); err != nil {
		t.Fatal(err)
	}
	if src.Lines[0].ID == 0 || src.Lines[0].WishlistID != src.ID {
		t.Fatalf("lines not persisted with wishlist: %+v", src.Lines)
	}
	dup, _ := model.NewWishlist("11", "Dup", model.VisibilityPrivate)
	dup.Key = src.Key
	dup.Lines = []model.Line{{ProductID: 3, Title: "C", Quantity: 1}}
	if err := r.Create(ctx, dup); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("duplicate key with lines: %v", err)
	}

	dst := ids[0]
	if _, err := r.AddLine(ctx, &model.Line{WishlistID: dst, ProductID: 1, Title: "A", Quantity: 5}); err != nil {
		t.Fatal(err)
	}
	moved, err := r.MoveLine(ctx, src.ID, src.Lines[0].ID, dst)
	if err != nil || moved.WishlistID != dst || moved.Quantity != 7 {
		t.Fatalf("move merges into existing line: %+v %v", moved, err)
	}
	if _, err := r.MoveLine(ctx, src.ID, src.Lines[0].ID, dst); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("moving an already moved line: %v", err)
	}
	if _, err := r.MoveLine(ctx, src.ID, src.Lines[1].ID, 999999); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("move to missing wishlist: %v", err)
	}
	still, _ := r.Get(ctx, src.ID)
	if len(still.Lines) != 1 {
		t.Fatalf("failed move must roll back: %+v", still.Lines)
	}

	found, err := r.FindProduct(ctx, "10", 1)
	if err != nil || len(found) != 1 || found[0].WishlistID != dst || found[0].Quantity != 7 || found[0].WishlistName != model.DefaultName {
		t.Fatalf("find product: %+v %v", found, err)
	}
	if found, _ := r.FindProduct(ctx, "11", 1); len(found) != 0 {
		t.Fatalf("other owner: %+v", found)
	}

	other, _ := model.NewWishlist("11", "Other", model.VisibilityPrivate)
	other.Lines = []model.Line{{ProductID: 2, Title: "B", Quantity: 4}}
	_ = r.Create(ctx, other)
	top, err := r.PopularProducts(ctx, 10)
	if err != nil || len(top) != 2 || top[0].ProductID != 2 || top[0].WishlistCount != 2 || top[0].TotalQuantity != 5 {
		t.Fatalf("popular: %+v %v", top, err)
	}

	n, err := r.ClearLines(ctx, src.ID)
	if err != nil || n != 1 {
		t.Fatalf("clear: %d %v", n, err)
	}
	if n, _ := r.ClearLines(ctx, src.ID); n != 0 {
		t.Fatalf("clear empty: %d", n)
	}
}
