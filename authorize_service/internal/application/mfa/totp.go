package mfa

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha1"
	"crypto/subtle"
	"encoding/base32"
	"encoding/binary"
	"fmt"
	"net/url"
	"strings"
	"time"
)

const (
	totpDigits    = 6
	totpPeriod    = 30
	totpSkew      = 1
	totpAlgorithm = "SHA1"
	secretBytes   = 20
)

var b32 = base32.StdEncoding.WithPadding(base32.NoPadding)

func newSecret() (string, error) {
	b := make([]byte, secretBytes)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return b32.EncodeToString(b), nil
}

func decodeSecret(secret string) ([]byte, error) {
	return b32.DecodeString(strings.ToUpper(strings.TrimRight(secret, "=")))
}

func timeStep(t time.Time) int64 {
	return t.Unix() / totpPeriod
}

func hotp(key []byte, counter int64) string {
	var msg [8]byte
	binary.BigEndian.PutUint64(msg[:], uint64(counter))
	mac := hmac.New(sha1.New, key)
	mac.Write(msg[:])
	sum := mac.Sum(nil)
	off := sum[len(sum)-1] & 0x0f
	bin := binary.BigEndian.Uint32(sum[off:off+4]) & 0x7fffffff
	return fmt.Sprintf("%0*d", totpDigits, bin%1_000_000)
}

func matchStep(key []byte, code string, now time.Time, lastUsed int64) (int64, bool) {
	cur := timeStep(now)
	for d := int64(-totpSkew); d <= totpSkew; d++ {
		step := cur + d
		if step <= lastUsed {
			continue
		}
		if subtle.ConstantTimeCompare([]byte(hotp(key, step)), []byte(code)) == 1 {
			return step, true
		}
	}
	return 0, false
}

func otpauthURI(issuer, account, secret string) string {
	q := url.Values{}
	q.Set("secret", secret)
	q.Set("issuer", issuer)
	q.Set("algorithm", totpAlgorithm)
	q.Set("digits", fmt.Sprint(totpDigits))
	q.Set("period", fmt.Sprint(totpPeriod))
	label := url.PathEscape(issuer) + ":" + url.PathEscape(account)
	return "otpauth://totp/" + label + "?" + q.Encode()
}

func isTOTPCode(code string) bool {
	if len(code) != totpDigits {
		return false
	}
	for _, c := range code {
		if c < '0' || c > '9' {
			return false
		}
	}
	return true
}
