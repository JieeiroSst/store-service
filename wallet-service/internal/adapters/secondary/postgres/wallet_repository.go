package postgres

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
)

type walletRepository struct{ db *sqlx.DB }

func NewWalletRepository(db *sqlx.DB) ports.WalletRepository {
	return &walletRepository{db: db}
}

func (r *walletRepository) Create(ctx context.Context, wallet *domain.Wallet) error {
	q := `INSERT INTO wallets (id, user_id, balance, frozen_amount, currency, status, version, created_at, updated_at)
	      VALUES (:id, :user_id, :balance, :frozen_amount, :currency, :status, :version, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, wallet)
	return err
}

func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Wallet, error) {
	var w domain.Wallet
	err := r.db.GetContext(ctx, &w, `SELECT * FROM wallets WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrWalletNotFound
	}
	return &w, err
}

func (r *walletRepository) GetByUserID(ctx context.Context, userID uuid.UUID) (*domain.Wallet, error) {
	var w domain.Wallet
	err := r.db.GetContext(ctx, &w, `SELECT * FROM wallets WHERE user_id = $1`, userID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrWalletNotFound
	}
	return &w, err
}

func (r *walletRepository) UpdateWithVersion(ctx context.Context, wallet *domain.Wallet) error {
	wallet.UpdatedAt = time.Now()
	nextVersion := wallet.Version + 1
	q := `UPDATE wallets
	      SET balance=$1, frozen_amount=$2, status=$3, version=$4, updated_at=$5
	      WHERE id=$6 AND version=$7`
	res, err := r.db.ExecContext(ctx, q,
		wallet.Balance, wallet.FrozenAmount, wallet.Status,
		nextVersion, wallet.UpdatedAt, wallet.ID, wallet.Version)
	if err != nil {
		return err
	}
	affected, _ := res.RowsAffected()
	if affected == 0 {
		return domain.ErrVersionConflict
	}
	wallet.Version = nextVersion
	return nil
}

func (r *walletRepository) List(ctx context.Context, page, pageSize int) ([]*domain.Wallet, int64, error) {
	offset := (page - 1) * pageSize
	var wallets []*domain.Wallet
	err := r.db.SelectContext(ctx, &wallets, `SELECT * FROM wallets ORDER BY created_at DESC LIMIT $1 OFFSET $2`, pageSize, offset)
	if err != nil {
		return nil, 0, err
	}
	var count int64
	r.db.GetContext(ctx, &count, `SELECT COUNT(*) FROM wallets`)
	return wallets, count, nil
}
