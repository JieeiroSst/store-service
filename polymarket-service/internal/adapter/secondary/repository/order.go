package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"gorm.io/gorm"
)

type orderRepository struct{ db *gorm.DB }

func NewOrderRepository(db *gorm.DB) *orderRepository { return &orderRepository{db: db} }

func (r *orderRepository) Create(ctx context.Context, o *model.Order) error {
	return mapWriteErr(conn(ctx, r.db).Create(o).Error)
}

func (r *orderRepository) GetByID(ctx context.Context, id int64) (*model.Order, error) {
	var o model.Order
	if err := conn(ctx, r.db).First(&o, id).Error; err != nil {
		return nil, notFound(err)
	}
	return &o, nil
}

func (r *orderRepository) GetByIDForUpdate(ctx context.Context, id int64) (*model.Order, error) {
	var o model.Order
	if err := conn(ctx, r.db).Clauses(forUpdate()).First(&o, id).Error; err != nil {
		return nil, notFound(err)
	}
	return &o, nil
}

func (r *orderRepository) GetByClientID(ctx context.Context, userID, clientOrderID string) (*model.Order, error) {
	var o model.Order
	err := conn(ctx, r.db).Where("user_id = ? AND client_order_id = ?", userID, clientOrderID).First(&o).Error
	if err != nil {
		return nil, notFound(err)
	}
	return &o, nil
}

func (r *orderRepository) Save(ctx context.Context, o *model.Order) error {
	return conn(ctx, r.db).Save(o).Error
}

func (r *orderRepository) ListCrossing(ctx context.Context, marketID int64, makerSide model.BookSide, yesPrice int64, limit int) ([]model.Order, error) {
	q := conn(ctx, r.db).Clauses(forUpdate()).
		Where("market_id = ? AND status = ? AND book_side = ?", marketID, model.OrderOpen, makerSide)
	if makerSide == model.Ask {
		q = q.Where("yes_price <= ?", yesPrice).Order("yes_price ASC, id ASC")
	} else {
		q = q.Where("yes_price >= ?", yesPrice).Order("yes_price DESC, id ASC")
	}
	var items []model.Order
	err := q.Limit(limit).Find(&items).Error
	return items, err
}

func (r *orderRepository) ListOpenByMarketForUpdate(ctx context.Context, marketID int64) ([]model.Order, error) {
	var items []model.Order
	err := conn(ctx, r.db).Clauses(forUpdate()).
		Where("market_id = ? AND status = ?", marketID, model.OrderOpen).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *orderRepository) ListOpenByMarket(ctx context.Context, marketID int64) ([]model.Order, error) {
	var items []model.Order
	err := conn(ctx, r.db).Where("market_id = ? AND status = ?", marketID, model.OrderOpen).Find(&items).Error
	return items, err
}

func (r *orderRepository) List(ctx context.Context, f port.OrderFilter) (*port.Page[model.Order], error) {
	cur, err := decodeCursor(f.Cursor, "orders")
	if err != nil {
		return nil, err
	}
	q := conn(ctx, r.db).Model(&model.Order{})
	if f.UserID != "" {
		q = q.Where("user_id = ?", f.UserID)
	}
	if f.MarketID != 0 {
		q = q.Where("market_id = ?", f.MarketID)
	}
	if f.Status != "" {
		q = q.Where("status = ?", f.Status)
	}
	if cur != nil {
		q = q.Where("id < ?", cur.ID)
	}
	var rows []model.Order
	if err := q.Order("id DESC").Limit(f.Limit + 1).Find(&rows).Error; err != nil {
		return nil, err
	}
	return paginate(rows, f.Limit, func(o model.Order) cursor { return cursor{S: "orders", ID: o.ID} }), nil
}

func (r *orderRepository) BookLevels(ctx context.Context, marketID int64, side model.BookSide) ([]port.BookLevel, error) {
	dir := "ASC"
	if side == model.Bid {
		dir = "DESC"
	}
	var levels []port.BookLevel
	err := conn(ctx, r.db).Model(&model.Order{}).
		Select("yes_price, SUM(size - filled) AS size").
		Where("market_id = ? AND status = ? AND book_side = ? AND (expires_at IS NULL OR expires_at > ?)",
			marketID, model.OrderOpen, side, time.Now()).
		Group("yes_price").Order("yes_price " + dir).Scan(&levels).Error
	return levels, err
}
