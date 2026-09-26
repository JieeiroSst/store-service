package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type EventRepository struct{ db *pgxpool.Pool }

var _ outbound.EventRepository = (*EventRepository)(nil)

func NewEventRepository(db *pgxpool.Pool) *EventRepository { return &EventRepository{db: db} }

func eventCols(a string) string {
	cols := []string{"id", "organizer_id", "title", "coalesce(description, '')", "category", "city", "venue",
		"coalesce(address, '')", "coalesce(banner_url, '')", "starts_at", "ends_at", "currency", "status",
		"coalesce(review_note, '')", "featured", "transferable", "resale_cap_percent", "coalesce(series_id, 0)", "coalesce(venue_id, 0)", "coalesce(wallet_id, '')", "refund_cutoff_hours", "version", "created_at", "updated_at"}
	for i, c := range cols {
		if strings.HasPrefix(c, "coalesce(") {
			cols[i] = strings.Replace(c, "(", "("+a, 1)
		} else {
			cols[i] = a + c
		}
	}
	return strings.Join(cols, ", ")
}

func scanEventInto(row pgx.Row, extra ...any) (domain.Event, error) {
	var e domain.Event
	dest := append([]any{&e.ID, &e.OrganizerID, &e.Title, &e.Description, &e.Category, &e.City, &e.Venue, &e.Address,
		&e.BannerURL, &e.StartsAt, &e.EndsAt, &e.Currency, &e.Status, &e.ReviewNote, &e.Featured, &e.Transferable, &e.ResaleCapPercent, &e.SeriesID, &e.VenueID, &e.WalletID,
		&e.RefundCutoffHours, &e.Version, &e.CreatedAt, &e.UpdatedAt}, extra...)
	return e, mapErr(row.Scan(dest...))
}

const typeCols = `id, event_id, name, coalesce(description, ''), price, total, available, sold, min_per_order, max_per_order,
	max_per_user, sale_starts_at, sale_ends_at, seated, active, sort_order, coalesce(session_id, 0)`

func scanType(row pgx.Row) (domain.TicketType, error) {
	var t domain.TicketType
	err := row.Scan(&t.ID, &t.EventID, &t.Name, &t.Description, &t.Price, &t.Total, &t.Available, &t.Sold, &t.MinPerOrder,
		&t.MaxPerOrder, &t.MaxPerUser, &t.SaleStartsAt, &t.SaleEndsAt, &t.Seated, &t.Active, &t.SortOrder, &t.SessionID)
	return t, mapErr(err)
}

func (r *EventRepository) Create(ctx context.Context, e domain.Event) (domain.Event, error) {
	var id int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			insert into event (organizer_id, title, description, category, city, venue, address, banner_url, starts_at, ends_at,
				currency, status, wallet_id, refund_cutoff_hours, transferable, resale_cap_percent)
			values ($1, $2, nullif($3, ''), $4, $5, $6, nullif($7, ''), nullif($8, ''), $9, $10, $11, $12, nullif($13, ''), $14, $15, $16)
			returning id`,
			e.OrganizerID, e.Title, e.Description, e.Category, e.City, e.Venue, e.Address, e.BannerURL, e.StartsAt, e.EndsAt,
			e.Currency, e.Status, e.WalletID, e.RefundCutoffHours, e.Transferable, e.ResaleCapPercent).Scan(&id); err != nil {
			return err
		}
		_, err := tx.Exec(ctx, `insert into event_session (event_id, starts_at, ends_at) values ($1, $2, $3)`, id, e.StartsAt, e.EndsAt)
		return err
	})
	if err != nil {
		return domain.Event{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *EventRepository) Update(ctx context.Context, e domain.Event) (domain.Event, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			update event set title = $2, description = nullif($3, ''), category = $4, city = $5, venue = $6, address = nullif($7, ''),
				banner_url = nullif($8, ''), starts_at = $9, ends_at = $10, currency = $11, wallet_id = nullif($12, ''),
				refund_cutoff_hours = $13, transferable = $14, resale_cap_percent = $15, version = version + 1, updated_at = now()
			where id = $1`,
			e.ID, e.Title, e.Description, e.Category, e.City, e.Venue, e.Address, e.BannerURL, e.StartsAt, e.EndsAt,
			e.Currency, e.WalletID, e.RefundCutoffHours, e.Transferable, e.ResaleCapPercent)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		if _, err := tx.Exec(ctx, `update event_session set starts_at = $2, ends_at = $3
			where event_id = $1 and (select count(*) from event_session where event_id = $1) = 1`, e.ID, e.StartsAt, e.EndsAt); err != nil {
			return err
		}
		return refreshEventTimes(ctx, tx, e.ID)
	})
	if err != nil {
		return domain.Event{}, mapErr(err)
	}
	return r.Get(ctx, e.ID)
}

