package application

import (
	"context"

	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
)

type publisherService struct {
	repo port.PublisherRepository
}

func NewPublisherService(repo port.PublisherRepository) port.PublisherUsecase {
	return &publisherService{repo: repo}
}

func (s *publisherService) CreatePublisher(ctx context.Context, publisher *model.Publisher) (*model.Publisher, error) {
	if err := s.repo.Create(ctx, publisher); err != nil {
		return nil, err
	}
	return publisher, nil
}

func (s *publisherService) GetPublisher(ctx context.Context, id int) (*model.Publisher, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *publisherService) ListPublishers(ctx context.Context) ([]model.Publisher, error) {
	return s.repo.List(ctx)
}

func (s *publisherService) UpdatePublisher(ctx context.Context, publisher *model.Publisher) (*model.Publisher, error) {
	if _, err := s.repo.GetByID(ctx, publisher.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, publisher); err != nil {
		return nil, err
	}
	return publisher, nil
}

func (s *publisherService) DeletePublisher(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
