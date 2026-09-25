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

type LeaseRepository struct{ db *pgxpool.Pool }

var _ outbound.LeaseRepository = (*LeaseRepository)(nil)

func NewLeaseRepository(db *pgxpool.Pool) *LeaseRepository { return &LeaseRepository{db: db} }

const leaseCols = `id, user_id, homestay_id, model, start_date, end_date, periods, billing_day, currency,
	rent::text, status, coalesce(note, ''), request_id, expires_at, coalesce(created_at, now())`

const invoiceCols = `id, lease_id, period_no, period_start, period_end, due_date, amount::text, status,
	coalesce(payment_method, ''), coalesce(payment_ref, ''), paid_at, version`

func scanLease(row pgx.Row) (domain.Lease, error) {
	var l domain.Lease
	err := row.Scan(&l.ID, &l.UserID, &l.HomestayID, &l.Model, &l.StartDate, &l.EndDate, &l.Periods, &l.BillingDay,
		&l.Currency, &l.Rent, &l.Status, &l.Note, &l.RequestID, &l.ExpiresAt, &l.CreatedAt)
	return l, mapErr(err)
}

func scanInvoice(row pgx.Row) (domain.Invoice, error) {
	var i domain.Invoice
	err := row.Scan(&i.ID, &i.LeaseID, &i.PeriodNo, &i.PeriodStart, &i.PeriodEnd, &i.DueDate, &i.Amount, &i.Status,
		&i.PaymentMethod, &i.PaymentRef, &i.PaidAt, &i.Version)
	return i, mapErr(err)
}

