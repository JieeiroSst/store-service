package domain

import (
	"context"
	"time"
)

const (
	ProviderPassword  = "password"
	ProviderSSO       = "sso"
	ProviderGoogle    = "google"
	ProviderFacebook  = "facebook"
	ProviderTwitter   = "twitter"
	ProviderMicrosoft = "microsoft"
)

const (
	RevokeLogout        = "logout"
	RevokeByUser        = "revoked_by_user"
	RevokeEvicted       = "evicted"
	RevokeAccountLocked = "account_locked"
)

type LoginSession struct {
	ID           string     `gorm:"primaryKey;size:64" json:"id"`
	UserID       int        `gorm:"index;not null" json:"user_id"`
	Provider     string     `gorm:"size:32" json:"provider"`
	IP           string     `gorm:"size:64" json:"ip"`
	UserAgent    string     `gorm:"size:512" json:"user_agent"`
	CreatedAt    time.Time  `json:"created_at"`
	LastActiveAt time.Time  `json:"last_active_at"`
	ExpiresAt    time.Time  `gorm:"index" json:"expires_at"`
	RevokedAt    *time.Time `json:"revoked_at,omitempty"`
	RevokeReason string     `gorm:"size:32" json:"revoke_reason,omitempty"`
}

func (LoginSession) TableName() string { return "user_sessions" }

type ClientInfo struct {
	IP        string
	UserAgent string
}

type clientInfoKey struct{}

func WithClientInfo(ctx context.Context, info ClientInfo) context.Context {
	return context.WithValue(ctx, clientInfoKey{}, info)
}

func ClientInfoFrom(ctx context.Context) ClientInfo {
	info, _ := ctx.Value(clientInfoKey{}).(ClientInfo)
	return info
}
