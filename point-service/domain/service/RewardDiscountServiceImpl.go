package service

import (
	"context"
	"log"

	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/infrastructure/persistence"
	"github.com/JIeeiroSst/point-service/infrastructure/repository"
)

type RewardDiscountServiceImpl struct {
	rewardDiscountReposiory repository.RewardDiscountRepository
}

func InitRewardDiscountServiceImpl(dns string) *RewardDiscountServiceImpl {
	dbHelper, err := persistence.InitDbHelper(dns)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &RewardDiscountServiceImpl{
		rewardDiscountReposiory: dbHelper.RewardDiscountRepository,
	}
}

func (s *RewardDiscountServiceImpl) Create(ctx context.Context, data dto.RewardDiscountDTO) error {
	entity := data.TransformDTOtoEntity()

	if err := s.rewardDiscountReposiory.Create(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (s *RewardDiscountServiceImpl) Update(ctx context.Context, data dto.RewardDiscountDTO) error {
	entity := data.TransformDTOtoEntity()

	if err := s.rewardDiscountReposiory.Update(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (s *RewardDiscountServiceImpl) GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*dto.ResponseDTO, error) {
	var responseDTO *dto.ResponseDTO
	resp, err := s.rewardDiscountReposiory.GetAll(ctx, perPage, sortOrder, cursor)
	if err != nil {
		return nil, err
	}
	return responseDTO.TransformEntityToDto(*resp), nil
}

func (s *RewardDiscountServiceImpl) GetByID(ctx context.Context, id string) (*dto.RewardDiscountDTO, error) {
	var rewardDiscountDTO *dto.RewardDiscountDTO

	resp, err := s.rewardDiscountReposiory.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return rewardDiscountDTO.TransformEntityToDto(*resp), nil
}
