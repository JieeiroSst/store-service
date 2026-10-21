package postgres

import (
	"context"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const cardColumns = `id, wallet_id, issuer_bank_id, card_number, card_number_hash, holder_name,
	network, expiry_month, expiry_year, status, created_at, updated_at`

type cardRepository struct{ base }

func NewCardRepository(db *sqlx.DB) port.CardRepository {
	return &cardRepository{base{db: db}}
}

func (r *cardRepository) Create(ctx context.Context, card *model.Card) error {
	_, err := r.named(ctx, `INSERT INTO cards (`+cardColumns+`)
		VALUES (:id, :wallet_id, :issuer_bank_id, :card_number, :card_number_hash, :holder_name,
		 :network, :expiry_month, :expiry_year, :status, :created_at, :updated_at)`, card)
	if uniqueViolation(err, "cards_card_number_hash_key") {
		return model.ErrCardNumberTaken
	}
	return err
}

func (r *cardRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Card, error) {
	var c model.Card
	if err := r.get(ctx, &c, model.ErrCardNotFound, `SELECT `+cardColumns+` FROM cards WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &c, nil
}

func (r *cardRepository) ListByWallet(ctx context.Context, walletID uuid.UUID) ([]*model.Card, error) {
	cards := make([]*model.Card, 0)
	err := sqlx.SelectContext(ctx, r.q(ctx), &cards,
		`SELECT `+cardColumns+` FROM cards WHERE wallet_id = $1 ORDER BY created_at DESC`, walletID)
	return cards, err
}

func (r *cardRepository) Update(ctx context.Context, card *model.Card) error {
	res, err := r.named(ctx, `UPDATE cards SET status = :status, holder_name = :holder_name, updated_at = :updated_at
		WHERE id = :id`, card)
	if err != nil {
		return err
	}
	return expectOne(res, model.ErrCardNotFound)
}

type bankRepository struct{ base }

func NewBankRepository(db *sqlx.DB) port.BankRepository {
	return &bankRepository{base{db: db}}
}

func (r *bankRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Bank, error) {
	var b model.Bank
	if err := r.get(ctx, &b, model.ErrBankNotFound, `SELECT id, name, code, country, created_at FROM banks WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &b, nil
}

type merchantRepository struct{ base }

func NewMerchantRepository(db *sqlx.DB) port.MerchantRepository {
	return &merchantRepository{base{db: db}}
}

func (r *merchantRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.Merchant, error) {
	var m model.Merchant
	if err := r.get(ctx, &m, model.ErrMerchantNotFound,
		`SELECT id, name, mcc, acquirer_bank_id, country, created_at FROM merchants WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &m, nil
}
