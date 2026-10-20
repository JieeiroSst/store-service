package postgres

import (
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type campaignRecord struct {
	ID             int64         `gorm:"column:id;primaryKey"`
	GameType       string        `gorm:"column:game_type"`
	Name           string        `gorm:"column:name"`
	Description    string        `gorm:"column:description"`
	StartAt        time.Time     `gorm:"column:start_at"`
	EndAt          time.Time     `gorm:"column:end_at"`
	Timezone       string        `gorm:"column:timezone"`
	Status         string        `gorm:"column:status"`
	WinRate        float64       `gorm:"column:win_rate"`
	DailySpins     int           `gorm:"column:daily_spins"`
	MaxSpinsPerDay int           `gorm:"column:max_spins_per_day"`
	MaxPointsDay   int64         `gorm:"column:max_points_per_user_day"`
	BudgetPoints   int64         `gorm:"column:budget_points"`
	PointsAwarded  int64         `gorm:"column:points_awarded"`
	ReviewPoints   int64         `gorm:"column:review_points"`
	SpinRules      string        `gorm:"column:spin_rules"`
	CheckinRewards string        `gorm:"column:checkin_rewards"`
	Settings       string        `gorm:"column:settings"`
	CreatedAt      time.Time     `gorm:"column:created_at"`
	UpdatedAt      time.Time     `gorm:"column:updated_at"`
	Prizes         []prizeRecord `gorm:"foreignKey:CampaignID"`
}

func (campaignRecord) TableName() string { return "lucky_campaigns" }

type prizeRecord struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	CampaignID int64  `gorm:"column:campaign_id"`
	Name       string `gorm:"column:name"`
	Type       string `gorm:"column:type"`
	Points     int64  `gorm:"column:points"`
	Weight     int    `gorm:"column:weight"`
	Quantity   *int64 `gorm:"column:quantity"`
	Awarded    int64  `gorm:"column:awarded"`
	Position   int    `gorm:"column:position"`
	ImageURL   string `gorm:"column:image_url"`
}

func (prizeRecord) TableName() string { return "lucky_prizes" }

type playRecord struct {
	ID             int64     `gorm:"column:id;primaryKey"`
	CampaignID     int64     `gorm:"column:campaign_id"`
	GameType       string    `gorm:"column:game_type"`
	UserID         string    `gorm:"column:user_id"`
	PrizeID        int64     `gorm:"column:prize_id"`
	PrizeName      string    `gorm:"column:prize_name"`
	PrizeType      string    `gorm:"column:prize_type"`
	Points         int64     `gorm:"column:points"`
	Position       int       `gorm:"column:position"`
	Choice         int       `gorm:"column:choice"`
	Source         string    `gorm:"column:source"`
	Day            time.Time `gorm:"column:day;type:date"`
	IP             string    `gorm:"column:ip"`
	DeviceID       string    `gorm:"column:device_id"`
	UserAgent      string    `gorm:"column:user_agent"`
	IdempotencyKey *string   `gorm:"column:idempotency_key"`
	CreatedAt      time.Time `gorm:"column:created_at"`
}

func (playRecord) TableName() string { return "lucky_plays" }

type creditRecord struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	CampaignID  int64     `gorm:"column:campaign_id"`
	UserID      string    `gorm:"column:user_id"`
	Event       string    `gorm:"column:event"`
	ReferenceID string    `gorm:"column:reference_id"`
	Spins       int       `gorm:"column:spins"`
	Day         time.Time `gorm:"column:day;type:date"`
	CreatedAt   time.Time `gorm:"column:created_at"`
}

func (creditRecord) TableName() string { return "lucky_spin_credits" }

type checkinRecord struct {
	ID         int64     `gorm:"column:id;primaryKey"`
	CampaignID int64     `gorm:"column:campaign_id"`
	UserID     string    `gorm:"column:user_id"`
	Day        time.Time `gorm:"column:day;type:date"`
	Streak     int       `gorm:"column:streak"`
	Points     int64     `gorm:"column:points"`
	IP         string    `gorm:"column:ip"`
	DeviceID   string    `gorm:"column:device_id"`
	UserAgent  string    `gorm:"column:user_agent"`
	CreatedAt  time.Time `gorm:"column:created_at"`
}

func (checkinRecord) TableName() string { return "lucky_checkins" }

type outboxRecord struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	EventID     string     `gorm:"column:event_id"`
	Source      string     `gorm:"column:source"`
	GameType    string     `gorm:"column:game_type"`
	UserID      string     `gorm:"column:user_id"`
	CampaignID  int64      `gorm:"column:campaign_id"`
	ReferenceID int64      `gorm:"column:reference_id"`
	Title       string     `gorm:"column:title"`
	Points      int64      `gorm:"column:points"`
	OccurredAt  time.Time  `gorm:"column:occurred_at"`
	Status      string     `gorm:"column:status"`
	HoldReason  string     `gorm:"column:hold_reason"`
	ReviewedBy  string     `gorm:"column:reviewed_by"`
	ReviewedAt  *time.Time `gorm:"column:reviewed_at"`
	PublishedAt *time.Time `gorm:"column:published_at"`
}

