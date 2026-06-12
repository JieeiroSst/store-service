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

	CreateRewardProgram(ctx context.Context, req CreateRewardProgramRequest) (*domain.RewardProgram, error)
	GetActiveRewardProgram(ctx context.Context) (*domain.RewardProgram, error)
	SetRewardProgramStatus(ctx context.Context, programID string, active bool) error
}

type GenerateLinkRequest struct {
	OwnerUserID string
	Channel     domain.Channel
	Platform    string // "ios" | "android" | "universal"
}

type GenerateLinkResponse struct {
	RefCode   string
	DeepLink  string
	ExpiresAt int64 // Unix ms
}

type TrackEventRequest struct {
	RefCode            string
	EventType          domain.EventType
	Platform           string
	Channel            domain.Channel
	NewUserID          string
	IPAddress          string
	DeviceID           string
	UserAgent          string
	FirebaseInstanceID string
}

type ConfirmInstallRequest struct {
	RefCode            string
	NewUserID          string
	Platform           string
	DeviceID           string
	FirebaseInstanceID string
}

type ConfirmInstallResponse struct {
	Attributed  bool
	OwnerUserID string
	RewardType  domain.RewardType
}

type ActivateReferralRequest struct {
	RefCode            string
	UserID             string
	Platform           string
	DeviceID           string
	FirebaseInstanceID string
}

type ActivateReferralResponse struct {
	Attributed  bool
	OwnerUserID string
	RewardType  domain.RewardType
}

type ReferralStatusResponse struct {
	RefCode          string                  `json:"ref_code"`
	Status           string                  `json:"status"`            // link status: active | used | expired
	InvitationStatus domain.InvitationStatus `json:"invitation_status"` // journey: pending→clicked→installed→rewarded
	OwnerUserID      string                  `json:"owner_user_id"`
	InvitedAt        *int64                  `json:"invited_at,omitempty"`    // Unix ms — first link_clicked event
	ActivatedAt      *int64                  `json:"activated_at,omitempty"`  // Unix ms — app_installed event
	TimeToInstallMs  *int64                  `json:"time_to_install_ms,omitempty"` // ActivatedAt - InvitedAt
	Platform         string                  `json:"platform,omitempty"`
	NewUserID        string                  `json:"new_user_id,omitempty"`
	RewardStatus     domain.RewardStatus     `json:"reward_status,omitempty"`
	RewardValue      float64                 `json:"reward_value,omitempty"`
}

type ReferralLinkRepository interface {
	Save(ctx context.Context, link *domain.ReferralLink) error
	FindByRefCode(ctx context.Context, refCode string) (*domain.ReferralLink, error)
	FindByOwnerUserID(ctx context.Context, ownerUserID string, limit int, cursor string) ([]*domain.ReferralLink, string, error)
	UpdateStatus(ctx context.Context, refCode string, status domain.ReferralStatus) error
	CountByOwnerUserIDSince(ctx context.Context, ownerUserID string, sinceMs int64) (int64, error)
}

type ReferralEventRepository interface {
	Save(ctx context.Context, event *domain.ReferralEvent) error
	FindByRefCode(ctx context.Context, refCode string) ([]*domain.ReferralEvent, error)
	FindByNewUserID(ctx context.Context, newUserID string) ([]*domain.ReferralEvent, error)
}

type ReferralAttributionRepository interface {
	Save(ctx context.Context, attr *domain.ReferralAttribution) error
	FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*domain.ReferralAttribution, error)
	FindByNewUserID(ctx context.Context, newUserID string) (*domain.ReferralAttribution, error)
}

type RewardRepository interface {
	Save(ctx context.Context, reward *domain.ReferralReward) error
	FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*domain.ReferralReward, error)
	FindByOwnerAndRefCode(ctx context.Context, ownerUserID, refCode string) (*domain.ReferralReward, error)
}

type UserStatsRepository interface {
	Get(ctx context.Context, userID string) (*domain.UserReferralStats, error)
	IncrementCounters(ctx context.Context, userID string, invited, installed, rewarded int64, rewardAmt float64) error
}

type RewardProgramRepository interface {
	Save(ctx context.Context, program *domain.RewardProgram) error
	FindByID(ctx context.Context, programID string) (*domain.RewardProgram, error)
	FindActive(ctx context.Context) (*domain.RewardProgram, error)
	UpdateStatus(ctx context.Context, programID string, status domain.RewardProgramStatus) error
}

type Cache interface {
	GetInt64(ctx context.Context, key string) (int64, bool, error)
	SetInt64(ctx context.Context, key string, value int64, ttl time.Duration) error
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
}

type RewardTierInput struct {
	MinCount    int64
	MaxCount    int64 // -1 = unlimited
	RewardValue float64
}

type CreateRewardProgramRequest struct {
	Name     string
	Tiers    []RewardTierInput
	Activate bool
}
