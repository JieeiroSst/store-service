package http

import (
	"encoding/base64"
	"errors"
	"testing"

	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

func TestIDCursorRoundTrip(t *testing.T) {
	if encodeIDCursor(0) != "" {
		t.Fatal("no next page means no cursor")
	}
	for _, id := range []int64{1, 42, 1 << 40} {
		got, err := decodeIDCursor(encodeIDCursor(id))
		if err != nil || got != id {
			t.Fatalf("%d -> %d, %v", id, got, err)
		}
	}
	if id, err := decodeIDCursor(""); id != 0 || err != nil {
		t.Fatal("an empty cursor is the first page")
	}
}

func TestIDCursorRejectsGarbage(t *testing.T) {
	for _, bad := range []string{
		"!!!",
		base64.RawURLEncoding.EncodeToString([]byte(`{"i":-5}`)),
		base64.RawURLEncoding.EncodeToString([]byte(`{"i":0}`)),
		base64.RawURLEncoding.EncodeToString([]byte(`{"i":1,"x":2}`)),
		base64.RawURLEncoding.EncodeToString([]byte(`nope`)),
	} {
		if _, err := decodeIDCursor(bad); !errors.Is(err, domain.ErrInvalid) {
			t.Errorf("%q: %v", bad, err)
		}
	}
}
