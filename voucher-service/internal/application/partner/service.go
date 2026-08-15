package partner

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"time"
)

var (
	errKeyRevoked      = errors.New("api key revoked")
	errInvalidSignature = errors.New("invalid api key signature")
	errStaleTimestamp  = errors.New("request timestamp outside allowed window")
)

const replayWindow = 5 * time.Minute

type Service struct {
	repo   APIKeyRepository
	encKey []byte // 32 bytes, AES-256-GCM
}

func NewService(repo APIKeyRepository, encKey []byte) PartnerService {
	return &Service{repo: repo, encKey: encKey}
}

func (s *Service) IssueKey(ctx context.Context, in IssueKeyInput) (*IssueKeyOutput, error) {
	prefix, err := randomHex(6)
	if err != nil {
		return nil, err
	}
	secret, err := randomHex(24)
	if err != nil {
		return nil, err
	}
	encrypted, err := s.encrypt(secret)
	if err != nil {
		return nil, err
	}

	rateLimit := in.RateLimitPerMin
	if rateLimit <= 0 {
		rateLimit = 60
	}
	key := &APIKey{
		PartnerID:       in.PartnerID,
		KeyPrefix:       prefix,
		Scopes:          in.Scopes,
		RateLimitPerMin: rateLimit,
		Status:          "active",
	}
	if err := s.repo.Create(ctx, key, encrypted); err != nil {
		return nil, err
	}
	return &IssueKeyOutput{KeyPrefix: prefix, Secret: secret}, nil
}

func (s *Service) RevokeKey(ctx context.Context, keyPrefix string) error {
	return s.repo.Revoke(ctx, keyPrefix)
}

func (s *Service) VerifySignature(ctx context.Context, keyPrefix, timestamp string, rawBody []byte, signature string) (*APIKey, error) {
	key, err := s.repo.FindByPrefix(ctx, keyPrefix)
	if err != nil {
		return nil, err
	}
	if key.Status != "active" {
		return nil, errKeyRevoked
	}

	ts, err := strconv.ParseInt(timestamp, 10, 64)
	if err != nil {
		return nil, errStaleTimestamp
	}
	if age := time.Since(time.Unix(ts, 0)); age > replayWindow || age < -replayWindow {
		return nil, errStaleTimestamp
	}

	secret, err := s.decrypt(key.EncryptedSecret)
	if err != nil {
		return nil, err
	}
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(timestamp + "."))
	mac.Write(rawBody)
	expected := hex.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(expected), []byte(signature)) {
		return nil, errInvalidSignature
	}

	_ = s.repo.TouchLastUsed(ctx, keyPrefix)
	return key, nil
}

func (s *Service) encrypt(plaintext string) (string, error) {
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return "", err
	}
	ciphertext := gcm.Seal(nonce, nonce, []byte(plaintext), nil)
	return hex.EncodeToString(ciphertext), nil
}

func (s *Service) decrypt(encoded string) (string, error) {
	ciphertext, err := hex.DecodeString(encoded)
	if err != nil {
		return "", err
	}
	block, err := aes.NewCipher(s.encKey)
	if err != nil {
		return "", err
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return "", err
	}
	if len(ciphertext) < gcm.NonceSize() {
		return "", fmt.Errorf("ciphertext too short")
	}
	nonce, data := ciphertext[:gcm.NonceSize()], ciphertext[gcm.NonceSize():]
	plaintext, err := gcm.Open(nil, nonce, data, nil)
	if err != nil {
		return "", err
	}
	return string(plaintext), nil
}

func randomHex(n int) (string, error) {
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}
