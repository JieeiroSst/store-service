package service

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/infrastructure/repository"
)

type RewardDiscountServiceImpl struct {
	rewardDiscountReposiory repository.RewardDiscountRepository
}

func InitRewardDiscountServiceImpl(rewardDiscountReposiory repository.RewardDiscountRepository) *RewardDiscountServiceImpl {
	return &RewardDiscountServiceImpl{
		rewardDiscountReposiory: rewardDiscountReposiory,
	}
}

func (s *RewardDiscountServiceImpl) Create(ctx context.Context, data dto.RewardDiscountDTO) error

func (s *RewardDiscountServiceImpl) Update(ctx context.Context, data dto.RewardDiscountDTO) error

func (s *RewardDiscountServiceImpl) GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*dto.ResponseDTO, error)

func (s *RewardDiscountServiceImpl) GetByID(ctx context.Context, id string) (*dto.RewardDiscountDTO, error)