func (outboxRecord) TableName() string { return "lucky_reward_outbox" }

type blockRecord struct {
	UserID    string    `gorm:"column:user_id;primaryKey"`
	Reason    string    `gorm:"column:reason"`
	BlockedBy string    `gorm:"column:blocked_by"`
	CreatedAt time.Time `gorm:"column:created_at"`
}

func (blockRecord) TableName() string { return "lucky_user_blocks" }

func translate(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.As(err, &pgErr) && pgErr.Code == "23P01":
		return model.ErrCampaignOverlap
	case errors.As(err, &pgErr) && pgErr.Code == "23514":
		return fmt.Errorf("%w: %s", model.ErrInvalid, pgErr.ConstraintName)
	case errors.As(err, &pgErr) && (pgErr.Code == "23505" || pgErr.Code == "23503"):
		return model.ErrConflict
	case errors.Is(err, gorm.ErrRecordNotFound):
		return model.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey), errors.Is(err, gorm.ErrForeignKeyViolated):
		return model.ErrConflict
	case errors.Is(err, gorm.ErrCheckConstraintViolated):
		return model.ErrInvalid
	}
	return err
}

func jsonText(v any) (string, error) {
	b, err := json.Marshal(v)
	return string(b), err
}

func toCampaignRecord(c *model.Campaign) (*campaignRecord, error) {
	rules, err := jsonText(c.SpinRules)
	if err != nil {
		return nil, err
	}
	rewards, err := jsonText(c.CheckinRewards)
	if err != nil {
		return nil, err
	}
	settings := string(c.Settings)
	if settings == "" {
		settings = "{}"
	}
	rec := &campaignRecord{
		ID:             c.ID,
		GameType:       string(c.GameType),
		Name:           c.Name,
		Description:    c.Description,
		StartAt:        c.StartAt,
		EndAt:          c.EndAt,
		Timezone:       c.Timezone,
		Status:         string(c.Status),
		WinRate:        c.WinRate,
		DailySpins:     c.DailySpins,
		MaxSpinsPerDay: c.MaxSpinsPerDay,
		MaxPointsDay:   c.MaxPointsDay,
		BudgetPoints:   c.BudgetPoints,
		PointsAwarded:  c.PointsAwarded,
		ReviewPoints:   c.ReviewPoints,
		SpinRules:      rules,
		CheckinRewards: rewards,
		Settings:       settings,
	}
	for _, p := range c.Prizes {
		rec.Prizes = append(rec.Prizes, prizeRecord{
			ID:         p.ID,
			CampaignID: c.ID,
			Name:       p.Name,
			Type:       string(p.Type),
			Points:     p.Points,
			Weight:     p.Weight,
			Quantity:   p.Quantity,
			Awarded:    p.Awarded,
			Position:   p.Position,
			ImageURL:   p.ImageURL,
		})
	}
	return rec, nil
}

func toCampaign(rec *campaignRecord) *model.Campaign {
	c := &model.Campaign{
		ID:             rec.ID,
		GameType:       model.GameType(rec.GameType),
		Name:           rec.Name,
		Description:    rec.Description,
		StartAt:        rec.StartAt.UTC(),
		EndAt:          rec.EndAt.UTC(),
		Timezone:       rec.Timezone,
		Status:         model.CampaignStatus(rec.Status),
		WinRate:        rec.WinRate,
		DailySpins:     rec.DailySpins,
		MaxSpinsPerDay: rec.MaxSpinsPerDay,
		MaxPointsDay:   rec.MaxPointsDay,
		BudgetPoints:   rec.BudgetPoints,
		PointsAwarded:  rec.PointsAwarded,
		ReviewPoints:   rec.ReviewPoints,
		SpinRules:      []model.SpinRule{},
		CheckinRewards: []int64{},
		Settings:       json.RawMessage(rec.Settings),
		Prizes:         make([]model.Prize, 0, len(rec.Prizes)),
		CreatedAt:      rec.CreatedAt.UTC(),
		UpdatedAt:      rec.UpdatedAt.UTC(),
	}
	_ = json.Unmarshal([]byte(rec.SpinRules), &c.SpinRules)
	_ = json.Unmarshal([]byte(rec.CheckinRewards), &c.CheckinRewards)
	if len(c.Settings) == 0 {
		c.Settings = json.RawMessage("{}")
	}
	for _, p := range rec.Prizes {
		c.Prizes = append(c.Prizes, model.Prize{
			ID:         p.ID,
			CampaignID: p.CampaignID,
			Name:       p.Name,
			Type:       model.PrizeType(p.Type),
			Points:     p.Points,
			Weight:     p.Weight,
			Quantity:   p.Quantity,
			Awarded:    p.Awarded,
			Position:   p.Position,
			ImageURL:   p.ImageURL,
		})
	}
	return c
}

