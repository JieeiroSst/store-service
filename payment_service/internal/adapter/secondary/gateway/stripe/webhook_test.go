package stripe

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"strconv"
	"testing"
	"time"
)

func stripeHeader(secret string, ts int64, payload []byte) map[string][]string {
	t := strconv.FormatInt(ts, 10)
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(t + "." + string(payload)))
	return map[string][]string{"Stripe-Signature": {"t=" + t + ",v1=" + hex.EncodeToString(mac.Sum(nil))}}
}

func TestParseWebhookSignatureChecks(t *testing.T) {
	payload := []byte(`{"type":"payment_intent.succeeded","data":{"object":{"id":"pi_1","status":"succeeded"}}}`)
	g := &gateway{webhookSecret: "whsec"}
	ctx := context.Background()

	if _, err := g.ParseWebhook(ctx, payload, stripeHeader("whsec", time.Now().Unix(), payload)); err != nil {
		t.Fatalf("fresh signed event rejected: %v", err)
	}
	if _, err := g.ParseWebhook(ctx, payload, stripeHeader("whsec", time.Now().Add(-time.Hour).Unix(), payload)); err == nil {
		t.Fatal("replayed hour-old event accepted")
	}
	if _, err := g.ParseWebhook(ctx, payload, stripeHeader("other", time.Now().Unix(), payload)); err == nil {
		t.Fatal("event signed with the wrong secret accepted")
	}
	unconfigured := &gateway{}
	if _, err := unconfigured.ParseWebhook(ctx, payload, stripeHeader("", time.Now().Unix(), payload)); err == nil {
		t.Fatal("event accepted with an empty webhook secret")
	}
}
