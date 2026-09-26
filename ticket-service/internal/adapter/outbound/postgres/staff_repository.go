package postgres

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type StaffRepository struct{ db *pgxpool.Pool }

var _ outbound.StaffRepository = (*StaffRepository)(nil)

func NewStaffRepository(db *pgxpool.Pool) *StaffRepository { return &StaffRepository{db: db} }

func (r *StaffRepository) Add(ctx context.Context, eventID, userID, addedBy int64) error {
	_, err := r.db.Exec(ctx, `insert into event_staff (event_id, user_id, added_by) values ($1, $2, $3) on conflict do nothing`, eventID, userID, addedBy)
	return mapErr(err)
}

func (r *StaffRepository) Remove(ctx context.Context, eventID, userID int64) error {
	_, err := r.db.Exec(ctx, `delete from event_staff where event_id = $1 and user_id = $2`, eventID, userID)
	return mapErr(err)
}

func (r *StaffRepository) List(ctx context.Context, eventID int64) ([]domain.StaffMember, error) {
	rows, err := r.db.Query(ctx, `select event_id, user_id, added_by, created_at from event_staff where event_id = $1 order by created_at`, eventID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.StaffMember{}
	for rows.Next() {
		var m domain.StaffMember
		if err := rows.Scan(&m.EventID, &m.UserID, &m.AddedBy, &m.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, mapErr(rows.Err())
}

func (r *StaffRepository) Has(ctx context.Context, eventID, userID int64) (bool, error) {
	var ok bool
	err := r.db.QueryRow(ctx, `select exists (select 1 from event_staff where event_id = $1 and user_id = $2)`, eventID, userID).Scan(&ok)
	return ok, mapErr(err)
}

func (r *StaffRepository) EventsFor(ctx context.Context, userID int64) ([]domain.Event, error) {
	rows, err := r.db.Query(ctx, `select `+eventCols("e.")+` from event_staff s join event e on e.id = s.event_id
		where s.user_id = $1 and e.status = $2 and e.ends_at > now() order by e.starts_at`, userID, domain.EventPublished)
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

// ---- waiting lists

type WaitlistRepository struct{ db *pgxpool.Pool }

var _ outbound.WaitlistRepository = (*WaitlistRepository)(nil)

func NewWaitlistRepository(db *pgxpool.Pool) *WaitlistRepository { return &WaitlistRepository{db: db} }

func (r *WaitlistRepository) Join(ctx context.Context, typeID, userID int64) error {
	tag, err := r.db.Exec(ctx, `insert into waitlist (ticket_type_id, user_id) values ($1, $2) on conflict do nothing`, typeID, userID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrConflict
	}
	return nil
}

func (r *WaitlistRepository) Leave(ctx context.Context, typeID, userID int64) error {
	_, err := r.db.Exec(ctx, `delete from waitlist where ticket_type_id = $1 and user_id = $2`, typeID, userID)
	return mapErr(err)
}

func (r *WaitlistRepository) Mine(ctx context.Context, userID int64) ([]domain.WaitlistEntry, error) {
	rows, err := r.db.Query(ctx, `select w.ticket_type_id, t.event_id, e.title, t.name, w.created_at
		from waitlist w join ticket_type t on t.id = w.ticket_type_id join event e on e.id = t.event_id
		where w.user_id = $1 order by w.created_at desc`, userID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.WaitlistEntry{}
	for rows.Next() {
		var w domain.WaitlistEntry
		if err := rows.Scan(&w.TicketTypeID, &w.EventID, &w.EventTitle, &w.TypeName, &w.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, mapErr(rows.Err())
}

func (r *WaitlistRepository) Pop(ctx context.Context, typeID int64, n int) ([]int64, error) {
	rows, err := r.db.Query(ctx, `
		delete from waitlist where (ticket_type_id, user_id) in (
			select ticket_type_id, user_id from waitlist where ticket_type_id = $1 order by created_at limit $2 for update skip locked)
		returning user_id`, typeID, n)
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
