package service

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/dto"
)

type RewardDiscountService interface {
	Create(ctx context.Context, data dto.RewardDiscountDTO) error
	Update(ctx context.Context, data dto.RewardDiscountDTO) error
	GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*dto.ResponseDTO, error)
	GetByID(ctx context.Context, id string) (*dto.RewardDiscountDTO, error)
}
