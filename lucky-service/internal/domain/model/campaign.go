package model

import (
	"bytes"
	"encoding/json"
	"math"
	"sort"
	"strings"
	"time"
)

type CampaignStatus string

const (
	StatusDraft  CampaignStatus = "draft"
	StatusActive CampaignStatus = "active"
	StatusPaused CampaignStatus = "paused"
)

func (s CampaignStatus) Valid() bool {
	return s == StatusDraft || s == StatusActive || s == StatusPaused
}

type PrizeType string

const (
	PrizePoints PrizeType = "points"
	PrizeNone   PrizeType = "none"
)

const (
	MaxPrizes        = 24
	MaxChoice        = 24
	MaxSpinRules     = 20
	MaxDailyPlays    = 100
	MaxCheckinCycle  = 31
	MaxSettingsBytes = 8 << 10
	winRateScale     = 100
	winRateBuckets   = 100 * winRateScale
)

type SpinRule struct {
	Event     string `json:"event"`
	Spins     int    `json:"spins"`
	MaxPerDay int    `json:"max_per_day"`
}

type Prize struct {
	ID         int64     `json:"id"`
	CampaignID int64     `json:"campaign_id"`
	Name       string    `json:"name"`
	Type       PrizeType `json:"type"`
	Points     int64     `json:"points"`
	Weight     int       `json:"weight"`
	Quantity   *int64    `json:"quantity"`
	Awarded    int64     `json:"awarded"`
	Position   int       `json:"position"`
	ImageURL   string    `json:"image_url"`
}

func (p Prize) InStock() bool {
	return p.Quantity == nil || p.Awarded < *p.Quantity
}

type Campaign struct {
	ID             int64           `json:"id"`
	GameType       GameType        `json:"game_type"`
	Name           string          `json:"name"`
	Description    string          `json:"description"`
	StartAt        time.Time       `json:"start_at"`
	EndAt          time.Time       `json:"end_at"`
	Timezone       string          `json:"timezone"`
	Status         CampaignStatus  `json:"status"`
	WinRate        float64         `json:"win_rate"`
	DailySpins     int             `json:"daily_spins"`
	MaxSpinsPerDay int             `json:"max_spins_per_day"`
	MaxPointsDay   int64           `json:"max_points_per_user_day"`
	BudgetPoints   int64           `json:"budget_points"`
	PointsAwarded  int64           `json:"points_awarded"`
	ReviewPoints   int64           `json:"review_points"`
	SpinRules      []SpinRule      `json:"spin_rules"`
	CheckinRewards []int64         `json:"checkin_rewards"`
	Settings       json.RawMessage `json:"settings"`
	Prizes         []Prize         `json:"prizes"`
	CreatedAt      time.Time       `json:"created_at"`
	UpdatedAt      time.Time       `json:"updated_at"`
}

func (c *Campaign) Normalize() {
	c.Name = strings.TrimSpace(c.Name)
	c.Description = strings.TrimSpace(c.Description)
	c.Timezone = strings.TrimSpace(c.Timezone)
	c.WinRate = math.Round(c.WinRate*winRateScale) / winRateScale
	if s := bytes.TrimSpace(c.Settings); len(s) == 0 || bytes.Equal(s, []byte("null")) {
		c.Settings = json.RawMessage("{}")
	}
	if c.SpinRules == nil {
		c.SpinRules = []SpinRule{}
	}
	if c.CheckinRewards == nil {
		c.CheckinRewards = []int64{}
	}
	if c.Prizes == nil {
		c.Prizes = []Prize{}
	}
	for i := range c.SpinRules {
		c.SpinRules[i].Event = strings.ToLower(strings.TrimSpace(c.SpinRules[i].Event))
	}
	for i := range c.Prizes {
		p := &c.Prizes[i]
		p.Name = strings.TrimSpace(p.Name)
		p.Type = PrizeType(strings.ToLower(strings.TrimSpace(string(p.Type))))
		p.ImageURL = strings.TrimSpace(p.ImageURL)
		if p.Type == PrizeNone {
			p.Points, p.Weight, p.Quantity = 0, 0, nil
		}
	}
	autoPosition := true
	for _, p := range c.Prizes {
		if p.Position != 0 {
			autoPosition = false
			break
		}
	}
	if autoPosition {
		for i := range c.Prizes {
			c.Prizes[i].Position = i + 1
		}
	}
	sort.SliceStable(c.Prizes, func(i, j int) bool { return c.Prizes[i].Position < c.Prizes[j].Position })
}

