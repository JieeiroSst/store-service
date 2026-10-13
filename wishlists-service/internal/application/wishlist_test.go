package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/wishlists-service/internal/adapter/secondary/memory"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
)

var (
	ctx    = context.Background()
	alice  = model.Actor{UserID: "1"}
	bob    = model.Actor{UserID: "2"}
	admin  = model.Actor{Admin: true}
	anon   = model.Actor{}
	strPtr = func(s string) *string { return &s }
)

func directory() *memory.Users {
	return memory.NewUsers(
		model.Owner{ID: "1", Username: "alice", Name: "Alice", Active: true},
		model.Owner{ID: "2", Username: "bob", Name: "Bob", Active: true},
		model.Owner{ID: "3", Username: "carol", Active: false},
	)
}

func newService(repo port.WishlistRepository) port.WishlistUsecase {
	return NewWishlistService(repo, directory())
}

func TestCreate(t *testing.T) {
	svc := newService(memory.New())
	w, err := svc.Create(ctx, alice, port.CreateWishlistInput{Name: "Birthday", Visibility: "shared"})
	if err != nil {
		t.Fatal(err)
	}
	if w.ID == 0 || w.OwnerID != "1" || w.Visibility != model.VisibilityShared || w.Key == "" {
		t.Fatalf("unexpected wishlist %+v", w)
	}
	if _, err := svc.Create(ctx, alice, port.CreateWishlistInput{OwnerID: "2"}); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("creating for another user should be forbidden, got %v", err)
	}
	w, err = svc.Create(ctx, admin, port.CreateWishlistInput{OwnerID: "2"})
	if err != nil || w.OwnerID != "2" || w.Name != model.DefaultName {
		t.Fatalf("admin create for user: %+v %v", w, err)
	}
	if _, err := svc.Create(ctx, admin, port.CreateWishlistInput{}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("service caller without owner_id should be invalid, got %v", err)
	}
	if _, err := svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "friends"}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("bad visibility should be invalid, got %v", err)
	}
}

func TestCreateRetriesKeyConflicts(t *testing.T) {
	repo := memory.New()
	repo.Conflicts = 2
	svc := newService(repo)
	if _, err := svc.Create(ctx, alice, port.CreateWishlistInput{}); err != nil {
		t.Fatalf("expected retry to succeed, got %v", err)
	}
	repo.Conflicts = keyAttempts
	if _, err := svc.Create(ctx, alice, port.CreateWishlistInput{}); !errors.Is(err, model.ErrConflict) {
		t.Fatalf("expected conflict after exhausting attempts, got %v", err)
	}
}

func TestReadAccess(t *testing.T) {
	svc := newService(memory.New())
	private, _ := svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "private"})
	shared, _ := svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "shared"})
	public, _ := svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "public"})

	if _, err := svc.Get(ctx, alice, private.ID); err != nil {
		t.Fatalf("owner get: %v", err)
	}
	if _, err := svc.Get(ctx, bob, private.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("stranger get private: %v", err)
	}
	if _, err := svc.Get(ctx, bob, shared.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("stranger get shared by id: %v", err)
	}
	if _, err := svc.Get(ctx, anon, public.ID); err != nil {
		t.Fatalf("anon get public: %v", err)
	}
	if _, err := svc.GetByKey(ctx, anon, shared.Key); err != nil {
		t.Fatalf("anon get shared by key: %v", err)
	}
	if _, err := svc.GetByKey(ctx, bob, private.Key); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("stranger get private by key: %v", err)
	}
	if _, err := svc.GetByKey(ctx, anon, ""); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("empty key: %v", err)
	}
}

