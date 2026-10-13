package model

import (
	"errors"
	"strings"
	"testing"
)

func TestParseVisibility(t *testing.T) {
	cases := map[string]Visibility{"": VisibilityPrivate, "private": VisibilityPrivate, "Shared": VisibilityShared, " PUBLIC ": VisibilityPublic}
	for in, want := range cases {
		got, err := ParseVisibility(in)
		if err != nil || got != want {
			t.Fatalf("ParseVisibility(%q) = %q, %v; want %q", in, got, err, want)
		}
	}
	if _, err := ParseVisibility("friends"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid, got %v", err)
	}
}

func TestNewWishlist(t *testing.T) {
	w, err := NewWishlist("7", "  ", VisibilityPrivate)
	if err != nil {
		t.Fatal(err)
	}
	if w.Name != DefaultName || len(w.Key) != KeyLength || w.OwnerID != "7" {
		t.Fatalf("unexpected wishlist %+v", w)
	}
	other, _ := NewWishlist("7", "x", VisibilityPrivate)
	if other.Key == w.Key {
		t.Fatal("keys should be random")
	}
	if _, err := NewWishlist("", "x", VisibilityPrivate); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid for missing owner, got %v", err)
	}
	if _, err := NewWishlist("7", strings.Repeat("a", 256), VisibilityPrivate); !errors.Is(err, ErrInvalid) {
		t.Fatalf("expected ErrInvalid for long name, got %v", err)
	}
}

func TestVisibility(t *testing.T) {
	owner := Actor{UserID: "7"}
	stranger := Actor{UserID: "8"}
	admin := Actor{Admin: true}
	anon := Actor{}
	for _, tc := range []struct {
		v     Visibility
		actor Actor
		byID  bool
		byKey bool
	}{
		{VisibilityPrivate, owner, true, true},
		{VisibilityPrivate, stranger, false, false},
		{VisibilityPrivate, admin, true, true},
		{VisibilityShared, stranger, false, true},
		{VisibilityShared, anon, false, true},
		{VisibilityPublic, anon, true, true},
	} {
		w := &Wishlist{OwnerID: "7", Visibility: tc.v}
		if got := w.VisibleTo(tc.actor); got != tc.byID {
			t.Errorf("%s/%+v VisibleTo = %v", tc.v, tc.actor, got)
		}
		if got := w.VisibleByKey(tc.actor); got != tc.byKey {
			t.Errorf("%s/%+v VisibleByKey = %v", tc.v, tc.actor, got)
		}
	}
}

func TestNewLine(t *testing.T) {
	l, err := NewLine(1, 42, " Book ", 0)
	if err != nil {
		t.Fatal(err)
	}
	if l.Quantity != 1 || l.Title != "Book" {
		t.Fatalf("unexpected line %+v", l)
	}
	for _, tc := range []struct {
		product int64
		title   string
		qty     int
	}{
		{0, "x", 1},
		{1, "", 1},
		{1, "x", -1},
		{1, "x", MaxQuantity + 1},
	} {
		if _, err := NewLine(1, tc.product, tc.title, tc.qty); !errors.Is(err, ErrInvalid) {
			t.Errorf("NewLine(%d, %q, %d) expected ErrInvalid, got %v", tc.product, tc.title, tc.qty, err)
		}
	}
}
