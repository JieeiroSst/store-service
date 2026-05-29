package domain

import "time"

type ReferralStatus string

const (
	StatusActive  ReferralStatus = "active"
	StatusUsed    ReferralStatus = "used"
	StatusExpired ReferralStatus = "expired"
)

type Channel string

const (
	ChannelCopy      Channel = "copy"
	ChannelWhatsApp  Channel = "whatsapp"
	ChannelFacebook  Channel = "facebook"
	ChannelInstagram Channel = "instagram"
	ChannelOther     Channel = "other"
)

type ReferralLink struct {
	RefCode     string         `dynamodbav:"ref_code"`    // PK — UUID v4
	CreatedAt   time.Time      `dynamodbav:"created_at"`  // SK
	OwnerUserID string         `dynamodbav:"owner_user_id"`
	Channel     Channel        `dynamodbav:"channel"`
	Status      ReferralStatus `dynamodbav:"status"`
	ExpiresAt   time.Time      `dynamodbav:"expires_at"`
	TTL         int64          `dynamodbav:"ttl"`          // Unix epoch — DynamoDB auto-delete
	DeepLink    string         `dynamodbav:"deep_link"`
	Platform    string         `dynamodbav:"platform"`     // "ios" | "android" | "universal"
}

func (r *ReferralLink) IsExpired() bool {
	return time.Now().After(r.ExpiresAt)
}

func (r *ReferralLink) IsActive() bool {
	return r.Status == StatusActive && !r.IsExpired()
}

type EventType string

const (
	EventLinkCopied   EventType = "link_copied"
	EventLinkClicked  EventType = "link_clicked"
	EventAppInstalled EventType = "app_installed"
	EventRegistered   EventType = "registered"
	EventRewardGiven  EventType = "reward_given"
)

type ReferralEvent struct {
	RefCode     string    `dynamodbav:"ref_code"`      // PK
	EventID     string    `dynamodbav:"event_id"`      // SK — KSUID (time-sortable)
	EventType   EventType `dynamodbav:"event_type"`
	OccurredAt  time.Time `dynamodbav:"occurred_at"`
	Platform    string    `dynamodbav:"platform"`      // "ios" | "android"
	NewUserID   string    `dynamodbav:"new_user_id"`   // populated after registration
	OwnerUserID string    `dynamodbav:"owner_user_id"` // denormalised for GSI queries
	IPAddress   string    `dynamodbav:"ip_address"`
	DeviceID    string    `dynamodbav:"device_id"`
	UserAgent   string    `dynamodbav:"user_agent"`
}

type RewardStatus string

const (
	RewardPending   RewardStatus = "pending"
	RewardApproved  RewardStatus = "approved"
	RewardPaid      RewardStatus = "paid"
	RewardRejected  RewardStatus = "rejected"
)

type RewardType string

const (
	RewardCash    RewardType = "cash"
	RewardCredit  RewardType = "credit"
	RewardCoupon  RewardType = "coupon"
)

type ReferralReward struct {
	OwnerUserID string       `dynamodbav:"owner_user_id"` // PK
	RefCode     string       `dynamodbav:"ref_code"`      // SK
	NewUserID   string       `dynamodbav:"new_user_id"`
	RewardType  RewardType   `dynamodbav:"reward_type"`
	RewardValue float64      `dynamodbav:"reward_value"`
	Status      RewardStatus `dynamodbav:"status"`
	CreatedAt   time.Time    `dynamodbav:"created_at"`
	UpdatedAt   time.Time    `dynamodbav:"updated_at"`
}

type UserReferralStats struct {
	UserID         string    `dynamodbav:"user_id"`          // PK
	SK             string    `dynamodbav:"sk"`               // SK = "STATS"
	TotalInvited   int64     `dynamodbav:"total_invited"`
	TotalInstalled int64     `dynamodbav:"total_installed"`
	TotalRewarded  int64     `dynamodbav:"total_rewarded"`
	TotalRewardAmt float64   `dynamodbav:"total_reward_amt"`
	LastActiveAt   time.Time `dynamodbav:"last_active_at"`
}
