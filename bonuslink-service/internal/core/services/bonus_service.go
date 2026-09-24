package services

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"go.uber.org/fx"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/bonuslink-service/internal/core/domain"
	"github.com/JIeeiroSst/bonuslink-service/internal/core/ports"
)

var Module = fx.Options(
	fx.Provide(fx.Annotate(NewBonusService, fx.As(new(ports.BonusService)))),
)

const defaultSource = "referral"

type bonusService struct {
	rewards ports.RewardRepository
	log     *zap.Logger
}

func NewBonusService(rewards ports.RewardRepository, log *zap.Logger) *bonusService {
	return &bonusService{rewards: rewards, log: log.Named("bonus-service")}
}

func (s *bonusService) RecordReward(ctx context.Context, req ports.RecordRewardRequest) (*domain.Reward, bool, error) {
	req.EventID = strings.TrimSpace(req.EventID)
	req.UserID = strings.TrimSpace(req.UserID)
	switch {
	case req.EventID == "":
		return nil, false, fmt.Errorf("%w: event_id is required", domain.ErrInvalidReward)
	case req.UserID == "":
		return nil, false, fmt.Errorf("%w: user_id is required", domain.ErrInvalidReward)
	case !req.RewardType.Valid():
		return nil, false, fmt.Errorf("%w: unknown reward_type %q", domain.ErrInvalidReward, req.RewardType)
	case req.Value <= 0:
		return nil, false, fmt.Errorf("%w: reward_value must be positive", domain.ErrInvalidReward)
	}
	if req.Source == "" {
		req.Source = defaultSource
	}

	reward := &domain.Reward{
		ID:         uuid.NewString(),
		EventID:    req.EventID,
		UserID:     req.UserID,
		RefCode:    req.RefCode,
		Source:     req.Source,
		RewardType: req.RewardType,
		Value:      req.Value,
	}
	created, err := s.rewards.Record(ctx, reward)
	if err != nil {
		return nil, false, fmt.Errorf("record reward: %w", err)
	}
	if !created {
		s.log.Info("duplicate reward event ignored", zap.String("event_id", req.EventID))
		return nil, false, nil
	}
	s.log.Info("reward recorded",
		zap.String("event_id", req.EventID),
		zap.String("user_id", req.UserID),
		zap.String("reward_type", string(req.RewardType)),
		zap.Float64("value", req.Value),
	)
	return reward, true, nil
}

func (s *bonusService) ListUserRewards(ctx context.Context, userID string, limit, offset int) ([]*domain.Reward, error) {
	return s.rewards.ListByUser(ctx, userID, limit, offset)
}

func (s *bonusService) GetUserBalances(ctx context.Context, userID string) ([]*domain.Balance, error) {
	return s.rewards.BalancesByUser(ctx, userID)
}