// refreshEventTimes keeps the event's dates equal to the first start and the last end of its scheduled sessions.
func refreshEventTimes(ctx context.Context, tx pgx.Tx, eventID int64) error {
	_, err := tx.Exec(ctx, `
		update event e set starts_at = x.s, ends_at = x.f, updated_at = now()
		from (select min(starts_at) as s, max(ends_at) as f from event_session where event_id = $1 and status = $2) x
		where e.id = $1 and x.s is not null`, eventID, int16(domain.SessionScheduled))
	return err
}

func (r *EventRepository) Get(ctx context.Context, id int64) (domain.Event, error) {
	e, err := scanEventInto(r.db.QueryRow(ctx, `select `+eventCols("")+` from event where id = $1`, id))
	if err != nil {
		return domain.Event{}, err
	}
	rows, err := r.db.Query(ctx, `select `+typeCols+` from ticket_type where event_id = $1 order by sort_order, id`, id)
	if err != nil {
		return domain.Event{}, mapErr(err)
	}
	defer rows.Close()
	e.TicketTypes = []domain.TicketType{}
	for rows.Next() {
		t, err := scanType(rows)
		if err != nil {
			return domain.Event{}, err
		}
		e.TicketTypes = append(e.TicketTypes, t)
	}
	if err := rows.Err(); err != nil {
		return domain.Event{}, mapErr(err)
	}
	if e.Sessions, err = r.sessions(ctx, id); err != nil {
		return domain.Event{}, err
	}
	e.SoldOut = false
	return e, nil
}

// ---- sessions

const sessionCols = `id, event_id, starts_at, ends_at, coalesce(label, ''), status, coalesce(cancel_reason, '')`

func scanSession(row pgx.Row) (domain.Session, error) {
	var s domain.Session
	err := row.Scan(&s.ID, &s.EventID, &s.StartsAt, &s.EndsAt, &s.Label, &s.Status, &s.CancelReason)
	return s, mapErr(err)
}

func (r *EventRepository) sessions(ctx context.Context, eventID int64) ([]domain.Session, error) {
	rows, err := r.db.Query(ctx, `select `+sessionCols+` from event_session where event_id = $1 order by starts_at, id`, eventID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Session{}
	for rows.Next() {
		s, err := scanSession(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, mapErr(rows.Err())
}

func (r *EventRepository) Session(ctx context.Context, id int64) (domain.Session, error) {
	return scanSession(r.db.QueryRow(ctx, `select `+sessionCols+` from event_session where id = $1`, id))
}

// copyTicketTypes gives a session the ticket types of another, with fresh stock: seated types bring their seats and
// layout, sale windows move by the difference between the two sessions' start times.
func copyTicketTypes(ctx context.Context, tx pgx.Tx, fromType []int64, newEvent, newSession int64, oldStart, newStart time.Time) error {
	for _, oldID := range fromType {
		var newType int64
		err := tx.QueryRow(ctx, `
			insert into ticket_type (event_id, session_id, name, description, price, total, available, min_per_order, max_per_order, max_per_user,
				sale_starts_at, sale_ends_at, seated, active, sort_order, layout)
			select $2, $3, name, description, price, total, total, min_per_order, max_per_order, max_per_user,
				sale_starts_at + ($5::timestamptz - $4::timestamptz), sale_ends_at + ($5::timestamptz - $4::timestamptz), seated, active, sort_order, layout
			from ticket_type where id = $1 returning id`, oldID, newEvent, newSession, oldStart, newStart).Scan(&newType)
		if err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `insert into seat (ticket_type_id, section, row_label, number, x, y, accessible)
			select $1, section, row_label, number, x, y, accessible from seat where ticket_type_id = $2`, newType, oldID); err != nil {
			return err
		}
	}
	return nil
}

func typeIDsOfSession(ctx context.Context, tx pgx.Tx, sessionID int64) ([]int64, error) {
	rows, err := tx.Query(ctx, `select id from ticket_type where session_id = $1 order by sort_order, id`, sessionID)
	if err != nil {
		return nil, err
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
	return out, rows.Err()
}

func (r *EventRepository) CreateSession(ctx context.Context, s domain.Session, copyFrom int64) (domain.Session, error) {
	var id int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `insert into event_session (event_id, starts_at, ends_at, label) values ($1, $2, $3, nullif($4, '')) returning id`,
			s.EventID, s.StartsAt, s.EndsAt, s.Label).Scan(&id); err != nil {
			return err
		}
		if copyFrom > 0 {
			var srcEvent int64
			var srcStart time.Time
			if err := tx.QueryRow(ctx, `select event_id, starts_at from event_session where id = $1`, copyFrom).Scan(&srcEvent, &srcStart); err != nil {
				return err
			}
			if srcEvent != s.EventID {
				return pgx.ErrNoRows
			}
			ids, err := typeIDsOfSession(ctx, tx, copyFrom)
			if err != nil {
				return err
			}
			if err := copyTicketTypes(ctx, tx, ids, s.EventID, id, srcStart, s.StartsAt); err != nil {
				return err
			}
		}
		return refreshEventTimes(ctx, tx, s.EventID)
	})
	if err != nil {
		return domain.Session{}, mapErr(err)
	}
	return r.Session(ctx, id)
}