func toPlay(rec playRecord) model.Play {
	p := model.Play{
		ID:         rec.ID,
		CampaignID: rec.CampaignID,
		GameType:   model.GameType(rec.GameType),
		UserID:     rec.UserID,
		PrizeID:    rec.PrizeID,
		PrizeName:  rec.PrizeName,
		PrizeType:  model.PrizeType(rec.PrizeType),
		Points:     rec.Points,
		Position:   rec.Position,
		Choice:     rec.Choice,
		Source:     model.PlaySource(rec.Source),
		Day:        rec.Day.Format(time.DateOnly),
		Client:     model.ClientInfo{IP: rec.IP, DeviceID: rec.DeviceID, UserAgent: rec.UserAgent},
		CreatedAt:  rec.CreatedAt.UTC(),
	}
	if rec.IdempotencyKey != nil {
		p.IdempotencyKey = *rec.IdempotencyKey
	}
	return p
}

func toPlays(recs []playRecord) []model.Play {
	out := make([]model.Play, 0, len(recs))
	for _, rec := range recs {
		out = append(out, toPlay(rec))
	}
	return out
}

func toGrant(rec creditRecord) model.CreditGrant {
	return model.CreditGrant{
		ID:          rec.ID,
		CampaignID:  rec.CampaignID,
		UserID:      rec.UserID,
		Event:       rec.Event,
		ReferenceID: rec.ReferenceID,
		Spins:       rec.Spins,
		Day:         rec.Day.Format(time.DateOnly),
		CreatedAt:   rec.CreatedAt.UTC(),
	}
}

func toCheckin(rec checkinRecord) model.Checkin {
	return model.Checkin{
		ID:         rec.ID,
		CampaignID: rec.CampaignID,
		UserID:     rec.UserID,
		Day:        rec.Day.Format(time.DateOnly),
		Streak:     rec.Streak,
		Points:     rec.Points,
		Client:     model.ClientInfo{IP: rec.IP, DeviceID: rec.DeviceID, UserAgent: rec.UserAgent},
		CreatedAt:  rec.CreatedAt.UTC(),
	}
}

func toOutbox(r model.Reward) outboxRecord {
	status := model.RewardPending
	if r.HoldReason != "" {
		status = model.RewardHeld
	}
	return outboxRecord{
		Status:      string(status),
		HoldReason:  r.HoldReason,
		EventID:     r.EventID,
		Source:      r.Source,
		GameType:    string(r.GameType),
		UserID:      r.UserID,
		CampaignID:  r.CampaignID,
		ReferenceID: r.ReferenceID,
		Title:       r.Title,
		Points:      r.Points,
		OccurredAt:  r.OccurredAt,
	}
}

func toReward(rec outboxRecord) model.Reward {
	return model.Reward{
		ID:          rec.ID,
		EventID:     rec.EventID,
		Source:      rec.Source,
		GameType:    model.GameType(rec.GameType),
		UserID:      rec.UserID,
		CampaignID:  rec.CampaignID,
		ReferenceID: rec.ReferenceID,
		Title:       rec.Title,
		Points:      rec.Points,
		OccurredAt:  rec.OccurredAt.UTC(),
		Status:      model.RewardStatus(rec.Status),
		HoldReason:  rec.HoldReason,
		PublishedAt: rec.PublishedAt,
	}
}

func toReviewReward(rec outboxRecord) model.ReviewReward {
	return model.ReviewReward{
		ID:          rec.ID,
		EventID:     rec.EventID,
		Source:      rec.Source,
		GameType:    model.GameType(rec.GameType),
		UserID:      rec.UserID,
		CampaignID:  rec.CampaignID,
		ReferenceID: rec.ReferenceID,
		Title:       rec.Title,
		Points:      rec.Points,
		Status:      model.RewardStatus(rec.Status),
		HoldReason:  rec.HoldReason,
		ReviewedBy:  rec.ReviewedBy,
		OccurredAt:  rec.OccurredAt.UTC(),
		PublishedAt: rec.PublishedAt,
	}
}

func parseDay(day string) (time.Time, error) {
	t, err := time.Parse(time.DateOnly, day)
	if err != nil {
		return time.Time{}, fmt.Errorf("%w: day %q", model.ErrInvalid, day)
	}
	return t, nil
}
