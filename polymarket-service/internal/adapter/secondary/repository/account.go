package repository

import (
	"context"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"gorm.io/gorm"
)

type balanceRepository struct{ db *gorm.DB }

func NewBalanceRepository(db *gorm.DB) *balanceRepository { return &balanceRepository{db: db} }

func (r *balanceRepository) GetForUpdate(ctx context.Context, userID string) (*model.Balance, error) {
	c := conn(ctx, r.db)
	if err := c.Exec("INSERT IGNORE INTO balances (user_id, available, locked) VALUES (?, 0, 0)", userID).Error; err != nil {
		return nil, err
	}
	var b model.Balance
	if err := c.Clauses(forUpdate()).Where("user_id = ?", userID).First(&b).Error; err != nil {
		return nil, notFound(err)
	}
	return &b, nil
}

func (r *balanceRepository) Get(ctx context.Context, userID string) (*model.Balance, error) {
	var b model.Balance
	err := conn(ctx, r.db).Where("user_id = ?", userID).First(&b).Error
	if err != nil {
		if notFound(err) == port.ErrNotFound {
			return &model.Balance{UserID: userID}, nil
		}
		return nil, err
	}
	return &b, nil
}

func (r *balanceRepository) Save(ctx context.Context, b *model.Balance) error {
	return conn(ctx, r.db).Save(b).Error
}

type ledgerRepository struct{ db *gorm.DB }

func NewLedgerRepository(db *gorm.DB) *ledgerRepository { return &ledgerRepository{db: db} }

func (r *ledgerRepository) Create(ctx context.Context, e *model.LedgerEntry) error {
	return conn(ctx, r.db).Create(e).Error
}

func (r *ledgerRepository) ListByUser(ctx context.Context, userID string, limit, offset int) ([]model.LedgerEntry, error) {
	var items []model.LedgerEntry
	err := conn(ctx, r.db).Where("user_id = ?", userID).Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error
	return items, err
}

type positionRepository struct{ db *gorm.DB }

func NewPositionRepository(db *gorm.DB) *positionRepository { return &positionRepository{db: db} }

func (r *positionRepository) GetForUpdate(ctx context.Context, marketID int64, userID string, outcome model.Outcome) (*model.Position, error) {
	var p model.Position
	err := conn(ctx, r.db).Clauses(forUpdate()).
		Where("market_id = ? AND user_id = ? AND outcome = ?", marketID, userID, outcome).First(&p).Error
	if err != nil {
		return nil, notFound(err)
	}
	return &p, nil
}

func (r *positionRepository) Save(ctx context.Context, p *model.Position) error {
	return conn(ctx, r.db).Save(p).Error
}

func (r *positionRepository) ListByUser(ctx context.Context, userID string, settled bool) ([]model.Position, error) {
	q := conn(ctx, r.db).Where("user_id = ? AND settled = ?", userID, settled)
	if !settled {
		q = q.Where("shares > 0")
	}
	var items []model.Position
	err := q.Order("updated_at DESC").Find(&items).Error
	return items, err
}

func (r *positionRepository) ListByMarketForUpdate(ctx context.Context, marketID int64) ([]model.Position, error) {
	var items []model.Position
	err := conn(ctx, r.db).Clauses(forUpdate()).Where("market_id = ?", marketID).Order("id ASC").Find(&items).Error
	return items, err
}

func (r *positionRepository) TopHolders(ctx context.Context, marketID int64, outcome model.Outcome, limit int) ([]port.Holder, error) {
	var items []port.Holder
	err := conn(ctx, r.db).Model(&model.Position{}).
		Select("user_id, shares").
		Where("market_id = ? AND outcome = ? AND shares > 0", marketID, outcome).
		Order("shares DESC").Limit(limit).Scan(&items).Error
	return items, err
}
