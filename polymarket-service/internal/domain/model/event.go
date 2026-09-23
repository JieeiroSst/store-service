package model

import "time"

type Outcome string

const (
	OutcomeYes   Outcome = "yes"
	OutcomeNo    Outcome = "no"
	OutcomeSplit Outcome = "split"
)

func (o Outcome) Valid() bool { return o == OutcomeYes || o == OutcomeNo }

func (o Outcome) Opposite() Outcome {
	if o == OutcomeYes {
		return OutcomeNo
	}
	return OutcomeYes
}

type EventStatus string

const (
	EventOpen     EventStatus = "open"
	EventResolved EventStatus = "resolved"
)

type Event struct {
	ID          int64       `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	Slug        string      `gorm:"column:slug" json:"slug"`
	Title       string      `gorm:"column:title" json:"title"`
	Description string      `gorm:"column:description" json:"description,omitempty"`
	Category    string      `gorm:"column:category" json:"category,omitempty"`
	Tags        []string    `gorm:"column:tags;serializer:json" json:"tags"`
	ImageURL    string      `gorm:"column:image_url" json:"image_url,omitempty"`
	Featured    bool        `gorm:"column:featured" json:"featured"`
	NegRisk     bool        `gorm:"column:neg_risk" json:"neg_risk"`
	Status      EventStatus `gorm:"column:status" json:"status"`
	EndDate     time.Time   `gorm:"column:end_date" json:"end_date"`
	Volume      int64       `gorm:"column:volume" json:"volume"`
	CreatedAt   time.Time   `gorm:"column:created_at" json:"created_at"`
	UpdatedAt   time.Time   `gorm:"column:updated_at" json:"updated_at"`

	Markets []Market `gorm:"-" json:"markets,omitempty"`
}

func (Event) TableName() string { return "events" }

type MarketStatus string

const (
	MarketOpen     MarketStatus = "open"
	MarketProposed MarketStatus = "proposed"
	MarketDisputed MarketStatus = "disputed"
	MarketResolved MarketStatus = "resolved"
)

type Market struct {
	ID             int64        `gorm:"column:id;primaryKey;autoIncrement" json:"id"`
	EventID        int64        `gorm:"column:event_id" json:"event_id"`
	Slug           string       `gorm:"column:slug" json:"slug"`
	Question       string       `gorm:"column:question" json:"question"`
	GroupItemTitle string       `gorm:"column:group_item_title" json:"group_item_title,omitempty"`
	Description    string       `gorm:"column:description" json:"description,omitempty"`
	ShareValue     int64        `gorm:"column:share_value" json:"share_value"`
	MinOrderSize   int64        `gorm:"column:min_order_size" json:"min_order_size"`
	EndTime        time.Time    `gorm:"column:end_time" json:"end_time"`
	Status         MarketStatus `gorm:"column:status" json:"status"`

	ProposedOutcome Outcome    `gorm:"column:proposed_outcome" json:"proposed_outcome,omitempty"`
	DisputeDeadline *time.Time `gorm:"column:dispute_deadline" json:"dispute_deadline,omitempty"`
	DisputedBy      string     `gorm:"column:disputed_by" json:"disputed_by,omitempty"`
	DisputeReason   string     `gorm:"column:dispute_reason" json:"dispute_reason,omitempty"`
	ResolvedOutcome Outcome    `gorm:"column:resolved_outcome" json:"resolved_outcome,omitempty"`

	RewardPool      int64 `gorm:"column:reward_pool" json:"reward_pool"`
	RewardMaxSpread int64 `gorm:"column:reward_max_spread" json:"reward_max_spread"`
	RewardMinSize   int64 `gorm:"column:reward_min_size" json:"reward_min_size"`

	DisputeBond int64 `gorm:"column:dispute_bond" json:"dispute_bond,omitempty"`
	BondSettled bool  `gorm:"column:bond_settled" json:"-"`

	BestBid   int64 `gorm:"column:best_bid" json:"best_bid"`
	BestAsk   int64 `gorm:"column:best_ask" json:"best_ask"`
	LastPrice int64 `gorm:"column:last_price" json:"last_price"`
	Volume    int64 `gorm:"column:volume" json:"volume"`

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

func (Market) TableName() string { return "markets" }

func (m *Market) Tradable(now time.Time) bool {
	return m.Status == MarketOpen && now.Before(m.EndTime)
}

func (m *Market) YesPrice() float64 {
	switch {
	case m.BestBid > 0 && m.BestAsk > 0:
		return float64(m.BestBid+m.BestAsk) / 2
	case m.LastPrice > 0:
		return float64(m.LastPrice)
	default:
		return float64(m.ShareValue) / 2
	}
}

func (m *Market) PriceOf(o Outcome) float64 {
	if o == OutcomeNo {
		return float64(m.ShareValue) - m.YesPrice()
	}
	return m.YesPrice()
}
