package repository

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/entity"
)

type ConvertedRewardPointRepository interface {
	Create(ctx context.Context, data entity.ConvertedRewardPoint) error
	Update(ctx context.Context, data entity.ConvertedRewardPoint) error
	GetAll(ctx context.Context) ([]entity.ConvertedRewardPoint, error)
}
