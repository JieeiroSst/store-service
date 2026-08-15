package partner

import (
	"context"
	"time"
)

type APIKey struct {
	ID              string
	PartnerID       string
	KeyPrefix       string
	EncryptedSecret string
	Scopes          []string
	RateLimitPerMin int
	Status          string
	LastUsedAt      *time.Time
}

type IssueKeyInput struct {
	PartnerID       string
	Scopes          []string
	RateLimitPerMin int
}

type IssueKeyOutput struct {
	KeyPrefix string
	Secret    string
}

type PartnerService interface {
	IssueKey(ctx context.Context, in IssueKeyInput) (*IssueKeyOutput, error)
	RevokeKey(ctx context.Context, keyPrefix string) error
	VerifySignature(ctx context.Context, keyPrefix, timestamp string, rawBody []byte, signature string) (*APIKey, error)
}

type APIKeyRepository interface {
	Create(ctx context.Context, key *APIKey, encryptedSecret string) error
	FindByPrefix(ctx context.Context, keyPrefix string) (*APIKey, error)
	Revoke(ctx context.Context, keyPrefix string) error
	TouchLastUsed(ctx context.Context, keyPrefix string) error
}
