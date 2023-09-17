package service

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/dto"
)

type RewardPointService interface {
	Create(ctx context.Context, data dto.RewardPointDTO) error
	Update(ctx context.Context, data dto.RewardPointDTO) error
	GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*dto.ResponseDTO, error)
	GetByID(ctx context.Context, id string) (*dto.RewardPointDTO, error)
}
