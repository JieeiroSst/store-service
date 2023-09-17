package service

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/dto"
)

type ConvertedRewardPointService interface {
	Create(ctx context.Context, data dto.ConvertedRewardPointDTO) error
	Update(ctx context.Context, data dto.ConvertedRewardPointDTO) error
	GetAll(ctx context.Context, perPage , sortOrder, cursor string) (*dto.ResponseDTO, error)
	GetByID(ctx context.Context, id string) (*dto.ConvertedRewardPointDTO, error)
}
