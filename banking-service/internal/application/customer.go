package application

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
)

type customerService struct {
	repo port.CustomerRepository
}

func NewCustomerService(repo port.CustomerRepository) port.CustomerUsecase {
	return &customerService{repo: repo}
}

func (s *customerService) CreateCustomer(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	if err := s.repo.Create(ctx, customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *customerService) GetCustomer(ctx context.Context, id int) (*model.Customer, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *customerService) ListCustomers(ctx context.Context) ([]model.Customer, error) {
	return s.repo.List(ctx)
}

func (s *customerService) UpdateCustomer(ctx context.Context, customer *model.Customer) (*model.Customer, error) {
	if _, err := s.repo.GetByID(ctx, customer.CustomerID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, customer); err != nil {
		return nil, err
	}
	return customer, nil
}

func (s *customerService) DeleteCustomer(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
