package repository

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/entity"
)

type RewardDiscountRepository interface {
	Create(ctx context.Context, data entity.RewardDiscount) error
	Update(ctx context.Context, data entity.RewardDiscount) error
	GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*entity.ResponseEntity, error)
	GetByID(ctx context.Context, id string) (*entity.RewardDiscount, error)
}
