package model

import (
	"strconv"
	"time"
)

type PlaySource string

const (
	SourceDaily  PlaySource = "daily"
	SourceCredit PlaySource = "credit"
)

type Play struct {
	ID             int64      `json:"id"`
	CampaignID     int64      `json:"campaign_id"`
	GameType       GameType   `json:"game_type"`
	UserID         string     `json:"user_id"`
	PrizeID        int64      `json:"prize_id"`
	PrizeName      string     `json:"prize_name"`
	PrizeType      PrizeType  `json:"prize_type"`
	Points         int64      `json:"points"`
	Position       int        `json:"position"`
	Choice         int        `json:"choice"`
	Source         PlaySource `json:"source"`
	Day            string     `json:"day"`
	Client         ClientInfo `json:"client"`
	IdempotencyKey string     `json:"-"`
	HoldReason     string     `json:"-"`
	CreatedAt      time.Time  `json:"created_at"`
}

func (p Play) Won() bool { return p.PrizeType == PrizePoints && p.Points > 0 }

type Usage struct {
	DailyToday    int `json:"daily_today"`
	CreditToday   int `json:"credit_today"`
	CreditBalance int `json:"credit_balance"`
}

type CreditGrant struct {
	ID          int64     `json:"id"`
	CampaignID  int64     `json:"campaign_id"`
	GameType    GameType  `json:"game_type"`
	UserID      string    `json:"user_id"`
	Event       string    `json:"event"`
	ReferenceID string    `json:"reference_id"`
	Spins       int       `json:"spins"`
	Day         string    `json:"day"`
	CreatedAt   time.Time `json:"created_at"`
}

type Checkin struct {
	ID         int64      `json:"id"`
	CampaignID int64      `json:"campaign_id"`
	UserID     string     `json:"user_id"`
	Day        string     `json:"day"`
	Streak     int        `json:"streak"`
	Points     int64      `json:"points"`
	Client     ClientInfo `json:"client"`
	HoldReason string     `json:"-"`
	CreatedAt  time.Time  `json:"created_at"`
}

func NextStreak(last *Checkin, today string) int {
	if last == nil {
		return 1
	}
	t, err := time.Parse(time.DateOnly, today)
	if err != nil {
		return 1
	}
	if last.Day == t.AddDate(0, 0, -1).Format(time.DateOnly) {
		return last.Streak + 1
	}
	return 1
}

func CurrentStreak(last *Checkin, today string) int {
	if last == nil {
		return 0
	}
	if last.Day == today {
		return last.Streak
	}
	if NextStreak(last, today) > 1 {
		return last.Streak
	}
	return 0
}

type Reward struct {
	ID          int64        `json:"-"`
	EventID     string       `json:"event_id"`
	Source      string       `json:"source"`
	GameType    GameType     `json:"game_type"`
	UserID      string       `json:"user_id"`
	CampaignID  int64        `json:"campaign_id"`
	ReferenceID int64        `json:"reference_id"`
	Title       string       `json:"title"`
	Points      int64        `json:"points"`
	OccurredAt  time.Time    `json:"occurred_at"`
	Status      RewardStatus `json:"-"`
	HoldReason  string       `json:"-"`
	PublishedAt *time.Time   `json:"-"`
}

type ReviewReward struct {
	ID          int64        `json:"id"`
	EventID     string       `json:"event_id"`
	Source      string       `json:"source"`
	GameType    GameType     `json:"game_type"`
	UserID      string       `json:"user_id"`
	CampaignID  int64        `json:"campaign_id"`
	ReferenceID int64        `json:"reference_id"`
	Title       string       `json:"title"`
	Points      int64        `json:"points"`
	Status      RewardStatus `json:"status"`
	HoldReason  string       `json:"hold_reason"`
	ReviewedBy  string       `json:"reviewed_by"`
	OccurredAt  time.Time    `json:"occurred_at"`
	PublishedAt *time.Time   `json:"published_at"`
}

func PlayReward(p Play) Reward {
	return Reward{
		EventID:     "lucky-play-" + strconv.FormatInt(p.ID, 10),
		Source:      "lucky-" + string(p.GameType),
		GameType:    p.GameType,
		UserID:      p.UserID,
		CampaignID:  p.CampaignID,
		ReferenceID: p.ID,
		Title:       p.PrizeName,
		Points:      p.Points,
		OccurredAt:  p.CreatedAt,
		HoldReason:  p.HoldReason,
	}
}

func CheckinReward(c Checkin) Reward {
	return Reward{
		EventID:     "lucky-checkin-" + strconv.FormatInt(c.ID, 10),
		Source:      "lucky-checkin",
		GameType:    GameCheckin,
		UserID:      c.UserID,
		CampaignID:  c.CampaignID,
		ReferenceID: c.ID,
		Title:       "Check-in day " + strconv.Itoa(c.Streak),
		Points:      c.Points,
		OccurredAt:  c.CreatedAt,
		HoldReason:  c.HoldReason,
	}
}

type Winner struct {
	CampaignID int64     `json:"campaign_id"`
	GameType   GameType  `json:"game_type"`
	User       string    `json:"user"`
	PrizeName  string    `json:"prize_name"`
	Points     int64     `json:"points"`
	At         time.Time `json:"at"`
}

func MaskUser(id string) string {
	r := []rune(id)
	if len(r) <= 2 {
		return string(r[:min(len(r), 1)]) + "***"
	}
	return string(r[:2]) + "***" + string(r[len(r)-1:])
}

func PlayWinner(p Play) Winner {
	return Winner{CampaignID: p.CampaignID, GameType: p.GameType, User: MaskUser(p.UserID), PrizeName: p.PrizeName, Points: p.Points, At: p.CreatedAt}
}
