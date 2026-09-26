package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type ReservationRepository struct{ db *pgxpool.Pool }

var _ outbound.ReservationRepository = (*ReservationRepository)(nil)

func NewReservationRepository(db *pgxpool.Pool) *ReservationRepository {
	return &ReservationRepository{db: db}
}

const reservationCols = `id, hotel_id, room_type_id, guest_id, start_date, end_date, rooms, adults, children, status, currency,
	room_total::text, addons_total::text, coalesce(promo_id, 0), coalesce(promo_code, ''), discount_amount::text, total_amount::text,
	fee_amount::text, refund_amount::text, stay_status, checked_in_at, checked_out_at,
	coalesce(special_requests, ''), request_id, coalesce(payment_method, ''), coalesce(payment_ref, ''), paid_at, expires_at,
	coalesce(status_note, ''), coalesce(version, 0), coalesce(created_at, now())`

func scanReservation(row pgx.Row) (domain.Reservation, error) {
	var x domain.Reservation
	err := row.Scan(&x.ID, &x.HotelID, &x.RoomTypeID, &x.GuestID, &x.Start, &x.End, &x.Rooms, &x.Adults, &x.Children, &x.Status,
		&x.Currency, &x.RoomTotal, &x.AddonsTotal, &x.PromoID, &x.PromoCode, &x.DiscountAmount, &x.TotalAmount,
		&x.FeeAmount, &x.RefundAmount, &x.Stay, &x.CheckedInAt, &x.CheckedOutAt,
		&x.SpecialRequests, &x.RequestID, &x.PaymentMethod, &x.PaymentRef, &x.PaidAt, &x.ExpiresAt,
		&x.StatusNote, &x.Version, &x.CreatedAt)
	return x, mapErr(err)
}

// writeEvent appends to a reservation's history.
func writeEvent(ctx context.Context, tx pgx.Tx, reservationID, actor int64, event string, from, to domain.ReservationStatus, note string) error {
	_, err := tx.Exec(ctx, `insert into reservation_event (reservation_id, actor, event, from_status, to_status, note)
		values ($1, $2, $3, nullif($4, 0), nullif($5, 0), nullif($6, ''))`, reservationID, actor, event, from, to, note)
	return err
}

func (r *ReservationRepository) Get(ctx context.Context, id int64) (domain.Reservation, error) {
	x, err := scanReservation(r.db.QueryRow(ctx, `select `+reservationCols+` from reservation where id = $1`, id))
	if err != nil {
		return x, err
	}
	rows, err := r.db.Query(ctx, `select service_id, name, quantity, unit_price::text, amount::text
		from reservation_service where reservation_id = $1 order by service_id`, id)
	if err != nil {
		return x, mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var a domain.ReservationAddon
		if err := rows.Scan(&a.ServiceID, &a.Name, &a.Quantity, &a.UnitPrice, &a.Amount); err != nil {
			return x, err
		}
		x.Addons = append(x.Addons, a)
	}
	return x, mapErr(rows.Err())
}

func (r *ReservationRepository) RoomsLeft(ctx context.Context, hotelID, roomTypeID int64, start, end time.Time) (int, error) {
	var priced, left int
	err := r.db.QueryRow(ctx, `select count(*), coalesce(min(total_inventory - total_reserved), 0) from room_type_inventory
		where hotel_id = $1 and room_type_id = $2 and date >= $3 and date < $4 and rate is not null`,
		hotelID, roomTypeID, start, end).Scan(&priced, &left)
	if err != nil {
		return 0, mapErr(err)
	}
	if priced != int(end.Sub(start).Hours()/24) {
		return 0, nil // a night is not for sale
	}
	return left, nil
}

func (r *ReservationRepository) RequestIDExists(ctx context.Context, requestID string) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `select exists (select 1 from reservation where request_id = $1)`, requestID).Scan(&ok)
	return ok, mapErr(err)
}

func (r *ReservationRepository) byRequestID(ctx context.Context, guestID int64, requestID string) (domain.Reservation, error) {
	x, err := scanReservation(r.db.QueryRow(ctx, `select `+reservationCols+` from reservation where request_id = $1`, requestID))
	if err != nil {
		return x, err
	}
	if x.GuestID != guestID {
		return domain.Reservation{}, domain.ErrConflict // that request id belongs to someone else
	}
	return r.Get(ctx, x.ID)
}

func (r *ReservationRepository) FindByRequestID(ctx context.Context, guestID int64, requestID string) (domain.Reservation, bool, error) {
	x, err := r.byRequestID(ctx, guestID, requestID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Reservation{}, false, nil
	}
	return x, err == nil, err
}

// errHoldCap is internal: the guest already holds their limit of unpaid reservations.
var errHoldCap = errors.New("hold cap")

