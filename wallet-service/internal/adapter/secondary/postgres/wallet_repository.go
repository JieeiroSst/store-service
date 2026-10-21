package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const walletColumns = `id, user_id, balance, frozen_amount, currency, status, tier,
	pin_hash, pin_failed_attempts, pin_locked_until, (pin_hash <> '') AS has_pin,
	version, created_at, updated_at`

type walletRepository struct{ base }

func NewWalletRepository(db *sqlx.DB) port.WalletRepository {
	return &walletRepository{base{db: db}}
}

func (r *walletRepository) Create(ctx context.Context, w *model.Wallet) error {
	_, err := r.named(ctx, `INSERT INTO wallets
		(id, user_id, balance, frozen_amount, currency, status, tier, pin_hash, pin_failed_attempts,
		 pin_locked_until, version, created_at, updated_at)
		VALUES (:id, :user_id, :balance, :frozen_amount, :currency, :status, :tier, :pin_hash,
		 :pin_failed_attempts, :pin_locked_until, :version, :created_at, :updated_at)`, w)
	if uniqueViolation(err, "wallets_user_currency_key") {
		return model.ErrWalletExists
	}
	return err
}

func (r *walletRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	var w model.Wallet
	if err := r.get(ctx, &w, model.ErrWalletNotFound, `SELECT `+walletColumns+` FROM wallets WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *walletRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Wallet, error) {
	var w model.Wallet
	if err := r.get(ctx, &w, model.ErrWalletNotFound, `SELECT `+walletColumns+` FROM wallets WHERE id = $1 FOR UPDATE`, id); err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *walletRepository) GetByUserAndCurrency(ctx context.Context, userID uuid.UUID, currency model.Currency) (*model.Wallet, error) {
	var w model.Wallet
	err := r.get(ctx, &w, model.ErrWalletNotFound,
		`SELECT `+walletColumns+` FROM wallets WHERE user_id = $1 AND currency = $2 AND status <> 'CLOSED'`, userID, currency)
	if err != nil {
		return nil, err
	}
	return &w, nil
}

func (r *walletRepository) ListByUser(ctx context.Context, userID uuid.UUID) ([]*model.Wallet, error) {
	wallets := make([]*model.Wallet, 0)
	err := sqlx.SelectContext(ctx, r.q(ctx), &wallets,
		`SELECT `+walletColumns+` FROM wallets WHERE user_id = $1 ORDER BY created_at ASC`, userID)
	return wallets, err
}

func (r *walletRepository) Update(ctx context.Context, w *model.Wallet) error {
	res, err := r.q(ctx).ExecContext(ctx, `UPDATE wallets
		SET balance = $1, frozen_amount = $2, status = $3, tier = $4, pin_hash = $5,
		    pin_failed_attempts = $6, pin_locked_until = $7, version = version + 1, updated_at = $8
		WHERE id = $9 AND version = $10`,
		w.Balance, w.FrozenAmount, w.Status, w.Tier, w.PinHash,
		w.PinFailedAttempts, w.PinLockedUntil, w.UpdatedAt, w.ID, w.Version)
	if err != nil {
		return err
	}
	if err := expectOne(res, model.ErrVersionConflict); err != nil {
		return err
	}
	w.Version++
	return nil
}