func TestList(t *testing.T) {
	svc := newService(memory.New())
	_, _ = svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "private"})
	_, _ = svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "public"})
	_, _ = svc.Create(ctx, bob, port.CreateWishlistInput{})

	mine, err := svc.List(ctx, alice, port.ListWishlistsInput{})
	if err != nil || len(mine.Items) != 2 || mine.Total != 2 || mine.Limit != defaultPageSize {
		t.Fatalf("alice own list: %+v %v", mine, err)
	}
	if mine.Items[0].Owner == nil || mine.Items[0].Owner.Username != "alice" {
		t.Fatalf("owner should be enriched from user-service: %+v", mine.Items[0].Owner)
	}
	theirs, err := svc.List(ctx, bob, port.ListWishlistsInput{OwnerID: "1"})
	if err != nil || len(theirs.Items) != 1 || theirs.Items[0].Visibility != model.VisibilityPublic {
		t.Fatalf("bob listing alice: %+v %v", theirs, err)
	}
	public, err := svc.List(ctx, anon, port.ListWishlistsInput{OwnerID: "1"})
	if err != nil || len(public.Items) != 1 {
		t.Fatalf("anon listing alice: %+v %v", public, err)
	}
	if _, err := svc.List(ctx, anon, port.ListWishlistsInput{}); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("anon list all: %v", err)
	}
	all, err := svc.List(ctx, admin, port.ListWishlistsInput{Limit: 1000})
	if err != nil || len(all.Items) != 3 || all.Limit != maxPageSize {
		t.Fatalf("admin list all: %+v %v", all, err)
	}
	page, _ := svc.List(ctx, admin, port.ListWishlistsInput{Limit: 1, Offset: 1})
	if len(page.Items) != 1 || page.Items[0].ID != all.Items[1].ID || page.Total != 3 {
		t.Fatalf("pagination: %+v", page)
	}
}

func TestUpdateAndDelete(t *testing.T) {
	svc := newService(memory.New())
	w, _ := svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "public"})

	if _, err := svc.Update(ctx, bob, w.ID, port.UpdateWishlistInput{Name: strPtr("hijack")}); !errors.Is(err, model.ErrForbidden) {
		t.Fatalf("stranger update public: %v", err)
	}
	updated, err := svc.Update(ctx, alice, w.ID, port.UpdateWishlistInput{Name: strPtr("Xmas"), Visibility: strPtr("private"), RegenerateKey: true})
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Xmas" || updated.Visibility != model.VisibilityPrivate || updated.Key == w.Key {
		t.Fatalf("unexpected update result %+v", updated)
	}
	if _, err := svc.Update(ctx, bob, w.ID, port.UpdateWishlistInput{}); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("stranger update private: %v", err)
	}
	if err := svc.Delete(ctx, bob, w.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("stranger delete: %v", err)
	}
	if err := svc.Delete(ctx, alice, w.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Get(ctx, alice, w.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("get after delete: %v", err)
	}
}

func TestLines(t *testing.T) {
	svc := newService(memory.New())
	w, _ := svc.Create(ctx, alice, port.CreateWishlistInput{})

	l, err := svc.AddLine(ctx, alice, w.ID, port.AddLineInput{ProductID: 42, Title: "Book"})
	if err != nil || l.Quantity != 1 {
		t.Fatalf("add line: %+v %v", l, err)
	}
	again, err := svc.AddLine(ctx, alice, w.ID, port.AddLineInput{ProductID: 42, Title: "Book", Quantity: 2})
	if err != nil || again.ID != l.ID || again.Quantity != 3 {
		t.Fatalf("re-adding a product should increase quantity: %+v %v", again, err)
	}
	if _, err := svc.AddLine(ctx, bob, w.ID, port.AddLineInput{ProductID: 1, Title: "x"}); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("stranger add line: %v", err)
	}
	if _, err := svc.AddLine(ctx, alice, w.ID, port.AddLineInput{Title: "x"}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("missing product: %v", err)
	}
	upd, err := svc.UpdateLine(ctx, alice, w.ID, l.ID, 5)
	if err != nil || upd.Quantity != 5 {
		t.Fatalf("update line: %+v %v", upd, err)
	}
	if _, err := svc.UpdateLine(ctx, alice, w.ID, l.ID, 0); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("zero quantity: %v", err)
	}
	if _, err := svc.UpdateLine(ctx, alice, w.ID, 9999, 1); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("unknown line: %v", err)
	}
	if err := svc.RemoveLine(ctx, alice, w.ID, l.ID); err != nil {
		t.Fatal(err)
	}
	got, _ := svc.Get(ctx, alice, w.ID)
	if len(got.Lines) != 0 {
		t.Fatalf("line should be removed: %+v", got.Lines)
	}
}

