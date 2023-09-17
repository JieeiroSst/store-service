package service

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/infrastructure/repository"
)

type RewardPointServiceImpl struct {
	rewardPointRepository repository.RewardPointRepository
}

func InitRewardPointServiceImpl(rewardPointRepository repository.RewardPointRepository) *RewardPointServiceImpl {
	return &RewardPointServiceImpl{
		rewardPointRepository: rewardPointRepository,
	}
}

func (s *RewardPointServiceImpl) Create(ctx context.Context, data dto.RewardPointDTO) error {
	entity := data.TransformDTOtoEntity()

	if err := s.rewardPointRepository.Create(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (s *RewardPointServiceImpl) Update(ctx context.Context, data dto.RewardPointDTO) error {
	entity := data.TransformDTOtoEntity()

	if err := s.rewardPointRepository.Update(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (s *RewardPointServiceImpl) GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*dto.ResponseDTO, error) {
	var responseDTO *dto.ResponseDTO
	resp, err := s.rewardPointRepository.GetAll(ctx, perPage, sortOrder, cursor)
	if err != nil {
		return nil, err
	}
	return responseDTO.TransformEntityToDto(*resp), nil
}

func (s *RewardPointServiceImpl) GetByID(ctx context.Context, id string) (*dto.RewardPointDTO, error) {
	var rewardPointDTO *dto.RewardPointDTO
	resp, err := s.rewardPointRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return rewardPointDTO.TransformEntityToDto(*resp), nil
}
