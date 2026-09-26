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

type WaitlistRepository struct{ db *pgxpool.Pool }

var _ outbound.WaitlistRepository = (*WaitlistRepository)(nil)

func NewWaitlistRepository(db *pgxpool.Pool) *WaitlistRepository { return &WaitlistRepository{db: db} }

const waitCols = `w.id, w.hotel_id, w.room_type_id, w.guest_id, w.start_date, w.end_date, w.rooms, w.adults, w.children, w.status,
	coalesce(w.reservation_id, 0), w.created_at, w.offered_at, coalesce(r.status, 0)`

const waitFrom = ` from waitlist w left join reservation r on r.id = w.reservation_id`

func scanWait(row pgx.Row) (domain.WaitlistEntry, error) {
	var e domain.WaitlistEntry
	err := row.Scan(&e.ID, &e.HotelID, &e.RoomTypeID, &e.GuestID, &e.Start, &e.End, &e.Rooms, &e.Adults, &e.Children, &e.Status,
		&e.ReservationID, &e.CreatedAt, &e.OfferedAt, &e.OfferStatus)
	return e, mapErr(err)
}

var errWaitFull = errors.New("waiting list full")

func (r *WaitlistRepository) Join(ctx context.Context, e domain.WaitlistEntry, maxWaiting int) (domain.WaitlistEntry, error) {
	var id int64
	var waiting int
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		// One guest's joins are serialised so parallel requests cannot slip past the limit or the duplicate check.
		if _, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtextextended('waitlist-guest-' || $1::bigint::text, 0))`, e.GuestID); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `select count(*) from waitlist where guest_id = $1 and status = $2`, e.GuestID, domain.WaitlistWaiting).Scan(&waiting); err != nil {
			return err
		}
		if waiting >= maxWaiting {
			return errWaitFull
		}
		var dup bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from waitlist where guest_id = $1 and status = $2 and room_type_id = $3
			and start_date = $4 and end_date = $5 and rooms = $6)`,
			e.GuestID, domain.WaitlistWaiting, e.RoomTypeID, e.Start, e.End, e.Rooms).Scan(&dup); err != nil {
			return err
		}
		if dup {
			return domain.ErrConflict
		}
		return tx.QueryRow(ctx, `insert into waitlist (hotel_id, room_type_id, guest_id, start_date, end_date, rooms, adults, children, status)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning id`,
			e.HotelID, e.RoomTypeID, e.GuestID, e.Start, e.End, e.Rooms, e.Adults, e.Children, domain.WaitlistWaiting).Scan(&id)
	})
	switch {
	case errors.Is(err, errWaitFull):
		return domain.WaitlistEntry{}, fmt.Errorf("%w: you are already waiting for %d stays; leave one first", domain.ErrConflict, waiting)
	case errors.Is(err, domain.ErrConflict):
		return domain.WaitlistEntry{}, fmt.Errorf("%w: you are already on the waiting list for this stay", domain.ErrConflict)
	case err != nil:
		return domain.WaitlistEntry{}, mapErr(err)
	}
	return scanWait(r.db.QueryRow(ctx, `select `+waitCols+waitFrom+` where w.id = $1`, id))
}

func (r *WaitlistRepository) Leave(ctx context.Context, guestID, id int64) error {
	tag, err := r.db.Exec(ctx, `update waitlist set status = $3 where id = $1 and guest_id = $2 and status = $4`,
		id, guestID, domain.WaitlistCancelled, domain.WaitlistWaiting)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *WaitlistRepository) List(ctx context.Context, guestID, afterID int64, limit int) ([]domain.WaitlistEntry, error) {
	rows, err := r.db.Query(ctx, `select `+waitCols+waitFrom+` where w.guest_id = $1 and ($2 = 0 or w.id < $2) order by w.id desc limit $3`, guestID, afterID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	return collectWait(rows)
}

func collectWait(rows pgx.Rows) ([]domain.WaitlistEntry, error) {
	out := []domain.WaitlistEntry{}
	for rows.Next() {
		e, err := scanWait(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, mapErr(rows.Err())
}

func (r *WaitlistRepository) Waiting(ctx context.Context, roomTypeID int64, today time.Time, limit int) ([]domain.WaitlistEntry, error) {
	rows, err := r.db.Query(ctx, `select `+waitCols+waitFrom+` where w.room_type_id = $1 and w.status = $2 and w.start_date >= $3
		order by w.id limit $4`, roomTypeID, domain.WaitlistWaiting, today, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	return collectWait(rows)
}

func (r *WaitlistRepository) RoomTypesWaiting(ctx context.Context, today time.Time, limit int) ([]int64, error) {
	rows, err := r.db.Query(ctx, `select distinct room_type_id from waitlist where status = $1 and start_date >= $2 limit $3`,
		domain.WaitlistWaiting, today, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	var out []int64
	for rows.Next() {
		var id int64
		if err := rows.Scan(&id); err != nil {
			return nil, err
		}
		out = append(out, id)
	}
	return out, mapErr(rows.Err())
}

func (r *WaitlistRepository) MarkOffered(ctx context.Context, id, reservationID int64) (outbound.OfferResult, error) {
	tag, err := r.db.Exec(ctx, `update waitlist set status = $3, reservation_id = $2, offered_at = now() where id = $1 and status = $4`,
		id, reservationID, domain.WaitlistOffered, domain.WaitlistWaiting)
	if err != nil {
		return outbound.OfferGone, mapErr(err)
	}
	if tag.RowsAffected() == 1 {
		return outbound.OfferMade, nil
	}
	// Not waiting any more. If it is offered with this very reservation, another replica did the same work.
	var status domain.WaitlistStatus
	var res int64
	err = r.db.QueryRow(ctx, `select status, coalesce(reservation_id, 0) from waitlist where id = $1`, id).Scan(&status, &res)
	if err != nil {
		return outbound.OfferGone, mapErr(err)
	}
	if status == domain.WaitlistOffered && res == reservationID {
		return outbound.OfferAlready, nil
	}
	return outbound.OfferGone, nil
}

func (r *WaitlistRepository) ExpireStale(ctx context.Context, today time.Time) (int, error) {
	tag, err := r.db.Exec(ctx, `update waitlist set status = $3 where status = $2 and start_date < $1`, today, domain.WaitlistWaiting, domain.WaitlistCancelled)
	return int(tag.RowsAffected()), mapErr(err)
}
