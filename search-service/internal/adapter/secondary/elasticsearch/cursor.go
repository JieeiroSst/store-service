package elasticsearch

import (
	"bytes"
	"crypto/sha1"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"strings"

	"github.com/JIeeiroSst/search-service/internal/domain"
)

type cursor struct {
	Kind        string `json:"k"`
	Fingerprint string `json:"f"`
	SearchAfter []any  `json:"a"`
}

func encodeCursor(c cursor) (string, error) {
	data, err := json.Marshal(c)
	if err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(data), nil
}

func decodeCursor(raw, kind, fingerprint string) (*cursor, error) {
	if raw == "" {
		return nil, nil
	}
	data, err := base64.RawURLEncoding.DecodeString(raw)
	if err != nil {
		return nil, domain.Invalid("invalid cursor")
	}
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.UseNumber()
	var c cursor
	if err := dec.Decode(&c); err != nil || len(c.SearchAfter) == 0 {
		return nil, domain.Invalid("invalid cursor")
	}
	if c.Kind != kind || c.Fingerprint != fingerprint {
		return nil, domain.Invalid("cursor does not match the request parameters")
	}
	return &c, nil
}

func fingerprintOf(kind, query string, indices []string) string {
	sum := sha1.Sum([]byte(kind + "\x00" + query + "\x00" + strings.Join(indices, ",")))
	return hex.EncodeToString(sum[:8])
}
