package application

import (
	"context"

	"github.com/JIeeiroSst/billing-service/internal/domain/model"
	"github.com/JIeeiroSst/billing-service/internal/domain/port"
)

type customerService struct {
	repo port.CustomerRepository
}

func NewCustomerService(repo port.CustomerRepository) port.CustomerUsecase {
	return &customerService{repo: repo}
}

func (s *customerService) Create(ctx context.Context, customer *model.Customer) error {
	return s.repo.Create(ctx, customer)
}

func (s *customerService) Get(ctx context.Context, id int) (*model.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *customerService) Update(ctx context.Context, customer *model.Customer) error {
	return s.repo.Update(ctx, customer)
}

func (s *customerService) Delete(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *customerService) List(ctx context.Context) ([]model.Customer, error) {
	return s.repo.List(ctx)
}
