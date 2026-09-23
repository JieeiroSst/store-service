package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"gorm.io/gorm"
)

type tradeRepository struct{ db *gorm.DB }

func NewTradeRepository(db *gorm.DB) *tradeRepository { return &tradeRepository{db: db} }

func (r *tradeRepository) Create(ctx context.Context, t *model.Trade) error {
	return conn(ctx, r.db).Create(t).Error
}

func (r *tradeRepository) ListByMarket(ctx context.Context, marketID int64, limit, offset int) ([]model.Trade, error) {
	var items []model.Trade
	err := conn(ctx, r.db).Where("market_id = ?", marketID).Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error
	return items, err
}

func (r *tradeRepository) ListByMarketSince(ctx context.Context, marketID int64, since time.Time, limit int) ([]model.Trade, error) {
	q := conn(ctx, r.db).Where("market_id = ?", marketID)
	if !since.IsZero() {
		q = q.Where("created_at >= ?", since)
	}
	var items []model.Trade
	err := q.Order("id ASC").Limit(limit).Find(&items).Error
	return items, err
}

func (r *tradeRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]model.Trade, error) {
	var items []model.Trade
	err := conn(ctx, r.db).Where("maker_user_id = ? OR taker_user_id = ?", userID, userID).
		Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error
	return items, err
}
