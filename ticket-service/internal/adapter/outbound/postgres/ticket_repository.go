package postgres

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type TicketRepository struct{ db *pgxpool.Pool }

var _ outbound.TicketRepository = (*TicketRepository)(nil)

func NewTicketRepository(db *pgxpool.Pool) *TicketRepository { return &TicketRepository{db: db} }

const ticketCols = `t.id, t.order_id, t.event_id, t.ticket_type_id, coalesce(t.seat_id, 0), tt.name,
	coalesce(s.section, ''), coalesce(s.row_label, ''), coalesce(s.number, 0), t.code, t.status, t.checked_in_at,
	coalesce(t.checked_in_by, 0), coalesce(t.holder_id, 0), coalesce(t.holder_name, ''), coalesce(t.holder_email, ''),
	t.transfer_count, t.created_at, ev.title, coalesce(ss.starts_at, ev.starts_at), coalesce(ss.ends_at, ev.ends_at), coalesce(t.session_id, 0), coalesce(ss.label, ''), ev.venue`

const ticketFrom = ` from ticket t join ticket_type tt on tt.id = t.ticket_type_id join event ev on ev.id = t.event_id
	left join event_session ss on ss.id = t.session_id left join seat s on s.id = t.seat_id `

func scanTicket(row pgx.Row, extra ...any) (domain.Ticket, error) {
	var t domain.Ticket
	var seat domain.Seat
	dest := append([]any{&t.ID, &t.OrderID, &t.EventID, &t.TicketTypeID, &t.SeatID, &t.TypeName, &seat.Section, &seat.Row,
		&seat.Number, &t.Code, &t.Status, &t.CheckedInAt, &t.CheckedInBy, &t.HolderID, &t.HolderName, &t.HolderEmail,
		&t.TransferCount, &t.CreatedAt, &t.EventTitle, &t.EventStartsAt, &t.EventEndsAt, &t.SessionID, &t.SessionLabel, &t.Venue}, extra...)
	if err := row.Scan(dest...); err != nil {
		return t, mapErr(err)
	}
	if t.SeatID != 0 {
		t.SeatLabel = seat.Label()
	}
	return t, nil
}

