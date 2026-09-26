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

type ResaleRepository struct{ db *pgxpool.Pool }

var _ outbound.ResaleRepository = (*ResaleRepository)(nil)

func NewResaleRepository(db *pgxpool.Pool) *ResaleRepository { return &ResaleRepository{db: db} }

const resaleCols = `r.id, r.ticket_id, t.event_id, r.seller_id, coalesce(r.buyer_id, 0), r.price, coalesce(i.unit_price, 0), r.fee,
	r.currency, r.status, coalesce(r.payment_ref, ''), tt.name, coalesce(s.section, ''), coalesce(s.row_label, ''), coalesce(s.number, 0),
	r.created_at, r.updated_at, coalesce(ss.starts_at, '9999-12-31'::timestamptz)`

const resaleFrom = ` from ticket_resale r join ticket t on t.id = r.ticket_id join ticket_type tt on tt.id = t.ticket_type_id
	left join order_item i on i.order_id = t.order_id and i.ticket_type_id = t.ticket_type_id left join seat s on s.id = t.seat_id
	left join event_session ss on ss.id = t.session_id `

func scanResale(row pgx.Row) (domain.ResaleListing, error) {
	var l domain.ResaleListing
	var seat domain.Seat
	err := row.Scan(&l.ID, &l.TicketID, &l.EventID, &l.SellerID, &l.BuyerID, &l.Price, &l.FacePrice, &l.Fee, &l.Currency, &l.Status,
		&l.PaymentRef, &l.TypeName, &seat.Section, &seat.Row, &seat.Number, &l.CreatedAt, &l.UpdatedAt, &l.StartsAt)
	if seat.Row != "" {
		l.SeatLabel = seat.Label()
	}
	return l, mapErr(err)
}

func (r *ResaleRepository) Get(ctx context.Context, id int64) (domain.ResaleListing, error) {
	return scanResale(r.db.QueryRow(ctx, `select `+resaleCols+resaleFrom+` where r.id = $1`, id))
}

