package mysql

import (
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"

	mysqldriver "github.com/go-sql-driver/mysql"
	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
)

const mysqlDupEntryErrNo = 1062

type referralLinkRepo struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewReferralLinkRepo(db *sqlx.DB, log *zap.Logger) ports.ReferralLinkRepository {
	return &referralLinkRepo{
		db:  db,
		log: log.Named("link-repo"),
	}
}

func (r *referralLinkRepo) Save(ctx context.Context, link *domain.ReferralLink) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO referral_links (ref_code, created_at, owner_user_id, channel, status, expires_at, deep_link, platform)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		link.RefCode, link.CreatedAt, link.OwnerUserID, link.Channel, link.Status, link.ExpiresAt, link.DeepLink, link.Platform,
	)
	if err != nil {
		var mysqlErr *mysqldriver.MySQLError
		if errors.As(err, &mysqlErr) && mysqlErr.Number == mysqlDupEntryErrNo {
			return domain.ErrDuplicateRefCode
		}
		return fmt.Errorf("insert referral link: %w", err)
	}

	r.log.Debug("referral link saved", zap.String("ref_code", link.RefCode))
	return nil
}

func (r *referralLinkRepo) FindByRefCode(ctx context.Context, refCode string) (*domain.ReferralLink, error) {
	var link domain.ReferralLink
	err := r.db.GetContext(ctx, &link,
		`SELECT ref_code, created_at, owner_user_id, channel, status, expires_at, deep_link, platform
		 FROM referral_links WHERE ref_code = ?`,
		refCode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("ref_code %s: %w", refCode, domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("query ref_code %s: %w", refCode, err)
	}
	return &link, nil
}

func (r *referralLinkRepo) FindByOwnerUserID(
	ctx context.Context,
	ownerUserID string,
	limit int,
	cursor string,
) ([]*domain.ReferralLink, string, error) {
	query := `SELECT ref_code, created_at, owner_user_id, channel, status, expires_at, deep_link, platform
		FROM referral_links WHERE owner_user_id = ?`
	args := []any{ownerUserID}

	if cursor != "" {
		if createdAt, refCode, err := decodeLinkCursor(cursor); err == nil {
			query += ` AND (created_at < ? OR (created_at = ? AND ref_code < ?))`
			args = append(args, createdAt, createdAt, refCode)
		}
	}

	query += ` ORDER BY created_at DESC, ref_code DESC LIMIT ?`
	args = append(args, limit)

	var links []*domain.ReferralLink
	if err := r.db.SelectContext(ctx, &links, query, args...); err != nil {
		return nil, "", fmt.Errorf("query links by owner %s: %w", ownerUserID, err)
	}

	nextCursor := ""
	if len(links) == limit {
		last := links[len(links)-1]
		nextCursor = encodeLinkCursor(last.CreatedAt, last.RefCode)
	}

	return links, nextCursor, nil
}

func (r *referralLinkRepo) UpdateStatus(ctx context.Context, refCode string, status domain.ReferralStatus) error {
	res, err := r.db.ExecContext(ctx,
		`UPDATE referral_links SET status = ? WHERE ref_code = ? AND status = ?`,
		status, refCode, domain.StatusActive,
	)
	if err != nil {
		return fmt.Errorf("update status for ref_code %s: %w", refCode, err)
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("update status for ref_code %s: %w", refCode, err)
	}
	if rows == 0 {
		return fmt.Errorf("update status for ref_code %s: %w", refCode, domain.ErrLinkNotActive)
	}
	return nil
}

func (r *referralLinkRepo) CountByOwnerUserIDSince(ctx context.Context, ownerUserID string, sinceMs int64) (int64, error) {
	var count int64
	err := r.db.GetContext(ctx, &count,
		`SELECT COUNT(*) FROM referral_links WHERE owner_user_id = ? AND created_at >= ?`,
		ownerUserID, sinceMs,
	)
	if err != nil {
		return 0, fmt.Errorf("count links by owner %s since %d: %w", ownerUserID, sinceMs, err)
	}
	return count, nil
}

// linkCursor encodes the keyset pagination position (created_at, ref_code) for ListUserLinks.
type linkCursor struct {
	CreatedAt int64  `json:"c"`
	RefCode   string `json:"r"`
}

func encodeLinkCursor(createdAt int64, refCode string) string {
	b, _ := json.Marshal(linkCursor{CreatedAt: createdAt, RefCode: refCode})
	return base64.StdEncoding.EncodeToString(b)
}

func decodeLinkCursor(cursor string) (int64, string, error) {
	decoded, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return 0, "", err
	}
	var c linkCursor
	if err := json.Unmarshal(decoded, &c); err != nil {
		return 0, "", err
	}
	return c.CreatedAt, c.RefCode, nil
}
