package ports

import (
	"context"
	"time"

	"github.com/referral/service/internal/core/domain"
)


type ReferralService interface {
	GenerateLink(ctx context.Context, req GenerateLinkRequest) (*GenerateLinkResponse, error)
	GetLink(ctx context.Context, refCode string) (*domain.ReferralLink, error)
	ListUserLinks(ctx context.Context, ownerUserID string, limit int, cursor string) ([]*domain.ReferralLink, string, error)
	TrackEvent(ctx context.Context, req TrackEventRequest) error
	ConfirmInstall(ctx context.Context, req ConfirmInstallRequest) (*ConfirmInstallResponse, error)
	ActivateReferral(ctx context.Context, req ActivateReferralRequest) (*ActivateReferralResponse, error)
	GetReferralStatus(ctx context.Context, refCode string) (*ReferralStatusResponse, error)
	GetUserStats(ctx context.Context, userID string) (*domain.UserReferralStats, error)
}

type GenerateLinkRequest struct {
	OwnerUserID string
	Channel     domain.Channel
	Platform    string // "ios" | "android" | "universal"
}

type GenerateLinkResponse struct {
	RefCode   string
	DeepLink  string
	ExpiresAt string // RFC3339
}

type TrackEventRequest struct {
	RefCode   string
	EventType domain.EventType
	Platform  string
	NewUserID string
	IPAddress string
	DeviceID  string
	UserAgent string
}

type ConfirmInstallRequest struct {
	RefCode   string
	NewUserID string
	Platform  string
	DeviceID  string
}

type ConfirmInstallResponse struct {
	Attributed  bool
	OwnerUserID string
	RewardType  domain.RewardType
}

type ActivateReferralRequest struct {
	RefCode  string
	UserID   string
	Platform string
	DeviceID string
}

type ActivateReferralResponse struct {
	Attributed  bool
	OwnerUserID string
	RewardType  domain.RewardType
}

type ReferralStatusResponse struct {
	RefCode     string     `json:"ref_code"`
	Status      string     `json:"status"`
	OwnerUserID string     `json:"owner_user_id"`
	ActivatedAt *time.Time `json:"activated_at,omitempty"`
	Platform    string     `json:"platform,omitempty"`
	NewUserID   string     `json:"new_user_id,omitempty"`
}

type ReferralLinkRepository interface {
	Save(ctx context.Context, link *domain.ReferralLink) error
	FindByRefCode(ctx context.Context, refCode string) (*domain.ReferralLink, error)
	FindByOwnerUserID(ctx context.Context, ownerUserID string, limit int, cursor string) ([]*domain.ReferralLink, string, error)
	UpdateStatus(ctx context.Context, refCode string, createdAt string, status domain.ReferralStatus) error
}

type ReferralEventRepository interface {
	Save(ctx context.Context, event *domain.ReferralEvent) error
	FindByRefCode(ctx context.Context, refCode string) ([]*domain.ReferralEvent, error)
	FindByNewUserID(ctx context.Context, newUserID string) ([]*domain.ReferralEvent, error)
}

type RewardRepository interface {
	Save(ctx context.Context, reward *domain.ReferralReward) error
	FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*domain.ReferralReward, error)
}

type UserStatsRepository interface {
	Get(ctx context.Context, userID string) (*domain.UserReferralStats, error)
	IncrementCounters(ctx context.Context, userID string, invited, installed, rewarded int64, rewardAmt float64) error
}