func (c *Campaign) Validate() error {
	switch {
	case !c.GameType.Valid():
		_, err := ParseGame(string(c.GameType))
		return err
	case c.Name == "":
		return invalid("name", "is required")
	case len(c.Name) > 255:
		return invalid("name", "must be at most 255 characters")
	case c.StartAt.IsZero():
		return invalid("start_at", "is required")
	case c.EndAt.IsZero():
		return invalid("end_at", "is required")
	case !c.EndAt.After(c.StartAt):
		return invalid("end_at", "must be after start_at")
	case len(c.Settings) > MaxSettingsBytes:
		return invalid("settings", "must be at most 8KB")
	}
	var obj map[string]any
	if err := json.Unmarshal(c.Settings, &obj); err != nil {
		return invalid("settings", "must be a JSON object")
	}
	if _, err := time.LoadLocation(c.Timezone); err != nil || c.Timezone == "" {
		return invalid("timezone", "must be an IANA time zone such as Asia/Ho_Chi_Minh")
	}
	if c.GameType == GameCheckin {
		return c.validateCheckin()
	}
	return c.validateInstant()
}

func (c *Campaign) validateCheckin() error {
	switch {
	case len(c.Prizes) > 0:
		return invalid("prizes", "a checkin campaign rewards by streak day, use checkin_rewards")
	case len(c.SpinRules) > 0:
		return invalid("spin_rules", "a checkin campaign has no extra plays")
	case len(c.CheckinRewards) == 0 || len(c.CheckinRewards) > MaxCheckinCycle:
		return invalid("checkin_rewards", "must list the points for 1 to 31 streak days")
	}
	for _, p := range c.CheckinRewards {
		if p <= 0 {
			return invalid("checkin_rewards", "every day must reward a positive number of points")
		}
	}
	return nil
}

func (c *Campaign) validateInstant() error {
	switch {
	case len(c.CheckinRewards) > 0:
		return invalid("checkin_rewards", "only checkin campaigns have streak rewards")
	case c.WinRate < 0 || c.WinRate > 100:
		return invalid("win_rate", "must be between 0 and 100")
	case c.DailySpins < 0 || c.DailySpins > MaxDailyPlays:
		return invalid("daily_spins", "must be between 0 and 100")
	case c.MaxSpinsPerDay < 0:
		return invalid("max_spins_per_day", "must not be negative")
	case c.MaxPointsDay < 0:
		return invalid("max_points_per_user_day", "must not be negative")
	case c.BudgetPoints < 0:
		return invalid("budget_points", "must not be negative")
	case c.BudgetPoints > 0 && c.BudgetPoints < c.PointsAwarded:
		return invalid("budget_points", "is below the points already awarded")
	case c.ReviewPoints < 0:
		return invalid("review_points", "must not be negative")
	case len(c.SpinRules) > MaxSpinRules:
		return invalid("spin_rules", "too many rules")
	}
	events := map[string]bool{}
	for _, r := range c.SpinRules {
		switch {
		case r.Event == "" || len(r.Event) > 64:
			return invalid("spin_rules.event", "must be 1-64 characters")
		case events[r.Event]:
			return invalid("spin_rules.event", "duplicate event "+r.Event)
		case r.Spins <= 0 || r.Spins > MaxDailyPlays:
			return invalid("spin_rules.spins", "must be between 1 and 100")
		case r.MaxPerDay < 0:
			return invalid("spin_rules.max_per_day", "must not be negative")
		}
		events[r.Event] = true
	}
	return c.validatePrizes()
}

