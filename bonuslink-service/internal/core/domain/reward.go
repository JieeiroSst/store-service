package domain

import "time"

type RewardType string

const (
	RewardPoints         RewardType = "POINTS"
	RewardBadge          RewardType = "BADGE"
	RewardDiscount       RewardType = "DISCOUNT"
	RewardExperience     RewardType = "EXPERIENCE"
	RewardPremiumContent RewardType = "PREMIUM_CONTENT"
)

func (t RewardType) Valid() bool {
	switch t {
	case RewardPoints, RewardBadge, RewardDiscount, RewardExperience, RewardPremiumContent:
		return true
	}
	return false
}

type Reward struct {
	ID         string     `json:"reward_id"`
	EventID    string     `json:"event_id"`
	UserID     string     `json:"user_id"`
	RefCode    string     `json:"ref_code,omitempty"`
	Source     string     `json:"source"`
	RewardType RewardType `json:"reward_type"`
	Value      float64    `json:"reward_value"`
	CreatedAt  time.Time  `json:"created_at"`
}

type Balance struct {
	UserID      string     `json:"user_id"`
	RewardType  RewardType `json:"reward_type"`
	TotalValue  float64    `json:"total_value"`
	RewardCount int64      `json:"reward_count"`
}
