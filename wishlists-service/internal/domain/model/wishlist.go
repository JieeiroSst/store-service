package model

import (
	"crypto/rand"
	"strings"
	"time"
	"unicode/utf8"
)

type Visibility string

const (
	VisibilityPrivate Visibility = "Private"
	VisibilityShared  Visibility = "Shared"
	VisibilityPublic  Visibility = "Public"
)

const (
	DefaultName    = "Default"
	MaxNameLength  = 255
	MaxTitleLength = 255
	KeyLength      = 12
	MaxQuantity    = 10000
	MaxLines       = 500
)

const keyAlphabet = "abcdefghijkmnpqrstuvwxyzABCDEFGHJKLMNPQRSTUVWXYZ23456789"

func ParseVisibility(s string) (Visibility, error) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "", "private":
		return VisibilityPrivate, nil
	case "shared":
		return VisibilityShared, nil
	case "public":
		return VisibilityPublic, nil
	}
	return "", invalid("visibility", "must be one of Private, Shared, Public")
}

type Actor struct {
	UserID string
	Admin  bool
}

func (a Actor) Owns(w *Wishlist) bool {
	return a.Admin || (a.UserID != "" && a.UserID == w.OwnerID)
}

type Wishlist struct {
	ID          int64      `json:"id"`
	Name        string     `json:"name"`
	Key         string     `json:"key"`
	Visibility  Visibility `json:"visibility"`
	DateCreated time.Time  `json:"date_created"`
	OwnerID     string     `json:"owner_id"`
	Owner       *Owner     `json:"owner,omitempty"`
	Lines       []Line     `json:"lines"`
}

func NewWishlist(ownerID, name string, visibility Visibility) (*Wishlist, error) {
	w := &Wishlist{OwnerID: strings.TrimSpace(ownerID), Visibility: visibility, Lines: []Line{}}
	if w.OwnerID == "" {
		return nil, invalid("owner_id", "is required")
	}
	if err := w.Rename(name); err != nil {
		return nil, err
	}
	if err := w.RegenerateKey(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *Wishlist) Rename(name string) error {
	name = strings.TrimSpace(name)
	if name == "" {
		name = DefaultName
	}
	if utf8.RuneCountInString(name) > MaxNameLength {
		return invalid("name", "must be at most 255 characters")
	}
	w.Name = name
	return nil
}

func (w *Wishlist) RegenerateKey() error {
	key, err := randomKey(KeyLength)
	if err != nil {
		return err
	}
	w.Key = key
	return nil
}

func (w *Wishlist) CopyFor(ownerID string) (*Wishlist, error) {
	name := []rune("Copy of " + w.Name)
	if len(name) > MaxNameLength {
		name = name[:MaxNameLength]
	}
	c, err := NewWishlist(ownerID, string(name), VisibilityPrivate)
	if err != nil {
		return nil, err
	}
	for _, l := range w.Lines {
		c.Lines = append(c.Lines, Line{ProductID: l.ProductID, Title: l.Title, Quantity: l.Quantity})
	}
	return c, nil
}

func (w *Wishlist) CanAdd(productID int64) error {
	for _, l := range w.Lines {
		if l.ProductID == productID {
			return nil
		}
	}
	if len(w.Lines) >= MaxLines {
		return invalid("lines", "wishlist already holds the maximum of 500 products")
	}
	return nil
}

func (w *Wishlist) VisibleTo(a Actor) bool {
	return w.Visibility == VisibilityPublic || a.Owns(w)
}

func (w *Wishlist) VisibleByKey(a Actor) bool {
	return w.Visibility != VisibilityPrivate || a.Owns(w)
}

type Line struct {
	ID         int64  `json:"id"`
	WishlistID int64  `json:"wishlist_id"`
	ProductID  int64  `json:"product_id"`
	Title      string `json:"title"`
	Quantity   int    `json:"quantity"`
}

func NewLine(wishlistID, productID int64, title string, quantity int) (*Line, error) {
	if productID <= 0 {
		return nil, invalid("product_id", "must be a positive integer")
	}
	title = strings.TrimSpace(title)
	if title == "" {
		return nil, invalid("title", "is required")
	}
	if utf8.RuneCountInString(title) > MaxTitleLength {
		return nil, invalid("title", "must be at most 255 characters")
	}
	if quantity == 0 {
		quantity = 1
	}
	if err := ValidateQuantity(quantity); err != nil {
		return nil, err
	}
	return &Line{WishlistID: wishlistID, ProductID: productID, Title: title, Quantity: quantity}, nil
}

func ValidateQuantity(q int) error {
	if q < 1 || q > MaxQuantity {
		return invalid("quantity", "must be between 1 and 10000")
	}
	return nil
}

func randomKey(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", err
	}
	for i, b := range buf {
		buf[i] = keyAlphabet[int(b)%len(keyAlphabet)]
	}
	return string(buf), nil
}
