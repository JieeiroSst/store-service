package application

import (
	"context"
	"time"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/Jieeirosst/account-transaction-service/internal/domain/model"
	"github.com/Jieeirosst/account-transaction-service/internal/domain/port"
	"github.com/Jieeirosst/account-transaction-service/pkg/pagination"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type accountService struct {
	repo port.AccountRepository
}

func NewAccountService(repo port.AccountRepository) port.AccountUsecase {
	return &accountService{repo: repo}
}

func (s *accountService) CreateAccount(ctx context.Context, firstName, lastName string) (*model.Account, error) {
	lg := logger.WithContext(ctx)

	account := &model.Account{
		ID:          uuid.NewString(),
		FirstName:   firstName,
		LastName:    lastName,
		DateCreated: time.Now(),
	}
	if err := s.repo.Create(ctx, account); err != nil {
		lg.Error("CreateAccount", zap.Error(err))
		return nil, err
	}
	return account, nil
}

func (s *accountService) GetAccount(ctx context.Context, id string) (*model.Account, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *accountService) ListAccounts(ctx context.Context, pageSize int32, pageToken string) ([]model.Account, string, error) {
	p := pagination.Parse(pageSize, pageToken)

	accounts, err := s.repo.List(ctx, p)
	if err != nil {
		return nil, "", err
	}
	return accounts, p.NextToken(len(accounts)), nil
}
