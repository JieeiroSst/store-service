package model

import (
	"encoding/base64"
	"encoding/json"
	"time"
)

type Cursor struct {
	CreatedAt time.Time `json:"t"`
	ID        string    `json:"i"`
}

func EncodeCursor(c Cursor) string {
	data, err := json.Marshal(c)
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(data)
}

func DecodeCursor(s string) (Cursor, bool) {
	if s == "" {
		return Cursor{}, false
	}
	data, err := base64.RawURLEncoding.DecodeString(s)
	if err != nil {
		return Cursor{}, false
	}
	var c Cursor
	if err := json.Unmarshal(data, &c); err != nil {
		return Cursor{}, false
	}
	return c, true
}

const (
	DefaultPageLimit = 20
	MaxPageLimit     = 100
)

func ClampLimit(limit int) int {
	switch {
	case limit <= 0:
		return DefaultPageLimit
	case limit > MaxPageLimit:
		return MaxPageLimit
	default:
		return limit
	}
}

type CursorKeyed interface {
	CursorKey() (createdAt time.Time, id string)
}

func PageFromRows[T CursorKeyed](rows []T, limit int) (page []T, nextCursor string) {
	hasMore := len(rows) > limit
	if hasMore {
		rows = rows[:limit]
	}
	if hasMore && len(rows) > 0 {
		createdAt, id := rows[len(rows)-1].CursorKey()
		nextCursor = EncodeCursor(Cursor{CreatedAt: createdAt, ID: id})
	}
	return rows, nextCursor
}

type Page[T any] struct {
	Rows       []T    `json:"rows"`
	NextCursor string `json:"next_cursor,omitempty"`
	HasMore    bool   `json:"has_more"`
}

func NewPage[T any](rows []T, nextCursor string) Page[T] {
	return Page[T]{Rows: rows, NextCursor: nextCursor, HasMore: nextCursor != ""}
}
