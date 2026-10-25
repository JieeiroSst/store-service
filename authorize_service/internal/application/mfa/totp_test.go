package mfa

import (
	"strings"
	"testing"
	"time"
)

func TestHOTPMatchesRFC6238Vectors(t *testing.T) {
	key := []byte("12345678901234567890")
	cases := []struct {
		unix int64
		want string
	}{
		{59, "287082"},
		{1111111109, "081804"},
		{1111111111, "050471"},
		{1234567890, "005924"},
		{2000000000, "279037"},
	}
	for _, c := range cases {
		if got := hotp(key, timeStep(time.Unix(c.unix, 0))); got != c.want {
			t.Errorf("t=%d: got %s want %s", c.unix, got, c.want)
		}
	}
}

func TestMatchStepSkewAndReplay(t *testing.T) {
	key := []byte("12345678901234567890")
	now := time.Unix(1111111111, 0)
	cur := timeStep(now)

	if step, ok := matchStep(key, hotp(key, cur-1), now, 0); !ok || step != cur-1 {
		t.Fatalf("previous step should match, got %d %v", step, ok)
	}
	if _, ok := matchStep(key, hotp(key, cur-2), now, 0); ok {
		t.Fatal("step outside skew must not match")
	}
	if _, ok := matchStep(key, hotp(key, cur), now, cur); ok {
		t.Fatal("already used step must not match")
	}
}

func TestSecretRoundTripAndURI(t *testing.T) {
	secret, err := newSecret()
	if err != nil {
		t.Fatal(err)
	}
	key, err := decodeSecret(secret)
	if err != nil || len(key) != secretBytes {
		t.Fatalf("decode secret: %v len=%d", err, len(key))
	}
	uri := otpauthURI("Keiki Store", "alice@example.com", secret)
	if !strings.HasPrefix(uri, "otpauth://totp/Keiki%20Store:alice@example.com?") || !strings.Contains(uri, "secret="+secret) || !strings.Contains(uri, "issuer=Keiki+Store") {
		t.Errorf("unexpected uri %s", uri)
	}
}

func TestRecoveryCodesConsumeOnce(t *testing.T) {
	plain, stored, err := newRecoveryCodes()
	if err != nil {
		t.Fatal(err)
	}
	if len(plain) != recoveryCodeCount {
		t.Fatalf("got %d codes", len(plain))
	}
	rest, ok := consumeRecoveryCode(stored, strings.ToUpper(plain[3]))
	if !ok || len(parseRecoveryCodes(rest)) != recoveryCodeCount-1 {
		t.Fatalf("consume failed ok=%v", ok)
	}
	if _, ok := consumeRecoveryCode(rest, plain[3]); ok {
		t.Fatal("recovery code reused")
	}
}
