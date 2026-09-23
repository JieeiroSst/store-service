package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"gorm.io/gorm"
)

type marketRepository struct{ db *gorm.DB }

func NewMarketRepository(db *gorm.DB) *marketRepository { return &marketRepository{db: db} }

func (r *marketRepository) Create(ctx context.Context, m *model.Market) error {
	return mapWriteErr(conn(ctx, r.db).Create(m).Error)
}

func (r *marketRepository) GetByID(ctx context.Context, id int64) (*model.Market, error) {
	var m model.Market
	if err := conn(ctx, r.db).First(&m, id).Error; err != nil {
		return nil, notFound(err)
	}
	return &m, nil
}

func (r *marketRepository) GetByIDForUpdate(ctx context.Context, id int64) (*model.Market, error) {
	var m model.Market
	if err := conn(ctx, r.db).Clauses(forUpdate()).First(&m, id).Error; err != nil {
		return nil, notFound(err)
	}
	return &m, nil
}

func (r *marketRepository) GetBySlug(ctx context.Context, slug string) (*model.Market, error) {
	var m model.Market
	if err := conn(ctx, r.db).Where("slug = ?", slug).First(&m).Error; err != nil {
		return nil, notFound(err)
	}
	return &m, nil
}

func (r *marketRepository) GetByIDs(ctx context.Context, ids []int64) ([]model.Market, error) {
	var items []model.Market
	err := conn(ctx, r.db).Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *marketRepository) ListByEvents(ctx context.Context, eventIDs []int64) ([]model.Market, error) {
	var items []model.Market
	err := conn(ctx, r.db).Where("event_id IN ?", eventIDs).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *marketRepository) ListByEventForUpdate(ctx context.Context, eventID int64) ([]model.Market, error) {
	var items []model.Market
	err := conn(ctx, r.db).Clauses(forUpdate()).Where("event_id = ?", eventID).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *marketRepository) ListRewarded(ctx context.Context) ([]model.Market, error) {
	var items []model.Market
	err := conn(ctx, r.db).Where("status = ? AND reward_pool > 0 AND end_time > ?", model.MarketOpen, time.Now()).Find(&items).Error
	return items, err
}

func (r *marketRepository) Search(ctx context.Context, f port.MarketFilter) ([]model.Market, int64, error) {
	q := conn(ctx, r.db).Model(&model.Market{})
	if f.Query != "" {
		q = q.Where("question LIKE ?", "%"+escapeLike(f.Query)+"%")
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Market
	err := q.Order("volume DESC, id DESC").Limit(f.Limit).Offset(f.Offset).Find(&items).Error
	return items, total, err
}

func (r *marketRepository) Save(ctx context.Context, m *model.Market) error {
	return conn(ctx, r.db).Save(m).Error
}
