package ports

import (
	"context"

	"github.com/JIeeiroSst/bonuslink-service/internal/core/domain"
)

type BonusService interface {
	RecordReward(ctx context.Context, req RecordRewardRequest) (*domain.Reward, bool, error)
	ListUserRewards(ctx context.Context, userID string, limit, offset int) ([]*domain.Reward, error)
	GetUserBalances(ctx context.Context, userID string) ([]*domain.Balance, error)
}

type RecordRewardRequest struct {
	EventID    string
	UserID     string
	RefCode    string
	Source     string
	RewardType domain.RewardType
	Value      float64
}

type RewardRepository interface {
	Record(ctx context.Context, reward *domain.Reward) (bool, error)
	ListByUser(ctx context.Context, userID string, limit, offset int) ([]*domain.Reward, error)
	BalancesByUser(ctx context.Context, userID string) ([]*domain.Balance, error)
}
