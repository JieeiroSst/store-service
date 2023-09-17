package repository

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/entity"
)

type RewardPointRepository interface {
	Create(ctx context.Context, data entity.RewardPoint) error
	Update(ctx context.Context, data entity.RewardPoint) error
	GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*entity.ResponseEntity, error)
	GetByID(ctx context.Context, id string) (*entity.RewardPoint, error)
}
