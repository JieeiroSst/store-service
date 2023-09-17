package service

import (
	"context"
	"log"

	"github.com/JIeeiroSst/point-service/domain/dto"
	"github.com/JIeeiroSst/point-service/infrastructure/persistence"
	"github.com/JIeeiroSst/point-service/infrastructure/repository"
)

type ConvertedRewardPointServiceImpl struct {
	convertedRewardPointRepository repository.ConvertedRewardPointRepository
}

func InitConvertedRewardPointServiceImpl(dns string) *ConvertedRewardPointServiceImpl {
	dbHelper, err := persistence.InitDbHelper(dns)
	if err != nil {
		log.Fatal(err.Error())
	}
	return &ConvertedRewardPointServiceImpl{
		convertedRewardPointRepository: dbHelper.ConvertedRewardPointRepository,
	}
}

func (s *ConvertedRewardPointServiceImpl) Create(ctx context.Context, data dto.ConvertedRewardPointDTO) error {
	entity := data.TransformDTOtoEntity()

	if err := s.convertedRewardPointRepository.Create(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (s *ConvertedRewardPointServiceImpl) Update(ctx context.Context, data dto.ConvertedRewardPointDTO) error {
	entity := data.TransformDTOtoEntity()

	if err := s.convertedRewardPointRepository.Update(ctx, entity); err != nil {
		return err
	}
	return nil
}

func (s *ConvertedRewardPointServiceImpl) GetAll(ctx context.Context, perPage, sortOrder, cursor string) (*dto.ResponseDTO, error) {
	var responseDTO *dto.ResponseDTO
	resp, err := s.convertedRewardPointRepository.GetAll(ctx, perPage, sortOrder, cursor)
	if err != nil {
		return nil, err
	}
	return responseDTO.TransformEntityToDto(*resp), nil
}

func (s *ConvertedRewardPointServiceImpl) GetByID(ctx context.Context, id string) (*dto.ConvertedRewardPointDTO, error) {
	var convertedRewardPointDTO *dto.ConvertedRewardPointDTO
	resp, err := s.convertedRewardPointRepository.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return convertedRewardPointDTO.TransformEntityToDto(*resp), nil
}