func (r *EventRepository) UpdateSession(ctx context.Context, s domain.Session) (domain.Session, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `update event_session set starts_at = $3, ends_at = $4, label = nullif($5, '') where id = $1 and event_id = $2`,
			s.ID, s.EventID, s.StartsAt, s.EndsAt, s.Label)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return refreshEventTimes(ctx, tx, s.EventID)
	})
	if err != nil {
		return domain.Session{}, mapErr(err)
	}
	return r.Session(ctx, s.ID)
}

func (r *EventRepository) DeleteSession(ctx context.Context, eventID, sessionID int64) error {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var n int
		if err := tx.QueryRow(ctx, `select count(*) from event_session where event_id = $1`, eventID).Scan(&n); err != nil {
			return err
		}
		var has bool
		if err := tx.QueryRow(ctx, `select exists (select 1 from ticket_type where session_id = $1)`, sessionID).Scan(&has); err != nil {
			return err
		}
		switch {
		case n <= 1:
			return fmt.Errorf("%w: an event needs at least one session", domain.ErrConflict)
		case has:
			return fmt.Errorf("%w: the session still has ticket types; remove or move them first (or cancel the session)", domain.ErrConflict)
		}
		tag, err := tx.Exec(ctx, `delete from event_session where id = $1 and event_id = $2`, sessionID, eventID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return refreshEventTimes(ctx, tx, eventID)
	})
	return mapErr(err)
}

func (r *EventRepository) CancelSession(ctx context.Context, eventID, sessionID int64, reason string) (domain.Session, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `update event_session set status = $3, cancel_reason = nullif($4, '') where id = $1 and event_id = $2 and status = $5`,
			sessionID, eventID, int16(domain.SessionCancelled), reason, int16(domain.SessionScheduled))
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			var one int
			if err := tx.QueryRow(ctx, `select 1 from event_session where id = $1 and event_id = $2`, sessionID, eventID).Scan(&one); err != nil {
				return err
			}
			return domain.ErrConflict // already cancelled
		}
		return refreshEventTimes(ctx, tx, eventID)
	})
	if err != nil {
		return domain.Session{}, mapErr(err)
	}
	return r.Session(ctx, sessionID)
}

func (r *EventRepository) SetStatus(ctx context.Context, id int64, from []domain.EventStatus, to domain.EventStatus, note string) (domain.Event, error) {
	f := make([]int16, len(from))
	for i, s := range from {
		f[i] = int16(s)
	}
	tag, err := r.db.Exec(ctx, `
		update event set status = $3, review_note = nullif($4, ''), version = version + 1, updated_at = now()
		where id = $1 and status = any($2)`, id, f, int16(to), note)
	if err != nil {
		return domain.Event{}, mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		var one int
		if err := r.db.QueryRow(ctx, `select 1 from event where id = $1`, id).Scan(&one); err != nil {
			return domain.Event{}, mapErr(err) // not found
		}
		return domain.Event{}, domain.ErrConflict
	}
	return r.Get(ctx, id)
}

