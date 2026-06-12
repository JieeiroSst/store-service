package mysql

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"go.uber.org/zap"

	"github.com/referral/service/internal/core/domain"
	"github.com/referral/service/internal/core/ports"
)

type referralEventRepo struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewReferralEventRepo(db *sqlx.DB, log *zap.Logger) ports.ReferralEventRepository {
	return &referralEventRepo{
		db:  db,
		log: log.Named("event-repo"),
	}
}

func (r *referralEventRepo) Save(ctx context.Context, event *domain.ReferralEvent) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO referral_events (event_id, ref_code, event_type, occurred_at, platform, channel, new_user_id, owner_user_id, ip_address, device_id, user_agent, firebase_instance_id)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		event.EventID, event.RefCode, event.EventType, event.OccurredAt, event.Platform, event.Channel,
		event.NewUserID, event.OwnerUserID, event.IPAddress, event.DeviceID, event.UserAgent, event.FirebaseInstanceID,
	)
	if err != nil {
		return fmt.Errorf("insert event: %w", err)
	}
	r.log.Debug("event saved",
		zap.String("ref_code", event.RefCode),
		zap.String("event_type", string(event.EventType)),
	)
	return nil
}

func (r *referralEventRepo) FindByRefCode(ctx context.Context, refCode string) ([]*domain.ReferralEvent, error) {
	var events []*domain.ReferralEvent
	err := r.db.SelectContext(ctx, &events,
		`SELECT event_id, ref_code, event_type, occurred_at, platform, channel, new_user_id, owner_user_id, ip_address, device_id, user_agent, firebase_instance_id
		 FROM referral_events WHERE ref_code = ? ORDER BY occurred_at DESC`,
		refCode,
	)
	if err != nil {
		return nil, fmt.Errorf("query events for ref_code %s: %w", refCode, err)
	}
	return events, nil
}

func (r *referralEventRepo) FindByNewUserID(ctx context.Context, newUserID string) ([]*domain.ReferralEvent, error) {
	var events []*domain.ReferralEvent
	err := r.db.SelectContext(ctx, &events,
		`SELECT event_id, ref_code, event_type, occurred_at, platform, channel, new_user_id, owner_user_id, ip_address, device_id, user_agent, firebase_instance_id
		 FROM referral_events WHERE new_user_id = ?`,
		newUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("query events for user %s: %w", newUserID, err)
	}
	return events, nil
}

type rewardRepo struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewRewardRepo(db *sqlx.DB, log *zap.Logger) ports.RewardRepository {
	return &rewardRepo{
		db:  db,
		log: log.Named("reward-repo"),
	}
}

func (r *rewardRepo) Save(ctx context.Context, reward *domain.ReferralReward) error {
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO referral_rewards (owner_user_id, ref_code, new_user_id, reward_type, reward_value, status, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		reward.OwnerUserID, reward.RefCode, reward.NewUserID, reward.RewardType,
		reward.RewardValue, reward.Status, reward.CreatedAt, reward.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert reward: %w", err)
	}
	r.log.Info("reward saved",
		zap.String("owner", reward.OwnerUserID),
		zap.String("ref_code", reward.RefCode),
		zap.Float64("value", reward.RewardValue),
	)
	return nil
}

