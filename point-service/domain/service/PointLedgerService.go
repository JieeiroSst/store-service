package service

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/point-service/domain/dto"
)

var ErrInvalidRewardEvent = errors.New("invalid reward event")

type PointLedgerService interface {
	Apply(ctx context.Context, event dto.RewardEventDTO) (bool, error)
	Balance(ctx context.Context, userID string) (*dto.PointAccountDTO, error)
	Transactions(ctx context.Context, userID string, limit, offset int) (*dto.PointTransactionPageDTO, error)
}