func (r *ResaleRepository) List(ctx context.Context, f outbound.ResaleFilter) ([]domain.ResaleListing, error) {
	order, after := `r.id desc`, f.AfterID
	if f.CheapestFirst {
		order, after = `r.price, r.id`, 0
	}
	rows, err := r.db.Query(ctx, `select `+resaleCols+resaleFrom+` where ($1::bigint = 0 or t.event_id = $1) and ($2::bigint = 0 or t.ticket_type_id = $2)
		and ($3::bigint = 0 or r.seller_id = $3) and ($4::smallint = 0 or r.status = $4) and ($5::bigint = 0 or r.id < $5)
		order by `+order+` limit $6`, f.EventID, f.TypeID, f.SellerID, int16(f.Status), after, f.Limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.ResaleListing{}
	for rows.Next() {
		l, err := scanResale(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, mapErr(rows.Err())
}

func (r *ResaleRepository) Create(ctx context.Context, p outbound.CreateListing) (domain.ResaleListing, error) {
	var id int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var status domain.TicketStatus
		var holder int64
		var count int
		var startsAt time.Time
		var evStatus domain.EventStatus
		var capPct int
		var currency string
		var face int64
		err := tx.QueryRow(ctx, `
			select t.status, coalesce(t.holder_id, 0), t.transfer_count, coalesce(ss.starts_at, e.starts_at), e.status, e.resale_cap_percent, e.currency, coalesce(i.unit_price, 0)
			from ticket t join event e on e.id = t.event_id left join event_session ss on ss.id = t.session_id
			left join order_item i on i.order_id = t.order_id and i.ticket_type_id = t.ticket_type_id
			where t.id = $1 for update of t`, p.TicketID).Scan(&status, &holder, &count, &startsAt, &evStatus, &capPct, &currency, &face)
		if err != nil {
			return err
		}
		switch {
		case holder != p.SellerID:
			return pgx.ErrNoRows // somebody else's ticket does not exist to you
		case status != domain.TicketValid:
			return fmt.Errorf("%w: the ticket is %s, only a valid ticket can be sold", domain.ErrConflict, status.Name())
		case evStatus != domain.EventPublished || !time.Now().Before(startsAt):
			return fmt.Errorf("%w: tickets can only be sold before a published event starts", domain.ErrConflict)
		case capPct <= 0:
			return fmt.Errorf("%w: the organizer does not allow tickets of this event to be resold", domain.ErrConflict)
		case p.MaxTransfers > 0 && count >= p.MaxTransfers:
			return fmt.Errorf("%w: a ticket can change hands at most %d times", domain.ErrConflict, p.MaxTransfers)
		}
		if ceiling := domain.ResaleCeiling(face, capPct); p.Price <= 0 || p.Price > ceiling {
			return fmt.Errorf("%w: the price must be between 1 and %d (%d%% of the face value %d)", domain.ErrInvalid, ceiling, capPct, face)
		}
		if _, err := tx.Exec(ctx, `update ticket set status = $2 where id = $1`, p.TicketID, int16(domain.TicketListed)); err != nil {
			return err
		}
		if err := tx.QueryRow(ctx, `insert into ticket_resale (ticket_id, seller_id, price, currency) values ($1, $2, $3, $4) returning id`,
			p.TicketID, p.SellerID, p.Price, currency).Scan(&id); err != nil {
			return err
		}
		return logEvent(ctx, tx, p.TicketID, "resale_listed", domain.TicketValid, domain.TicketListed, p.SellerID, fmt.Sprintf("price %d", p.Price))
	})
	if err != nil {
		return domain.ResaleListing{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *ResaleRepository) Cancel(ctx context.Context, id, sellerID int64) (domain.ResaleListing, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var ticketID int64
		err := tx.QueryRow(ctx, `update ticket_resale set status = $3, updated_at = now() where id = $1 and seller_id = $2 and status = $4 returning ticket_id`,
			id, sellerID, int16(domain.ResaleCancelled), int16(domain.ResaleOpen)).Scan(&ticketID)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update ticket set status = $2 where id = $1 and status = $3`, ticketID, int16(domain.TicketValid), int16(domain.TicketListed)); err != nil {
			return err
		}
		return logEvent(ctx, tx, ticketID, "resale_cancelled", domain.TicketListed, domain.TicketValid, sellerID, "")
	})
	if errors.Is(err, pgx.ErrNoRows) {
		if l, gerr := r.Get(ctx, id); gerr != nil || l.SellerID != sellerID {
			return domain.ResaleListing{}, domain.ErrNotFound
		}
		return domain.ResaleListing{}, domain.ErrConflict // sold, claimed by a buyer, or already down
	}
	if err != nil {
		return domain.ResaleListing{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *ResaleRepository) Claim(ctx context.Context, id, buyerID int64) (domain.ResaleListing, error) {
	// One statement: of many buyers reaching for the listing at once, exactly one matches "status = open".
	tag, err := r.db.Exec(ctx, `update ticket_resale set status = $3, buyer_id = $2, updated_at = now()
		where id = $1 and status = $4 and seller_id <> $2`, id, buyerID, int16(domain.ResaleProcessing), int16(domain.ResaleOpen))
	if err != nil {
		return domain.ResaleListing{}, mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		if _, gerr := r.Get(ctx, id); gerr != nil {
			return domain.ResaleListing{}, gerr
		}
		return domain.ResaleListing{}, domain.ErrConflict
	}
	return r.Get(ctx, id)
}

func (r *ResaleRepository) Release(ctx context.Context, id int64) error {
	// Back on the market, unless its ticket was taken out of it meanwhile (revoked, expired): then the listing ends.
	_, err := r.db.Exec(ctx, `
		update ticket_resale set buyer_id = null, payment_ref = null, updated_at = now(),
			status = case when exists (select 1 from ticket t where t.id = ticket_id and t.status = $4) then $2::smallint else $5::smallint end
		where id = $1 and status = $3`,
		id, int16(domain.ResaleOpen), int16(domain.ResaleProcessing), int16(domain.TicketListed), int16(domain.ResaleCancelled))
	return mapErr(err)
}

func (r *ResaleRepository) SetPayment(ctx context.Context, id int64, ref string, fee int64) error {
	tag, err := r.db.Exec(ctx, `update ticket_resale set payment_ref = $2, fee = $3, updated_at = now() where id = $1 and status = $4`,
		id, ref, fee, int16(domain.ResaleProcessing))
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *ResaleRepository) Complete(ctx context.Context, id, buyerID int64, email, newCode string) (domain.Ticket, error) {
	var ticketID int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var seller, price int64
		err := tx.QueryRow(ctx, `update ticket_resale set status = $3, updated_at = now() where id = $1 and buyer_id = $2 and status = $4
			returning ticket_id, seller_id, price`, id, buyerID, int16(domain.ResaleSold), int16(domain.ResaleProcessing)).Scan(&ticketID, &seller, &price)
		if err != nil {
			return err
		}
		var one int64
		err = tx.QueryRow(ctx, `update ticket set holder_id = $2, holder_name = null, holder_email = nullif($3, ''), code = $4, status = $5,
				transfer_count = transfer_count + 1
			where id = $1 and status = $6 returning id`,
			ticketID, buyerID, email, newCode, int16(domain.TicketValid), int16(domain.TicketListed)).Scan(&one)
		if err != nil {
			return domain.ErrConflict // the ticket is not on sale any more (voided, expired): the sale is undone by the rollback
		}
		return logEvent(ctx, tx, ticketID, "resold", domain.TicketListed, domain.TicketValid, buyerID, fmt.Sprintf("from user %d for %d", seller, price))
	})
	if errors.Is(err, pgx.ErrNoRows) {
		// not processing: fine only if this buyer already completed it
		var got int64
		if qerr := r.db.QueryRow(ctx, `select ticket_id from ticket_resale where id = $1 and buyer_id = $2 and status = $3`,
			id, buyerID, int16(domain.ResaleSold)).Scan(&got); qerr == nil {
			return oneTicket(ctx, r.db, `where t.id = $1`, got)
		}
		return domain.Ticket{}, domain.ErrConflict
	}
	if err != nil {
		return domain.Ticket{}, mapErr(err)
	}
	return oneTicket(ctx, r.db, `where t.id = $1`, ticketID)
}

func (r *ResaleRepository) Stale(ctx context.Context, before time.Time, limit int) ([]domain.ResaleListing, error) {
	rows, err := r.db.Query(ctx, `select `+resaleCols+resaleFrom+` where r.status = $1 and r.updated_at < $2 order by r.id limit $3`,
		int16(domain.ResaleProcessing), before, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.ResaleListing{}
	for rows.Next() {
		l, err := scanResale(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, l)
	}
	return out, mapErr(rows.Err())
}

func (r *ResaleRepository) ExpireStarted(ctx context.Context, now time.Time, limit int) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `
		with due as (
			select r.id, r.ticket_id from ticket_resale r join ticket t on t.id = r.ticket_id join event_session e on e.id = t.session_id
			where r.status = $3 and e.starts_at < $1 order by r.id limit $2 for update of r skip locked
		), u as (
			update ticket_resale x set status = $4, updated_at = $1 from due where x.id = due.id returning x.ticket_id
		), k as (
			update ticket set status = $5 where id in (select ticket_id from u) and status = $6 returning id
		), ev as (
			insert into ticket_event (ticket_id, event, from_status, to_status, note) select id, 'resale_expired', $6, $5, 'the event started' from k
		)
		select count(*) from u`, now, limit, int16(domain.ResaleOpen), int16(domain.ResaleCancelled), int16(domain.TicketValid), int16(domain.TicketListed)).Scan(&n)
	return n, mapErr(err)
}
