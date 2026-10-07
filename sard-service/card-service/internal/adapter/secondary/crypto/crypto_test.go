package crypto

import (
	"bytes"
	"testing"

	"github.com/JIeeiroSst/card-service/internal/domain"
)

var keys = DeriveKeys([]byte("unit-test-master-key-0123456789abcdef"))

func TestVault(t *testing.T) {
	v, err := NewVault(keys)
	if err != nil {
		t.Fatal(err)
	}
	pan := "7301001234567897"
	a, _ := v.EncryptPAN(pan)
	b, _ := v.EncryptPAN(pan)
	if bytes.Equal(a, b) {
		t.Fatal("ciphertexts must use a fresh nonce")
	}
	if bytes.Contains(a, []byte(pan)) {
		t.Fatal("pan leaked into ciphertext")
	}
	got, err := v.DecryptPAN(a)
	if err != nil || got != pan {
		t.Fatalf("decrypt = %q, %v", got, err)
	}
	a[len(a)-1] ^= 1
	if _, err := v.DecryptPAN(a); err == nil {
		t.Fatal("tampered ciphertext must not decrypt")
	}
	if v.HashPAN(pan) != v.HashPAN(pan) || v.HashPAN(pan) == v.HashPAN("7301001234567898") {
		t.Fatal("pan hash must be deterministic and distinct")
	}
	other, _ := NewVault(DeriveKeys([]byte("another-master-key-0123456789abcdef")))
	if other.HashPAN(pan) == v.HashPAN(pan) {
		t.Fatal("pan hash must depend on the master key")
	}
}

func TestCVV(t *testing.T) {
	s := NewSecurity(keys)
	e := domain.Expiry{Month: 10, Year: 2031}
	cvv := s.CVV("7301001234567897", e, "201")
	if len(cvv) != 3 || cvv != s.CVV("7301001234567897", e, "201") {
		t.Fatalf("cvv %q must be 3 deterministic digits", cvv)
	}
	diff := 0
	for i := 0; i < 50; i++ {
		if s.CVV("7301001234567897", domain.Expiry{Month: 1 + i%12, Year: 2030 + i/12}, "201") != cvv {
			diff++
		}
	}
	if diff < 40 {
		t.Fatalf("cvv barely changes with expiry: %d/50", diff)
	}
}

func TestPIN(t *testing.T) {
	s := NewSecurity(keys)
	h1, err := s.HashPIN("2580")
	if err != nil {
		t.Fatal(err)
	}
	h2, _ := s.HashPIN("2580")
	if h1 == h2 {
		t.Fatal("pin hashes must be salted")
	}
	if !s.VerifyPIN("2580", h1) || s.VerifyPIN("2581", h1) || s.VerifyPIN("2580", "garbage") {
		t.Fatal("pin verification is wrong")
	}
}
