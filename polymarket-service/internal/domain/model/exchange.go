package model

import "time"

type PnLKind string

const (
	PnLTrade      PnLKind = "trade"      // realised on a sell
	PnLSettlement PnLKind = "settlement" // realised when a market resolves
	PnLConversion PnLKind = "conversion" // neg-risk conversion
	PnLFee        PnLKind = "fee"        // taker fee paid (negative)
	PnLRebate     PnLKind = "rebate"     // maker rebate earned
	PnLReferral   PnLKind = "referral"   // referral commission earned
	PnLReward     PnLKind = "reward"     // liquidity reward earned
	PnLBond       PnLKind = "bond"       // dispute bond forfeited (negative)
)

type PnLEntry struct {
	ID        int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	UserID    string    `gorm:"column:user_id" json:"user_id"`
	MarketID  int64     `gorm:"column:market_id" json:"market_id"`
	Kind      PnLKind   `gorm:"column:kind" json:"kind"`
	Amount    int64     `gorm:"column:amount" json:"amount"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (PnLEntry) TableName() string { return "pnl_entries" }

type ExchangeBucket string

const (
	BucketRevenue ExchangeBucket = "revenue"
	BucketBond    ExchangeBucket = "bond"
)

type ExchangeEntry struct {
	ID        int64          `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Bucket    ExchangeBucket `gorm:"column:bucket" json:"bucket"`
	Kind      string         `gorm:"column:kind" json:"kind"`
	Amount    int64          `gorm:"column:amount" json:"amount"` // signed
	MarketID  int64          `gorm:"column:market_id" json:"market_id,omitempty"`
	UserID    string         `gorm:"column:user_id" json:"user_id,omitempty"`
	CreatedAt time.Time      `gorm:"column:created_at" json:"created_at"`
}

func (ExchangeEntry) TableName() string { return "exchange_entries" }

type RewardPayout struct {
	ID         int64     `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EpochStart time.Time `gorm:"column:epoch_start" json:"epoch_start"`
	MarketID   int64     `gorm:"column:market_id" json:"market_id"`
	UserID     string    `gorm:"column:user_id" json:"user_id"`
	Score      float64   `gorm:"column:score" json:"score"`
	Amount     int64     `gorm:"column:amount" json:"amount"`
	CreatedAt  time.Time `gorm:"column:created_at" json:"created_at"`
}

func (RewardPayout) TableName() string { return "reward_payouts" }

type ReferralCode struct {
	UserID    string    `gorm:"column:user_id;primaryKey" json:"user_id"`
	RefCode   string    `gorm:"column:ref_code" json:"ref_code"`
	DeepLink  string    `gorm:"column:deep_link" json:"deep_link"`
	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
}

func (ReferralCode) TableName() string { return "referral_codes" }

type Referral struct {
	RefereeUserID  string    `gorm:"column:referee_user_id;primaryKey" json:"referee_user_id"`
	ReferrerUserID string    `gorm:"column:referrer_user_id" json:"referrer_user_id"`
	RefCode        string    `gorm:"column:ref_code" json:"ref_code"`
	CreatedAt      time.Time `gorm:"column:created_at" json:"created_at"`
}

func (Referral) TableName() string { return "referrals" }
