package otp

import (
	"errors"
	"testing"
	"time"
)

func TestGenerateAndAuthorize(t *testing.T) {
	o := NewOtp("unit-test-secret")
	got, err := o.CreateOtpByUser("alice")
	if err != nil {
		t.Fatal(err)
	}
	if len(got.OTP) != 6 {
		t.Fatalf("otp %q", got.OTP)
	}
	if err := o.Authorize(got.OTP, "alice"); err != nil {
		t.Fatalf("valid code rejected: %v", err)
	}
	if err := o.Authorize(got.OTP, "ALICE"); err != nil {
		t.Fatalf("usernames are case-insensitive: %v", err)
	}
	wrong := "000000"
	if got.OTP == wrong {
		wrong = "111111"
	}
	if err := o.Authorize(wrong, "alice"); !errors.Is(err, ErrInvalid) {
		t.Fatalf("wrong code must be rejected: %v", err)
	}
	other, _ := NewOtp("another-secret").CreateOtpByUser("alice")
	if other.OTP != got.OTP {
		if err := o.Authorize(other.OTP, "alice"); !errors.Is(err, ErrInvalid) {
			t.Fatalf("code from another secret must be rejected: %v", err)
		}
	}
}

func TestExpiresAt(t *testing.T) {
	now := time.Unix(1_800_000_007, 0)
	if got := ExpiresAt(now); got.Unix() != 1_800_000_060 {
		t.Fatalf("expires at %d", got.Unix())
	}
	for i := int64(0); i < 30; i++ {
		ttl := ExpiresAt(time.Unix(1_800_000_000+i, 0)).Unix() - (1_800_000_000 + i)
		if ttl <= 30 || ttl > 60 {
			t.Fatalf("ttl %d out of (30,60]", ttl)
		}
	}
}
