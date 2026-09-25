package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type BookingRepository struct{ db *pgxpool.Pool }

var _ outbound.BookingRepository = (*BookingRepository)(nil)

func NewBookingRepository(db *pgxpool.Pool) *BookingRepository { return &BookingRepository{db: db} }

const bookingCols = `id, user_id, homestay_id, checkin_date, checkout_date, guests, status, currency,
	coalesce(subtotal::text, '0'), coalesce(discount::text, '0'), total_amount::text,
	coalesce(note, ''), request_id, coalesce(created_at, now()),
	coalesce(payment_method, ''), coalesce(payment_ref, ''), paid_at, expires_at, coalesce(version, 0)`

func scanBooking(row pgx.Row) (domain.Booking, error) {
	var b domain.Booking
	err := row.Scan(&b.ID, &b.UserID, &b.HomestayID, &b.CheckIn, &b.CheckOut, &b.Guests, &b.Status, &b.Currency,
		&b.Subtotal, &b.Discount, &b.TotalAmount, &b.Note, &b.RequestID, &b.CreatedAt,
		&b.PaymentMethod, &b.PaymentRef, &b.PaidAt, &b.ExpiresAt, &b.Version)
	return b, mapErr(err)
}

func (r *BookingRepository) Get(ctx context.Context, id int64) (domain.Booking, error) {
	return scanBooking(r.db.QueryRow(ctx, `select `+bookingCols+` from booking where id = $1`, id))
}

func (r *BookingRepository) byRequestID(ctx context.Context, userID int64, requestID string) (domain.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx, `select `+bookingCols+` from booking where request_id = $1`, requestID))
	if err != nil {
		return b, err
	}
	if b.UserID != userID {
		return domain.Booking{}, domain.ErrConflict
	}
	return b, nil
}

func (r *BookingRepository) Reserve(ctx context.Context, p outbound.ReserveParams) (domain.Booking, error) {
	if b, err := r.byRequestID(ctx, p.UserID, p.RequestID); err == nil {
		return b, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.Booking{}, err
	}

	nights := int(p.CheckOut.Sub(p.CheckIn).Hours() / 24)
	var booking domain.Booking
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var claimed int
		var subtotal string
		if err := tx.QueryRow(ctx, `
			with claimed as (
				update homestay_availability set status = $4
				where homestay_id = $1 and date >= $2 and date < $3 and status = $5 and price is not null
				returning price
			)
			select count(*), coalesce(sum(price), 0)::text from claimed`,
			p.HomestayID, p.CheckIn, p.CheckOut, domain.SlotBooked, domain.SlotAvailable,
		).Scan(&claimed, &subtotal); err != nil {
			return err
		}
		if claimed != nights {
			return domain.ErrDatesUnavailable
		}
		var err error
		booking, err = scanBooking(tx.QueryRow(ctx, `
			insert into booking (user_id, homestay_id, checkin_date, checkout_date, guests, status, currency,
				subtotal, discount, total_amount, price_detail, note, request_id, expires_at,
				version, created_at, created_by, updated_at, updated_by)
			values ($1, $2, $3, $4, $5, $6, $7, $8::numeric, 0, $8::numeric(18, 4),
				jsonb_build_object('nights', $9::int), nullif($10, ''), $11, $12, 1, now(), $1, now(), $1)
			returning `+bookingCols,
			p.UserID, p.HomestayID, p.CheckIn, p.CheckOut, p.Guests, domain.BookingPending, p.Currency,
			subtotal, nights, p.Note, p.RequestID, p.ExpiresAt))
		return err
	})
	if errors.Is(err, domain.ErrConflict) {
		return r.byRequestID(ctx, p.UserID, p.RequestID)
	}
	return booking, err
}

func (r *BookingRepository) Cancel(ctx context.Context, id, actor int64) (domain.Booking, error) {
	var out domain.Booking
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		b, err := scanBooking(tx.QueryRow(ctx, `select `+bookingCols+` from booking where id = $1 for update`, id))
		if err != nil {
			return err
		}
		if b.Status == domain.BookingCancelled {
			out = b
			return nil
		}
		if out, err = scanBooking(tx.QueryRow(ctx, `
			update booking set status = $2, version = coalesce(version, 0) + 1, updated_at = now(), updated_by = $3
			where id = $1 returning `+bookingCols, id, domain.BookingCancelled, actor)); err != nil {
			return err
		}
		_, err = tx.Exec(ctx, `
			update homestay_availability set status = $4
			where homestay_id = $1 and date >= $2 and date < $3 and status = $5`,
			b.HomestayID, b.CheckIn, b.CheckOut, domain.SlotAvailable, domain.SlotBooked)
		return err
	})
	return out, err
}

func (r *BookingRepository) List(ctx context.Context, userID, hostID int64, limit, offset int) ([]domain.Booking, error) {
	rows, err := r.db.Query(ctx, `
		select `+bookingCols+` from booking
		where ($1 = 0 or user_id = $1)
			and ($4 = 0 or homestay_id in (select id from homestay where created_by = $4))
		order by id desc limit $2 offset $3`, userID, limit, offset, hostID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Booking{}
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, mapErr(rows.Err())
}

func (r *BookingRepository) SetPaymentAttempt(ctx context.Context, id int64, m domain.PaymentMethod, ref string) (domain.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx, `
		update booking set payment_method = $2, payment_ref = $3, version = coalesce(version, 0) + 1, updated_at = now()
		where id = $1 and status = $4
		returning `+bookingCols, id, string(m), ref, domain.BookingPending))
	if errors.Is(err, domain.ErrNotFound) {
		return b, domain.ErrConflict // no longer pending
	}
	return b, err
}

func (r *BookingRepository) MarkPaid(ctx context.Context, id int64, m domain.PaymentMethod, ref string) (domain.Booking, error) {
	b, err := scanBooking(r.db.QueryRow(ctx, `
		update booking set status = $4, payment_method = $2, payment_ref = $3, paid_at = now(),
			version = coalesce(version, 0) + 1, updated_at = now()
		where id = $1 and status = $5
		returning `+bookingCols, id, string(m), ref, domain.BookingConfirmed, domain.BookingPending))
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		return b, err
	}
	cur, err := r.Get(ctx, id)
	if err != nil {
		return domain.Booking{}, err
	}
	if cur.Status == domain.BookingConfirmed && cur.PaymentRef == ref && cur.PaymentMethod == m {
		return cur, nil
	}
	return domain.Booking{}, domain.ErrConflict
}

func (r *BookingRepository) Expire(ctx context.Context, id int64) (bool, error) {
	released := false
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var h int64
		var in, out time.Time
		err := tx.QueryRow(ctx, `
			update booking set status = $2, version = coalesce(version, 0) + 1, updated_at = now()
			where id = $1 and status = $3
			returning homestay_id, checkin_date, checkout_date`, id, domain.BookingCancelled, domain.BookingPending).Scan(&h, &in, &out)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		released = true
		_, err = tx.Exec(ctx, `
			update homestay_availability set status = $4
			where homestay_id = $1 and date >= $2 and date < $3 and status = $5`,
			h, in, out, domain.SlotAvailable, domain.SlotBooked)
		return err
	})
	return released, err
}

func (r *BookingRepository) ListExpired(ctx context.Context, before time.Time, limit int) ([]domain.Booking, error) {
	rows, err := r.db.Query(ctx, `
		select `+bookingCols+` from booking
		where status = $1 and expires_at < $2
		order by expires_at limit $3`, domain.BookingPending, before, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Booking{}
	for rows.Next() {
		b, err := scanBooking(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, b)
	}
	return out, mapErr(rows.Err())
}
