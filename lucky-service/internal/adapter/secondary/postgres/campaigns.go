package postgres

import (
	"context"
	_ "embed"
	"fmt"
	"time"

	"github.com/JIeeiroSst/lucky-service/internal/domain/model"
	"github.com/JIeeiroSst/lucky-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

//go:embed schema.sql
var schema string

const migrationLockID = 7_432_021_001

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", migrationLockID).Error; err != nil {
			return err
		}
		return tx.Exec(schema).Error
	})
}

func (r *Repository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (r *Repository) Name() string { return "postgres" }

func (r *Repository) Critical() bool { return true }

func (r *Repository) Create(ctx context.Context, c *model.Campaign) error {
	rec, err := toCampaignRecord(c)
	if err != nil {
		return err
	}
	err = r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		prizes := rec.Prizes
		rec.Prizes = nil
		if err := tx.Omit("created_at", "updated_at").Create(rec).Error; err != nil {
			return err
		}
		if len(prizes) > 0 {
			for i := range prizes {
				prizes[i].CampaignID = rec.ID
			}
			if err := tx.Create(&prizes).Error; err != nil {
				return err
			}
		}
		rec.Prizes = prizes
		return tx.Where("id = ?", rec.ID).Take(rec).Error
	})
	if err != nil {
		return translate(err)
	}
	*c = *toCampaign(rec)
	return nil
}

func (r *Repository) Update(ctx context.Context, c *model.Campaign) error {
	rec, err := toCampaignRecord(c)
	if err != nil {
		return err
	}
	return translate(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current campaignRecord
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("id = ?", c.ID).Take(&current).Error; err != nil {
			return err
		}
		if current.Status == string(model.StatusActive) {
			return fmt.Errorf("%w: pause the campaign before editing it", model.ErrConflict)
		}
		err := tx.Model(&campaignRecord{}).Where("id = ?", c.ID).Updates(map[string]any{
			"name":                    rec.Name,
			"description":             rec.Description,
			"start_at":                rec.StartAt,
			"end_at":                  rec.EndAt,
			"timezone":                rec.Timezone,
			"win_rate":                rec.WinRate,
			"daily_spins":             rec.DailySpins,
			"max_spins_per_day":       rec.MaxSpinsPerDay,
			"max_points_per_user_day": rec.MaxPointsDay,
			"budget_points":           rec.BudgetPoints,
			"review_points":           rec.ReviewPoints,
			"spin_rules":              rec.SpinRules,
			"checkin_rewards":         rec.CheckinRewards,
			"settings":                rec.Settings,
			"updated_at":              gorm.Expr("now()"),
		}).Error
		if err != nil {
			return err
		}
		keep := []int64{0}
		for _, p := range rec.Prizes {
			if p.ID != 0 {
				keep = append(keep, p.ID)
			}
		}
		if err := tx.Where("campaign_id = ? AND id NOT IN ?", c.ID, keep).Delete(&prizeRecord{}).Error; err != nil {
			return err
		}
		for _, p := range rec.Prizes {
			p.CampaignID = c.ID
			if p.ID == 0 {
				if err := tx.Create(&p).Error; err != nil {
					return err
				}
				continue
			}
			res := tx.Model(&prizeRecord{}).Where("id = ? AND campaign_id = ?", p.ID, c.ID).Updates(map[string]any{
				"name":      p.Name,
				"type":      p.Type,
				"points":    p.Points,
				"weight":    p.Weight,
				"quantity":  p.Quantity,
				"position":  p.Position,
				"image_url": p.ImageURL,
			})
			if res.Error != nil {
				return res.Error
			}
			if res.RowsAffected == 0 {
				return model.ErrNotFound
			}
		}
		return nil
	}))
}

func orderPrizes(db *gorm.DB) *gorm.DB { return db.Order("position") }

func (r *Repository) Get(ctx context.Context, id int64) (*model.Campaign, error) {
	var rec campaignRecord
	if err := r.db.WithContext(ctx).Preload("Prizes", orderPrizes).Where("id = ?", id).Take(&rec).Error; err != nil {
		return nil, translate(err)
	}
	return toCampaign(&rec), nil
}

func (r *Repository) List(ctx context.Context, f port.CampaignFilter) ([]model.Campaign, int64, error) {
	q := r.db.WithContext(ctx).Model(&campaignRecord{})
	if f.Status != "" {
		q = q.Where("status = ?", string(f.Status))
	}
	if f.GameType != "" {
		q = q.Where("game_type = ?", string(f.GameType))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	var recs []campaignRecord
	if err := q.Preload("Prizes", orderPrizes).Order("id DESC").Limit(f.Limit).Offset(f.Offset).Find(&recs).Error; err != nil {
		return nil, 0, translate(err)
	}
	out := make([]model.Campaign, 0, len(recs))
	for i := range recs {
		out = append(out, *toCampaign(&recs[i]))
	}
	return out, total, nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Where("id = ?", id).Delete(&campaignRecord{})
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *Repository) SetStatus(ctx context.Context, id int64, status model.CampaignStatus) error {
	res := r.db.WithContext(ctx).Model(&campaignRecord{}).Where("id = ?", id).Updates(map[string]any{
		"status":     string(status),
		"updated_at": gorm.Expr("now()"),
	})
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *Repository) Running(ctx context.Context, at time.Time) ([]model.Campaign, error) {
	var recs []campaignRecord
	err := r.db.WithContext(ctx).Preload("Prizes", orderPrizes).
		Where("status = ? AND start_at <= ? AND end_at > ?", string(model.StatusActive), at, at).
		Order("game_type, start_at").Find(&recs).Error
	if err != nil {
		return nil, translate(err)
	}
	out := make([]model.Campaign, 0, len(recs))
	for i := range recs {
		out = append(out, *toCampaign(&recs[i]))
	}
	return out, nil
}
