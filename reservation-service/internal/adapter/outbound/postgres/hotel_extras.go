package postgres

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

// ---- promotions

const promotionCols = `id, hotel_id, code, coalesce(description, ''), coalesce(percent_off, 0), coalesce(amount_off::text, ''),
	min_nights, coalesce(max_uses, 0), used_count, valid_from, valid_to, active`

func scanPromotion(row pgx.Row) (domain.Promotion, error) {
	var p domain.Promotion
	err := row.Scan(&p.ID, &p.HotelID, &p.Code, &p.Description, &p.PercentOff, &p.AmountOff, &p.MinNights, &p.MaxUses,
		&p.UsedCount, &p.ValidFrom, &p.ValidTo, &p.Active)
	return p, mapErr(err)
}

func (r *HotelRepository) SavePromotion(ctx context.Context, p domain.Promotion) (domain.Promotion, error) {
	if p.ID == 0 {
		return scanPromotion(r.db.QueryRow(ctx, `
			insert into promotion (hotel_id, code, description, percent_off, amount_off, min_nights, max_uses, valid_from, valid_to, active)
			values ($1, $2, nullif($3, ''), nullif($4, 0), nullif($5, '')::numeric, $6, nullif($7, 0), $8, $9, $10)
			returning `+promotionCols,
			p.HotelID, p.Code, p.Description, p.PercentOff, p.AmountOff, p.MinNights, p.MaxUses, p.ValidFrom, p.ValidTo, p.Active))
	}
	// The number of uses is never edited by hand: it is the count of guests who took the code.
	return scanPromotion(r.db.QueryRow(ctx, `
		update promotion set code = $3, description = nullif($4, ''), percent_off = nullif($5, 0), amount_off = nullif($6, '')::numeric,
			min_nights = $7, max_uses = nullif($8, 0), valid_from = $9, valid_to = $10, active = $11
		where id = $1 and hotel_id = $2 returning `+promotionCols,
		p.ID, p.HotelID, p.Code, p.Description, p.PercentOff, p.AmountOff, p.MinNights, p.MaxUses, p.ValidFrom, p.ValidTo, p.Active))
}

func (r *HotelRepository) Promotions(ctx context.Context, hotelID int64) ([]domain.Promotion, error) {
	rows, err := r.db.Query(ctx, `select `+promotionCols+` from promotion where hotel_id = $1 order by id`, hotelID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Promotion{}
	for rows.Next() {
		p, err := scanPromotion(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, mapErr(rows.Err())
}

func (r *HotelRepository) PromotionByCode(ctx context.Context, hotelID int64, code string) (domain.Promotion, error) {
	return scanPromotion(r.db.QueryRow(ctx, `select `+promotionCols+` from promotion where hotel_id = $1 and lower(code) = lower($2)`, hotelID, code))
}

// ---- report

func (r *HotelRepository) Report(ctx context.Context, hotelID int64, from, to time.Time) (domain.HotelReport, error) {
	rep := domain.HotelReport{From: from, To: to, StatusCounts: map[domain.ReservationStatus]int{}}

	rows, err := r.db.Query(ctx, `select date, sum(total_inventory)::int, sum(total_reserved)::int from room_type_inventory
		where hotel_id = $1 and date >= $2 and date < $3 group by date order by date`, hotelID, from, to)
	if err != nil {
		return rep, mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var d domain.OccupancyDay
		if err := rows.Scan(&d.Date, &d.Rooms, &d.Reserved); err != nil {
			return rep, err
		}
		rep.Days = append(rep.Days, d)
	}
	if err := rows.Err(); err != nil {
		return rep, mapErr(err)
	}
	rows.Close()

	err = r.db.QueryRow(ctx, `
		select (select currency from hotel where id = $1),
			count(*) filter (where status = $4),
			coalesce(sum(total_amount) filter (where status = $4), 0)::text,
			count(*) filter (where status = $5),
			coalesce(sum(refund_amount) filter (where status = $5), 0)::text,
			coalesce(sum(fee_amount) filter (where status = $5), 0)::text,
			count(*) filter (where stay_status = 3)
		from reservation where hotel_id = $1 and start_date >= $2 and start_date < $3`,
		hotelID, from, to, domain.ReservationPaid, domain.ReservationRefunded).
		Scan(&rep.Currency, &rep.PaidCount, &rep.PaidRevenue, &rep.RefundedCount, &rep.RefundedAmount, &rep.FeesKept, &rep.NoShows)
	if err != nil {
		return rep, mapErr(err)
	}
	counts, err := r.db.Query(ctx, `select status, count(*)::int from reservation
		where hotel_id = $1 and start_date >= $2 and start_date < $3 group by status`, hotelID, from, to)
	if err != nil {
		return rep, mapErr(err)
	}
	defer counts.Close()
	for counts.Next() {
		var st domain.ReservationStatus
		var n int
		if err := counts.Scan(&st, &n); err != nil {
			return rep, err
		}
		rep.StatusCounts[st] = n
	}
	return rep, mapErr(counts.Err())
}

// ---- search filters, added to the hotel search query

// amenityFilter keeps hotels that have all the wanted amenities, compared case-insensitively.
func amenityFilter(arg func(any) int, wanted []string) string {
	seen := map[string]bool{}
	var lower []string
	for _, a := range wanted {
		if a = strings.ToLower(strings.TrimSpace(a)); a != "" && !seen[a] {
			seen[a] = true
			lower = append(lower, a)
		}
	}
	if len(lower) == 0 {
		return ""
	}
	n := arg(lower)
	return " and (select count(distinct lower(a)) from unnest(h.amenities) a where lower(a) = any($" + itoa(n) + "::text[])) = " + itoa(len(lower))
}

func itoa(n int) string {
	const digits = "0123456789"
	if n == 0 {
		return "0"
	}
	var b []byte
	for ; n > 0; n /= 10 {
		b = append([]byte{digits[n%10]}, b...)
	}
	return string(b)
}

// ---- wishlist

type WishlistRepository struct {
	db     *pgxpool.Pool
	hotels *HotelRepository
}

var _ outbound.WishlistRepository = (*WishlistRepository)(nil)

func NewWishlistRepository(db *pgxpool.Pool) *WishlistRepository {
	return &WishlistRepository{db: db, hotels: NewHotelRepository(db)}
}

func (r *WishlistRepository) Add(ctx context.Context, userID, hotelID int64) error {
	_, err := r.db.Exec(ctx, `insert into wishlist (user_id, hotel_id) values ($1, $2) on conflict do nothing`, userID, hotelID)
	if mapped := mapErr(err); errors.Is(mapped, domain.ErrInvalid) {
		return domain.ErrNotFound // foreign key: no such hotel
	} else {
		return mapped
	}
}

func (r *WishlistRepository) Remove(ctx context.Context, userID, hotelID int64) error {
	_, err := r.db.Exec(ctx, `delete from wishlist where user_id = $1 and hotel_id = $2`, userID, hotelID)
	return mapErr(err)
}

func (r *WishlistRepository) List(ctx context.Context, userID, afterID int64, limit int) ([]domain.Hotel, error) {
	rows, err := r.db.Query(ctx, `select `+hotelCols+hotelFrom+`
		join wishlist w on w.hotel_id = h.id
		where w.user_id = $1 and h.status = $2 and ($3 = 0 or h.id < $3)
		order by h.id desc limit $4`, userID, domain.HotelActive, afterID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Hotel{}
	for rows.Next() {
		h, err := scanHotel(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, mapErr(rows.Err())
}
