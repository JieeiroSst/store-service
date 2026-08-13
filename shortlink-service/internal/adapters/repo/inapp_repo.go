package repo

import (
	"context"
	"errors"
	"strings"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

type InAppEventRepo struct{ db *gorm.DB }

func NewInAppEventRepo(db *gorm.DB) *InAppEventRepo { return &InAppEventRepo{db: db} }

func (r *InAppEventRepo) Insert(ctx context.Context, event *domain.InAppEvent) (string, error) {
	m := &InAppEventModel{
		InstallID:         event.InstallID,
		EventName:         event.EventName,
		EventData:         mapToJSON(event.EventData),
		EventTimestamp:    event.EventTimestamp,
		AttributedLinkID:  event.AttributedLinkID,
		AttributedClickID: event.AttributedClickID,
		AttributedAt:      event.AttributedAt,
		SessionID:         event.SessionID,
		SDKName:           event.SDKName,
		SDKVersion:        event.SDKVersion,
	}

	err := r.db.WithContext(ctx).Create(m).Error
	if err == nil {
		return m.ID, nil
	}

	if isAttributedLinkFKViolation(err) {
		m.AttributedLinkID = nil
		if retryErr := r.db.WithContext(ctx).Create(m).Error; retryErr != nil {
			return "", retryErr
		}
		return m.ID, nil
	}

	return "", err
}

func isAttributedLinkFKViolation(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23503" && strings.Contains(pgErr.ConstraintName, "attributed_link_id")
}
