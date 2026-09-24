package postgres

import (
	"context"
	"database/sql"
	"time"

	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/bonuslink-service/internal/core/domain"
)

type rewardRepo struct {
	db *sqlx.DB
}

func NewRewardRepo(db *sqlx.DB) *rewardRepo {
	return &rewardRepo{db: db}
}

func (r *rewardRepo) Record(ctx context.Context, reward *domain.Reward) (bool, error) {
	tx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return false, err
	}
	defer func() { _ = tx.Rollback() }()

	var createdAt time.Time
	err = tx.QueryRowxContext(ctx, `
		INSERT INTO bonus_rewards (reward_id, event_id, user_id, ref_code, source, reward_type, reward_value)
		VALUES ($1, $2, $3, NULLIF($4, ''), $5, $6, $7)
		ON CONFLICT (event_id) DO NOTHING
		RETURNING created_at`,
		reward.ID, reward.EventID, reward.UserID, reward.RefCode, reward.Source, reward.RewardType, reward.Value,
	).Scan(&createdAt)
	if err == sql.ErrNoRows {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	reward.CreatedAt = createdAt

	_, err = tx.ExecContext(ctx, `
		INSERT INTO bonus_balances (user_id, reward_type, total_value, reward_count)
		VALUES ($1, $2, $3, 1)
		ON CONFLICT (user_id, reward_type) DO UPDATE
		SET total_value  = bonus_balances.total_value + EXCLUDED.total_value,
		    reward_count = bonus_balances.reward_count + 1,
		    updated_at   = CURRENT_TIMESTAMP`,
		reward.UserID, reward.RewardType, reward.Value)
	if err != nil {
		return false, err
	}
	return true, tx.Commit()
}

func (r *rewardRepo) ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Reward, error) {
	rows, err := r.db.QueryxContext(ctx, `
		SELECT reward_id, event_id, user_id, COALESCE(ref_code, ''), source, reward_type, reward_value, created_at
		FROM bonus_rewards
		WHERE user_id = $1
		ORDER BY created_at DESC, reward_id
		LIMIT $2 OFFSET $3`, userID, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*domain.Reward{}
	for rows.Next() {
		var rw domain.Reward
		if err := rows.Scan(&rw.ID, &rw.EventID, &rw.UserID, &rw.RefCode, &rw.Source, &rw.RewardType, &rw.Value, &rw.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, &rw)
	}
	return out, rows.Err()
}

func (r *rewardRepo) BalancesByUser(ctx context.Context, userID string) ([]*domain.Balance, error) {
	rows, err := r.db.QueryxContext(ctx, `
		SELECT user_id, reward_type, total_value, reward_count
		FROM bonus_balances
		WHERE user_id = $1
		ORDER BY reward_type`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	out := []*domain.Balance{}
	for rows.Next() {
		var b domain.Balance
		if err := rows.Scan(&b.UserID, &b.RewardType, &b.TotalValue, &b.RewardCount); err != nil {
			return nil, err
		}
		out = append(out, &b)
	}
	return out, rows.Err()
}
