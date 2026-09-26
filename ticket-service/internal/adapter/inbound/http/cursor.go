package http

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

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
