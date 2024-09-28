package usecase

import (
	"context"

	"github.com/JIeeiroSst/consumer-service/internal/dto"
	"github.com/JIeeiroSst/consumer-service/internal/repository"
	"github.com/JieeiroSst/logger"
)

type Consumers interface {
	Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error)
	Create(ctx context.Context, consumer dto.Consumer) error
}

type ConsumerUsecase struct {
	ConsumerRepository repository.Consumers
}

func NewConsumerUsecase(ConsumerRepository repository.Consumers) *ConsumerUsecase {
	return &ConsumerUsecase{
		ConsumerRepository: ConsumerRepository,
	}
}

func (u *ConsumerUsecase) Find(ctx context.Context, pagination logger.Pagination) (logger.Pagination, error) {
	consumers, err := u.ConsumerRepository.Find(ctx, pagination)
	if err != nil {
		return logger.Pagination{}, nil
	}
	return consumers, nil
}

func (u *ConsumerUsecase) Create(ctx context.Context, consumer dto.Consumer) error {
	model := consumer.Build()

	if err := u.ConsumerRepository.Create(ctx, model); err != nil {
		return err
	}
	return nil
}
