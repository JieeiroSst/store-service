package crypto

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/pbkdf2"
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"

	"github.com/JIeeiroSst/card-service/internal/domain"
)

const (
	pinIterations = 210_000
	pinSaltLen    = 16
	pinKeyLen     = 32
	pinScheme     = "pbkdf2-sha256"
	cvvDigits     = 3
)

type Keys struct {
	encryption []byte
	panHash    []byte
	cvv        []byte
}

func DeriveKeys(master []byte) Keys {
	return Keys{
		encryption: derive(master, "card-service/pan-encryption"),
		panHash:    derive(master, "card-service/pan-hash"),
		cvv:        derive(master, "card-service/cvv"),
	}
}

func derive(master []byte, label string) []byte {
	m := hmac.New(sha256.New, master)
	m.Write([]byte(label))
	return m.Sum(nil)
}

type Vault struct {
	aead    cipher.AEAD
	hashKey []byte
}

func NewVault(keys Keys) (*Vault, error) {
	block, err := aes.NewCipher(keys.encryption)
	if err != nil {
		return nil, fmt.Errorf("pan cipher: %w", err)
	}
	aead, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("pan cipher: %w", err)
	}
	return &Vault{aead: aead, hashKey: keys.panHash}, nil
}

func (v *Vault) HashPAN(pan string) string {
	m := hmac.New(sha256.New, v.hashKey)
	m.Write([]byte(pan))
	return hex.EncodeToString(m.Sum(nil))
}

func (v *Vault) EncryptPAN(pan string) ([]byte, error) {
	nonce := make([]byte, v.aead.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}
	return v.aead.Seal(nonce, nonce, []byte(pan), nil), nil
}

func (v *Vault) DecryptPAN(sealed []byte) (string, error) {
	n := v.aead.NonceSize()
	if len(sealed) < n {
		return "", errors.New("pan ciphertext is too short")
	}
	plain, err := v.aead.Open(nil, sealed[:n], sealed[n:], nil)
	if err != nil {
		return "", fmt.Errorf("decrypt pan: %w", err)
	}
	return string(plain), nil
}

type Security struct {
	cvvKey []byte
}

func NewSecurity(keys Keys) *Security {
	return &Security{cvvKey: keys.cvv}
}

func (s *Security) CVV(pan string, expiry domain.Expiry, serviceCode string) string {
	m := hmac.New(sha256.New, s.cvvKey)
	m.Write([]byte(pan + "|" + expiry.YYMM() + "|" + serviceCode))
	sum := m.Sum(nil)
	out := make([]byte, 0, cvvDigits)
	for _, b := range sum {
		if b >= 250 {
			continue
		}
		out = append(out, '0'+b%10)
		if len(out) == cvvDigits {
			break
		}
	}
	return string(out)
}

func (s *Security) HashPIN(pin string) (string, error) {
	salt := make([]byte, pinSaltLen)
	if _, err := rand.Read(salt); err != nil {
		return "", err
	}
	key, err := pbkdf2.Key(sha256.New, pin, salt, pinIterations, pinKeyLen)
	if err != nil {
		return "", err
	}
	return strings.Join([]string{
		pinScheme,
		strconv.Itoa(pinIterations),
		base64.RawStdEncoding.EncodeToString(salt),
		base64.RawStdEncoding.EncodeToString(key),
	}, "$"), nil
}

func (s *Security) VerifyPIN(pin, encoded string) bool {
	parts := strings.Split(encoded, "$")
	if len(parts) != 4 || parts[0] != pinScheme {
		return false
	}
	iter, err := strconv.Atoi(parts[1])
	if err != nil || iter <= 0 {
		return false
	}
	salt, err := base64.RawStdEncoding.DecodeString(parts[2])
	if err != nil {
		return false
	}
	want, err := base64.RawStdEncoding.DecodeString(parts[3])
	if err != nil {
		return false
	}
	got, err := pbkdf2.Key(sha256.New, pin, salt, iter, len(want))
	if err != nil {
		return false
	}
	return subtle.ConstantTimeCompare(got, want) == 1
}
