package http_test

import (
	"encoding/json"
	"io"
	"testing"
)

func mustJSON(t *testing.T, r io.Reader, v any) {
	t.Helper()
	if err := json.NewDecoder(r).Decode(v); err != nil {
		t.Fatal(err)
	}
}
