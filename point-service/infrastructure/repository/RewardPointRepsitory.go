package repository

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/entity"
)

type RewardPointRepository interface {
	Create(ctx context.Context, data entity.RewardDiscount) error
	Update(ctx context.Context, data entity.RewardDiscount) error
	GetAll(ctx context.Context, perPage int, sortOrder, cursor string) ([]entity.RewardDiscount, error)
	GetByID(ctx context.Context, id string) (*entity.RewardDiscount, error)
}
