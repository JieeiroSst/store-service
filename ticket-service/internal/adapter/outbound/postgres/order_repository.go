package postgres

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type OrderRepository struct{ db *pgxpool.Pool }

var _ outbound.OrderRepository = (*OrderRepository)(nil)

func NewOrderRepository(db *pgxpool.Pool) *OrderRepository { return &OrderRepository{db: db} }

func orderCols(a string) string {
	return a + "id, " + a + "user_id, " + a + "event_id, " + a + "status, " + a + "currency, " + a + "subtotal, " + a + "discount, " +
		a + "total, coalesce(" + a + "promo_id, 0), coalesce(" + a + "promo_code, ''), " + a + "buyer_name, " + a + "buyer_email, " +
		"coalesce(" + a + "buyer_phone, ''), " + a + "request_id, " + a + "expires_at, coalesce(" + a + "payment_method, ''), " +
		"coalesce(" + a + "payment_ref, ''), " + a + "paid_at, " + a + "refund_amount, coalesce(" + a + "status_note, ''), " +
		a + "version, " + a + "created_at, " + a + "updated_at, coalesce(" + a + "session_id, 0)"
}

func scanOrder(row pgx.Row) (domain.Order, error) {
	var x domain.Order
	var method string
	err := row.Scan(&x.ID, &x.UserID, &x.EventID, &x.Status, &x.Currency, &x.Subtotal, &x.Discount, &x.Total, &x.PromoID,
		&x.PromoCode, &x.BuyerName, &x.BuyerEmail, &x.BuyerPhone, &x.RequestID, &x.ExpiresAt, &method, &x.PaymentRef,
		&x.PaidAt, &x.RefundAmount, &x.StatusNote, &x.Version, &x.CreatedAt, &x.UpdatedAt, &x.SessionID)
	x.PaymentMethod = domain.PaymentMethod(method)
	return x, mapErr(err)
}

