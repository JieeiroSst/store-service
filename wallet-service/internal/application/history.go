package application

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const maxStatementRange = 366 * 24 * time.Hour

type historyService struct {
	core
}

func NewHistoryService(d Deps) port.HistoryService {
	return &historyService{core: newCore(d)}
}

func (s *historyService) Transaction(ctx context.Context, txID uuid.UUID) (*model.Transaction, error) {
	key := fmt.Sprintf(transactionCacheKey, txID)
	if tx, ok := cacheGet[model.Transaction](ctx, &s.core, key); ok {
		return tx, nil
	}
	tx, err := s.Deps.Transactions.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	cacheSet(ctx, &s.core, key, tx, transactionCacheTTL)
	return tx, nil
}

func (s *historyService) Transactions(ctx context.Context, filter model.TransactionFilter) ([]*model.Transaction, int64, error) {
	if _, err := s.Wallets.GetByID(ctx, filter.WalletID); err != nil {
		return nil, 0, err
	}
	if err := validateRange(filter.From, filter.To); err != nil {
		return nil, 0, err
	}
	filter.Page, filter.PageSize = model.NormalizePage(filter.Page, filter.PageSize)
	return s.Deps.Transactions.List(ctx, filter)
}

func (s *historyService) Ledger(ctx context.Context, walletID uuid.UUID, from, to *time.Time, page, pageSize int) ([]*model.LedgerEntry, int64, error) {
	if _, err := s.Wallets.GetByID(ctx, walletID); err != nil {
		return nil, 0, err
	}
	if err := validateRange(from, to); err != nil {
		return nil, 0, err
	}
	page, pageSize = model.NormalizePage(page, pageSize)
	return s.Deps.Ledger.List(ctx, walletID, from, to, page, pageSize)
}

func (s *historyService) Statement(ctx context.Context, walletID uuid.UUID, from, to time.Time) (*model.Statement, error) {
	w, err := s.Wallets.GetByID(ctx, walletID)
	if err != nil {
		return nil, err
	}
	if !from.Before(to) {
		return nil, model.Invalid("from must be before to")
	}
	if to.Sub(from) > maxStatementRange {
		return nil, model.Invalid("statement range must not exceed one year")
	}
	summary, err := s.Deps.Ledger.Summarize(ctx, walletID, from, to)
	if err != nil {
		return nil, err
	}
	st := model.NewStatement(w, from, to, summary)
	return &st, nil
}

func validateRange(from, to *time.Time) error {
	if from != nil && to != nil && from.After(*to) {
		return model.Invalid("from must not be after to")
	}
	return nil
}