func TestCreateForOtherUserChecksUserService(t *testing.T) {
	users := directory()
	svc := NewWishlistService(memory.New(), users)
	w, err := svc.Create(ctx, admin, port.CreateWishlistInput{OwnerID: "2"})
	if err != nil || w.Owner == nil || w.Owner.Name != "Bob" {
		t.Fatalf("admin create for bob: %+v %v", w, err)
	}
	var ve *model.ValidationError
	if _, err := svc.Create(ctx, admin, port.CreateWishlistInput{OwnerID: "99"}); !errors.As(err, &ve) || ve.Field != "owner_id" {
		t.Fatalf("unknown owner: %v", err)
	}
	if _, err := svc.Create(ctx, admin, port.CreateWishlistInput{OwnerID: "3"}); !errors.As(err, &ve) {
		t.Fatalf("inactive owner: %v", err)
	}
	users.Err = model.ErrUnavailable
	if _, err := svc.Create(ctx, admin, port.CreateWishlistInput{OwnerID: "2"}); !errors.Is(err, model.ErrUnavailable) {
		t.Fatalf("user-service down: %v", err)
	}
	w, err = svc.Create(ctx, alice, port.CreateWishlistInput{})
	if err != nil || w.Owner != nil {
		t.Fatalf("own create must not depend on user-service: %+v %v", w, err)
	}
}

func TestDefaultAndQuickAdd(t *testing.T) {
	svc := newService(memory.New())
	d, err := svc.Default(ctx, alice)
	if err != nil || d.Name != model.DefaultName || d.Visibility != model.VisibilityPrivate {
		t.Fatalf("default: %+v %v", d, err)
	}
	again, _ := svc.Default(ctx, alice)
	if again.ID != d.ID {
		t.Fatalf("default should be stable: %d vs %d", again.ID, d.ID)
	}
	l, err := svc.AddToDefault(ctx, alice, port.AddLineInput{ProductID: 5, Title: "Mug"})
	if err != nil || l.WishlistID != d.ID {
		t.Fatalf("quick add: %+v %v", l, err)
	}
	if _, err := svc.Default(ctx, admin); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("service default: %v", err)
	}
	found, err := svc.FindProduct(ctx, alice, 5)
	if err != nil || len(found) != 1 || found[0].WishlistID != d.ID || found[0].LineID != l.ID {
		t.Fatalf("find product: %+v %v", found, err)
	}
	none, _ := svc.FindProduct(ctx, bob, 5)
	if len(none) != 0 {
		t.Fatalf("bob should not see alice's product: %+v", none)
	}
	if _, err := svc.FindProduct(ctx, alice, 0); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("bad product: %v", err)
	}
}

func TestCopyShared(t *testing.T) {
	svc := newService(memory.New())
	src, _ := svc.Create(ctx, alice, port.CreateWishlistInput{Name: "Wedding", Visibility: "shared"})
	_, _ = svc.AddLine(ctx, alice, src.ID, port.AddLineInput{ProductID: 1, Title: "Plate", Quantity: 4})
	priv, _ := svc.Create(ctx, alice, port.CreateWishlistInput{})

	cp, err := svc.CopyShared(ctx, bob, src.Key)
	if err != nil {
		t.Fatal(err)
	}
	if cp.OwnerID != "2" || cp.Name != "Copy of Wedding" || cp.Visibility != model.VisibilityPrivate || len(cp.Lines) != 1 || cp.Lines[0].Quantity != 4 {
		t.Fatalf("unexpected copy %+v", cp)
	}
	if cp.Key == src.Key {
		t.Fatal("copy must get its own key")
	}
	if _, err := svc.CopyShared(ctx, bob, priv.Key); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("copying private list: %v", err)
	}
	if _, err := svc.CopyShared(ctx, anon, src.Key); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("anonymous copy: %v", err)
	}
}