func collectOrders(rows pgx.Rows) ([]domain.Order, error) {
	defer rows.Close()
	out := []domain.Order{}
	for rows.Next() {
		x, err := scanOrder(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, mapErr(rows.Err())
}

type querier interface {
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
}

func withItems(ctx context.Context, q querier, orders []domain.Order) error {
	if len(orders) == 0 {
		return nil
	}
	ids := make([]int64, len(orders))
	at := make(map[int64]int, len(orders))
	for i, o := range orders {
		ids[i], at[o.ID] = o.ID, i
	}
	rows, err := q.Query(ctx, `select id, order_id, ticket_type_id, name, quantity, unit_price, coalesce(seat_ids, '{}')
		from order_item where order_id = any($1) order by id`, ids)
	if err != nil {
		return mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var it domain.OrderItem
		if err := rows.Scan(&it.ID, &it.OrderID, &it.TicketTypeID, &it.Name, &it.Quantity, &it.UnitPrice, &it.SeatIDs); err != nil {
			return err
		}
		o := &orders[at[it.OrderID]]
		o.Items = append(o.Items, it)
	}
	return mapErr(rows.Err())
}

func (r *OrderRepository) Get(ctx context.Context, id int64) (domain.Order, error) {
	x, err := scanOrder(r.db.QueryRow(ctx, `select `+orderCols("")+` from ticket_order where id = $1`, id))
	if err != nil {
		return domain.Order{}, err
	}
	one := []domain.Order{x}
	if err := withItems(ctx, r.db, one); err != nil {
		return domain.Order{}, err
	}
	x = one[0]
	tickets, err := queryTickets(ctx, r.db, `where t.order_id = $1 order by t.id`, id)
	if err != nil {
		return domain.Order{}, err
	}
	x.Tickets = tickets
	return x, nil
}

func (r *OrderRepository) byRequestID(ctx context.Context, userID int64, requestID string) (domain.Order, error) {
	var id int64
	if err := r.db.QueryRow(ctx, `select id from ticket_order where user_id = $1 and request_id = $2`, userID, requestID).Scan(&id); err != nil {
		return domain.Order{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *OrderRepository) FindByRequestID(ctx context.Context, userID int64, requestID string) (domain.Order, bool, error) {
	x, err := r.byRequestID(ctx, userID, requestID)
	if errors.Is(err, domain.ErrNotFound) {
		return domain.Order{}, false, nil
	}
	return x, err == nil, err
}

func (r *OrderRepository) RequestIDExists(ctx context.Context, requestID string) (bool, error) {
	var exists bool
	err := r.db.QueryRow(ctx, `select exists (select 1 from ticket_order where request_id = $1)`, requestID).Scan(&exists)
	return exists, mapErr(err)
}

func (r *OrderRepository) Available(ctx context.Context, typeIDs []int64) (map[int64]int, error) {
	rows, err := r.db.Query(ctx, `select id, available from ticket_type where id = any($1) and active`, typeIDs)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := make(map[int64]int, len(typeIDs))
	for rows.Next() {
		var id int64
		var n int
		if err := rows.Scan(&id, &n); err != nil {
			return nil, err
		}
		out[id] = n
	}
	return out, mapErr(rows.Err())
}

var errHoldCap = errors.New("hold cap")

type typeInfo struct {
	name        string
	seated      bool
	active      bool
	available   int
	minPerOrder int
	maxPerOrder int
	maxPerUser  int
	startsAt    *time.Time
	endsAt      *time.Time
	sessionID   int64
	sessionStat domain.SessionStatus
	sessionEnds time.Time
}

func (r *OrderRepository) Reserve(ctx context.Context, p outbound.ReserveParams) (domain.Order, error) {
	if x, err := r.byRequestID(ctx, p.UserID, p.RequestID); err == nil {
		return x, nil
	} else if !errors.Is(err, domain.ErrNotFound) {
		return domain.Order{}, err
	}

	items := slices.Clone(p.Items)
	slices.SortFunc(items, func(a, b outbound.ReserveItem) int { return int(a.TicketTypeID - b.TicketTypeID) })
	typeIDs := make([]int64, len(items))
	tickets := 0
	for i, it := range items {
		typeIDs[i] = it.TicketTypeID
		tickets += it.Quantity
	}

	var orderID int64
	var held int
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `set local lock_timeout = '800ms'`); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `set local statement_timeout = '3s'`); err != nil {
			return err
		}

		var status domain.EventStatus
		var endsAt time.Time
		var currency string
		var sessionID int64
		err := tx.QueryRow(ctx, `select status, ends_at, currency from event where id = $1`, p.EventID).Scan(&status, &endsAt, &currency)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		switch {
		case status != domain.EventPublished && status != domain.EventCancelled:
			return domain.ErrNotFound // drafts and events under review do not exist to buyers
		case status == domain.EventCancelled || !time.Now().Before(endsAt):
			return fmt.Errorf("%w: the event is not on sale", domain.ErrConflict)
		}

		if _, err := tx.Exec(ctx, `select pg_advisory_xact_lock(hashtextextended('ticket-user-' || $1::bigint::text, 0))`, p.UserID); err != nil {
			return err
		}
		if p.MaxPending > 0 && !p.Comp {
			if err := tx.QueryRow(ctx, `select count(*) from ticket_order where user_id = $1 and status = $2`, p.UserID, domain.OrderPending).Scan(&held); err != nil {
				return err
			}
			if held >= p.MaxPending {
				return errHoldCap
			}
		}

		infos := make(map[int64]typeInfo, len(items))
		rows, err := tx.Query(ctx, `select tt.id, tt.name, tt.seated, tt.active, tt.available, tt.min_per_order, tt.max_per_order, tt.max_per_user,
				tt.sale_starts_at, tt.sale_ends_at, coalesce(tt.session_id, 0), coalesce(s.status, 1), coalesce(s.ends_at, '9999-12-31'::timestamptz)
			from ticket_type tt left join event_session s on s.id = tt.session_id where tt.id = any($1) and tt.event_id = $2`, typeIDs, p.EventID)
		if err != nil {
			return err
		}
		for rows.Next() {
			var id int64
			var t typeInfo
			if err := rows.Scan(&id, &t.name, &t.seated, &t.active, &t.available, &t.minPerOrder, &t.maxPerOrder, &t.maxPerUser, &t.startsAt, &t.endsAt, &t.sessionID, &t.sessionStat, &t.sessionEnds); err != nil {
				rows.Close()
				return err
			}
			infos[id] = t
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		now := time.Now()
		for _, it := range items {
			t, ok := infos[it.TicketTypeID]
			if ok && sessionID == 0 {
				sessionID = t.sessionID
			}
			switch {
			case !ok || !t.active:
				return fmt.Errorf("%w: ticket type %d is not offered by this event", domain.ErrNotFound, it.TicketTypeID)
			case t.sessionStat != domain.SessionScheduled:
				return fmt.Errorf("%w: the showtime of %s was cancelled", domain.ErrConflict, t.name)
			case !now.Before(t.sessionEnds):
				return fmt.Errorf("%w: the showtime of %s is over", domain.ErrConflict, t.name)
			case sessionID != 0 && t.sessionID != sessionID:
				return fmt.Errorf("%w: one order is for one showtime; the ticket types you chose belong to different ones", domain.ErrInvalid)
			case !p.Comp && t.startsAt != nil && now.Before(*t.startsAt):
				return fmt.Errorf("%w: sales of %s have not started", domain.ErrConflict, t.name)
			case !p.Comp && t.endsAt != nil && !now.Before(*t.endsAt):
				return fmt.Errorf("%w: sales of %s have ended", domain.ErrConflict, t.name)
			case !p.Comp && (it.Quantity < t.minPerOrder || it.Quantity > t.maxPerOrder):
				return fmt.Errorf("%w: %s can be bought %d to %d at a time", domain.ErrInvalid, t.name, t.minPerOrder, t.maxPerOrder)
			case t.seated && len(it.SeatIDs) != it.Quantity:
				return fmt.Errorf("%w: %s is seated: pick %d seat(s)", domain.ErrInvalid, t.name, it.Quantity)
			case !t.seated && len(it.SeatIDs) != 0:
				return fmt.Errorf("%w: %s has no seats to pick", domain.ErrInvalid, t.name)
			case t.available < it.Quantity:
				return &domain.SoldOutError{TicketTypeID: it.TicketTypeID, Name: t.name} // hopeless: do not even queue for the row
			}
			if t.maxPerUser > 0 && !p.Comp {
				var have int
				if err := tx.QueryRow(ctx, `select coalesce(sum(i.quantity), 0) from order_item i join ticket_order o on o.id = i.order_id
					where o.user_id = $1 and i.ticket_type_id = $2 and o.status in ($3, $4)`,
					p.UserID, it.TicketTypeID, domain.OrderPending, domain.OrderPaid).Scan(&have); err != nil {
					return err
				}
				if have+it.Quantity > t.maxPerUser {
					return fmt.Errorf("%w: at most %d %s tickets per account (you already have %d)", domain.ErrConflict, t.maxPerUser, t.name, have)
				}
			}
		}

		// 1. seats
		var seatIDs []int64
		for _, it := range items {
			if len(it.SeatIDs) == 0 {
				continue
			}
			tag, err := tx.Exec(ctx, `
				with picked as (
					select id from seat where id = any($1) and ticket_type_id = $2 and status = $3
					order by id for update skip locked
				)
				update seat s set status = $4 from picked where s.id = picked.id`,
				it.SeatIDs, it.TicketTypeID, domain.SeatAvailable, domain.SeatHeld)
			if err != nil {
				return err
			}
			if int(tag.RowsAffected()) != len(it.SeatIDs) {
				return fmt.Errorf("%w: a seat you picked is no longer available", domain.ErrConflict)
			}
			seatIDs = append(seatIDs, it.SeatIDs...)
		}

		// 2. tickets
		var subtotal int64
		unit := make(map[int64]int64, len(items))
		for _, it := range items {
			var price int64
			err := tx.QueryRow(ctx, `
				update ticket_type set available = available - $3
				where id = $1 and event_id = $2 and active and available >= $3
					and ($4 or ((sale_starts_at is null or sale_starts_at <= now()) and (sale_ends_at is null or sale_ends_at > now())))
				returning price`, it.TicketTypeID, p.EventID, it.Quantity, p.Comp).Scan(&price)
			if errors.Is(err, pgx.ErrNoRows) {
				return &domain.SoldOutError{TicketTypeID: it.TicketTypeID, Name: infos[it.TicketTypeID].name} // rolls everything back
			}
			if err != nil {
				return err
			}
			if p.Comp {
				price = 0
			}
			unit[it.TicketTypeID] = price
			subtotal += price * int64(it.Quantity)
		}

		// 3. promo code
		var discount int64
		var promoID, promoCode any
		if p.Promo != nil {
			var kind domain.PromoKind
			var value int64
			err := tx.QueryRow(ctx, `
				update promotion set used = used + 1
				where id = $1 and event_id = $2 and active and (max_uses = 0 or used < max_uses)
					and (valid_from is null or valid_from <= now()) and (valid_to is null or valid_to >= now()) and min_tickets <= $3
				returning kind, value`, p.Promo.ID, p.EventID, tickets).Scan(&kind, &value)
			if errors.Is(err, pgx.ErrNoRows) {
				return fmt.Errorf("%w: the promo code is no longer available", domain.ErrInvalid)
			}
			if err != nil {
				return err
			}
			discount = domain.Promotion{Kind: kind, Value: value}.DiscountFor(subtotal)
			promoID, promoCode = p.Promo.ID, p.Promo.Code
		}

		// 4. the order
		if err := tx.QueryRow(ctx, `
			insert into ticket_order (user_id, event_id, status, currency, subtotal, discount, total, promo_id, promo_code,
				buyer_name, buyer_email, buyer_phone, request_id, expires_at, invited_by, session_id)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, nullif($12, ''), $13, $14, nullif($15, 0), nullif($16, 0))
			returning id`,
			p.UserID, p.EventID, domain.OrderPending, currency, subtotal, discount, subtotal-discount, promoID, promoCode,
			p.BuyerName, p.BuyerEmail, p.BuyerPhone, p.RequestID, p.ExpiresAt, p.InvitedBy, sessionID).Scan(&orderID); err != nil {
			return err
		}
		for _, it := range items {
			if _, err := tx.Exec(ctx, `insert into order_item (order_id, ticket_type_id, name, quantity, unit_price, seat_ids)
				values ($1, $2, $3, $4, $5, $6)`,
				orderID, it.TicketTypeID, infos[it.TicketTypeID].name, it.Quantity, unit[it.TicketTypeID], nilIfEmpty(it.SeatIDs)); err != nil {
				return err
			}
		}
		if len(seatIDs) > 0 {
			if _, err := tx.Exec(ctx, `update seat set order_id = $2 where id = any($1)`, seatIDs, orderID); err != nil {
				return err
			}
		}
		return nil
	})
	if errors.Is(err, errHoldCap) {
		return domain.Order{}, fmt.Errorf("%w: you already hold %d unpaid orders; pay or cancel one first", domain.ErrConflict, held)
	}
	var pgConflict = errors.Is(mapErr(err), domain.ErrConflict) && !errors.Is(err, domain.ErrConflict)
	if pgConflict {
		// a unique violation: a retry of this very request won the race to insert
		if x, ferr := r.byRequestID(ctx, p.UserID, p.RequestID); ferr == nil {
			return x, nil
		}
	}
	if err != nil {
		return domain.Order{}, mapErr(err)
	}
	return r.Get(ctx, orderID)
}

func nilIfEmpty(ids []int64) any {
	if len(ids) == 0 {
		return nil
	}
	return ids
}

func (r *OrderRepository) List(ctx context.Context, f outbound.OrderFilter) ([]domain.Order, error) {
	rows, err := r.db.Query(ctx, `select `+orderCols("")+` from ticket_order
		where ($1::bigint = 0 or user_id = $1) and ($2::bigint = 0 or event_id = $2) and ($3::smallint = 0 or status = $3) and ($4::bigint = 0 or id < $4)
		order by id desc limit $5`, f.UserID, f.EventID, int16(f.Status), f.AfterID, f.Limit)
	if err != nil {
		return nil, mapErr(err)
	}
	out, err := collectOrders(rows)
	if err != nil {
		return nil, err
	}
	return out, withItems(ctx, r.db, out)
}

func (r *OrderRepository) SetPaymentAttempt(ctx context.Context, id int64, m domain.PaymentMethod, ref string) (domain.Order, error) {
	err := r.db.QueryRow(ctx, `
		update ticket_order set payment_method = $2, payment_ref = $3, version = version + 1, updated_at = now()
		where id = $1 and status = $4 returning id`, id, string(m), ref, domain.OrderPending).Scan(&id)
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, domain.ErrConflict // no longer pending
	}
	if err != nil {
		return domain.Order{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *OrderRepository) MarkPaid(ctx context.Context, id int64, m domain.PaymentMethod, ref string, codes []string) (domain.Order, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `set local lock_timeout = '2s'`); err != nil {
			return err
		}
		var eventID int64
		err := tx.QueryRow(ctx, `
			update ticket_order set status = $4, payment_method = $2, payment_ref = $3, paid_at = now(), version = version + 1, updated_at = now()
			where id = $1 and status = $5 returning event_id`, id, string(m), ref, domain.OrderPaid, domain.OrderPending).Scan(&eventID)
		if err != nil {
			return err // no row: not pending any more
		}
		rows, err := tx.Query(ctx, `select ticket_type_id, quantity, coalesce(seat_ids, '{}') from order_item where order_id = $1 order by ticket_type_id`, id)
		if err != nil {
			return err
		}
		type item struct {
			typeID int64
			qty    int
			seats  []int64
		}
		var items []item
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.typeID, &it.qty, &it.seats); err != nil {
				rows.Close()
				return err
			}
			items = append(items, it)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}

		var typeIDs, seatIDs []int64
		for _, it := range items {
			if _, err := tx.Exec(ctx, `update ticket_type set sold = sold + $2 where id = $1`, it.typeID, it.qty); err != nil {
				return err
			}
			for k := 0; k < it.qty; k++ {
				typeIDs = append(typeIDs, it.typeID)
				if k < len(it.seats) {
					seatIDs = append(seatIDs, it.seats[k])
				} else {
					seatIDs = append(seatIDs, 0)
				}
			}
			if len(it.seats) > 0 {
				if _, err := tx.Exec(ctx, `update seat set status = $2 where id = any($1)`, it.seats, domain.SeatSold); err != nil {
					return err
				}
			}
		}
		if len(codes) != len(typeIDs) {
			return fmt.Errorf("need %d ticket codes, got %d", len(typeIDs), len(codes))
		}
		_, err = tx.Exec(ctx, `
			with ins as (
				insert into ticket (order_id, event_id, ticket_type_id, seat_id, code, holder_id, holder_name, holder_email, session_id)
				select o.id, $2, u.t, nullif(u.s, 0), u.c, o.user_id, o.buyer_name, o.buyer_email, o.session_id
				from unnest($3::bigint[], $4::bigint[], $5::text[]) as u(t, s, c), ticket_order o where o.id = $1
				returning id, holder_id
			)
			insert into ticket_event (ticket_id, event, to_status, actor) select id, 'issued', $6, holder_id from ins`,
			id, eventID, typeIDs, seatIDs, codes, int16(domain.TicketValid))
		return err
	})
	if err == nil {
		return r.Get(ctx, id)
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return domain.Order{}, mapErr(err)
	}
	cur, err := r.Get(ctx, id)
	if err != nil {
		return domain.Order{}, err
	}
	if cur.Status == domain.OrderPaid && cur.PaymentRef == ref && cur.PaymentMethod == m {
		return cur, nil
	}
	return domain.Order{}, domain.ErrConflict
}

