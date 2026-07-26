package application

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
)

type accountService struct {
	repo port.AccountRepository
}

func NewAccountService(repo port.AccountRepository) port.AccountUsecase {
	return &accountService{repo: repo}
}

func (s *accountService) CreateAccount(ctx context.Context, account *model.Account) (*model.Account, error) {
	if err := s.repo.Create(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *accountService) GetAccount(ctx context.Context, id int) (*model.Account, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *accountService) ListAccounts(ctx context.Context) ([]model.Account, error) {
	return s.repo.List(ctx)
}

func (s *accountService) UpdateAccount(ctx context.Context, account *model.Account) (*model.Account, error) {
	if _, err := s.repo.GetByID(ctx, account.AccountID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, account); err != nil {
		return nil, err
	}
	return account, nil
}

func (s *accountService) DeleteAccount(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
