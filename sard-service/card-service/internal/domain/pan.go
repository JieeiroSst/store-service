package domain

import (
	"fmt"
	"io"
	"strings"
)

const (
	PANLength = 16
	BINLength = 6
)

func GeneratePAN(bin string, r io.Reader) (string, error) {
	if len(bin) != BINLength || !isDigits(bin) {
		return "", Invalid("bin must be %d digits, got %q", BINLength, bin)
	}
	random, err := randomDigits(r, PANLength-BINLength-1)
	if err != nil {
		return "", fmt.Errorf("generate pan: %w", err)
	}
	payload := bin + random
	return payload + string('0'+LuhnCheckDigit(payload)), nil
}

func NewAuthCode(r io.Reader) (string, error) {
	return randomDigits(r, 6)
}

func LuhnCheckDigit(payload string) byte {
	sum := 0
	double := true
	for i := len(payload) - 1; i >= 0; i-- {
		d := int(payload[i] - '0')
		if double {
			d *= 2
			if d > 9 {
				d -= 9
			}
		}
		sum += d
		double = !double
	}
	return byte((10 - sum%10) % 10)
}

func ValidPAN(pan string) bool {
	if len(pan) != PANLength || !isDigits(pan) {
		return false
	}
	return LuhnCheckDigit(pan[:len(pan)-1]) == pan[len(pan)-1]-'0'
}

func NormalizePAN(s string) string {
	return strings.NewReplacer(" ", "", "-", "").Replace(strings.TrimSpace(s))
}

func MaskPAN(bin, last4 string) string {
	return bin + strings.Repeat("*", PANLength-len(bin)-len(last4)) + last4
}

func randomDigits(r io.Reader, n int) (string, error) {
	out := make([]byte, 0, n)
	buf := make([]byte, n*2)
	for len(out) < n {
		if _, err := io.ReadFull(r, buf); err != nil {
			return "", err
		}
		for _, b := range buf {
			if b >= 250 {
				continue
			}
			out = append(out, '0'+b%10)
			if len(out) == n {
				break
			}
		}
	}
	return string(out), nil
}

func isDigits(s string) bool {
	if s == "" {
		return false
	}
	for i := 0; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}
