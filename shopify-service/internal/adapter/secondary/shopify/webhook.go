package shopify

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"

	"github.com/JIeeiroSst/shopify-service/config"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
)

type WebhookVerifier struct {
	secret []byte
}

func NewWebhookVerifier(cfg *config.Config) port.WebhookVerifier {
	return &WebhookVerifier{secret: []byte(cfg.Shopify.APISecret)}
}

func (v *WebhookVerifier) Verify(payload []byte, signature string) bool {
	if len(v.secret) == 0 || signature == "" {
		return false
	}
	expected, err := base64.StdEncoding.DecodeString(signature)
	if err != nil {
		return false
	}
	mac := hmac.New(sha256.New, v.secret)
	mac.Write(payload)
	return hmac.Equal(mac.Sum(nil), expected)
}
