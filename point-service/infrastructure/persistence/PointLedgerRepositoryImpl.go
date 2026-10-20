package persistence

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/point-service/domain/entity"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type PointLedgerRepositoryImpl struct {
	db *gorm.DB
}

func NewPointLedgerRepositoryImpl(db *gorm.DB) *PointLedgerRepositoryImpl {
	return &PointLedgerRepositoryImpl{db: db}
}

func (r *PointLedgerRepositoryImpl) Credit(ctx context.Context, tx *entity.PointTransaction) (bool, error) {
	applied := false
	err := r.db.WithContext(ctx).Transaction(func(db *gorm.DB) error {
		now := time.Now().UTC()
		tx.CreatedAt = now
		res := db.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "event_id"}}, DoNothing: true}).Create(tx)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return nil
		}
		account := entity.PointAccount{UserID: tx.UserID, Balance: tx.Points, TotalEarned: tx.Points, UpdatedAt: now}
		err := db.Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "user_id"}},
			DoUpdates: clause.Assignments(map[string]any{
				"balance":      gorm.Expr("balance + ?", tx.Points),
				"total_earned": gorm.Expr("total_earned + ?", tx.Points),
				"updated_at":   now,
			}),
		}).Create(&account).Error
		if err != nil {
			return err
		}
		if err := db.Clauses(clause.Locking{Strength: "UPDATE"}).Where("user_id = ?", tx.UserID).Take(&account).Error; err != nil {
			return err
		}
		tx.BalanceAfter = account.Balance
		if err := db.Model(&entity.PointTransaction{}).Where("id = ?", tx.ID).Update("balance_after", account.Balance).Error; err != nil {
			return err
		}
		applied = true
		return nil
	})
	return applied, err
}

func (r *PointLedgerRepositoryImpl) GetAccount(ctx context.Context, userID string) (*entity.PointAccount, error) {
	var a entity.PointAccount
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).Take(&a).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return &entity.PointAccount{UserID: userID}, nil
	}
	if err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *PointLedgerRepositoryImpl) ListTransactions(ctx context.Context, userID string, limit, offset int) ([]entity.PointTransaction, int64, error) {
	q := r.db.WithContext(ctx).Model(&entity.PointTransaction{}).Where("user_id = ?", userID)
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []entity.PointTransaction
	if err := q.Order("id DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
