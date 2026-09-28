package postgres

import (
	"context"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/jackc/pgx/v5"
)

func (d *DB) ids(ctx context.Context, sql string, args ...any) ([]string, error) {
	rows, err := d.q(ctx).Query(ctx, sql, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(row pgx.Row) (string, error) {
		var id string
		err := row.Scan(&id)
		return id, err
	})
}

func updated(tag interface{ RowsAffected() int64 }, err error) error {
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return mapErr(err)
}

const sessionColumns = `session_id, machine_id, started_at, expires_at, status, created_at, updated_at`

type sessionRepository struct{ db *DB }

func NewSessionRepository(db *DB) port.SessionRepository { return &sessionRepository{db: db} }

func (r *sessionRepository) get(ctx context.Context, sql, id string) (*domain.Session, error) {
	var s domain.Session
	err := r.db.q(ctx).QueryRow(ctx, sql, id).
		Scan(&s.ID, &s.MachineID, &s.StartedAt, &s.ExpiresAt, &s.Status, &s.CreatedAt, &s.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &s, nil
}

func (r *sessionRepository) Create(ctx context.Context, s *domain.Session) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO sessions (`+sessionColumns+`) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		s.ID, s.MachineID, s.StartedAt, s.ExpiresAt, s.Status, s.CreatedAt, s.UpdatedAt)
	return mapErr(err)
}

func (r *sessionRepository) Get(ctx context.Context, id string) (*domain.Session, error) {
	return r.get(ctx, `SELECT `+sessionColumns+` FROM sessions WHERE session_id = $1`, id)
}

func (r *sessionRepository) GetForUpdate(ctx context.Context, id string) (*domain.Session, error) {
	return r.get(ctx, `SELECT `+sessionColumns+` FROM sessions WHERE session_id = $1 FOR UPDATE`, id)
}

func (r *sessionRepository) Update(ctx context.Context, s *domain.Session) error {
	return updated(r.db.q(ctx).Exec(ctx, `UPDATE sessions SET expires_at = $2, status = $3, updated_at = $4
		WHERE session_id = $1`, s.ID, s.ExpiresAt, s.Status, s.UpdatedAt))
}

func (r *sessionRepository) ListExpired(ctx context.Context, now time.Time, limit int) ([]string, error) {
	return r.db.ids(ctx, `SELECT session_id FROM sessions WHERE status = 'active' AND expires_at <= $1
		ORDER BY expires_at LIMIT $2`, now, limit)
}

const reservationColumns = `reservation_id, session_id, inventory_id, expires_at, status, created_at, updated_at`

type reservationRepository struct{ db *DB }

func NewReservationRepository(db *DB) port.ReservationRepository {
	return &reservationRepository{db: db}
}

func (r *reservationRepository) Create(ctx context.Context, res *domain.Reservation) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO reservations (`+reservationColumns+`) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		res.ID, res.SessionID, res.InventoryID, res.ExpiresAt, res.Status, res.CreatedAt, res.UpdatedAt)
	return mapErr(err)
}

func (r *reservationRepository) Get(ctx context.Context, id string) (*domain.Reservation, error) {
	return r.get(ctx, `SELECT `+reservationColumns+` FROM reservations WHERE reservation_id = $1`, id)
}

func (r *reservationRepository) GetForUpdate(ctx context.Context, id string) (*domain.Reservation, error) {
	return r.get(ctx, `SELECT `+reservationColumns+` FROM reservations WHERE reservation_id = $1 FOR UPDATE`, id)
}

