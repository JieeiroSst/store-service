package mfa

import (
	"crypto/rand"
	"crypto/sha256"
	"encoding/base32"
	"encoding/hex"
	"encoding/json"
	"strings"
)

const recoveryCodeCount = 10

var recoveryEncoding = base32.NewEncoding("abcdefghijkmnpqrstuvwxyz23456789").WithPadding(base32.NoPadding)

func newRecoveryCodes() ([]string, string, error) {
	plain := make([]string, recoveryCodeCount)
	hashes := make([]string, recoveryCodeCount)
	for i := range plain {
		b := make([]byte, 7)
		if _, err := rand.Read(b); err != nil {
			return nil, "", err
		}
		code := recoveryEncoding.EncodeToString(b)[:10]
		plain[i] = code[:5] + "-" + code[5:]
		hashes[i] = hashRecoveryCode(code)
	}
	encoded, err := json.Marshal(hashes)
	if err != nil {
		return nil, "", err
	}
	return plain, string(encoded), nil
}

func normalizeRecoveryCode(code string) string {
	return strings.ToLower(strings.NewReplacer("-", "", " ", "").Replace(code))
}

func hashRecoveryCode(code string) string {
	sum := sha256.Sum256([]byte(normalizeRecoveryCode(code)))
	return hex.EncodeToString(sum[:])
}

func parseRecoveryCodes(stored string) []string {
	var hashes []string
	if stored == "" {
		return nil
	}
	_ = json.Unmarshal([]byte(stored), &hashes)
	return hashes
}

func consumeRecoveryCode(stored, code string) (string, bool) {
	want := hashRecoveryCode(code)
	hashes := parseRecoveryCodes(stored)
	for i, h := range hashes {
		if h == want {
			rest := append(append([]string{}, hashes[:i]...), hashes[i+1:]...)
			encoded, _ := json.Marshal(rest)
			return string(encoded), true
		}
	}
	return "", false
}
