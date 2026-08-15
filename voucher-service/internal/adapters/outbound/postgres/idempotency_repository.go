package postgres

import (
	"context"
	"time"

	"github.com/JIeeiroSst/voucher-service/internal/platform/idempotency"
	"github.com/JIeeiroSst/voucher-service/internal/platform/txmanager"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type idempotencyKeyModel struct {
	Key            string    `gorm:"column:key;primaryKey"`
	RequestHash    string    `gorm:"column:request_hash"`
	Status         string    `gorm:"column:status"`
	ResponseStatus *int      `gorm:"column:response_status"`
	ResponseBody   []byte    `gorm:"column:response_body"`
	LockedAt       time.Time `gorm:"column:locked_at"`
	ExpiresAt      time.Time `gorm:"column:expires_at"`
	CreatedAt      time.Time `gorm:"column:created_at"`
	UpdatedAt      time.Time `gorm:"column:updated_at"`
}

func (idempotencyKeyModel) TableName() string { return "idempotency_keys" }

type IdempotencyRepository struct {
	db *gorm.DB
}

func NewIdempotencyRepository(db *gorm.DB) idempotency.Store {
	return &IdempotencyRepository{db: db}
}

func (r *IdempotencyRepository) Claim(ctx context.Context, key, requestHash string, ttl time.Duration) (bool, error) {
	now := time.Now().UTC()
	model := idempotencyKeyModel{
		Key:         key,
		RequestHash: requestHash,
		Status:      string(idempotency.StatusInProgress),
		LockedAt:    now,
		ExpiresAt:   now.Add(ttl),
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	tx := txmanager.DBFromContext(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).Create(&model)
	if tx.Error != nil {
		return false, tx.Error
	}
	return tx.RowsAffected > 0, nil
}

func (r *IdempotencyRepository) Get(ctx context.Context, key string) (*idempotency.Record, error) {
	var m idempotencyKeyModel
	if err := txmanager.DBFromContext(ctx, r.db).First(&m, "key = ?", key).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	rec := &idempotency.Record{Key: m.Key, Status: idempotency.Status(m.Status), ResponseBody: m.ResponseBody}
	if m.ResponseStatus != nil {
		rec.ResponseStatus = *m.ResponseStatus
	}
	return rec, nil
}

func (r *IdempotencyRepository) Complete(ctx context.Context, key string, responseStatus int, responseBody []byte) error {
	return txmanager.DBFromContext(ctx, r.db).Model(&idempotencyKeyModel{}).
		Where("key = ?", key).
		Updates(map[string]any{
			"status":          string(idempotency.StatusCompleted),
			"response_status": responseStatus,
			"response_body":   responseBody,
			"updated_at":      time.Now().UTC(),
		}).Error
}

func (r *IdempotencyRepository) Fail(ctx context.Context, key string) error {
	return txmanager.DBFromContext(ctx, r.db).Model(&idempotencyKeyModel{}).
		Where("key = ?", key).
		Updates(map[string]any{
			"status":     string(idempotency.StatusFailed),
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *IdempotencyRepository) Release(ctx context.Context, key string) error {
	return txmanager.DBFromContext(ctx, r.db).Delete(&idempotencyKeyModel{}, "key = ?", key).Error
}