func (r *OrderRepository) Transition(ctx context.Context, p outbound.TransitionParams) (domain.Order, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `set local lock_timeout = '2s'`); err != nil {
			return err
		}
		var status domain.OrderStatus
		var promoID *int64
		err := tx.QueryRow(ctx, `select status, promo_id from ticket_order where id = $1 for update`, p.ID).Scan(&status, &promoID)
		if errors.Is(err, pgx.ErrNoRows) {
			return domain.ErrNotFound
		}
		if err != nil {
			return err
		}
		if !slices.Contains(p.From, status) {
			return domain.ErrConflict
		}
		if p.To == domain.OrderRefunded && !p.Force {
			var used bool
			if err := tx.QueryRow(ctx, `select exists (select 1 from ticket where order_id = $1 and status = $2)`, p.ID, int16(domain.TicketUsed)).Scan(&used); err != nil {
				return err
			}
			if used {
				return domain.ErrConflict
			}
		}
		if _, err := tx.Exec(ctx, `
			update ticket_order set status = $2, status_note = nullif($3, ''), refund_amount = $4, version = version + 1, updated_at = now()
			where id = $1`, p.ID, p.To, p.Note, p.Refund); err != nil {
			return err
		}

		rows, err := tx.Query(ctx, `select ticket_type_id, quantity, coalesce(seat_ids, '{}') from order_item where order_id = $1 order by ticket_type_id`, p.ID)
		if err != nil {
			return err
		}
		type item struct {
			typeID int64
			qty    int
			seats  []int64
		}
		var items []item
		for rows.Next() {
			var it item
			if err := rows.Scan(&it.typeID, &it.qty, &it.seats); err != nil {
				rows.Close()
				return err
			}
			items = append(items, it)
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		for _, it := range items {
			q := `update ticket_type set available = available + $2 where id = $1`
			if status == domain.OrderPaid {
				q = `update ticket_type set available = available + $2, sold = sold - $2 where id = $1`
			}
			if _, err := tx.Exec(ctx, q, it.typeID, it.qty); err != nil {
				return err
			}
			if len(it.seats) > 0 {
				if _, err := tx.Exec(ctx, `update seat set status = $2, order_id = null where id = any($1)`, it.seats, domain.SeatAvailable); err != nil {
					return err
				}
			}
		}
		if p.To == domain.OrderRefunded {
			if _, err := tx.Exec(ctx, `update ticket_transfer set status = $2, resolved_at = now()
				where status = $3 and ticket_id in (select id from ticket where order_id = $1)`, p.ID, int16(domain.TransferCancelled), int16(domain.TransferPending)); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `update ticket_resale set status = $2, updated_at = now()
				where status = $3 and ticket_id in (select id from ticket where order_id = $1)`, p.ID, int16(domain.ResaleCancelled), int16(domain.ResaleOpen)); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
				with old as (select id, status from ticket where order_id = $1 and status <> $2 for update),
				u as (update ticket set status = $2 where id in (select id from old))
				insert into ticket_event (ticket_id, event, from_status, to_status, note) select id, 'voided', status, $2, $3 from old`,
				p.ID, int16(domain.TicketVoid), "order "+p.To.Name()); err != nil {
				return err
			}
		}
		if promoID != nil {
			if _, err := tx.Exec(ctx, `update promotion set used = greatest(used - 1, 0) where id = $1`, *promoID); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.Order{}, mapErr(err)
	}
	return r.Get(ctx, p.ID)
}

func (r *OrderRepository) listOrders(ctx context.Context, sql string, args ...any) ([]domain.Order, error) {
	rows, err := r.db.Query(ctx, sql, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	out, err := collectOrders(rows)
	if err != nil {
		return nil, err
	}
	return out, withItems(ctx, r.db, out)
}

func (r *OrderRepository) ListExpired(ctx context.Context, before time.Time, limit int) ([]domain.Order, error) {
	return r.listOrders(ctx, `select `+orderCols("")+` from ticket_order where status = $1 and expires_at < $2 order by expires_at limit $3`,
		domain.OrderPending, before, limit)
}

func (r *OrderRepository) ListOpenOfCancelledEvents(ctx context.Context, limit int) ([]domain.Order, error) {
	return r.listOrders(ctx, `select `+orderCols("o.")+` from ticket_order o join event e on e.id = o.event_id
		left join event_session s on s.id = o.session_id
		where (e.status = $1 or s.status = $5) and o.status in ($2, $3) order by o.id limit $4`,
		domain.EventCancelled, domain.OrderPending, domain.OrderPaid, limit, int16(domain.SessionCancelled))
}

func (r *OrderRepository) DueForReminder(ctx context.Context, from, to time.Time, limit int) ([]domain.Order, error) {
	return r.listOrders(ctx, `select `+orderCols("o.")+` from ticket_order o join event e on e.id = o.event_id
		join event_session s on s.id = o.session_id
		where o.status = $1 and o.reminded_at is null and e.status = $2 and s.status = $6 and s.starts_at > $3 and s.starts_at <= $4
		order by o.id limit $5`, domain.OrderPaid, domain.EventPublished, from, to, limit, int16(domain.SessionScheduled))
}

func (r *OrderRepository) MarkReminded(ctx context.Context, id int64) (bool, error) {
	tag, err := r.db.Exec(ctx, `update ticket_order set reminded_at = now() where id = $1 and reminded_at is null`, id)
	return tag.RowsAffected() == 1, mapErr(err)
}

func (r *OrderRepository) BuyersOfSession(ctx context.Context, sessionID int64, limit int) ([]int64, error) {
	rows, err := r.db.Query(ctx, `select distinct user_id from ticket_order where session_id = $1 and status in ($2, $3) limit $4`,
		sessionID, int16(domain.OrderPending), int16(domain.OrderPaid), limit)
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
