package repository

import (
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
)

type cursor struct {
	S  string `json:"s"`
	ID int64  `json:"i"`
	N  int64  `json:"n,omitempty"`
	T  int64  `json:"t,omitempty"`
}

func (c cursor) encode() string {
	raw, _ := json.Marshal(c)
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeCursor(raw, sort string) (*cursor, error) {
	if raw == "" {
		return nil, nil
	}
	b, err := base64.RawURLEncoding.DecodeString(raw)
	var c cursor
	if err == nil {
		err = json.Unmarshal(b, &c)
	}
	if err != nil || c.ID <= 0 || c.S != sort {
		return nil, fmt.Errorf("%w: invalid cursor", port.ErrInvalidInput)
	}
	return &c, nil
}

func paginate[T any](rows []T, limit int, next func(last T) cursor) *port.Page[T] {
	page := &port.Page[T]{Items: rows, IsLastPage: true}
	if len(rows) > limit {
		page.Items = rows[:limit]
		page.IsLastPage = false
		page.NextCursor = next(page.Items[limit-1]).encode()
	}
	if page.Items == nil {
		page.Items = []T{}
	}
	return page
}
