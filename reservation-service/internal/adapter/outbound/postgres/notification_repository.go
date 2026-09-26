package postgres

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type NotificationRepository struct{ db *pgxpool.Pool }

var _ outbound.NotificationRepository = (*NotificationRepository)(nil)

func NewNotificationRepository(db *pgxpool.Pool) *NotificationRepository {
	return &NotificationRepository{db: db}
}

const notificationCols = `id, user_id, kind, title, body, coalesce(reservation_id, 0), coalesce(hotel_id, 0), created_at, read_at`

func (r *NotificationRepository) Add(ctx context.Context, n domain.Notification) error {
	_, err := r.db.Exec(ctx, `insert into notification (user_id, kind, title, body, reservation_id, hotel_id)
		values ($1, $2, $3, $4, nullif($5, 0), nullif($6, 0)) on conflict do nothing`,
		n.UserID, n.Kind, n.Title, n.Body, n.ReservationID, n.HotelID)
	return mapErr(err)
}

func (r *NotificationRepository) List(ctx context.Context, userID int64, unreadOnly bool, afterID int64, limit int) ([]domain.Notification, error) {
	rows, err := r.db.Query(ctx, `select `+notificationCols+` from notification
		where user_id = $1 and (not $2 or read_at is null) and ($3 = 0 or id < $3)
		order by id desc limit $4`, userID, unreadOnly, afterID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Kind, &n.Title, &n.Body, &n.ReservationID, &n.HotelID, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, mapErr(rows.Err())
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
	defer rows.Close()
	out := []domain.Notification{}
	for rows.Next() {
		var n domain.Notification
		if err := rows.Scan(&n.ID, &n.UserID, &n.Kind, &n.Title, &n.Body, &n.ReservationID, &n.HotelID, &n.CreatedAt, &n.ReadAt); err != nil {
			return nil, err
		}
		out = append(out, n)
	}
	return out, mapErr(rows.Err())
}

func (r *NotificationRepository) MarkPushed(ctx context.Context, id int64) error {
	_, err := r.db.Exec(ctx, `update notification set pushed_at = now() where id = $1`, id)
	return mapErr(err)
}
