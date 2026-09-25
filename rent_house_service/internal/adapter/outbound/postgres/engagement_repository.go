package postgres

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type ReviewRepository struct{ db *pgxpool.Pool }

var _ outbound.ReviewRepository = (*ReviewRepository)(nil)

func NewReviewRepository(db *pgxpool.Pool) *ReviewRepository { return &ReviewRepository{db: db} }

func (r *ReviewRepository) Create(ctx context.Context, rv domain.Review) (domain.Review, error) {
	var q string
	var stay int64
	if rv.BookingID != 0 {
		stay = rv.BookingID
		q = `insert into review (homestay_id, user_id, booking_id, rating, comment)
			select b.homestay_id, b.user_id, b.id, $4, nullif($5, '') from booking b
			where b.id = $1 and b.user_id = $2 and b.homestay_id = $3 and b.status = 1 and b.checkout_date <= current_date
			returning id, created_at`
	} else {
		stay = rv.LeaseID
		q = `insert into review (homestay_id, user_id, lease_id, rating, comment)
			select l.homestay_id, l.user_id, l.id, $4, nullif($5, '') from lease l
			where l.id = $1 and l.user_id = $2 and l.homestay_id = $3 and l.status in (2, 3) and l.start_date <= current_date
			returning id, created_at`
	}
	err := r.db.QueryRow(ctx, q, stay, rv.UserID, rv.HomestayID, rv.Rating, rv.Comment).Scan(&rv.ID, &rv.CreatedAt)
	if err == pgx.ErrNoRows {
		return domain.Review{}, domain.ErrForbidden
	}
	return rv, mapErr(err)
}

func (r *ReviewRepository) ListByHomestay(ctx context.Context, homestayID int64, limit, offset int) ([]domain.Review, error) {
	rows, err := r.db.Query(ctx, `
		select id, homestay_id, user_id, coalesce(booking_id, 0), coalesce(lease_id, 0), rating, coalesce(comment, ''), created_at
		from review where homestay_id = $1 order by id desc limit $2 offset $3`, homestayID, limit, offset)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Review{}
	for rows.Next() {
		var rv domain.Review
		if err := rows.Scan(&rv.ID, &rv.HomestayID, &rv.UserID, &rv.BookingID, &rv.LeaseID, &rv.Rating, &rv.Comment, &rv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, rv)
	}
	return out, mapErr(rows.Err())
}

type WishlistRepository struct {
	db        *pgxpool.Pool
	homestays *HomestayRepository
}

var _ outbound.WishlistRepository = (*WishlistRepository)(nil)

func NewWishlistRepository(db *pgxpool.Pool) *WishlistRepository {
	return &WishlistRepository{db: db, homestays: NewHomestayRepository(db)}
}

func (r *WishlistRepository) Add(ctx context.Context, userID, homestayID int64) error {
	_, err := r.db.Exec(ctx, `insert into wishlist (user_id, homestay_id) values ($1, $2) on conflict do nothing`, userID, homestayID)
	if mapped := mapErr(err); errors.Is(mapped, domain.ErrInvalid) {
		return domain.ErrNotFound
	} else {
		return mapped
	}
}

func (r *WishlistRepository) Remove(ctx context.Context, userID, homestayID int64) error {
	_, err := r.db.Exec(ctx, `delete from wishlist where user_id = $1 and homestay_id = $2`, userID, homestayID)
	return mapErr(err)
}

func (r *WishlistRepository) List(ctx context.Context, userID int64, limit, offset int) ([]domain.Homestay, error) {
	rows, err := r.db.Query(ctx, `select `+homestayCols+homestayFrom+`
		join wishlist w on w.homestay_id = h.id where w.user_id = $1 and h.status = $2
		order by w.created_at desc limit $3 offset $4`, userID, domain.HomestayActive, limit, offset)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Homestay{}
	for rows.Next() {
		h, err := scanHomestay(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, mapErr(err)
	}
	return out, r.homestays.attachRates(ctx, out)
}
