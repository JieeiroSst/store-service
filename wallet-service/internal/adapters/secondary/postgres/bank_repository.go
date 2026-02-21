package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)

type bankRepository struct{ db *sqlx.DB }

func NewBankRepository(db *sqlx.DB) ports.BankRepository {
	return &bankRepository{db: db}
}

func (r *bankRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Bank, error) {
	var bank domain.Bank
	err := r.db.GetContext(ctx, &bank, `SELECT * FROM banks WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("bank not found")
	}
	return &bank, err
}

type merchantRepository struct{ db *sqlx.DB }

func NewMerchantRepository(db *sqlx.DB) ports.MerchantRepository {
	return &merchantRepository{db: db}
}

func (r *merchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Merchant, error) {
	var m domain.Merchant
	err := r.db.GetContext(ctx, &m, `SELECT * FROM merchants WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, errors.New("merchant not found")
	}
	return &m, err
}
