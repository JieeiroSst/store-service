package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"gorm.io/gorm"
)

const playerLockSpace = 74_321

type PlayRepository struct {
	db *gorm.DB
}

func NewPlayRepository(db *gorm.DB) *PlayRepository {
	return &PlayRepository{db: db}
}

func (r *PlayRepository) WithPlayer(ctx context.Context, campaignID int64, userID string, fn func(port.PlayerLedger) error) error {
	return translate(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		key := fmt.Sprintf("%d:%s", campaignID, userID)
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?, hashtext(?))", playerLockSpace, key).Error; err != nil {
			return err
		}
		return fn(&ledger{tx: tx, campaignID: campaignID, userID: userID})
	}))
}

func (r *PlayRepository) Usage(ctx context.Context, campaignID int64, userID, day string) (model.Usage, error) {
	u, err := usage(r.db.WithContext(ctx), campaignID, userID, day)
	return u, translate(err)
}

func (r *PlayRepository) LastCheckin(ctx context.Context, campaignID int64, userID string) (*model.Checkin, error) {
	c, err := lastCheckin(r.db.WithContext(ctx), campaignID, userID)
	return c, translate(err)
}

func (r *PlayRepository) List(ctx context.Context, f port.PlayFilter) ([]model.Play, int64, error) {
	q := r.db.WithContext(ctx).Model(&playRecord{})
	if f.CampaignID != 0 {
		q = q.Where("campaign_id = ?", f.CampaignID)
	}
	if f.GameType != "" {
		q = q.Where("game_type = ?", string(f.GameType))
	}
	if f.UserID != "" {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.DeviceID != "" {
		q = q.Where("device_id = ?", f.DeviceID)
	}
	if f.IP != "" {
		q = q.Where("ip = ?", f.IP)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	var recs []playRecord
	if err := q.Order("id DESC").Limit(f.Limit).Offset(f.Offset).Find(&recs).Error; err != nil {
		return nil, 0, translate(err)
	}
	return toPlays(recs), total, nil
}

func (r *PlayRepository) Checkins(ctx context.Context, f port.CheckinFilter) ([]model.Checkin, int64, error) {
	q := r.db.WithContext(ctx).Model(&checkinRecord{})
	if f.CampaignID != 0 {
		q = q.Where("campaign_id = ?", f.CampaignID)
	}
	if f.UserID != "" {
		q = q.Where("user_id = ?", f.UserID)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	var recs []checkinRecord
	if err := q.Order("id DESC").Limit(f.Limit).Offset(f.Offset).Find(&recs).Error; err != nil {
		return nil, 0, translate(err)
	}
	out := make([]model.Checkin, 0, len(recs))
	for _, rec := range recs {
		out = append(out, toCheckin(rec))
	}
	return out, total, nil
}

func (r *PlayRepository) RecentWins(ctx context.Context, campaignID int64, limit int) ([]model.Play, error) {
	var recs []playRecord
	err := r.db.WithContext(ctx).Where("campaign_id = ? AND points > 0", campaignID).Order("id DESC").Limit(limit).Find(&recs).Error
	if err != nil {
		return nil, translate(err)
	}
	return toPlays(recs), nil
}

func (r *PlayRepository) Pending(ctx context.Context, before time.Time, limit int) ([]model.Reward, error) {
	var recs []outboxRecord
	err := r.db.WithContext(ctx).Where("published_at IS NULL AND status = ? AND occurred_at < ?", string(model.RewardPending), before).Order("id").Limit(limit).Find(&recs).Error
	if err != nil {
		return nil, translate(err)
	}
	out := make([]model.Reward, 0, len(recs))
	for _, rec := range recs {
		out = append(out, toReward(rec))
	}
	return out, nil
}

func (r *PlayRepository) MarkPublished(ctx context.Context, id int64, at time.Time) error {
	return translate(r.db.WithContext(ctx).Model(&outboxRecord{}).
		Where("id = ? AND published_at IS NULL", id).
		Update("published_at", at).Error)
}

type ledger struct {
	tx         *gorm.DB
	campaignID int64
	userID     string
}

func (l *ledger) PlayByKey(ctx context.Context, key string) (*model.Play, error) {
	var rec playRecord
	err := l.tx.WithContext(ctx).Where("user_id = ? AND idempotency_key = ?", l.userID, key).Take(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p := toPlay(rec)
	return &p, nil
}

func (l *ledger) Usage(ctx context.Context, day string) (model.Usage, error) {
	return usage(l.tx.WithContext(ctx), l.campaignID, l.userID, day)
}

func (l *ledger) ClaimPrize(ctx context.Context, prizeID int64, points int64) (bool, error) {
	db := l.tx.WithContext(ctx)
	if err := db.SavePoint("claim").Error; err != nil {
		return false, err
	}
	res := db.Exec(
		"UPDATE lucky_prizes SET awarded = awarded + 1 WHERE id = ? AND campaign_id = ? AND (quantity IS NULL OR awarded < quantity)",
		prizeID, l.campaignID)
	if res.Error != nil {
		return false, res.Error
	}
	if res.RowsAffected != 1 {
		return false, nil
	}
	if points > 0 {
		res = db.Exec(
			"UPDATE lucky_campaigns SET points_awarded = points_awarded + ? WHERE id = ? AND (budget_points = 0 OR points_awarded + ? <= budget_points)",
			points, l.campaignID, points)
		if res.Error != nil {
			return false, res.Error
		}
		if res.RowsAffected != 1 {
			return false, db.RollbackTo("claim").Error
		}
	}
	return true, nil
}

func (l *ledger) PointsWonOn(ctx context.Context, day string) (int64, error) {
	var won int64
	err := l.tx.WithContext(ctx).Model(&playRecord{}).
		Select("COALESCE(SUM(points), 0)").
		Where("campaign_id = ? AND user_id = ? AND day = CAST(? AS date)", l.campaignID, l.userID, day).
		Scan(&won).Error
	return won, err
}

func (l *ledger) SavePlay(ctx context.Context, p *model.Play) error {
	day, err := parseDay(p.Day)
	if err != nil {
		return err
	}
	rec := playRecord{
		CampaignID: l.campaignID,
		GameType:   string(p.GameType),
		UserID:     l.userID,
		PrizeID:    p.PrizeID,
		PrizeName:  p.PrizeName,
		PrizeType:  string(p.PrizeType),
		Points:     p.Points,
		Position:   p.Position,
		Choice:     p.Choice,
		Source:     string(p.Source),
		Day:        day,
		IP:         p.Client.IP,
		DeviceID:   p.Client.DeviceID,
		UserAgent:  p.Client.UserAgent,
		CreatedAt:  p.CreatedAt,
	}
	if p.IdempotencyKey != "" {
		rec.IdempotencyKey = &p.IdempotencyKey
	}
	db := l.tx.WithContext(ctx)
	if err := db.Create(&rec).Error; err != nil {
		return err
	}
	p.ID, p.CampaignID, p.UserID = rec.ID, rec.CampaignID, rec.UserID
	if !p.Won() {
		return nil
	}
	out := toOutbox(model.PlayReward(*p))
	return db.Create(&out).Error
}

func (l *ledger) Grant(ctx context.Context, event, referenceID string) (*model.CreditGrant, error) {
	var rec creditRecord
	err := l.tx.WithContext(ctx).
		Where("campaign_id = ? AND user_id = ? AND event = ? AND reference_id = ?", l.campaignID, l.userID, event, referenceID).
		Take(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	g := toGrant(rec)
	return &g, nil
}

func (l *ledger) GrantsOn(ctx context.Context, event, day string) (int, error) {
	var n int64
	err := l.tx.WithContext(ctx).Model(&creditRecord{}).
		Where("campaign_id = ? AND user_id = ? AND event = ? AND day = CAST(? AS date)", l.campaignID, l.userID, event, day).
		Count(&n).Error
	return int(n), err
}

func (l *ledger) SaveGrant(ctx context.Context, g *model.CreditGrant) error {
	day, err := parseDay(g.Day)
	if err != nil {
		return err
	}
	rec := creditRecord{
		CampaignID:  l.campaignID,
		UserID:      l.userID,
		Event:       g.Event,
		ReferenceID: g.ReferenceID,
		Spins:       g.Spins,
		Day:         day,
		CreatedAt:   g.CreatedAt,
	}
	if err := l.tx.WithContext(ctx).Create(&rec).Error; err != nil {
		return err
	}
	g.ID = rec.ID
	return nil
}

func (l *ledger) LastCheckin(ctx context.Context) (*model.Checkin, error) {
	return lastCheckin(l.tx.WithContext(ctx), l.campaignID, l.userID)
}

func (l *ledger) SaveCheckin(ctx context.Context, c *model.Checkin) error {
	day, err := parseDay(c.Day)
	if err != nil {
		return err
	}
	rec := checkinRecord{
		CampaignID: l.campaignID,
		UserID:     l.userID,
		Day:        day,
		Streak:     c.Streak,
		Points:     c.Points,
		IP:         c.Client.IP,
		DeviceID:   c.Client.DeviceID,
		UserAgent:  c.Client.UserAgent,
		CreatedAt:  c.CreatedAt,
	}
	db := l.tx.WithContext(ctx)
	if err := db.Create(&rec).Error; err != nil {
		return err
	}
	c.ID, c.CampaignID, c.UserID = rec.ID, rec.CampaignID, rec.UserID
	out := toOutbox(model.CheckinReward(*c))
	return db.Create(&out).Error
}

func lastCheckin(db *gorm.DB, campaignID int64, userID string) (*model.Checkin, error) {
	var rec checkinRecord
	err := db.Where("campaign_id = ? AND user_id = ?", campaignID, userID).Order("day DESC").Take(&rec).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	c := toCheckin(rec)
	return &c, nil
}

func usage(db *gorm.DB, campaignID int64, userID, day string) (model.Usage, error) {
	var row struct {
		DailyToday  int
		CreditToday int
		Granted     int
		CreditUsed  int
	}
	err := db.Raw(`
SELECT
    COUNT(*) FILTER (WHERE day = CAST(@day AS date) AND source = 'daily')  AS daily_today,
    COUNT(*) FILTER (WHERE day = CAST(@day AS date) AND source = 'credit') AS credit_today,
    COUNT(*) FILTER (WHERE source = 'credit')                              AS credit_used,
    (SELECT COALESCE(SUM(spins), 0) FROM lucky_spin_credits WHERE campaign_id = @campaign AND user_id = @user) AS granted
FROM lucky_plays
WHERE campaign_id = @campaign AND user_id = @user`,
		map[string]any{"campaign": campaignID, "user": userID, "day": day}).Scan(&row).Error
	if err != nil {
		return model.Usage{}, err
	}
	return model.Usage{
		DailyToday:    row.DailyToday,
		CreditToday:   row.CreditToday,
		CreditBalance: max(row.Granted-row.CreditUsed, 0),
	}, nil
}