func (r *rewardRepo) FindByOwnerAndRefCode(ctx context.Context, ownerUserID, refCode string) (*domain.ReferralReward, error) {
	var reward domain.ReferralReward
	err := r.db.GetContext(ctx, &reward,
		`SELECT owner_user_id, ref_code, new_user_id, reward_type, reward_value, status, created_at, updated_at
		 FROM referral_rewards WHERE owner_user_id = ? AND ref_code = ?`,
		ownerUserID, refCode,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("reward %s/%s: %w", ownerUserID, refCode, domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get reward %s/%s: %w", ownerUserID, refCode, err)
	}
	return &reward, nil
}

func (r *rewardRepo) FindByOwnerUserID(ctx context.Context, ownerUserID string) ([]*domain.ReferralReward, error) {
	var rewards []*domain.ReferralReward
	err := r.db.SelectContext(ctx, &rewards,
		`SELECT owner_user_id, ref_code, new_user_id, reward_type, reward_value, status, created_at, updated_at
		 FROM referral_rewards WHERE owner_user_id = ? ORDER BY created_at DESC`,
		ownerUserID,
	)
	if err != nil {
		return nil, fmt.Errorf("query rewards for owner %s: %w", ownerUserID, err)
	}
	return rewards, nil
}

type rewardProgramRepo struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewRewardProgramRepo(db *sqlx.DB, log *zap.Logger) ports.RewardProgramRepository {
	return &rewardProgramRepo{
		db:  db,
		log: log.Named("program-repo"),
	}
}

// rewardProgramRow mirrors reward_programs columns for scanning; Tiers is stored as
// a JSON column and converted by hand since domain.RewardProgram.Tiers is tagged db:"-".
type rewardProgramRow struct {
	ProgramID string                     `db:"program_id"`
	Name      string                     `db:"name"`
	Status    domain.RewardProgramStatus `db:"status"`
	Tiers     []byte                     `db:"tiers"`
	CreatedAt int64                      `db:"created_at"`
	UpdatedAt int64                      `db:"updated_at"`
}

func (row *rewardProgramRow) toDomain() (*domain.RewardProgram, error) {
	var tiers []domain.RewardTier
	if err := json.Unmarshal(row.Tiers, &tiers); err != nil {
		return nil, fmt.Errorf("unmarshal tiers: %w", err)
	}
	return &domain.RewardProgram{
		ProgramID: row.ProgramID,
		Name:      row.Name,
		Status:    row.Status,
		Tiers:     tiers,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *rewardProgramRepo) Save(ctx context.Context, program *domain.RewardProgram) error {
	tiersJSON, err := json.Marshal(program.Tiers)
	if err != nil {
		return fmt.Errorf("marshal tiers: %w", err)
	}

	_, err = r.db.ExecContext(ctx,
		`INSERT INTO reward_programs (program_id, name, status, tiers, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		program.ProgramID, program.Name, program.Status, tiersJSON, program.CreatedAt, program.UpdatedAt,
	)
	if err != nil {
		return fmt.Errorf("insert reward program: %w", err)
	}
	r.log.Info("reward program saved", zap.String("program_id", program.ProgramID))
	return nil
}

func (r *rewardProgramRepo) FindByID(ctx context.Context, programID string) (*domain.RewardProgram, error) {
	var row rewardProgramRow
	err := r.db.GetContext(ctx, &row,
		`SELECT program_id, name, status, tiers, created_at, updated_at FROM reward_programs WHERE program_id = ?`,
		programID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("reward program %s: %w", programID, domain.ErrNotFound)
	}
	if err != nil {
		return nil, fmt.Errorf("get reward program %s: %w", programID, err)
	}
	return row.toDomain()
}

// FindActive queries the status index for the single active program.
func (r *rewardProgramRepo) FindActive(ctx context.Context) (*domain.RewardProgram, error) {
	var row rewardProgramRow
	err := r.db.GetContext(ctx, &row,
		`SELECT program_id, name, status, tiers, created_at, updated_at FROM reward_programs WHERE status = ? LIMIT 1`,
		domain.ProgramStatusActive,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("query active reward program: %w", err)
	}
	return row.toDomain()
}

func (r *rewardProgramRepo) UpdateStatus(ctx context.Context, programID string, status domain.RewardProgramStatus) error {
	_, err := r.db.ExecContext(ctx,
		`UPDATE reward_programs SET status = ?, updated_at = ? WHERE program_id = ?`,
		status, time.Now().UnixMilli(), programID,
	)
	if err != nil {
		return fmt.Errorf("update reward program status %s: %w", programID, err)
	}
	return nil
}

type userStatsRepo struct {
	db  *sqlx.DB
	log *zap.Logger
}

func NewUserStatsRepo(db *sqlx.DB, log *zap.Logger) ports.UserStatsRepository {
	return &userStatsRepo{
		db:  db,
		log: log.Named("stats-repo"),
	}
}

func (r *userStatsRepo) Get(ctx context.Context, userID string) (*domain.UserReferralStats, error) {
	var stats domain.UserReferralStats
	err := r.db.GetContext(ctx, &stats,
		`SELECT user_id, total_invited, total_installed, total_rewarded, total_reward_amt, last_active_at
		 FROM user_referral_stats WHERE user_id = ?`,
		userID,
	)
	if errors.Is(err, sql.ErrNoRows) {
		return &domain.UserReferralStats{UserID: userID}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("get stats for %s: %w", userID, err)
	}
	return &stats, nil
}

func (r *userStatsRepo) IncrementCounters(
	ctx context.Context,
	userID string,
	invited, installed, rewarded int64,
	rewardAmt float64,
) error {
	nowMs := time.Now().UnixMilli()

	_, err := r.db.ExecContext(ctx,
		`INSERT INTO user_referral_stats (user_id, total_invited, total_installed, total_rewarded, total_reward_amt, last_active_at)
		 VALUES (?, ?, ?, ?, ?, ?)
		 ON DUPLICATE KEY UPDATE
		   total_invited = total_invited + VALUES(total_invited),
		   total_installed = total_installed + VALUES(total_installed),
		   total_rewarded = total_rewarded + VALUES(total_rewarded),
		   total_reward_amt = total_reward_amt + VALUES(total_reward_amt),
		   last_active_at = VALUES(last_active_at)`,
		userID, invited, installed, rewarded, rewardAmt, nowMs,
	)
	if err != nil {
		return fmt.Errorf("increment stats for %s: %w", userID, err)
	}
	return nil
}