func (r *reservationRepository) get(ctx context.Context, sql, id string) (*domain.Reservation, error) {
	var res domain.Reservation
	err := r.db.q(ctx).QueryRow(ctx, sql, id).
		Scan(&res.ID, &res.SessionID, &res.InventoryID, &res.ExpiresAt, &res.Status, &res.CreatedAt, &res.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &res, nil
}

func (r *reservationRepository) Update(ctx context.Context, res *domain.Reservation) error {
	return updated(r.db.q(ctx).Exec(ctx, `UPDATE reservations SET status = $2, updated_at = $3 WHERE reservation_id = $1`,
		res.ID, res.Status, res.UpdatedAt))
}

func (r *reservationRepository) ListExpiredPending(ctx context.Context, now time.Time, limit int) ([]string, error) {
	return r.db.ids(ctx, `SELECT reservation_id FROM reservations WHERE status = 'pending' AND expires_at <= $1
		ORDER BY expires_at LIMIT $2`, now, limit)
}

func (r *reservationRepository) ListPendingBySession(ctx context.Context, sessionID string) ([]string, error) {
	return r.db.ids(ctx, `SELECT reservation_id FROM reservations WHERE session_id = $1 AND status = 'pending'
		ORDER BY created_at`, sessionID)
}

const paymentColumns = `payment_id, session_id, reservation_id, amount_cents, discount_cents, coupon_code, currency,
	payment_method, payment_status, transaction_id, payment_metadata, completed_at, created_at, updated_at`

type paymentRepository struct{ db *DB }

func NewPaymentRepository(db *DB) port.PaymentRepository { return &paymentRepository{db: db} }

func (r *paymentRepository) Create(ctx context.Context, p *domain.Payment) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO payments (`+paymentColumns+`)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)`,
		p.ID, p.SessionID, p.ReservationID, p.AmountCents, p.DiscountCents, p.CouponCode, p.Currency, p.Method, p.Status,
		p.TransactionID, p.Metadata, p.CompletedAt, p.CreatedAt, p.UpdatedAt)
	return mapErr(err)
}

func (r *paymentRepository) Get(ctx context.Context, id string) (*domain.Payment, error) {
	return r.get(ctx, `SELECT `+paymentColumns+` FROM payments WHERE payment_id = $1`, id)
}

func (r *paymentRepository) GetForUpdate(ctx context.Context, id string) (*domain.Payment, error) {
	return r.get(ctx, `SELECT `+paymentColumns+` FROM payments WHERE payment_id = $1 FOR UPDATE`, id)
}

func (r *paymentRepository) ListStalePending(ctx context.Context, before time.Time, limit int) ([]string, error) {
	return r.db.ids(ctx, `SELECT payment_id FROM payments WHERE payment_status = 'pending' AND created_at <= $1
		ORDER BY created_at LIMIT $2`, before, limit)
}

func (r *paymentRepository) get(ctx context.Context, sql, id string) (*domain.Payment, error) {
	var p domain.Payment
	err := r.db.q(ctx).QueryRow(ctx, sql, id).
		Scan(&p.ID, &p.SessionID, &p.ReservationID, &p.AmountCents, &p.DiscountCents, &p.CouponCode, &p.Currency, &p.Method, &p.Status,
			&p.TransactionID, &p.Metadata, &p.CompletedAt, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return nil, mapErr(err)
	}
	return &p, nil
}

func (r *paymentRepository) Update(ctx context.Context, p *domain.Payment) error {
	return updated(r.db.q(ctx).Exec(ctx, `UPDATE payments
		SET payment_status = $2, transaction_id = $3, payment_metadata = $4, completed_at = $5, updated_at = $6
		WHERE payment_id = $1`,
		p.ID, p.Status, p.TransactionID, p.Metadata, p.CompletedAt, p.UpdatedAt))
}

const orderColumns = `order_id, order_no, session_id, reservation_id, payment_id, status, fulfilled_at, created_at, updated_at`

type orderRepository struct{ db *DB }

func NewOrderRepository(db *DB) port.OrderRepository { return &orderRepository{db: db} }

func scanOrder(row pgx.Row) (domain.Order, error) {
	var o domain.Order
	err := row.Scan(&o.ID, &o.OrderNo, &o.SessionID, &o.ReservationID, &o.PaymentID, &o.Status, &o.FulfilledAt, &o.CreatedAt, &o.UpdatedAt)
	return o, err
}

func (r *orderRepository) get(ctx context.Context, sql, id string) (*domain.Order, error) {
	o, err := scanOrder(r.db.q(ctx).QueryRow(ctx, sql, id))
	if err != nil {
		return nil, mapErr(err)
	}
	return &o, nil
}

func (r *orderRepository) Create(ctx context.Context, o *domain.Order) error {
	err := r.db.q(ctx).QueryRow(ctx, `INSERT INTO orders
		(order_id, session_id, reservation_id, payment_id, status, fulfilled_at, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8) RETURNING order_no`,
		o.ID, o.SessionID, o.ReservationID, o.PaymentID, o.Status, o.FulfilledAt, o.CreatedAt, o.UpdatedAt).Scan(&o.OrderNo)
	return mapErr(err)
}

func (r *orderRepository) ListBySession(ctx context.Context, sessionID string) ([]domain.Order, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+orderColumns+` FROM orders WHERE session_id = $1 ORDER BY created_at`, sessionID)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, scanOrder)
}

func (r *orderRepository) Get(ctx context.Context, id string) (*domain.Order, error) {
	return r.get(ctx, `SELECT `+orderColumns+` FROM orders WHERE order_id = $1`, id)
}

func (r *orderRepository) GetByPayment(ctx context.Context, paymentID string) (*domain.Order, error) {
	return r.get(ctx, `SELECT `+orderColumns+` FROM orders WHERE payment_id = $1`, paymentID)
}

func (r *orderRepository) GetForUpdate(ctx context.Context, id string) (*domain.Order, error) {
	return r.get(ctx, `SELECT `+orderColumns+` FROM orders WHERE order_id = $1 FOR UPDATE`, id)
}

func (r *orderRepository) Update(ctx context.Context, o *domain.Order) error {
	return updated(r.db.q(ctx).Exec(ctx, `UPDATE orders SET status = $2, fulfilled_at = $3, updated_at = $4
		WHERE order_id = $1`, o.ID, o.Status, o.FulfilledAt, o.UpdatedAt))
}