func (r *LeaseRepository) invoices(ctx context.Context, q interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, leaseID int64) ([]domain.Invoice, error) {
	rows, err := q.Query(ctx, `select `+invoiceCols+` from rent_invoice where lease_id = $1 order by period_no`, leaseID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Invoice{}
	for rows.Next() {
		i, err := scanInvoice(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, mapErr(rows.Err())
}

func (r *LeaseRepository) Get(ctx context.Context, id int64) (domain.Lease, error) {
	l, err := scanLease(r.db.QueryRow(ctx, `select `+leaseCols+` from lease where id = $1`, id))
	if err != nil {
		return l, err
	}
	l.Invoices, err = r.invoices(ctx, r.db, id)
	return l, err
}

func (r *LeaseRepository) GetInvoice(ctx context.Context, id int64) (domain.Invoice, error) {
	return scanInvoice(r.db.QueryRow(ctx, `select `+invoiceCols+` from rent_invoice where id = $1`, id))
}

func (r *LeaseRepository) byRequestID(ctx context.Context, userID int64, requestID string) (domain.Lease, error) {
	l, err := scanLease(r.db.QueryRow(ctx, `select `+leaseCols+` from lease where request_id = $1`, requestID))
	if err != nil {
		return l, err
	}
	if l.UserID != userID {
		return domain.Lease{}, domain.ErrConflict
	}
	return r.Get(ctx, l.ID)
}

func (r *LeaseRepository) Create(ctx context.Context, p outbound.CreateLeaseParams) (domain.Lease, error) {
	if l, err := r.byRequestID(ctx, p.UserID, p.RequestID); err == nil {
		return l, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.Lease{}, err
	}

	nights := int(p.End.Sub(p.Start).Hours() / 24)
	var id int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var claimed int
		if err := tx.QueryRow(ctx, `
			with claimed as (
				insert into homestay_availability (homestay_id, date, price, status)
				select $1, d::date, null, $4 from generate_series($2::date, $3::date - 1, interval '1 day') d
				on conflict (homestay_id, date) do update set status = $4 where homestay_availability.status = $5
				returning 1
			) select count(*) from claimed`,
			p.HomestayID, p.Start, p.End, domain.SlotBooked, domain.SlotAvailable).Scan(&claimed); err != nil {
			return err
		}
		if claimed != nights {
			return domain.ErrDatesUnavailable
		}
		if err := tx.QueryRow(ctx, `
			insert into lease (user_id, homestay_id, model, start_date, end_date, periods, billing_day, currency, rent,
				status, note, request_id, expires_at, version, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9::numeric, $10, nullif($11, ''), $12, $13, 1, now(), now())
			returning id`,
			p.UserID, p.HomestayID, string(p.Model), p.Start, p.End, p.Periods, p.BillingDay, p.Currency, p.Rent,
			domain.LeasePending, p.Note, p.RequestID, p.ExpiresAt).Scan(&id); err != nil {
			return err
		}
		n := len(p.Schedule)
		nos, ps, pe, due := make([]int32, n), make([]time.Time, n), make([]time.Time, n), make([]time.Time, n)
		for i, s := range p.Schedule {
			nos[i], ps[i], pe[i], due[i] = int32(s.PeriodNo), s.PeriodStart, s.PeriodEnd, s.DueDate
		}
		_, err := tx.Exec(ctx, `
			insert into rent_invoice (lease_id, period_no, period_start, period_end, due_date, amount, status)
			select $1, t.n, t.ps, t.pe, t.dd, $6::numeric, $7
			from unnest($2::int[], $3::date[], $4::date[], $5::date[]) as t(n, ps, pe, dd)`,
			id, nos, ps, pe, due, p.Rent, domain.InvoiceUnpaid)
		return err
	})
	if errors.Is(err, domain.ErrConflict) {
		return r.byRequestID(ctx, p.UserID, p.RequestID)
	}
	if err != nil {
		return domain.Lease{}, err
	}
	return r.Get(ctx, id)
}

func (r *LeaseRepository) List(ctx context.Context, userID, hostID int64, limit, offset int) ([]domain.Lease, error) {
	rows, err := r.db.Query(ctx, `select `+leaseCols+` from lease
		where ($1 = 0 or user_id = $1) and ($4 = 0 or homestay_id in (select id from homestay where created_by = $4))
		order by id desc limit $2 offset $3`, userID, limit, offset, hostID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Lease{}
	for rows.Next() {
		l, err := scanLease(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, mapErr(rows.Err())
}

func (r *LeaseRepository) ListInvoices(ctx context.Context, f outbound.InvoiceFilter) ([]domain.Invoice, error) {
	var due any
	if !f.DueBefore.IsZero() {
		due = f.DueBefore
	}
	rows, err := r.db.Query(ctx, `
		select i.id, i.lease_id, i.period_no, i.period_start, i.period_end, i.due_date, i.amount::text, i.status,
			coalesce(i.payment_method, ''), coalesce(i.payment_ref, ''), i.paid_at, i.version, l.user_id, l.homestay_id
		from rent_invoice i join lease l on l.id = i.lease_id
		where l.status in ($1, $2)
			and ($3 = 0 or i.status = $3)
			and ($4::date is null or i.due_date < $4::date)
			and ($5 = 0 or l.user_id = $5)
			and ($8 = 0 or l.homestay_id in (select id from homestay where created_by = $8))
		order by i.due_date, i.id limit $6 offset $7`,
		domain.LeaseActive, domain.LeaseEnded, f.Status, due, f.UserID, f.Limit, f.Offset, f.HostID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Invoice{}
	for rows.Next() {
		var i domain.Invoice
		if err := rows.Scan(&i.ID, &i.LeaseID, &i.PeriodNo, &i.PeriodStart, &i.PeriodEnd, &i.DueDate, &i.Amount, &i.Status,
			&i.PaymentMethod, &i.PaymentRef, &i.PaidAt, &i.Version, &i.UserID, &i.HomestayID); err != nil {
			return nil, err
		}
		out = append(out, i)
	}
	return out, mapErr(rows.Err())
}

func (r *LeaseRepository) SetInvoiceAttempt(ctx context.Context, id int64, m domain.PaymentMethod, ref string) (domain.Invoice, error) {
	i, err := scanInvoice(r.db.QueryRow(ctx, `
		update rent_invoice set payment_method = $2, payment_ref = $3, version = version + 1
		where id = $1 and status = $4 returning `+invoiceCols, id, string(m), ref, domain.InvoiceUnpaid))
	if errors.Is(err, domain.ErrNotFound) {
		return i, domain.ErrConflict // paid or voided meanwhile
	}
	return i, err
}

func (r *LeaseRepository) MarkInvoicePaid(ctx context.Context, id int64, m domain.PaymentMethod, ref string) (domain.Invoice, error) {
	var inv domain.Invoice
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var err error
		inv, err = scanInvoice(tx.QueryRow(ctx, `
			update rent_invoice set status = $4, payment_method = $2, payment_ref = $3, paid_at = now(), version = version + 1
			where id = $1 and status = $5 returning `+invoiceCols, id, string(m), ref, domain.InvoicePaid, domain.InvoiceUnpaid))
		if err != nil {
			return err
		}
		// The first payment of a pending lease starts it.
		_, err = tx.Exec(ctx, `update lease set status = $2, version = coalesce(version, 0) + 1, updated_at = now()
			where id = $1 and status = $3`, inv.LeaseID, domain.LeaseActive, domain.LeasePending)
		return err
	})
	if errors.Is(err, domain.ErrNotFound) {
		// Not unpaid: fine only if this exact payment already settled it.
		cur, gerr := r.GetInvoice(ctx, id)
		if gerr != nil {
			return domain.Invoice{}, gerr
		}
		if cur.Status == domain.InvoicePaid && cur.PaymentRef == ref && cur.PaymentMethod == m {
			return cur, nil
		}
		return domain.Invoice{}, domain.ErrConflict
	}
	return inv, err
}

func release(ctx context.Context, tx pgx.Tx, homestayID int64, from, to time.Time) error {
	if _, err := tx.Exec(ctx, `delete from homestay_availability
		where homestay_id = $1 and date >= $2 and date < $3 and status = $4 and price is null`,
		homestayID, from, to, domain.SlotBooked); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `update homestay_availability set status = $4
		where homestay_id = $1 and date >= $2 and date < $3 and status = $5`,
		homestayID, from, to, domain.SlotAvailable, domain.SlotBooked)
	return err
}

func (r *LeaseRepository) Expire(ctx context.Context, id int64) (bool, error) {
	done := false
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var h int64
		var from, to time.Time
		err := tx.QueryRow(ctx, `update lease set status = $2, version = coalesce(version, 0) + 1, updated_at = now()
			where id = $1 and status = $3 returning homestay_id, start_date, end_date`,
			id, domain.LeaseCancelled, domain.LeasePending).Scan(&h, &from, &to)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil
		}
		if err != nil {
			return err
		}
		done = true
		if _, err := tx.Exec(ctx, `update rent_invoice set status = $2 where lease_id = $1 and status = $3`,
			id, domain.InvoiceVoid, domain.InvoiceUnpaid); err != nil {
			return err
		}
		return release(ctx, tx, h, from, to)
	})
	return done, err
}

