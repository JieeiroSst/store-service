package postgres

import (
	"context"
	"encoding/json"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type WishlistRepository struct{ db *pgxpool.Pool }

var _ outbound.WishlistRepository = (*WishlistRepository)(nil)

func NewWishlistRepository(db *pgxpool.Pool) *WishlistRepository { return &WishlistRepository{db: db} }

func (r *WishlistRepository) Add(ctx context.Context, userID, eventID int64) error {
	_, err := r.db.Exec(ctx, `insert into wishlist (user_id, event_id) values ($1, $2) on conflict do nothing`, userID, eventID)
	return mapErr(err)
}

func (r *WishlistRepository) Remove(ctx context.Context, userID, eventID int64) error {
	_, err := r.db.Exec(ctx, `delete from wishlist where user_id = $1 and event_id = $2`, userID, eventID)
	return mapErr(err)
}

func (r *WishlistRepository) List(ctx context.Context, userID int64, limit, offset int) ([]domain.Event, error) {
	rows, err := r.db.Query(ctx, `select `+eventCols("e.")+` from wishlist w join event e on e.id = w.event_id
		where w.user_id = $1 and e.status = $2 order by w.created_at desc, e.id desc limit $3 offset $4`,
		userID, domain.EventPublished, limit, offset)
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

// ---- notifications

type NotificationRepository struct{ db *pgxpool.Pool }

var _ outbound.NotificationRepository = (*NotificationRepository)(nil)

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

const notificationCols = `id, user_id, kind, title, body, coalesce(order_id, 0), coalesce(event_id, 0), created_at, read_at`

func scanNotifications(rows interface {
	Next() bool
	Scan(...any) error
	Err() error
	Close()
}) ([]domain.Notification, error) {
	defer rows.Close()
	out := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Kind, &n.Title, &n.Body, &n.OrderID, &n.EventID, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, mapErr(rows.Err())
}

func (r *NotificationRepository) Add(ctx context.Context, n domain.Notification) error {
	var to, tmpl *string
	var data []byte
	if n.Email != nil {
		to, tmpl = &n.Email.To, &n.Email.Template
		data, _ = json.Marshal(n.Email.Data)
	}
	_, err := r.db.Exec(ctx, `insert into notification (user_id, kind, title, body, order_id, event_id, email_to, email_template, email_data)
		values ($1, $2, $3, $4, nullif($5, 0), nullif($6, 0), $7, $8, $9) on conflict do nothing`,
		n.UserID, n.Kind, n.Title, n.Body, n.OrderID, n.EventID, to, tmpl, data)
	return mapErr(err)
}

func (r *NotificationRepository) ClaimUnemailed(ctx context.Context, limit int, lease time.Duration) ([]domain.Notification, error) {
	rows, err := r.db.Query(ctx, `
		update notification set email_attempts = email_attempts + 1, email_claimed_until = now() + make_interval(secs => $2)
		where id in (
			select id from notification
			where email_template is not null and emailed_at is null and email_attempts < 5
				and (email_claimed_until is null or email_claimed_until < now())
			order by id limit $1 for update skip locked)
		returning id, user_id, kind, title, body, coalesce(order_id, 0), coalesce(event_id, 0), created_at, read_at, email_to, email_template, email_data`,
		limit, lease.Seconds())
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		var to, tmpl string
		var data []byte
		if err := rows.Scan(&n.ID, &n.UserID, &n.Kind, &n.Title, &n.Body, &n.OrderID, &n.EventID, &n.CreatedAt, &n.ReadAt, &to, &tmpl, &data); err != nil {
			return nil, err
		}
		n.Email = &domain.Email{To: to, Template: tmpl, Data: map[string]string{}}
		_ = json.Unmarshal(data, &n.Email.Data)
		out = append(out, n)
	}
	return out, mapErr(rows.Err())
}

func (r *NotificationRepository) MarkEmailed(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `update notification set emailed_at = now() where id = $1`, id)
	return mapErr(err)
}

func (r *NotificationRepository) List(ctx context.Context, userID int64, unreadOnly bool, afterID int64, limit int) ([]domain.Notification, error) {
	rows, err := r.db.Query(ctx, `select `+notificationCols+` from notification
		where user_id = $1 and (not $2 or read_at is null) and ($3::bigint = 0 or id < $3) order by id desc limit $4`,
		userID, unreadOnly, afterID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	return scanNotifications(rows)
}

func (r *NotificationRepository) UnreadCount(ctx context.Context, userID int64) (int, error) {
	var n int
	err := r.db.QueryRow(ctx, `select count(*) from notification where user_id = $1 and read_at is null`, userID).Scan(&n)
	return n, mapErr(err)
}

func (r *NotificationRepository) MarkRead(ctx context.Context, userID, id int64) error {
	tag, err := r.db.Exec(ctx, `update notification set read_at = coalesce(read_at, now()) where id = $1 and user_id = $2`, id, userID)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *NotificationRepository) MarkAllRead(ctx context.Context, userID int64) (int, error) {
	tag, err := r.db.Exec(ctx, `update notification set read_at = now() where user_id = $1 and read_at is null`, userID)
	return int(tag.RowsAffected()), mapErr(err)
}

func (r *NotificationRepository) ClaimUnpushed(ctx context.Context, limit int, lease time.Duration) ([]domain.Notification, error) {
	rows, err := r.db.Query(ctx, `
		update notification set push_attempts = push_attempts + 1, claimed_until = now() + make_interval(secs => $2)
		where id in (
			select id from notification
			where pushed_at is null and push_attempts < 5 and (claimed_until is null or claimed_until < now())
			order by id limit $1 for update skip locked)
		returning `+notificationCols, limit, lease.Seconds())
	if err != nil {
		return nil, mapErr(err)
	}
	return scanNotifications(rows)
}

func (r *NotificationRepository) MarkPushed(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `update notification set pushed_at = now() where id = $1`, id)
	return mapErr(err)
}
