package postgres

import (
	"context"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"

	"github.com/JIeeiroSst/wallet-service/internal/domain/model"
	"github.com/JIeeiroSst/wallet-service/internal/domain/port"
)

const paymentRequestColumns = `id, requester_user_id, requester_wallet_id, payer_user_id, payer_wallet_id,
	transaction_id, amount, currency, COALESCE(note, '') AS note, status, expires_at, created_at, updated_at`

type paymentRequestRepository struct{ base }

func NewPaymentRequestRepository(db *sqlx.DB) port.PaymentRequestRepository {
	return &paymentRequestRepository{base{db: db}}
}

func (r *paymentRequestRepository) Create(ctx context.Context, p *model.PaymentRequest) error {
	_, err := r.named(ctx, `INSERT INTO payment_requests
		(id, requester_user_id, requester_wallet_id, payer_user_id, payer_wallet_id, transaction_id,
		 amount, currency, note, status, expires_at, created_at, updated_at)
		VALUES (:id, :requester_user_id, :requester_wallet_id, :payer_user_id, :payer_wallet_id,
		 :transaction_id, :amount, :currency, NULLIF(:note, ''), :status, :expires_at, :created_at, :updated_at)`, p)
	return err
}

func (r *paymentRequestRepository) GetByID(ctx context.Context, id uuid.UUID) (*model.PaymentRequest, error) {
	var p model.PaymentRequest
	if err := r.get(ctx, &p, model.ErrPaymentRequestNotFound,
		`SELECT `+paymentRequestColumns+` FROM payment_requests WHERE id = $1`, id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRequestRepository) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.PaymentRequest, error) {
	var p model.PaymentRequest
	if err := r.get(ctx, &p, model.ErrPaymentRequestNotFound,
		`SELECT `+paymentRequestColumns+` FROM payment_requests WHERE id = $1 FOR UPDATE`, id); err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *paymentRequestRepository) Update(ctx context.Context, p *model.PaymentRequest) error {
	res, err := r.named(ctx, `UPDATE payment_requests
		SET status = :status, payer_wallet_id = :payer_wallet_id, transaction_id = :transaction_id,
		    updated_at = :updated_at
		WHERE id = :id`, p)
	if err != nil {
		return err
	}
	return expectOne(res, model.ErrPaymentRequestNotFound)
}

func (r *paymentRequestRepository) List(ctx context.Context, f model.PaymentRequestFilter) ([]*model.PaymentRequest, int64, error) {
	column := "payer_user_id"
	if f.Role == model.RoleOutgoing {
		column = "requester_user_id"
	}
	where := []string{column + " = $1"}
	args := []any{f.UserID}
	switch f.Status {
	case "":
	case model.PaymentRequestPending:
		where = append(where, "status = 'PENDING' AND expires_at > NOW()")
	case model.PaymentRequestExpired:
		where = append(where, "(status = 'EXPIRED' OR (status = 'PENDING' AND expires_at <= NOW()))")
	default:
		args = append(args, f.Status)
		where = append(where, fmt.Sprintf("status = $%d", len(args)))
	}
	cond := strings.Join(where, " AND ")
	total, err := r.count(ctx, `SELECT COUNT(*) FROM payment_requests WHERE `+cond, args...)
	if err != nil {
		return nil, 0, err
	}
	items := make([]*model.PaymentRequest, 0)
	query := fmt.Sprintf(`SELECT %s FROM payment_requests WHERE %s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`,
		paymentRequestColumns, cond, len(args)+1, len(args)+2)
	args = append(args, f.PageSize, (f.Page-1)*f.PageSize)
	if err := sqlx.SelectContext(ctx, r.q(ctx), &items, query, args...); err != nil {
		return nil, 0, err
	}
	return items, total, nil
}
