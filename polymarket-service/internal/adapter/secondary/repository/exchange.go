package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/polymarket-service/internal/domain/model"
	"github.com/JIeeiroSst/polymarket-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type pnlRepository struct{ db *gorm.DB }

func NewPnLRepository(db *gorm.DB) *pnlRepository { return &pnlRepository{db: db} }

func (r *pnlRepository) CreateMany(ctx context.Context, entries []model.PnLEntry) error {
	if len(entries) == 0 {
		return nil
	}
	return conn(ctx, r.db).Create(&entries).Error
}

type exchangeAccountRepository struct{ db *gorm.DB }

func NewExchangeAccountRepository(db *gorm.DB) *exchangeAccountRepository {
	return &exchangeAccountRepository{db: db}
}

func (r *exchangeAccountRepository) Add(ctx context.Context, entries ...model.ExchangeEntry) error {
	if len(entries) == 0 {
		return nil
	}
	return conn(ctx, r.db).Create(&entries).Error
}

func (r *exchangeAccountRepository) Balance(ctx context.Context, bucket model.ExchangeBucket) (int64, error) {
	var sum int64
	err := conn(ctx, r.db).Model(&model.ExchangeEntry{}).Where("bucket = ?", bucket).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

type rewardRepository struct{ db *gorm.DB }

func NewRewardRepository(db *gorm.DB) *rewardRepository { return &rewardRepository{db: db} }

func (r *rewardRepository) ClaimEpoch(ctx context.Context, start time.Time) (bool, error) {
	res := conn(ctx, r.db).Exec("INSERT IGNORE INTO reward_epochs (epoch_start) VALUES (?)", start)
	return res.RowsAffected == 1, res.Error
}

func (r *rewardRepository) CreatePayouts(ctx context.Context, payouts []model.RewardPayout) error {
	if len(payouts) == 0 {
		return nil
	}
	return conn(ctx, r.db).Create(&payouts).Error
}

func (r *rewardRepository) ListPayouts(ctx context.Context, userID string, limit int, cursorStr string) (*port.Page[model.RewardPayout], error) {
	cur, err := decodeCursor(cursorStr, "rewards")
	if err != nil {
		return nil, err
	}
	q := conn(ctx, r.db).Where("user_id = ?", userID)
	if cur != nil {
		q = q.Where("id < ?", cur.ID)
	}
	var rows []model.RewardPayout
	if err := q.Order("id DESC").Limit(limit + 1).Find(&rows).Error; err != nil {
		return nil, err
	}
	return paginate(rows, limit, func(p model.RewardPayout) cursor { return cursor{S: "rewards", ID: p.ID} }), nil
}

func (r *rewardRepository) TotalPaid(ctx context.Context, userID string) (int64, error) {
	var sum int64
	err := conn(ctx, r.db).Model(&model.RewardPayout{}).Where("user_id = ?", userID).
		Select("COALESCE(SUM(amount), 0)").Scan(&sum).Error
	return sum, err
}

type referralRepository struct{ db *gorm.DB }

func NewReferralRepository(db *gorm.DB) *referralRepository { return &referralRepository{db: db} }

func (r *referralRepository) GetCode(ctx context.Context, userID string) (*model.ReferralCode, error) {
	var c model.ReferralCode
	if err := conn(ctx, r.db).Where("user_id = ?", userID).First(&c).Error; err != nil {
		return nil, notFound(err)
	}
	return &c, nil
}

// SaveCode keeps the first code a user was issued if two requests race.
func (r *referralRepository) SaveCode(ctx context.Context, code *model.ReferralCode) error {
	return conn(ctx, r.db).Clauses(clause.OnConflict{DoNothing: true}).Create(code).Error
}

func (r *referralRepository) GetReferral(ctx context.Context, refereeUserID string) (*model.Referral, error) {
	var ref model.Referral
	if err := conn(ctx, r.db).Where("referee_user_id = ?", refereeUserID).First(&ref).Error; err != nil {
		return nil, notFound(err)
	}
	return &ref, nil
}

func (r *referralRepository) SaveReferral(ctx context.Context, ref *model.Referral) error {
	return mapWriteErr(conn(ctx, r.db).Create(ref).Error)
}

func (r *referralRepository) CountReferees(ctx context.Context, referrerUserID string) (int64, error) {
	var n int64
	err := conn(ctx, r.db).Model(&model.Referral{}).Where("referrer_user_id = ?", referrerUserID).Count(&n).Error
	return n, err
}

func (r *referralRepository) Earnings(ctx context.Context, referrerUserID string) (int64, error) {
	var sum int64
	err := conn(ctx, r.db).Model(&model.Trade{}).Where("referrer_user_id = ?", referrerUserID).
		Select("COALESCE(SUM(referral_fee), 0)").Scan(&sum).Error
	return sum, err
}
