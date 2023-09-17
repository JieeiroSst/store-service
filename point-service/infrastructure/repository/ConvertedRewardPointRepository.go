package repository

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/entity"
)

type ConvertedRewardPointRepository interface {
	Create(ctx context.Context, data entity.ConvertedRewardPoint) error
	Update(ctx context.Context, data entity.ConvertedRewardPoint) error
	GetAll(ctx context.Context, perPage int, sortOrder, cursor string) ([]entity.ConvertedRewardPoint, error)
	GetByID(ctx context.Context, id int) (entity.ConvertedRewardPoint, error)
}