func (r *ReservationRepository) Reserve(ctx context.Context, p outbound.ReserveParams) (domain.Reservation, error) {
	if x, err := r.byRequestID(ctx, p.GuestID, p.RequestID); err == nil {
		return x, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.Reservation{}, err
	}

	nights := int(p.End.Sub(p.Start).Hours() / 24)
	var id int64
	var held int
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		// A rush on the last room must not park a thousand connections behind one row: waiting for a
		// lock is capped, and a request that is refused for it is told to come back (ErrBusy).
		if _, err := tx.Exec(ctx, `set local lock_timeout = '800ms'`); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `set local statement_timeout = '3s'`); err != nil {
			return err
		}
		if p.MaxPending > 0 {
			// One guest's requests are serialised, so parallel requests cannot slip past the cap.
			if _, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtextextended('reservation-guest-' || $1::bigint::text, 0))`, p.GuestID); err != nil {
				return err
			}
			if err := tx.QueryRow(ctx, `select count(*) from reservation where guest_id = $1 and status = $2`,
				p.GuestID, domain.ReservationPending).Scan(&held); err != nil {
				return err
			}
			if held >= p.MaxPending {
				return errHoldCap
			}
		}
		// One statement takes the rooms and prices the nights, so two guests racing for the last room
		// cannot both succeed. The rows are locked first, in date order, so overlapping ranges never wait
		// on each other in a cycle; and the availability test is re-checked on each row after its lock
		// is won, so whoever arrives second sees the room already gone and matches fewer nights than asked.
		var taken int
		var nightlySum string
		if err := tx.QueryRow(ctx, `
			with locked as (
				select date from room_type_inventory
				where hotel_id = $1 and room_type_id = $2 and date >= $3 and date < $4
					and rate is not null and total_inventory - total_reserved >= $5
				order by date for update
			), taken as (
				update room_type_inventory i set total_reserved = i.total_reserved + $5
				from locked l where i.room_type_id = $2 and i.date = l.date
				returning i.rate
			) select count(*), coalesce(sum(rate), 0)::text from taken`,
			p.HotelID, p.RoomTypeID, p.Start, p.End, p.Rooms).Scan(&taken, &nightlySum); err != nil {
			return err
		}
		if taken != nights {
			return domain.ErrDatesUnavailable // rolls the update back
		}
		roomTotal, err := domain.MulAmount(nightlySum, p.Rooms)
		if err != nil {
			return err
		}

		// One use of the promo code, taken atomically: the conditions are in the UPDATE, so of a hundred guests
		// racing for the last use of a code exactly one wins it.
		discount := "0.0000"
		var promoID any
		var promoCode any
		if p.Promo != nil {
			var pct int
			var amountOff string
			err := tx.QueryRow(ctx, `
				update promotion set used_count = used_count + 1
				where id = $1 and hotel_id = $2 and active and (max_uses is null or used_count < max_uses)
					and (valid_from is null or valid_from <= $3::date) and (valid_to is null or valid_to >= $3::date) and min_nights <= $4
				returning coalesce(percent_off, 0), coalesce(amount_off::text, '')`,
				p.Promo.ID, p.HotelID, p.Start, nights).Scan(&pct, &amountOff)
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: the promo code is no longer available", domain.ErrInvalid)
			}
			if err != nil {
				return err
			}
			pr := domain.Promotion{PercentOff: pct, AmountOff: amountOff}
			if discount, err = pr.DiscountFor(roomTotal); err != nil {
				return err
			}
			promoID, promoCode = p.Promo.ID, p.Promo.Code
		}
		afterDiscount, err := domain.SubAmounts(roomTotal, discount)
		if err != nil {
			return err
		}
		total, err := domain.SumAmounts(afterDiscount, p.AddonsTotal)
		if err != nil {
			return err
		}

		if err := tx.QueryRow(ctx, `
			insert into reservation (hotel_id, room_type_id, guest_id, start_date, end_date, rooms, adults, children, status,
				currency, room_total, addons_total, promo_id, promo_code, discount_amount, total_amount, special_requests,
				request_id, expires_at, version, created_at, updated_at)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::numeric, $12::numeric, $13, $14, $15::numeric, $16::numeric,
				nullif($17, ''), $18, $19, 1, now(), now())
			returning id`,
			p.HotelID, p.RoomTypeID, p.GuestID, p.Start, p.End, p.Rooms, p.Adults, p.Children, domain.ReservationPending,
			p.Currency, roomTotal, p.AddonsTotal, promoID, promoCode, discount, total, p.SpecialRequests, p.RequestID, p.ExpiresAt).Scan(&id); err != nil {
			return err
		}
		for _, a := range p.Addons {
			if _, err := tx.Exec(ctx, `insert into reservation_service (reservation_id, service_id, name, quantity, unit_price, amount)
				values ($1, $2, $3, $4, $5::numeric, $6::numeric)`, id, a.ServiceID, a.Name, a.Quantity, a.UnitPrice, a.Amount); err != nil {
				return err
			}
		}
		note := ""
		if p.Actor == 0 {
			note = "created from the waiting list"
		}
		return writeEvent(ctx, tx, id, p.Actor, "created", 0, domain.ReservationPending, note)
	})
	if errors.Is(err, errHoldCap) {
		return domain.Reservation{}, fmt.Errorf("%w: you already hold %d unpaid reservations; pay or cancel one first", domain.ErrConflict, held)
	}
	if errors.Is(err, domain.ErrConflict) {
		return r.byRequestID(ctx, p.GuestID, p.RequestID) // lost a race with a retry of the same request
	}
	if err != nil {
		return domain.Reservation{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *ReservationRepository) List(ctx context.Context, f outbound.ReservationFilter) ([]domain.Reservation, error) {
	rows, err := r.db.Query(ctx, `select `+reservationCols+` from reservation
		where ($1 = 0 or guest_id = $1)
			and ($2 = 0 or hotel_id in (select id from hotel where owner_id = $2))
			and ($3 = 0 or hotel_id = $3)
			and ($4 = 0 or status = $4)
			and ($7 = 0 or id < $7)
		order by id desc limit $5 offset $6`, f.GuestID, f.OwnerID, f.HotelID, f.Status, f.Limit, f.Offset, f.AfterID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	return collectReservations(rows)
}

func collectReservations(rows pgx.Rows) ([]domain.Reservation, error) {
	out := []domain.Reservation{}
	for rows.Next() {
		x, err := scanReservation(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, mapErr(rows.Err())
}

func (r *ReservationRepository) SetPaymentAttempt(ctx context.Context, id int64, m domain.PaymentMethod, ref string) (domain.Reservation, error) {
	x, err := scanReservation(r.db.QueryRow(ctx, `
		update reservation set payment_method = $2, payment_ref = $3, version = coalesce(version, 0) + 1, updated_at = now()
		where id = $1 and status = $4 returning `+reservationCols, id, string(m), ref, domain.ReservationPending))
	if errors.Is(err, domain.ErrNotFound) {
		return x, domain.ErrConflict // no longer pending
	}
	return x, err
}

func (r *ReservationRepository) MarkPaid(ctx context.Context, id int64, m domain.PaymentMethod, ref string) (domain.Reservation, error) {
	var x domain.Reservation
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var err error
		x, err = scanReservation(tx.QueryRow(ctx, `
			update reservation set status = $4, payment_method = $2, payment_ref = $3, paid_at = now(),
				version = coalesce(version, 0) + 1, updated_at = now()
			where id = $1 and status = $5 returning `+reservationCols, id, string(m), ref, domain.ReservationPaid, domain.ReservationPending))
		if err != nil {
			return err
		}
		return writeEvent(ctx, tx, id, x.GuestID, "paid", domain.ReservationPending, domain.ReservationPaid, string(m))
	})
	if err == nil || !errors.Is(err, domain.ErrNotFound) {
		return x, err
	}
	// Not pending: fine only if this exact payment already settled it.
	cur, err := r.Get(ctx, id)
	if err != nil {
		return domain.Reservation{}, err
	}
	if cur.Status == domain.ReservationPaid && cur.PaymentRef == ref && cur.PaymentMethod == m {
		return cur, nil
	}
	return domain.Reservation{}, domain.ErrConflict
}

func (r *ReservationRepository) Transition(ctx context.Context, p outbound.TransitionParams) (domain.Reservation, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var status domain.ReservationStatus
		var typeID, rooms int
		var promoID *int64
		var start, end time.Time
		err := tx.QueryRow(ctx, `select status, room_type_id, start_date, end_date, rooms, promo_id from reservation where id = $1 for update`, p.ID).
			Scan(&status, &typeID, &start, &end, &rooms, &promoID)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		allowed := false
		for _, f := range p.From {
			allowed = allowed || f == status
		}
		if !allowed {
			return domain.ErrConflict
		}
		if _, err := tx.Exec(ctx, `
			update reservation set status = $2, status_note = nullif($3, ''), version = coalesce(version, 0) + 1, updated_at = now(),
				fee_amount = coalesce(nullif($4, '')::numeric, 0), refund_amount = coalesce(nullif($5, '')::numeric, 0)
			where id = $1`, p.ID, p.To, p.Note, p.Fee, p.Refund); err != nil {
			return err
		}
		// The rooms go back on sale.
		if _, err := tx.Exec(ctx, `update room_type_inventory set total_reserved = greatest(total_reserved - $4, 0)
			where room_type_id = $1 and date >= $2 and date < $3`, typeID, start, end, rooms); err != nil {
			return err
		}
		// A promo code use is given back when nothing was bought with it.
		if promoID != nil && status == domain.ReservationPending {
			if _, err := tx.Exec(ctx, `update promotion set used_count = greatest(used_count - 1, 0) where id = $1`, *promoID); err != nil {
				return err
			}
		}
		return writeEvent(ctx, tx, p.ID, p.Actor, p.To.Name(), status, p.To, p.Note)
	})
	if err != nil {
		return domain.Reservation{}, err
	}
	return r.Get(ctx, p.ID)
}

func (r *ReservationRepository) ListExpired(ctx context.Context, before time.Time, limit int) ([]domain.Reservation, error) {
	rows, err := r.db.Query(ctx, `select `+reservationCols+` from reservation where status = $1 and expires_at < $2
		order by expires_at limit $3`, domain.ReservationPending, before, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	return collectReservations(rows)
}

func (r *ReservationRepository) SetStay(ctx context.Context, id int64, from, to domain.StayStatus, actor int64) (domain.Reservation, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			update reservation set stay_status = $3::smallint, version = coalesce(version, 0) + 1, updated_at = now(),
				checked_in_at = case when $3::smallint = 1 then now() else checked_in_at end,
				checked_out_at = case when $3::smallint = 2 then now() else checked_out_at end
			where id = $1 and status = $4 and stay_status = $2`, id, from, to, domain.ReservationPaid)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			var exists bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from reservation where id = $1)`, id).Scan(&exists); err != nil {
				return err
			}
			if !exists {
				return domain.ErrNotFound
			}
			return domain.ErrConflict
		}
		return writeEvent(ctx, tx, id, actor, to.Name(), 0, 0, "")
	})
	if err != nil {
		return domain.Reservation{}, err
	}
	return r.Get(ctx, id)
}

func (r *ReservationRepository) Events(ctx context.Context, id int64) ([]domain.ReservationEvent, error) {
	rows, err := r.db.Query(ctx, `select id, reservation_id, at, actor, event, coalesce(from_status, 0), coalesce(to_status, 0), coalesce(note, '')
		from reservation_event where reservation_id = $1 order by id`, id)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.ReservationEvent{}
	for rows.Next() {
		var e domain.ReservationEvent
		if err := rows.Scan(&e.ID, &e.ReservationID, &e.At, &e.Actor, &e.Event, &e.FromStatus, &e.ToStatus, &e.Note); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, mapErr(rows.Err())
}

func (r *ReservationRepository) DueForReminder(ctx context.Context, from, to time.Time, limit int) ([]domain.Reservation, error) {
	rows, err := r.db.Query(ctx, `select `+reservationCols+` from reservation
		where status = $1 and stay_status = 0 and reminded_at is null and start_date >= $2 and start_date < $3
		order by start_date, id limit $4`, domain.ReservationPaid, from, to, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	return collectReservations(rows)
}

func (r *ReservationRepository) MarkReminded(ctx context.Context, id int64) (bool, error) {
	tag, err := r.db.Exec(ctx, `update reservation set reminded_at = now() where id = $1 and reminded_at is null`, id)
	return tag.RowsAffected() == 1, mapErr(err)
}

type ReviewRepository struct{ db *pgxpool.Pool }

var _ outbound.ReviewRepository = (*ReviewRepository)(nil)

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository { return &ReviewRepository{db: db} }

func (r *ReviewRepository) Create(ctx context.Context, rv domain.Review) (domain.Review, error) {
	// Eligibility is part of the insert, so a review can never be stored for a stay the reviewer
	// did not have, even under concurrent requests. A no-show did not stay.
	err := r.db.QueryRow(ctx, `
		insert into review (hotel_id, user_id, reservation_id, rating, comment)
		select x.hotel_id, x.guest_id, x.id, $4, nullif($5, '') from reservation x
		where x.id = $1 and x.guest_id = $2 and x.hotel_id = $3 and x.status = $6 and x.end_date <= current_date and x.stay_status <> 3
		returning id, created_at`, rv.ReservationID, rv.UserID, rv.HotelID, rv.Rating, rv.Comment, domain.ReservationPaid).Scan(&rv.ID, &rv.CreatedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Review{}, domain.ErrForbidden
	}
	return rv, mapErr(err)
}

func (r *ReviewRepository) ListByHotel(ctx context.Context, hotelID, afterID int64, limit int) ([]domain.Review, error) {
	rows, err := r.db.Query(ctx, `select id, hotel_id, user_id, reservation_id, rating, coalesce(comment, ''), created_at
		from review where hotel_id = $1 and ($2 = 0 or id < $2) order by id desc limit $3`, hotelID, afterID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Review{}
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.HotelID, &rv.UserID, &rv.ReservationID, &rv.Rating, &rv.Comment, &rv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	return out, mapErr(rows.Err())
}
