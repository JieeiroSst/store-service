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

type cardRepository struct{ db *sqlx.DB }

func NewCardRepository(db *sqlx.DB) ports.CardRepository {
	return &cardRepository{db: db}
}

func (r *cardRepository) Create(ctx context.Context, card *domain.Card) error {
	q := `INSERT INTO cards
		(id, wallet_id, issuer_bank_id, card_number, card_number_hash, holder_name,
		 network, expiry_month, expiry_year, status, created_at, updated_at)
		VALUES (:id, :wallet_id, :issuer_bank_id, :card_number, :card_number_hash, :holder_name,
		 :network, :expiry_month, :expiry_year, :status, :created_at, :updated_at)`
	_, err := r.db.NamedExecContext(ctx, q, card)
	return err
}

func (r *cardRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Card, error) {
	var c domain.Card
	err := r.db.GetContext(ctx, &c, `SELECT * FROM cards WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, domain.ErrCardNotFound
	}
	return &c, err
}

func (r *cardRepository) GetByWalletID(ctx context.Context, walletID uuid.UUID) ([]*domain.Card, error) {
	var cards []*domain.Card
	err := r.db.SelectContext(ctx, &cards, `SELECT * FROM cards WHERE wallet_id = $1 ORDER BY created_at DESC`, walletID)
	return cards, err
}

func (r *cardRepository) Update(ctx context.Context, card *domain.Card) error {
	card.UpdatedAt = time.Now()
	q := `UPDATE cards SET status=:status, holder_name=:holder_name, updated_at=:updated_at WHERE id=:id`
	_, err := r.db.NamedExecContext(ctx, q, card)
	return err
}
