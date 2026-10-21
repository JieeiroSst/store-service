package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const bankAccountColumns = `id, user_id, bank_id, account_number, account_number_hash, account_name,
	is_default, status, created_at, updated_at`

type bankAccountRepository struct{ base }

func NewBankAccountRepository(db *sqlx.DB) port.BankAccountRepository {
	return &bankAccountRepository{base{db: db}}
}

func (r *bankAccountRepository) Create(ctx context.Context, a *model.BankAccount) error {
	_, err := r.named(ctx, `INSERT INTO bank_accounts (`+bankAccountColumns+`)
		VALUES (:id, :user_id, :bank_id, :account_number, :account_number_hash, :account_name,
		 :is_default, :status, :created_at, :updated_at)`, a)
	if uniqueViolation(err, "idx_bank_accounts_active_number") {
		return model.ErrBankAccountFound
	}
	return err
}

func (r *bankAccountRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.BankAccount, error) {
	var a model.BankAccount
	if err := r.get(ctx, &a, model.ErrBankAccountAbsent, `SELECT `+bankAccountColumns+` FROM bank_accounts WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &a, nil
}

func (r *bankAccountRepository) ListActiveByUser(ctx context.Context, userID uuid.UUID) ([]*model.BankAccount, error) {
	accounts := make([]*model.BankAccount, 0)
	err := sqlx.SelectContext(ctx, r.q(ctx), &accounts, `SELECT `+bankAccountColumns+` FROM bank_accounts
		WHERE user_id = $1 AND status = 'ACTIVE' ORDER BY is_default DESC, created_at ASC`, userID)
	return accounts, err
}

func (r *bankAccountRepository) Update(ctx context.Context, a *model.BankAccount) error {
	res, err := r.named(ctx, `UPDATE bank_accounts SET is_default = :is_default, status = :status, updated_at = :updated_at
		WHERE id = :id`, a)
	if err != nil {
		return err
	}
	return expectOne(res, model.ErrBankAccountAbsent)
}

func (r *bankAccountRepository) ClearDefault(ctx context.Context, userID uuid.UUID) error {
	_, err := r.q(ctx).ExecContext(ctx, `UPDATE bank_accounts SET is_default = FALSE, updated_at = NOW()
		WHERE user_id = $1 AND is_default`, userID)
	return err
}
