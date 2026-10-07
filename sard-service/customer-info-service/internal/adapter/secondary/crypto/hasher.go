package crypto

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"

	"github.com/JIeeiroSst/customer-info-service/config"
	"github.com/JIeeiroSst/customer-info-service/internal/port"
	"go.uber.org/fx"
)

var Module = fx.Options(
	fx.Provide(func(cfg *config.Config) port.DocumentHasher { return NewHasher([]byte(cfg.Security.DocumentHashKey)) }),
)

type Hasher struct{ key []byte }

func NewHasher(key []byte) *Hasher {
	m := hmac.New(sha256.New, key)
	m.Write([]byte("customer-info-service/document-hash"))
	return &Hasher{key: m.Sum(nil)}
}

func (h *Hasher) Hash(documentNumber string) string {
	m := hmac.New(sha256.New, h.key)
	m.Write([]byte(documentNumber))
	return hex.EncodeToString(m.Sum(nil))
}