func TestMoveAndClear(t *testing.T) {
	svc := newService(memory.New())
	a, _ := svc.Create(ctx, alice, port.CreateWishlistInput{})
	b, _ := svc.Create(ctx, alice, port.CreateWishlistInput{})
	other, _ := svc.Create(ctx, bob, port.CreateWishlistInput{})
	l, _ := svc.AddLine(ctx, alice, a.ID, port.AddLineInput{ProductID: 9, Title: "Lamp", Quantity: 2})
	_, _ = svc.AddLine(ctx, alice, b.ID, port.AddLineInput{ProductID: 9, Title: "Lamp", Quantity: 1})

	if _, err := svc.MoveLine(ctx, alice, a.ID, l.ID, other.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("move into someone else's list: %v", err)
	}
	if _, err := svc.MoveLine(ctx, alice, a.ID, l.ID, a.ID); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("move into same list: %v", err)
	}
	if _, err := svc.MoveLine(ctx, alice, a.ID, 12345, b.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("move unknown line: %v", err)
	}
	moved, err := svc.MoveLine(ctx, alice, a.ID, l.ID, b.ID)
	if err != nil || moved.WishlistID != b.ID || moved.Quantity != 3 {
		t.Fatalf("move should merge quantities: %+v %v", moved, err)
	}
	got, _ := svc.Get(ctx, alice, a.ID)
	if len(got.Lines) != 0 {
		t.Fatalf("source should be empty: %+v", got.Lines)
	}
	n, err := svc.ClearLines(ctx, alice, b.ID)
	if err != nil || n != 1 {
		t.Fatalf("clear: %d %v", n, err)
	}
	if _, err := svc.ClearLines(ctx, bob, b.ID); !errors.Is(err, model.ErrNotFound) {
		t.Fatalf("stranger clear: %v", err)
	}
}

func TestLineLimitAndPopular(t *testing.T) {
	svc := newService(memory.New())
	w, _ := svc.Create(ctx, alice, port.CreateWishlistInput{})
	for i := 1; i <= model.MaxLines; i++ {
		if _, err := svc.AddLine(ctx, alice, w.ID, port.AddLineInput{ProductID: int64(i), Title: "p"}); err != nil {
			t.Fatalf("line %d: %v", i, err)
		}
	}
	if _, err := svc.AddLine(ctx, alice, w.ID, port.AddLineInput{ProductID: 100000, Title: "p"}); !errors.Is(err, model.ErrInvalid) {
		t.Fatalf("over the cap: %v", err)
	}
	if _, err := svc.AddLine(ctx, alice, w.ID, port.AddLineInput{ProductID: 1, Title: "p"}); err != nil {
		t.Fatalf("existing product at the cap should still increment: %v", err)
	}
	b, _ := svc.Create(ctx, bob, port.CreateWishlistInput{})
	_, _ = svc.AddLine(ctx, bob, b.ID, port.AddLineInput{ProductID: 7, Title: "p"})
	top, err := svc.PopularProducts(ctx, 0)
	if err != nil || len(top) != defaultPopularSize || top[0].ProductID != 7 || top[0].WishlistCount != 2 {
		t.Fatalf("popular: %+v %v", top, err)
	}
}

func TestShareKeyHiddenFromNonOwners(t *testing.T) {
	svc := newService(memory.New())
	pub, _ := svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "public"})
	if pub.Key == "" {
		t.Fatal("owner must see the key")
	}
	got, _ := svc.Get(ctx, bob, pub.ID)
	if got.Key != "" {
		t.Fatalf("stranger must not receive the share key: %q", got.Key)
	}
	page, _ := svc.List(ctx, anon, port.ListWishlistsInput{OwnerID: "1"})
	if page.Items[0].Key != "" {
		t.Fatalf("public listing leaked the key: %q", page.Items[0].Key)
	}
	mine, _ := svc.Get(ctx, alice, pub.ID)
	if mine.Key != pub.Key {
		t.Fatal("owner read lost the key")
	}
}

func TestEnrichmentToleratesMissingCredentials(t *testing.T) {
	users := directory()
	users.Err = model.ErrUnauthenticated
	svc := NewWishlistService(memory.New(), users)
	w, _ := svc.Create(ctx, alice, port.CreateWishlistInput{Visibility: "public"})
	got, err := svc.Get(ctx, anon, w.ID)
	if err != nil || got.Owner != nil {
		t.Fatalf("anonymous read without user-service access: %+v %v", got, err)
	}
	if _, err := svc.Create(ctx, admin, port.CreateWishlistInput{OwnerID: "2"}); !errors.Is(err, model.ErrUnauthenticated) {
		t.Fatalf("owner check needs the caller's credentials: %v", err)
	}
}
