package service

import (
	"context"

	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/infrastructure/repository"
)

type ConvertedRewardPointServiceImpl struct {
	convertedRewardPointRepository repository.ConvertedRewardPointRepository
}

func InitConvertedRewardPointServiceImpl(convertedRewardPointRepository repository.ConvertedRewardPointRepository) *ConvertedRewardPointServiceImpl {

	return &ConvertedRewardPointServiceImpl{
		convertedRewardPointRepository: convertedRewardPointRepository,
	}
}

func (s *ConvertedRewardPointServiceImpl) Create(ctx context.Context, data dto.ConvertedRewardPointDTO) error

func (s *ConvertedRewardPointServiceImpl) Update(ctx context.Context, data dto.ConvertedRewardPointDTO) error

func (s *ConvertedRewardPointServiceImpl) GetAll(ctx context.Context, perPage int, sortOrder, cursor string) (*dto.ResponseDTO, error)

func (s *ConvertedRewardPointServiceImpl) GetByID(ctx context.Context, id string) (*dto.ConvertedRewardPointDTO, error)
