package pg

import (
	"context"
	"time"

	"github.com/JIeeiroSst/user-service/internal/domain"
	"gorm.io/gorm"
)

type SessionRepository struct {
	db *gorm.DB
}

func NewSessionRepository(db *gorm.DB) *SessionRepository {
	return &SessionRepository{db: db}
}

func (r *SessionRepository) Create(ctx context.Context, s domain.LoginSession) error {
	return r.db.WithContext(ctx).Create(&s).Error
}

func (r *SessionRepository) Get(ctx context.Context, id string) (domain.LoginSession, error) {
	var s domain.LoginSession
	res := r.db.WithContext(ctx).Where("id = ?", id).Limit(1).Find(&s)
	if res.Error != nil {
		return domain.LoginSession{}, res.Error
	}
	if res.RowsAffected == 0 {
		return domain.LoginSession{}, domain.ErrSessionNotFound
	}
	return s, nil
}

func (r *SessionRepository) Touch(ctx context.Context, id string, lastActive, expiresAt time.Time) error {
	return r.db.WithContext(ctx).Model(&domain.LoginSession{}).
		Where("id = ? AND revoked_at IS NULL", id).
		Updates(map[string]any{"last_active_at": lastActive, "expires_at": expiresAt}).Error
}

func (r *SessionRepository) ListActive(ctx context.Context, userID int, now time.Time) ([]domain.LoginSession, error) {
	var out []domain.LoginSession
	err := r.db.WithContext(ctx).
		Where("user_id = ? AND revoked_at IS NULL AND expires_at > ?", userID, now).
		Order("last_active_at DESC").
		Find(&out).Error
	return out, err
}

func (r *SessionRepository) Revoke(ctx context.Context, ids []string, reason string, at time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domain.LoginSession{}).
		Where("id IN ? AND revoked_at IS NULL", ids).
		Updates(map[string]any{"revoked_at": at, "revoke_reason": reason}).Error
}