func (r *EventRepository) SetFeatured(ctx context.Context, id int64, featured bool) error {
	tag, err := r.db.Exec(ctx, `update event set featured = $2, updated_at = now() where id = $1`, id, featured)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// likePattern turns user input into a substring pattern that treats % and _ literally.
func likePattern(s string) string {
	s = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
	return "%" + s + "%"
}

func (r *EventRepository) Search(ctx context.Context, f domain.EventFilter) ([]domain.Event, error) {
	var (
		where = []string{"e.status = 3", "n.next_start is not null"}
		args  []any
	)
	arg := func(v any) string {
		args = append(args, v)
		return fmt.Sprintf("$%d", len(args))
	}
	if f.Query != "" {
		p := arg(likePattern(f.Query))
		where = append(where, fmt.Sprintf("(e.title ilike %[1]s or e.venue ilike %[1]s or e.city ilike %[1]s)", p))
	}
	if f.City != "" {
		where = append(where, "lower(e.city) = lower("+arg(f.City)+")")
	}
	if f.Category != "" {
		where = append(where, "e.category = "+arg(f.Category))
	}
	// the date filters ask whether some showtime still to come falls in the range
	if f.From != nil || f.To != nil {
		cond := []string{"s.event_id = e.id", "s.status = 1", "s.ends_at > now()"}
		if f.From != nil {
			cond = append(cond, "s.starts_at >= "+arg(*f.From))
		}
		if f.To != nil {
			cond = append(cond, "s.starts_at < "+arg(*f.To))
		}
		where = append(where, "exists (select 1 from event_session s where "+strings.Join(cond, " and ")+")")
	}
	if f.Featured {
		where = append(where, "e.featured")
	}
	if f.MinPrice != nil || f.MaxPrice != nil {
		cond := []string{"t.event_id = e.id", "t.active"}
		if f.MinPrice != nil {
			cond = append(cond, "t.price >= "+arg(*f.MinPrice))
		}
		if f.MaxPrice != nil {
			cond = append(cond, "t.price <= "+arg(*f.MaxPrice))
		}
		where = append(where, "exists (select 1 from ticket_type t where "+strings.Join(cond, " and ")+")")
	}
	order := "n.next_start, e.id"
	switch f.Sort {
	case "popular":
		order = "coalesce(p.sold, 0) desc, n.next_start, e.id"
	case "newest":
		order = "e.id desc"
	}
	q := `select ` + eventCols("e.") + `, p.min_price, coalesce(p.all_sold, false), n.next_start
		from event e
		left join lateral (
			select min(starts_at) as next_start from event_session where event_id = e.id and status = 1 and ends_at > now()
		) n on true
		left join lateral (
			select min(price) as min_price, bool_and(available = 0) as all_sold, sum(sold) as sold
			from ticket_type t where t.event_id = e.id and t.active
		) p on true
		where ` + strings.Join(where, " and ") + ` order by ` + order +
		fmt.Sprintf(" limit %s offset %s", arg(f.Limit), arg(f.Offset))
	rows, err := r.db.Query(ctx, q, args...)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Event{}
	for rows.Next() {
		var minPrice *int64
		var allSold bool
		var next *time.Time
		e, err := scanEventInto(rows, &minPrice, &allSold, &next)
		if err != nil {
			return nil, err
		}
		e.MinPrice, e.SoldOut, e.NextSession = minPrice, allSold && minPrice != nil, next
		out = append(out, e)
	}
	return out, mapErr(rows.Err())
}

func (r *EventRepository) list(ctx context.Context, where string, arg any, afterID int64, limit int) ([]domain.Event, error) {
	rows, err := r.db.Query(ctx, `select `+eventCols("")+` from event where `+where+` and ($2::bigint = 0 or id < $2) order by id desc limit $3`, arg, afterID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Event{}
	for rows.Next() {
		e, err := scanEventInto(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, e)
	}
	return out, mapErr(rows.Err())
}

func (r *EventRepository) ListByOrganizer(ctx context.Context, organizerID, afterID int64, limit int) ([]domain.Event, error) {
	return r.list(ctx, "organizer_id = $1", organizerID, afterID, limit)
}

func (r *EventRepository) ListByStatus(ctx context.Context, status domain.EventStatus, afterID int64, limit int) ([]domain.Event, error) {
	return r.list(ctx, "status = $1", int16(status), afterID, limit)
}

func (r *EventRepository) Duplicate(ctx context.Context, srcID int64, spec outbound.DuplicateSpec) (domain.Event, error) {
	var newID int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		var series int64
		var srcStart time.Time
		if err := tx.QueryRow(ctx, `select coalesce(series_id, id), starts_at from event where id = $1`, srcID).Scan(&series, &srcStart); err != nil {
			return err
		}
		if _, err := tx.Exec(ctx, `update event set series_id = $2 where id = $1 and series_id is null`, srcID, series); err != nil {
			return err
		}
		err := tx.QueryRow(ctx, `
			insert into event (organizer_id, title, description, category, city, venue, address, banner_url, starts_at, ends_at,
				currency, status, wallet_id, refund_cutoff_hours, transferable, resale_cap_percent, series_id, venue_id, seat_map)
			select organizer_id, coalesce(nullif($2, ''), title), description, category, city, venue, address, banner_url, $3, $4,
				currency, $5, wallet_id, refund_cutoff_hours, transferable, resale_cap_percent, $6, venue_id, seat_map
			from event where id = $1 returning id`, srcID, spec.Title, spec.StartsAt, spec.EndsAt, int16(domain.EventDraft), series).Scan(&newID)
		if err != nil {
			return err
		}
		// Every session moves by the same shift (the first one starts at spec.StartsAt); an event of one session takes
		// spec.EndsAt as its end, so the copy may be longer or shorter.
		delta := spec.StartsAt.Sub(srcStart)
		rows, err := tx.Query(ctx, `select id, starts_at, ends_at, coalesce(label, ''), status, coalesce(cancel_reason, '') from event_session where event_id = $1 order by starts_at, id`, srcID)
		if err != nil {
			return err
		}
		type old struct {
			id         int64
			start, end time.Time
			label      string
		}
		var sessions []old
		for rows.Next() {
			var o old
			var st domain.SessionStatus
			var reason string
			if err := rows.Scan(&o.id, &o.start, &o.end, &o.label, &st, &reason); err != nil {
				rows.Close()
				return err
			}
			if st == domain.SessionScheduled { // a cancelled showtime is not repeated
				sessions = append(sessions, o)
			}
		}
		rows.Close()
		if err := rows.Err(); err != nil {
			return err
		}
		if len(sessions) == 0 {
			return fmt.Errorf("%w: the event has no scheduled session to copy", domain.ErrConflict)
		}
		for _, o := range sessions {
			start, end := o.start.Add(delta), o.end.Add(delta)
			if len(sessions) == 1 {
				end = spec.EndsAt
			}
			var newSession int64
			if err := tx.QueryRow(ctx, `insert into event_session (event_id, starts_at, ends_at, label) values ($1, $2, $3, nullif($4, '')) returning id`,
				newID, start, end, o.label).Scan(&newSession); err != nil {
				return err
			}
			ids, err := typeIDsOfSession(ctx, tx, o.id)
			if err != nil {
				return err
			}
			if err := copyTicketTypes(ctx, tx, ids, newID, newSession, o.start, start); err != nil {
				return err
			}
		}
		if err := refreshEventTimes(ctx, tx, newID); err != nil {
			return err
		}
		if spec.CopyPromotions {
			if _, err := tx.Exec(ctx, `
				insert into promotion (event_id, code, kind, value, max_uses, min_tickets, valid_from, valid_to, active)
				select $2, code, kind, value, max_uses, min_tickets, valid_from + ($3::timestamptz - $4::timestamptz), valid_to + ($3::timestamptz - $4::timestamptz), active
				from promotion where event_id = $1`, srcID, newID, spec.StartsAt, srcStart); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		return domain.Event{}, mapErr(err)
	}
	return r.Get(ctx, newID)
}

func (r *EventRepository) Series(ctx context.Context, eventID int64) ([]domain.Event, error) {
	rows, err := r.db.Query(ctx, `select `+eventCols("e.")+`, p.min_price, coalesce(p.all_sold, false)
		from event e
		left join lateral (
			select min(price) as min_price, bool_and(available = 0) as all_sold from ticket_type t where t.event_id = e.id and t.active
		) p on true
		where e.series_id = (select series_id from event where id = $1) and e.status = $2 and e.ends_at > now()
		order by e.starts_at, e.id`, eventID, int16(domain.EventPublished))
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Event{}
	for rows.Next() {
		var minPrice *int64
		var allSold bool
		e, err := scanEventInto(rows, &minPrice, &allSold)
		if err != nil {
			return nil, err
		}
		e.MinPrice, e.SoldOut = minPrice, allSold && minPrice != nil
		out = append(out, e)
	}
	return out, mapErr(rows.Err())
}

// ---- ticket types

func (r *EventRepository) CreateTicketType(ctx context.Context, t domain.TicketType) (domain.TicketType, error) {
	out, err := scanType(r.db.QueryRow(ctx, `
		insert into ticket_type (event_id, name, description, price, total, available, min_per_order, max_per_order, max_per_user,
			sale_starts_at, sale_ends_at, seated, active, sort_order, session_id)
		values ($1, $2, nullif($3, ''), $4, $5, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14)
		returning `+typeCols,
		t.EventID, t.Name, t.Description, t.Price, t.Total, t.MinPerOrder, t.MaxPerOrder, t.MaxPerUser,
		t.SaleStartsAt, t.SaleEndsAt, t.Seated, t.Active, t.SortOrder, t.SessionID))
	if errors.Is(err, domain.ErrConflict) {
		return out, fmt.Errorf("%w: this showtime already has a ticket type named %q", domain.ErrConflict, t.Name)
	}
	return out, err
}

func (r *EventRepository) UpdateTicketType(ctx context.Context, t domain.TicketType) (domain.TicketType, error) {
	// available moves with total in the same statement (the right-hand sides see the old row), and the table's CHECKs
	// refuse a total that would leave fewer than zero tickets available.
	return scanType(r.db.QueryRow(ctx, `
		update ticket_type set name = $3, description = nullif($4, ''), price = $5, available = available + ($6 - total), total = $6,
			min_per_order = $7, max_per_order = $8, max_per_user = $9, sale_starts_at = $10, sale_ends_at = $11,
			active = $12, sort_order = $13
		where id = $1 and event_id = $2
		returning `+typeCols,
		t.ID, t.EventID, t.Name, t.Description, t.Price, t.Total, t.MinPerOrder, t.MaxPerOrder, t.MaxPerUser,
		t.SaleStartsAt, t.SaleEndsAt, t.Active, t.SortOrder))
}

func (r *EventRepository) TicketType(ctx context.Context, id int64) (domain.TicketType, error) {
	return scanType(r.db.QueryRow(ctx, `select `+typeCols+` from ticket_type where id = $1`, id))
}

func (r *EventRepository) AddSeats(ctx context.Context, typeID int64, seats []domain.Seat) (int, error) {
	sections, rowsL, numbers := make([]string, len(seats)), make([]string, len(seats)), make([]int32, len(seats))
	xs, ys, acc := make([]float64, len(seats)), make([]float64, len(seats)), make([]bool, len(seats))
	for i, s := range seats {
		sections[i], rowsL[i], numbers[i], acc[i] = s.Section, s.Row, int32(s.Number), s.Accessible
		if s.X != nil && s.Y != nil {
			xs[i], ys[i] = *s.X, *s.Y
		}
	}
	var added int
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			insert into seat (ticket_type_id, section, row_label, number, x, y, accessible)
			select $1, s, r, n, x, y, a from unnest($2::text[], $3::text[], $4::int[], $5::float8[], $6::float8[], $7::bool[]) as u(s, r, n, x, y, a)
			on conflict do nothing`, typeID, sections, rowsL, numbers, xs, ys, acc)
		if err != nil {
			return err
		}
		added = int(tag.RowsAffected())
		// every seat is one more ticket on sale
		_, err = tx.Exec(ctx, `update ticket_type set total = total + $2, available = available + $2 where id = $1`, typeID, added)
		return err
	})
	return added, mapErr(err)
}

func (r *EventRepository) Seats(ctx context.Context, typeID int64) ([]domain.Seat, error) {
	rows, err := r.db.Query(ctx, `select id, ticket_type_id, section, row_label, number, status, x, y, accessible from seat
		where ticket_type_id = $1 order by section, length(row_label), row_label, number`, typeID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Seat{}
	for rows.Next() {
		var s domain.Seat
		if err := rows.Scan(&s.ID, &s.TicketTypeID, &s.Section, &s.Row, &s.Number, &s.Status, &s.X, &s.Y, &s.Accessible); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, mapErr(rows.Err())
}

func (r *EventRepository) SetSeatMap(ctx context.Context, typeID int64, l domain.SeatMapLayout, seats []domain.SeatPosition) (int, error) {
	raw, err := json.Marshal(l)
	if err != nil {
		return 0, err
	}
	secs, rowsL, numbers, xs, ys := make([]string, len(seats)), make([]string, len(seats)), make([]int32, len(seats)), make([]float64, len(seats)), make([]float64, len(seats))
	for i, s := range seats {
		secs[i], rowsL[i], numbers[i], xs[i], ys[i] = s.Section, s.Row, int32(s.Number), s.X, s.Y
	}
	var placed int
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		if _, err := tx.Exec(ctx, `update ticket_type set layout = $2 where id = $1`, typeID, raw); err != nil {
			return err
		}
		// a seat is found by row and number; when a section is named it must match too
		tag, err := tx.Exec(ctx, `update seat s set x = u.x, y = u.y
			from unnest($2::text[], $3::text[], $4::int[], $5::float8[], $6::float8[]) as u(sec, r, n, x, y)
			where s.ticket_type_id = $1 and s.row_label = u.r and s.number = u.n and (u.sec = '' or s.section = u.sec)`, typeID, secs, rowsL, numbers, xs, ys)
		placed = int(tag.RowsAffected())
		return err
	})
	return placed, mapErr(err)
}

func (r *EventRepository) SeatLayout(ctx context.Context, typeID int64) (domain.SeatMapLayout, error) {
	var raw []byte
	if err := r.db.QueryRow(ctx, `select layout from ticket_type where id = $1`, typeID).Scan(&raw); err != nil {
		return domain.SeatMapLayout{}, mapErr(err)
	}
	var l domain.SeatMapLayout
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &l); err != nil {
			return domain.SeatMapLayout{}, err
		}
	}
	return l, nil
}

// ---- promotions

const promoCols = `id, event_id, code, kind, value, max_uses, used, min_tickets, valid_from, valid_to, active`

func scanPromo(row pgx.Row) (domain.Promotion, error) {
	var p domain.Promotion
	err := row.Scan(&p.ID, &p.EventID, &p.Code, &p.Kind, &p.Value, &p.MaxUses, &p.Used, &p.MinTickets, &p.ValidFrom, &p.ValidTo, &p.Active)
	return p, mapErr(err)
}

func (r *EventRepository) Promotions(ctx context.Context, eventID int64) ([]domain.Promotion, error) {
	rows, err := r.db.Query(ctx, `select `+promoCols+` from promotion where event_id = $1 order by id`, eventID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Promotion{}
	for rows.Next() {
		p, err := scanPromo(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, mapErr(rows.Err())
}

func (r *EventRepository) PromotionByCode(ctx context.Context, eventID int64, code string) (domain.Promotion, error) {
	return scanPromo(r.db.QueryRow(ctx, `select `+promoCols+` from promotion where event_id = $1 and code = $2`, eventID, code))
}

func (r *EventRepository) CreatePromotion(ctx context.Context, p domain.Promotion) (domain.Promotion, error) {
	return scanPromo(r.db.QueryRow(ctx, `
		insert into promotion (event_id, code, kind, value, max_uses, min_tickets, valid_from, valid_to, active)
		values ($1, $2, $3, $4, $5, $6, $7, $8, $9) returning `+promoCols,
		p.EventID, p.Code, p.Kind, p.Value, p.MaxUses, p.MinTickets, p.ValidFrom, p.ValidTo, p.Active))
}

func (r *EventRepository) UpdatePromotion(ctx context.Context, p domain.Promotion) (domain.Promotion, error) {
	return scanPromo(r.db.QueryRow(ctx, `
		update promotion set code = $3, kind = $4, value = $5, max_uses = $6, min_tickets = $7, valid_from = $8, valid_to = $9, active = $10
		where id = $1 and event_id = $2 returning `+promoCols,
		p.ID, p.EventID, p.Code, p.Kind, p.Value, p.MaxUses, p.MinTickets, p.ValidFrom, p.ValidTo, p.Active))
}

// ---- reporting

func (r *EventRepository) Report(ctx context.Context, eventID int64) (domain.EventReport, error) {
	rep := domain.EventReport{EventID: eventID}
	err := r.db.QueryRow(ctx, `
		select count(*) filter (where status = 2), coalesce(sum(total) filter (where status = 2), 0),
			coalesce(sum(refund_amount) filter (where status = 5), 0)
		from ticket_order where event_id = $1`, eventID).Scan(&rep.Orders, &rep.Revenue, &rep.Refunded)
	if err != nil {
		return rep, mapErr(err)
	}
	if err := r.db.QueryRow(ctx, `select count(*) from ticket where event_id = $1 and status = $2`, eventID, int16(domain.TicketUsed)).
		Scan(&rep.CheckedIn); err != nil {
		return rep, mapErr(err)
	}
	if err := r.db.QueryRow(ctx, `select coalesce(sum(i.quantity), 0) from order_item i join ticket_order o on o.id = i.order_id
		where o.event_id = $1 and o.status = 2 and o.invited_by is not null`, eventID).Scan(&rep.Invited); err != nil {
		return rep, mapErr(err)
	}
	drows, err := r.db.Query(ctx, `
		select date_trunc('day', o.paid_at at time zone 'UTC') as day, count(*), coalesce(sum(q.n), 0), coalesce(sum(o.total), 0)
		from ticket_order o join lateral (select sum(quantity) as n from order_item where order_id = o.id) q on true
		where o.event_id = $1 and o.status = 2 and o.paid_at is not null and o.invited_by is null group by 1 order by 1`, eventID)
	if err != nil {
		return rep, mapErr(err)
	}
	defer drows.Close()
	rep.Daily = []domain.DailySales{}
	for drows.Next() {
		var d domain.DailySales
		if err := drows.Scan(&d.Day, &d.Orders, &d.Tickets, &d.Revenue); err != nil {
			return rep, err
		}
		rep.Daily = append(rep.Daily, d)
	}
	if err := drows.Err(); err != nil {
		return rep, mapErr(err)
	}
	rows, err := r.db.Query(ctx, `
		select t.id, t.name, t.total, t.sold, t.total - t.available - t.sold,
			coalesce((select sum(oi.quantity * oi.unit_price) from order_item oi join ticket_order o on o.id = oi.order_id
				where oi.ticket_type_id = t.id and o.status = 2), 0)
		from ticket_type t where t.event_id = $1 order by t.sort_order, t.id`, eventID)
	if err != nil {
		return rep, mapErr(err)
	}
	defer rows.Close()
	rep.ByType = []domain.TypeSales{}
	for rows.Next() {
		var s domain.TypeSales
		if err := rows.Scan(&s.TicketTypeID, &s.Name, &s.Total, &s.Sold, &s.Held, &s.Revenue); err != nil {
			return rep, err
		}
		rep.TicketsSold += s.Sold
		rep.ByType = append(rep.ByType, s)
	}
	return rep, mapErr(rows.Err())
}

// ---- venue maps

func (r *EventRepository) ApplyVenueMap(ctx context.Context, eventID, venueID int64, drawing domain.SeatMapLayout, groups []domain.SeatGroup, replace bool) (int, error) {
	raw, err := json.Marshal(drawing)
	if err != nil {
		return 0, err
	}
	total := 0
	err = pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		for _, g := range groups {
			var owner int64
			var seated bool
			var sold, avail, tot int
			err := tx.QueryRow(ctx, `select event_id, seated, sold, available, total from ticket_type where id = $1 for update`, g.TicketTypeID).Scan(&owner, &seated, &sold, &avail, &tot)
			if err != nil {
				return err
			}
			if owner != eventID {
				return pgx.ErrNoRows
			}
			if !seated {
				return fmt.Errorf("%w: ticket type %d is general admission and has no seats", domain.ErrInvalid, g.TicketTypeID)
			}
			var have int
			if err := tx.QueryRow(ctx, `select count(*) from seat where ticket_type_id = $1`, g.TicketTypeID).Scan(&have); err != nil {
				return err
			}
			if have > 0 {
				if !replace {
					return fmt.Errorf("%w: ticket type %d already has seats; pass replace to swap them", domain.ErrConflict, g.TicketTypeID)
				}
				if sold > 0 || tot-avail-sold > 0 {
					return fmt.Errorf("%w: ticket type %d has tickets sold or held, so its seats cannot be replaced", domain.ErrConflict, g.TicketTypeID)
				}
				if _, err := tx.Exec(ctx, `delete from seat where ticket_type_id = $1`, g.TicketTypeID); err != nil {
					return err
				}
				if _, err := tx.Exec(ctx, `update ticket_type set total = 0, available = 0 where id = $1`, g.TicketTypeID); err != nil {
					return err
				}
			}
			n := len(g.Seats)
			secs, rowsL, nums, xs, ys, acc := make([]string, n), make([]string, n), make([]int32, n), make([]float64, n), make([]float64, n), make([]bool, n)
			for i, s := range g.Seats {
				secs[i], rowsL[i], nums[i], xs[i], ys[i], acc[i] = s.Section, s.Row, int32(s.Number), s.X, s.Y, s.Accessible
			}
			tag, err := tx.Exec(ctx, `
				insert into seat (ticket_type_id, section, row_label, number, x, y, accessible)
				select $1, s, r, n, x, y, a from unnest($2::text[], $3::text[], $4::int[], $5::float8[], $6::float8[], $7::bool[]) as u(s, r, n, x, y, a)
				on conflict do nothing`, g.TicketTypeID, secs, rowsL, nums, xs, ys, acc)
			if err != nil {
				return err
			}
			added := int(tag.RowsAffected())
			if _, err := tx.Exec(ctx, `update ticket_type set total = total + $2, available = available + $2 where id = $1`, g.TicketTypeID, added); err != nil {
				return err
			}
			total += added
		}
		tag, err := tx.Exec(ctx, `update event set venue_id = $2, seat_map = $3, version = version + 1, updated_at = now() where id = $1`, eventID, venueID, raw)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return nil
	})
	return total, mapErr(err)
}

func (r *EventRepository) EventSeatMap(ctx context.Context, eventID int64) (domain.EventSeatMap, error) {
	var out domain.EventSeatMap
	var raw []byte
	if err := r.db.QueryRow(ctx, `select seat_map from event where id = $1`, eventID).Scan(&raw); err != nil {
		return out, mapErr(err)
	}
	if len(raw) > 0 {
		if err := json.Unmarshal(raw, &out.Layout); err != nil {
			return out, err
		}
	}
	trows, err := r.db.Query(ctx, `select `+typeCols+` from ticket_type where event_id = $1 and seated and active order by sort_order, id`, eventID)
	if err != nil {
		return out, mapErr(err)
	}
	defer trows.Close()
	for trows.Next() {
		t, err := scanType(trows)
		if err != nil {
			return out, err
		}
		out.Types = append(out.Types, t)
	}
	if err := trows.Err(); err != nil {
		return out, mapErr(err)
	}
	rows, err := r.db.Query(ctx, `select s.id, s.ticket_type_id, s.section, s.row_label, s.number, s.status, s.x, s.y, s.accessible
		from seat s join ticket_type t on t.id = s.ticket_type_id where t.event_id = $1 and t.seated and t.active
		order by s.ticket_type_id, s.section, length(s.row_label), s.row_label, s.number`, eventID)
	if err != nil {
		return out, mapErr(err)
	}
	defer rows.Close()
	out.Seats = []domain.Seat{}
	for rows.Next() {
		var s domain.Seat
		if err := rows.Scan(&s.ID, &s.TicketTypeID, &s.Section, &s.Row, &s.Number, &s.Status, &s.X, &s.Y, &s.Accessible); err != nil {
			return out, err
		}
		out.Seats = append(out.Seats, s)
	}
	return out, mapErr(rows.Err())
}