func (c *Campaign) validatePrizes() error {
	if len(c.Prizes) < 2 || len(c.Prizes) > MaxPrizes {
		return invalid("prizes", "a game needs between 2 and 24 prize slots")
	}
	blanks, winners := 0, 0
	positions := map[int]bool{}
	for _, p := range c.Prizes {
		switch {
		case p.Name == "" || len(p.Name) > 255:
			return invalid("prizes.name", "must be 1-255 characters")
		case p.Position <= 0:
			return invalid("prizes.position", "must be positive")
		case positions[p.Position]:
			return invalid("prizes.position", "must be unique")
		}
		positions[p.Position] = true
		switch p.Type {
		case PrizeNone:
			blanks++
		case PrizePoints:
			winners++
			switch {
			case p.Points <= 0:
				return invalid("prizes.points", "must be positive for a points prize")
			case p.Weight <= 0:
				return invalid("prizes.weight", "must be positive for a points prize")
			case p.Quantity != nil && *p.Quantity < 0:
				return invalid("prizes.quantity", "must not be negative")
			case p.Quantity != nil && *p.Quantity < p.Awarded:
				return invalid("prizes.quantity", "is below the number already awarded")
			}
		default:
			return invalid("prizes.type", "must be points or none")
		}
	}
	if blanks != 1 {
		return invalid("prizes", "exactly one slot must be the no-prize (none) slot")
	}
	if winners == 0 {
		return invalid("prizes", "at least one points prize is required")
	}
	return nil
}

func (c *Campaign) Location() *time.Location {
	loc, err := time.LoadLocation(c.Timezone)
	if err != nil {
		return time.UTC
	}
	return loc
}

func (c *Campaign) Day(now time.Time) string {
	return now.In(c.Location()).Format(time.DateOnly)
}

func (c *Campaign) RunningAt(now time.Time) bool {
	return c.Status == StatusActive && !now.Before(c.StartAt) && now.Before(c.EndAt)
}

func (c *Campaign) Rule(event string) (SpinRule, bool) {
	event = strings.ToLower(strings.TrimSpace(event))
	for _, r := range c.SpinRules {
		if r.Event == event {
			return r, true
		}
	}
	return SpinRule{}, false
}

func (c *Campaign) Blank() *Prize {
	for i := range c.Prizes {
		if c.Prizes[i].Type == PrizeNone {
			return &c.Prizes[i]
		}
	}
	return nil
}

func (c *Campaign) Draw(roll func(n int) int) *Prize {
	blank := c.Blank()
	if roll(winRateBuckets) >= int(math.Round(c.WinRate*winRateScale)) {
		return blank
	}
	var pool []*Prize
	total := 0
	for i := range c.Prizes {
		p := &c.Prizes[i]
		if p.Type == PrizePoints && p.Weight > 0 && p.InStock() {
			pool = append(pool, p)
			total += p.Weight
		}
	}
	if total == 0 {
		return blank
	}
	n := roll(total)
	for _, p := range pool {
		if n < p.Weight {
			return p
		}
		n -= p.Weight
	}
	return blank
}

func (c *Campaign) ValidateChoice(choice int) error {
	if c.GameType.NeedsChoice() {
		if choice < 1 || choice > MaxChoice {
			return invalid("choice", "pick a box or card between 1 and 24")
		}
		return nil
	}
	if choice != 0 {
		return invalid("choice", "this game does not take a choice")
	}
	return nil
}

func (c *Campaign) Allowance(u Usage) (PlaySource, bool) {
	if c.MaxSpinsPerDay > 0 && u.DailyToday+u.CreditToday >= c.MaxSpinsPerDay {
		return "", false
	}
	if u.DailyToday < c.DailySpins {
		return SourceDaily, true
	}
	if u.CreditBalance > 0 {
		return SourceCredit, true
	}
	return "", false
}

func (c *Campaign) Remaining(u Usage) int {
	n := max(c.DailySpins-u.DailyToday, 0) + max(u.CreditBalance, 0)
	if c.MaxSpinsPerDay > 0 {
		n = min(n, max(c.MaxSpinsPerDay-u.DailyToday-u.CreditToday, 0))
	}
	return n
}

func (c *Campaign) WithinDailyCap(wonToday, points int64) bool {
	return c.MaxPointsDay == 0 || wonToday+points <= c.MaxPointsDay
}

func (c *Campaign) NeedsReview(points int64) bool {
	return c.ReviewPoints > 0 && points >= c.ReviewPoints
}

func (c *Campaign) CheckinReward(streak int) int64 {
	if len(c.CheckinRewards) == 0 || streak <= 0 {
		return 0
	}
	return c.CheckinRewards[(streak-1)%len(c.CheckinRewards)]
}