func queryTickets(ctx context.Context, q querier, tail string, args ...any) ([]domain.Ticket, error) {
	rows, err := q.Query(ctx, `select `+ticketCols+ticketFrom+tail, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Ticket{}
	for rows.Next() {
		t, err := scanTicket(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, mapErr(rows.Err())
}

func oneTicket(ctx context.Context, q querier, tail string, args ...any) (domain.Ticket, error) {
	ts, err := queryTickets(ctx, q, tail, args...)
	if err != nil {
		return domain.Ticket{}, err
	}
	if len(ts) == 0 {
		return domain.Ticket{}, domain.ErrNotFound
	}
	return ts[0], nil
}

func (r *TicketRepository) Get(ctx context.Context, id int64) (domain.Ticket, error) {
	return oneTicket(ctx, r.db, `where t.id = $1`, id)
}

func (r *TicketRepository) ByCode(ctx context.Context, eventID int64, code string) (domain.Ticket, error) {
	return oneTicket(ctx, r.db, `where t.code = $1 and ($2::bigint = 0 or t.event_id = $2)`, code, eventID)
}

func (r *TicketRepository) List(ctx context.Context, f outbound.TicketFilter) ([]domain.Ticket, error) {
	return queryTickets(ctx, r.db, `where ($1::bigint = 0 or t.holder_id = $1) and ($2::bigint = 0 or t.event_id = $2)
		and ($3::smallint = 0 or t.status = $3) and ($4::bigint = 0 or t.id < $4) order by t.id desc limit $5`,
		f.HolderID, f.EventID, int16(f.Status), f.AfterID, f.Limit)
}

func (r *TicketRepository) History(ctx context.Context, id int64) ([]domain.TicketEvent, error) {
	rows, err := r.db.Query(ctx, `select id, ticket_id, event, coalesce(from_status, 0), coalesce(to_status, 0), coalesce(actor, 0),
		coalesce(note, ''), created_at from ticket_event where ticket_id = $1 order by id`, id)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.TicketEvent{}
	for rows.Next() {
		var e domain.TicketEvent
		if err := rows.Scan(&e.ID, &e.TicketID, &e.Event, &e.FromStatus, &e.ToStatus, &e.Actor, &e.Note, &e.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, mapErr(rows.Err())
}

func (r *TicketRepository) Attendees(ctx context.Context, eventID, afterID int64, limit int) ([]domain.Attendee, error) {
	rows, err := r.db.Query(ctx, `select `+ticketCols+`, o.buyer_name, o.buyer_email`+ticketFrom+`
		join ticket_order o on o.id = t.order_id where t.event_id = $1 and ($2::bigint = 0 or t.id < $2) order by t.id desc limit $3`,
		eventID, afterID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Attendee{}
	for rows.Next() {
		var a domain.Attendee
		t, err := scanTicket(rows, &a.BuyerName, &a.BuyerEmail)
		if err != nil {
			return nil, err
		}
		a.Ticket = t
		out = append(out, a)
	}
	return out, mapErr(rows.Err())
}

// logEvent appends to a ticket's history. Statuses and actor of 0 are stored as NULL.
func logEvent(ctx context.Context, tx pgx.Tx, ticketID int64, event string, from, to domain.TicketStatus, actor int64, note string) error {
	_, err := tx.Exec(ctx, `insert into ticket_event (ticket_id, event, from_status, to_status, actor, note)
		values ($1, $2, nullif($3, 0), nullif($4, 0), nullif($5, 0), nullif($6, ''))`, ticketID, event, int16(from), int16(to), actor, note)
	return err
}

// ---- the gate

func (r *TicketRepository) CheckIn(ctx context.Context, eventID int64, code string, by int64, at time.Time) (domain.Ticket, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		with u as (
			update ticket set status = $5, checked_in_at = $4, checked_in_by = $3
			where event_id = $1 and code = $2 and status = $6 returning id
		), ev as (
			insert into ticket_event (ticket_id, event, from_status, to_status, actor)
			select id, 'checked_in', $6, $5, nullif($3, 0) from u
		)
		select id from u`, eventID, code, by, at, int16(domain.TicketUsed), int16(domain.TicketValid)).Scan(&id)
	if err == nil {
		return oneTicket(ctx, r.db, `where t.id = $1`, id)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Ticket{}, mapErr(err)
	}
	t, err := oneTicket(ctx, r.db, `where t.event_id = $1 and t.code = $2`, eventID, code)
	if err != nil {
		return domain.Ticket{}, err
	}
	switch t.Status {
	case domain.TicketUsed:
		return t, domain.ErrAlreadyCheckedIn
	case domain.TicketVoid:
		return t, fmt.Errorf("%w: the ticket is void", domain.ErrConflict)
	case domain.TicketExpired:
		return t, fmt.Errorf("%w: the ticket has expired", domain.ErrConflict)
	case domain.TicketTransferring:
		return t, fmt.Errorf("%w: the ticket is being transferred to someone else", domain.ErrConflict)
	case domain.TicketListed:
		return t, fmt.Errorf("%w: the ticket is on sale in the resale market", domain.ErrConflict)
	}
	return t, domain.ErrConflict // it became valid-and-used between the two statements; scan again
}

func (r *TicketRepository) RevertCheckIn(ctx context.Context, id, by int64) (domain.Ticket, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var one int64
		err := tx.QueryRow(ctx, `update ticket set status = $2, checked_in_at = null, checked_in_by = null
			where id = $1 and status = $3 returning id`, id, int16(domain.TicketValid), int16(domain.TicketUsed)).Scan(&one)
		if err != nil {
			return err
		}
		return logEvent(ctx, tx, id, "check_in_reverted", domain.TicketUsed, domain.TicketValid, by, "")
	})
	return r.afterChange(ctx, id, err)
}

func (r *TicketRepository) afterChange(ctx context.Context, id int64, err error) (domain.Ticket, error) {
	if errors.Is(err, pgx.ErrNoRows) {
		if _, gerr := r.Get(ctx, id); gerr != nil {
			return domain.Ticket{}, gerr
		}
		return domain.Ticket{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *TicketRepository) SetHolder(ctx context.Context, id int64, name, email string, by int64) (domain.Ticket, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var one int64
		err := tx.QueryRow(ctx, `update ticket set holder_name = nullif($2, ''), holder_email = nullif($3, '')
			where id = $1 and status = $4 returning id`, id, name, email, int16(domain.TicketValid)).Scan(&one)
		if err != nil {
			return err
		}
		return logEvent(ctx, tx, id, "holder_changed", domain.TicketValid, domain.TicketValid, by, "")
	})
	return r.afterChange(ctx, id, err)
}

func (r *TicketRepository) Reissue(ctx context.Context, id int64, newCode string, by int64) (domain.Ticket, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var one int64
		err := tx.QueryRow(ctx, `update ticket set code = $2 where id = $1 and status = $3 returning id`,
			id, newCode, int16(domain.TicketValid)).Scan(&one)
		if err != nil {
			return err
		}
		return logEvent(ctx, tx, id, "reissued", domain.TicketValid, domain.TicketValid, by, "")
	})
	return r.afterChange(ctx, id, err)
}

func (r *TicketRepository) Void(ctx context.Context, id, by int64, note string) (domain.Ticket, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var from domain.TicketStatus
		err := tx.QueryRow(ctx, `select status from ticket where id = $1 for update`, id).Scan(&from)
		if err != nil {
			return err
		}
		if from != domain.TicketValid && from != domain.TicketTransferring && from != domain.TicketListed {
			return domain.ErrConflict
		}
		if _, err := tx.Exec(ctx, `update ticket set status = $2 where id = $1`, id, int16(domain.TicketVoid)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update ticket_transfer set status = $2, resolved_at = now() where ticket_id = $1 and status = $3`,
			id, int16(domain.TransferCancelled), int16(domain.TransferPending)); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update ticket_resale set status = $2, updated_at = now() where ticket_id = $1 and status = $3`,
			id, int16(domain.ResaleCancelled), int16(domain.ResaleOpen)); err != nil {
			return err
		}
		return logEvent(ctx, tx, id, "voided", from, domain.TicketVoid, by, note)
	})
	if errors.Is(err, domain.ErrConflict) {
		return domain.Ticket{}, err
	}
	return r.afterChange(ctx, id, err)
}

// ---- transfers

const transferCols = `id, ticket_id, from_user, coalesce(to_user, 0), coalesce(to_email, ''), status, coalesce(message, ''), created_at, expires_at, resolved_at`

func scanTransfer(row pgx.Row) (domain.Transfer, error) {
	var t domain.Transfer
	err := row.Scan(&t.ID, &t.TicketID, &t.FromUser, &t.ToUser, &t.ToEmail, &t.Status, &t.Message, &t.CreatedAt, &t.ExpiresAt, &t.ResolvedAt)
	return t, mapErr(err)
}

func (r *TicketRepository) GetTransfer(ctx context.Context, id int64) (domain.Transfer, error) {
	t, err := scanTransfer(r.db.QueryRow(ctx, `select `+transferCols+` from ticket_transfer where id = $1`, id))
	if err != nil {
		return t, err
	}
	tk, err := r.Get(ctx, t.TicketID)
	if err == nil {
		t.Ticket = &tk
	}
	return t, nil
}

func (r *TicketRepository) Offer(ctx context.Context, p outbound.OfferParams) (domain.Transfer, error) {
	var id int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var status domain.TicketStatus
		var holder int64
		var count int
		err := tx.QueryRow(ctx, `select status, coalesce(holder_id, 0), transfer_count from ticket where id = $1 for update`, p.TicketID).
			Scan(&status, &holder, &count)
		if err != nil {
			return err
		}
		switch {
		case holder != p.FromUser:
			return pgx.ErrNoRows // somebody else's ticket does not exist to you
		case status != domain.TicketValid:
			return fmt.Errorf("%w: the ticket is %s, only a valid ticket can be transferred", domain.ErrConflict, status.Name())
		case p.MaxTransfers > 0 && count >= p.MaxTransfers:
			return fmt.Errorf("%w: a ticket can change hands at most %d times", domain.ErrConflict, p.MaxTransfers)
		}
		if _, err := tx.Exec(ctx, `update ticket set status = $2 where id = $1`, p.TicketID, int16(domain.TicketTransferring)); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `insert into ticket_transfer (ticket_id, from_user, to_user, to_email, message, expires_at)
			values ($1, $2, nullif($3, 0), nullif($4, ''), nullif($5, ''), $6) returning id`,
			p.TicketID, p.FromUser, p.ToUser, p.ToEmail, p.Message, p.ExpiresAt).Scan(&id); err != nil {
			return err
		}
		return logEvent(ctx, tx, p.TicketID, "transfer_offered", domain.TicketValid, domain.TicketTransferring, p.FromUser, "")
	})
	if err != nil {
		return domain.Transfer{}, mapErr(err)
	}
	return r.GetTransfer(ctx, id)
}

func (r *TicketRepository) transfers(ctx context.Context, where string, args ...any) ([]domain.Transfer, error) {
	rows, err := r.db.Query(ctx, `select `+transferCols+` from ticket_transfer where `+where, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Transfer{}
	var ids []int64
	for rows.Next() {
		t, err := scanTransfer(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
		ids = append(ids, t.TicketID)
	}
	if err := rows.Err(); err != nil {
		return nil, mapErr(err)
	}
	if len(ids) == 0 {
		return out, nil
	}
	tickets, err := queryTickets(ctx, r.db, `where t.id = any($1)`, ids)
	if err != nil {
		return nil, err
	}
	by := make(map[int64]domain.Ticket, len(tickets))
	for _, t := range tickets {
		by[t.ID] = t
	}
	for i := range out {
		if t, ok := by[out[i].TicketID]; ok {
			out[i].Ticket = &t
		}
	}
	return out, nil
}

func (r *TicketRepository) Incoming(ctx context.Context, userID int64, email string, afterID int64, limit int) ([]domain.Transfer, error) {
	return r.transfers(ctx, `status = $1 and expires_at > now() and (to_user = $2 or (to_email is not null and lower(to_email) = lower($3)))
		and ($4::bigint = 0 or id < $4) order by id desc limit $5`, int16(domain.TransferPending), userID, email, afterID, limit)
}

func (r *TicketRepository) Outgoing(ctx context.Context, userID, afterID int64, limit int) ([]domain.Transfer, error) {
	return r.transfers(ctx, `from_user = $1 and ($2::bigint = 0 or id < $2) order by id desc limit $3`, userID, afterID, limit)
}

func (r *TicketRepository) AcceptTransfer(ctx context.Context, id, newHolder int64, name, email, newCode string) (domain.Ticket, error) {
	var ticketID int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var from int64
		err := tx.QueryRow(ctx, `update ticket_transfer set status = $2, resolved_at = now()
			where id = $1 and status = $3 and expires_at > now() returning ticket_id, from_user`,
			id, int16(domain.TransferAccepted), int16(domain.TransferPending)).Scan(&ticketID, &from)
		if err != nil {
			return err
		}
		var one int64
		err = tx.QueryRow(ctx, `update ticket set holder_id = $2, holder_name = nullif($3, ''), holder_email = nullif($4, ''), code = $5,
				status = $6, transfer_count = transfer_count + 1
			where id = $1 and status = $7 returning id`,
			ticketID, newHolder, name, email, newCode, int16(domain.TicketValid), int16(domain.TicketTransferring)).Scan(&one)
		if err != nil {
			return domain.ErrConflict
		}
		return logEvent(ctx, tx, ticketID, "transfer_accepted", domain.TicketTransferring, domain.TicketValid, newHolder, fmt.Sprintf("from user %d", from))
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Ticket{}, domain.ErrConflict // no longer open
	}
	if err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	return r.Get(ctx, ticketID)
}

func (r *TicketRepository) CloseTransfer(ctx context.Context, id int64, as domain.TransferStatus, by int64) (domain.Transfer, error) {
	event := "transfer_declined"
	if as == domain.TransferCancelled {
		event = "transfer_cancelled"
	}
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var ticketID int64
		if err := tx.QueryRow(ctx, `update ticket_transfer set status = $2, resolved_at = now() where id = $1 and status = $3 returning ticket_id`,
			id, int16(as), int16(domain.TransferPending)).Scan(&ticketID); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update ticket set status = $2 where id = $1 and status = $3`,
			ticketID, int16(domain.TicketValid), int16(domain.TicketTransferring)); err != nil {
			return err
		}
		return logEvent(ctx, tx, ticketID, event, domain.TicketTransferring, domain.TicketValid, by, "")
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Transfer{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Transfer{}, mapErr(err)
	}
	return r.GetTransfer(ctx, id)
}

// ---- expiry

func (r *TicketRepository) ExpireTickets(ctx context.Context, now time.Time, limit int) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		with picked as (
			select t.id, t.status from ticket t join event_session e on e.id = t.session_id
			where t.status in ($3, $4, $8) and e.ends_at < $1 order by t.id limit $2 for update of t skip locked
		), u as (
			update ticket t set status = $5 from picked p where t.id = p.id returning t.id
		), c as (
			update ticket_transfer x set status = $6, resolved_at = $1 where x.status = $7 and x.ticket_id in (select id from picked)
		), l as (
			update ticket_resale x set status = $9, updated_at = $1 where x.status = $10 and x.ticket_id in (select id from picked)
		), ev as (
			insert into ticket_event (ticket_id, event, from_status, to_status, note)
			select id, 'expired', status, $5, 'the event is over' from picked
		)
		select count(*) from u`, now, limit, int16(domain.TicketValid), int16(domain.TicketTransferring), int16(domain.TicketExpired),
		int16(domain.TransferExpired), int16(domain.TransferPending), int16(domain.TicketListed), int16(domain.ResaleCancelled), int16(domain.ResaleOpen)).Scan(&n)
	return n, mapErr(err)
}

func (r *TicketRepository) ExpireTransfers(ctx context.Context, now time.Time, limit int) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		with due as (
			select id from ticket_transfer where status = $3 and expires_at < $1 order by id limit $2 for update skip locked
		), t as (
			update ticket_transfer x set status = $4, resolved_at = $1 from due where x.id = due.id returning x.ticket_id
		), k as (
			update ticket set status = $5 where id in (select ticket_id from t) and status = $6 returning id
		), ev as (
			insert into ticket_event (ticket_id, event, from_status, to_status, note) select id, 'transfer_expired', $6, $5, 'nobody answered' from k
		)
		select count(*) from t`, now, limit, int16(domain.TransferPending), int16(domain.TransferExpired),
		int16(domain.TicketValid), int16(domain.TicketTransferring)).Scan(&n)
	return n, mapErr(err)
}
