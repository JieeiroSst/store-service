package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math"

	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

// The cursor a client sees is an opaque token. It only says where a page ended, so it holds no
// secret and needs no signature: a forged one can move the client's own position, nothing more.
type cursorToken struct {
	Sort   string  `json:"s"`
	ID     int64   `json:"i"`
	Rating float64 `json:"r,omitempty"`
	Count  int     `json:"c,omitempty"`
	Score  float64 `json:"x,omitempty"`
}

func encodeCursor(c *domain.HotelCursor) string {
	if c == nil {
		return ""
	}
	b, _ := json.Marshal(cursorToken{Sort: c.Sort, ID: c.ID, Rating: c.Rating, Count: c.Count, Score: c.Score})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeCursor(s string) (*domain.HotelCursor, error) {
	bad := fmt.Errorf("%w: invalid cursor", domain.ErrInvalid)
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return nil, bad
	}
	var t cursorToken
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&t); err != nil {
		return nil, bad
	}
	switch t.Sort {
	case "recommended", "rating", "newest", "oldest":
	default:
		return nil, bad
	}
	if t.ID <= 0 || t.Count < 0 || math.IsNaN(t.Rating) || math.IsInf(t.Rating, 0) || math.IsNaN(t.Score) || math.IsInf(t.Score, 0) {
		return nil, bad
	}
	return &domain.HotelCursor{Sort: t.Sort, ID: t.ID, Rating: t.Rating, Count: t.Count, Score: t.Score}, nil
}

// Lists ordered by id (reservations, reviews, notifications, ...) page on the id of the last item.
type idCursorToken struct {
	ID int64 `json:"i"`
}

func encodeIDCursor(id int64) string {
	if id == 0 {
		return ""
	}
	b, _ := json.Marshal(idCursorToken{ID: id})
	return base64.RawURLEncoding.EncodeToString(b)
}

func decodeIDCursor(s string) (int64, error) {
	if s == "" {
		return 0, nil
	}
	raw, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return 0, fmt.Errorf("%w: invalid cursor", domain.ErrInvalid)
	}
	var t idCursorToken
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&t); err != nil || t.ID <= 0 {
		return 0, fmt.Errorf("%w: invalid cursor", domain.ErrInvalid)
	}
	return t.ID, nil
}
