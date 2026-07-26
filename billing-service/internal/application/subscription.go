package application

import (
	"context"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
)

type subscriptionService struct {
	repo port.SubscriptionRepository
}

func NewSubscriptionService(repo port.SubscriptionRepository) port.SubscriptionUsecase {
	return &subscriptionService{repo: repo}
}

func (s *subscriptionService) Create(ctx context.Context, subscription *model.Subscription) error {
	return s.repo.Create(ctx, subscription)
}

func (s *subscriptionService) Get(ctx context.Context, id int) (*model.Subscription, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *subscriptionService) Update(ctx context.Context, subscription *model.Subscription) error {
	return s.repo.Update(ctx, subscription)
}

func (s *subscriptionService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *subscriptionService) List(ctx context.Context) ([]model.Subscription, error) {
	return s.repo.List(ctx)
}
