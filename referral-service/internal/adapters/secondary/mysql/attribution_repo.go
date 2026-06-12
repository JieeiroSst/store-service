package mysql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
)

type attributionRepo struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewAttributionRepo(db *sqlx.DB, log *zap.Logger) ports.ReferralAttributionRepository {
	return &attributionRepo{
		db:  db,
		log: log.Named("attribution-repo"),
	}
}

func (r *attributionRepo) Save(ctx context.Context, attr *domain.ReferralAttribution) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO referral_attributions (owner_user_id, new_user_id, ref_code, platform, device_id, attributed_at, firebase_instance_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		attr.OwnerUserID, attr.NewUserID, attr.RefCode, attr.Platform, attr.DeviceID, attr.AttributedAt, attr.FirebaseInstanceID,
	)
	if err != nil {
		return fmt.Errorf("insert attribution: %w", err)
	}

	r.log.Debug("attribution saved",
		zap.String("owner_user_id", attr.OwnerUserID),
		zap.String("new_user_id", attr.NewUserID),
		zap.String("ref_code", attr.RefCode),
	)
	return nil
}

func (r *attributionRepo) FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*domain.ReferralAttribution, error) {
	var attrs []*domain.ReferralAttribution
	err := r.db.SelectContext(ctx, &attrs,
		`SELECT owner_user_id, new_user_id, ref_code, platform, device_id, attributed_at, firebase_instance_id
		 FROM referral_attributions WHERE owner_user_id = ? ORDER BY attributed_at DESC`,
		ownerUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("query attributions for owner %s: %w", ownerUserID, err)
	}
	return attrs, nil
}

func (r *attributionRepo) FindByNewUserID(ctx context.Context, newUserID string) (*domain.ReferralAttribution, error) {
	var attr domain.ReferralAttribution
	err := r.db.GetContext(ctx, &attr,
		`SELECT owner_user_id, new_user_id, ref_code, platform, device_id, attributed_at, firebase_instance_id
		 FROM referral_attributions WHERE new_user_id = ? LIMIT 1`,
		newUserID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("new_user_id %s: %w", newUserID, domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("query attribution for new_user %s: %w", newUserID, err)
	}
	return &attr, nil
}