func (r *LeaseRepository) ListExpired(ctx context.Context, before time.Time, limit int) ([]domain.Lease, error) {
	rows, err := r.db.Query(ctx, `select `+leaseCols+` from lease where status = $1 and expires_at < $2
		order by expires_at limit $3`, domain.LeasePending, before, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Lease{}
	for rows.Next() {
		l, err := scanLease(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, mapErr(rows.Err())
}

func (r *LeaseRepository) Terminate(ctx context.Context, id int64, from time.Time) (domain.Lease, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		l, err := scanLease(tx.QueryRow(ctx, `select `+leaseCols+` from lease where id = $1 for update`, id))
		if err != nil {
			return err
		}
		if l.Status != domain.LeaseActive {
			return domain.ErrConflict
		}
		if from.Before(l.StartDate) {
			from = l.StartDate
		}
		if from.After(l.EndDate) {
			from = l.EndDate
		}
		if _, err := tx.Exec(ctx, `update lease set status = $2, end_date = $3, version = coalesce(version, 0) + 1, updated_at = now()
			where id = $1`, id, domain.LeaseEnded, from); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update rent_invoice set status = $2 where lease_id = $1 and status = $3 and period_start >= $4`,
			id, domain.InvoiceVoid, domain.InvoiceUnpaid, from); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update rent_invoice set status = $2, version = version + 1
			where lease_id = $1 and status = $3 and period_start >= $4`,
			id, domain.InvoiceRefunded, domain.InvoicePaid, from); err != nil {
			return err
		}
		return release(ctx, tx, l.HomestayID, from, l.EndDate)
	})
	if err != nil {
		return domain.Lease{}, err
	}
	return r.Get(ctx, id)
}
