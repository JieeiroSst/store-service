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
	ChannelZalo      Channel = "zalo"
	ChannelOther     Channel = "other"
)

type ReferralLink struct {
	RefCode     string         `json:"ref_code"      db:"ref_code"`      // primary key
	CreatedAt   int64          `json:"created_at"    db:"created_at"`    // Unix ms — indexed with owner_user_id for pagination
	OwnerUserID string         `json:"owner_user_id" db:"owner_user_id"`
	Channel     Channel        `json:"channel"       db:"channel"`
	Status      ReferralStatus `json:"status"        db:"status"`
	ExpiresAt   int64          `json:"expires_at"    db:"expires_at"`    // Unix ms
	DeepLink    string         `json:"deep_link"     db:"deep_link"`
	Platform    string         `json:"platform"      db:"platform"`      // "ios" | "android" | "universal"
}

func (r *ReferralLink) IsExpired() bool {
	return time.Now().UnixMilli() > r.ExpiresAt
}

func (r *ReferralLink) IsActive() bool {
	return r.Status == StatusActive && !r.IsExpired()
}

type EventType string

const (
	EventLinkCreated  EventType = "link_created"  // fired when a new referral link is generated
	EventLinkCopied   EventType = "link_copied"   // fired by client when user explicitly copies/shares
	EventLinkClicked  EventType = "link_clicked"
	EventAppInstalled EventType = "app_installed"
	EventRegistered   EventType = "registered"
	EventRewardGiven  EventType = "reward_given"
)

type ReferralEvent struct {
	RefCode             string    `json:"ref_code"      db:"ref_code"`      // indexed with occurred_at
	EventID             string    `json:"event_id"      db:"event_id"`      // primary key — UUID
	EventType           EventType `json:"event_type"    db:"event_type"`
	OccurredAt          int64     `json:"occurred_at"   db:"occurred_at"`   // Unix ms
	Platform            string    `json:"platform"      db:"platform"`      // "ios" | "android"
	Channel             Channel   `json:"channel,omitempty"      db:"channel"`             // share/click channel — may differ from the link's default channel
	NewUserID           string    `json:"new_user_id,omitempty"   db:"new_user_id"`   // populated after registration
	OwnerUserID         string    `json:"owner_user_id" db:"owner_user_id"` // denormalised for lookups
	IPAddress           string    `json:"ip_address"    db:"ip_address"`
	DeviceID            string    `json:"device_id"     db:"device_id"`
	UserAgent           string    `json:"user_agent"    db:"user_agent"`
	FirebaseInstanceID  string    `json:"firebase_instance_id,omitempty" db:"firebase_instance_id"` // Firebase Analytics App Instance ID, for cross-referencing with GA4/BigQuery
}

// ReferralAttribution records the confirmed relationship between an owner and a new user
// after a successful install. Primary key: (owner_user_id, new_user_id).
type ReferralAttribution struct {
	OwnerUserID        string `json:"owner_user_id" db:"owner_user_id"`
	NewUserID          string `json:"new_user_id"   db:"new_user_id"`   // indexed for reverse lookups
	RefCode            string `json:"ref_code"      db:"ref_code"`
	Platform           string `json:"platform"      db:"platform"`
	DeviceID           string `json:"device_id"     db:"device_id"`
	AttributedAt       int64  `json:"attributed_at" db:"attributed_at"` // Unix ms
	FirebaseInstanceID string `json:"firebase_instance_id,omitempty" db:"firebase_instance_id"` // Firebase Analytics App Instance ID, for cross-referencing with GA4/BigQuery
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

// Primary key: (owner_user_id, ref_code).
type ReferralReward struct {
	OwnerUserID string       `json:"owner_user_id" db:"owner_user_id"`
	RefCode     string       `json:"ref_code"      db:"ref_code"`
	NewUserID   string       `json:"new_user_id"   db:"new_user_id"`
	RewardType  RewardType   `json:"reward_type"   db:"reward_type"`
	RewardValue float64      `json:"reward_value"  db:"reward_value"`
	Status      RewardStatus `json:"status"        db:"status"`
	CreatedAt   int64        `json:"created_at"    db:"created_at"`    // Unix ms
	UpdatedAt   int64        `json:"updated_at"    db:"updated_at"`    // Unix ms
}

// Primary key: user_id.
type UserReferralStats struct {
	UserID         string  `json:"user_id"          db:"user_id"`
	TotalInvited   int64   `json:"total_invited"    db:"total_invited"`
	TotalInstalled int64   `json:"total_installed"  db:"total_installed"`
	TotalRewarded  int64   `json:"total_rewarded"   db:"total_rewarded"`
	TotalRewardAmt float64 `json:"total_reward_amt" db:"total_reward_amt"`
	LastActiveAt   int64   `json:"last_active_at"   db:"last_active_at"`   // Unix ms
}

// Progression: pending → clicked → installed → rewarded
type InvitationStatus string

const (
	InvitationPending   InvitationStatus = "pending"   
	InvitationClicked   InvitationStatus = "clicked"   
	InvitationInstalled InvitationStatus = "installed" 
	InvitationRewarded  InvitationStatus = "rewarded" 
	InvitationExpired   InvitationStatus = "expired"  
)

type RewardProgramStatus string

const (
	ProgramStatusActive   RewardProgramStatus = "active"
	ProgramStatusInactive RewardProgramStatus = "inactive"
)

type RewardTier struct {
	MinCount    int64   `json:"min_count"    db:"min_count"`
	MaxCount    int64   `json:"max_count"    db:"max_count"`
	RewardValue float64 `json:"reward_value" db:"reward_value"`
}

// Primary key: program_id; status is indexed for FindActive queries.
// Tiers is persisted as a JSON column — marshaled/unmarshaled by the repo, not scanned directly.
type RewardProgram struct {
	ProgramID string              `json:"program_id" db:"program_id"`
	Name      string              `json:"name"       db:"name"`
	Status    RewardProgramStatus `json:"status"     db:"status"`
	Tiers     []RewardTier        `json:"tiers"      db:"-"`
	CreatedAt int64               `json:"created_at" db:"created_at"`
	UpdatedAt int64               `json:"updated_at" db:"updated_at"`
}

func (p *RewardProgram) RewardForInstall(installNumber int64) float64 {
	for _, t := range p.Tiers {
		if installNumber >= t.MinCount && (t.MaxCount < 0 || installNumber <= t.MaxCount) {
			return t.RewardValue
		}
	}
	return 0
}
