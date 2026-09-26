package postgres

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSSt/reservation-service/internal/application/port/outbound"
	"github.com/JIeeiroSSt/reservation-service/internal/domain"
)

type HotelRepository struct{ db *pgxpool.Pool }

var _ outbound.HotelRepository = (*HotelRepository)(nil)

func NewHotelRepository(db *pgxpool.Pool) *HotelRepository { return &HotelRepository{db: db} }

const hotelCols = `h.id, h.name, coalesce(h.description, ''), h.stars, h.city, h.address, h.phone_number,
	coalesce(h.email, ''), coalesce(h.images, '{}'), coalesce(h.amenities, '{}'), h.check_in_time, h.check_out_time,
	h.currency, h.free_cancel_hours, h.status, coalesce(h.review_note, ''), h.owner_id, coalesce(h.wallet_id, ''),
	coalesce(r.avg, 0)::float8, coalesce(r.cnt, 0)::int, coalesce(h.version, 0),
	coalesce(h.created_at, now()), coalesce(h.updated_at, h.created_at, now()), h.cancellation_tiers`

// hotelFrom joins each hotel's review summary.
const hotelFrom = ` from hotel h left join (
	select hotel_id, avg(rating) avg, count(*) cnt from review group by hotel_id
) r on r.hotel_id = h.id`

type tierJSON struct {
	HoursBefore int `json:"hours_before"`
	FeePercent  int `json:"fee_percent"`
}

// tiersJSON encodes a cancellation policy for the jsonb column; no tiers is NULL ("use free_cancel_hours").
func tiersJSON(tiers []domain.CancellationTier) any {
	if len(tiers) == 0 {
		return nil
	}
	out := make([]tierJSON, len(tiers))
	for i, t := range tiers {
		out[i] = tierJSON{t.HoursBefore, t.FeePercent}
	}
	b, _ := json.Marshal(out)
	return string(b)
}

func scanHotel(row pgx.Row) (domain.Hotel, error) {
	var h domain.Hotel
	var tiers []byte
	err := row.Scan(&h.ID, &h.Name, &h.Description, &h.Stars, &h.City, &h.Address, &h.PhoneNumber, &h.Email,
		&h.Images, &h.Amenities, &h.CheckInTime, &h.CheckOutTime, &h.Currency, &h.FreeCancelHours, &h.Status,
		&h.ReviewNote, &h.OwnerID, &h.WalletID, &h.Rating, &h.ReviewCount, &h.Version, &h.CreatedAt, &h.UpdatedAt, &tiers)
	if err == nil && len(tiers) > 0 {
		var raw []tierJSON
		if jerr := json.Unmarshal(tiers, &raw); jerr == nil {
			for _, t := range raw {
				h.CancellationTiers = append(h.CancellationTiers, domain.CancellationTier{HoursBefore: t.HoursBefore, FeePercent: t.FeePercent})
			}
		}
	}
	return h, mapErr(err)
}

func (r *HotelRepository) Get(ctx context.Context, id int64) (domain.Hotel, error) {
	h, err := scanHotel(r.db.QueryRow(ctx, `select `+hotelCols+hotelFrom+` where h.id = $1`, id))
	if err != nil {
		return h, err
	}
	h.RoomTypes, err = r.roomTypes(ctx, id)
	return h, err
}

func (r *HotelRepository) Create(ctx context.Context, h domain.Hotel) (domain.Hotel, error) {
	var id int64
	err := r.db.QueryRow(ctx, `
		insert into hotel (name, description, stars, city, address, phone_number, email, images, amenities,
			check_in_time, check_out_time, currency, free_cancel_hours, status, owner_id, wallet_id,
			cancellation_tiers, version, created_at, updated_at)
		values ($1, $2, $3, $4, $5, $6, nullif($7, ''), $8, $9, $10, $11, $12, $13, $14, $15, nullif($16, ''), $17::jsonb, 1, now(), now())
		returning id`,
		h.Name, h.Description, h.Stars, h.City, h.Address, h.PhoneNumber, h.Email, nonNil(h.Images), nonNil(h.Amenities),
		h.CheckInTime, h.CheckOutTime, h.Currency, h.FreeCancelHours, h.Status, h.OwnerID, h.WalletID, tiersJSON(h.CancellationTiers)).Scan(&id)
	if err != nil {
		return domain.Hotel{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *HotelRepository) Update(ctx context.Context, h domain.Hotel) (domain.Hotel, error) {
	tag, err := r.db.Exec(ctx, `
		update hotel set name = $2, description = $3, stars = $4, city = $5, address = $6, phone_number = $7,
			email = nullif($8, ''), images = $9, amenities = $10, check_in_time = $11, check_out_time = $12,
			currency = $13, free_cancel_hours = $14, status = $15::smallint, wallet_id = nullif($16, ''),
			cancellation_tiers = $17::jsonb,
			review_note = case when $15::smallint = 4 then review_note else null end,
			version = coalesce(version, 0) + 1, updated_at = now()
		where id = $1`,
		h.ID, h.Name, h.Description, h.Stars, h.City, h.Address, h.PhoneNumber, h.Email, nonNil(h.Images), nonNil(h.Amenities),
		h.CheckInTime, h.CheckOutTime, h.Currency, h.FreeCancelHours, h.Status, h.WalletID, tiersJSON(h.CancellationTiers))
	if err != nil {
		return domain.Hotel{}, mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.Hotel{}, domain.ErrNotFound
	}
	return r.Get(ctx, h.ID)
}

func (r *HotelRepository) SetStatus(ctx context.Context, id int64, s domain.HotelStatus) error {
	return r.SetReview(ctx, id, s, "")
}

func (r *HotelRepository) SetReview(ctx context.Context, id int64, s domain.HotelStatus, note string) error {
	tag, err := r.db.Exec(ctx, `update hotel set status = $2, review_note = nullif($3, ''), version = coalesce(version, 0) + 1,
		updated_at = now() where id = $1`, id, s, note)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (r *HotelRepository) List(ctx context.Context, f domain.HotelFilter) ([]domain.Hotel, error) {
	q := `select ` + hotelCols + hotelFrom + ` where true`
	var args []any
	arg := func(v any) int {
		args = append(args, v)
		return len(args)
	}
	switch {
	case f.Status != 0:
		q += fmt.Sprintf(" and h.status = $%d", arg(f.Status))
	case !f.IncludeInactive:
		q += fmt.Sprintf(" and h.status = $%d", arg(domain.HotelActive))
	}
	if f.OwnerID > 0 {
		q += fmt.Sprintf(" and h.owner_id = $%d", arg(f.OwnerID))
	}
	if f.City != "" {
		q += fmt.Sprintf(" and lower(h.city) = lower($%d)", arg(f.City))
	}
	if f.Query != "" {
		n := arg("%" + likeEscaper.Replace(f.Query) + "%")
		q += fmt.Sprintf(" and (h.name ilike $%d or h.city ilike $%d or h.address ilike $%d)", n, n, n)
	}
	q += amenityFilter(arg, f.Amenities)
	rateBounds := "" // added to conditions on a night's rate
	rateParams := func() string {
		s := ""
		if f.MinRate != "" {
			s += fmt.Sprintf(" and i.rate >= $%d::numeric", arg(f.MinRate))
		}
		if f.MaxRate != "" {
			s += fmt.Sprintf(" and i.rate <= $%d::numeric", arg(f.MaxRate))
		}
		return s
	}
	if !f.CheckIn.IsZero() && f.CheckOut.After(f.CheckIn) {
		nights := int(f.CheckOut.Sub(f.CheckIn).Hours() / 24)
		in, out, rooms, guests := arg(f.CheckIn), arg(f.CheckOut), arg(max(f.Rooms, 1)), arg(max(f.Guests, 1))
		rateBounds = rateParams()
		q += fmt.Sprintf(` and exists (select 1 from room_type rt where rt.hotel_id = h.id and rt.active
			and rt.capacity * $%[3]d >= $%[4]d
			and (select count(*) from room_type_inventory i where i.room_type_id = rt.id
				and i.date >= $%[1]d and i.date < $%[2]d and i.rate is not null
				and i.total_inventory - i.total_reserved >= $%[3]d%[6]s) = %[5]d)`, in, out, rooms, guests, nights, rateBounds)
	} else if f.MinRate != "" || f.MaxRate != "" {
		// No dates: the hotel has some upcoming night, of a room type on sale, within the rate range.
		q += fmt.Sprintf(` and exists (select 1 from room_type_inventory i join room_type rt on rt.id = i.room_type_id and rt.active
			where i.hotel_id = h.id and i.date >= current_date and i.rate is not null%s)`, rateParams())
	}

	// Ordering and keyset paging. Every order ends in "h.id desc" so it is total, and a cursor (the
	// sort keys of the last row) selects exactly the rows after it.
	c := f.After
	rating := "coalesce(r.avg, 0)::float8"
	switch f.Sort {
	case "newest":
		if c != nil {
			q += fmt.Sprintf(" and h.id < $%d", arg(c.ID))
		}
		q += " order by h.id desc"
	case "oldest":
		if c != nil {
			q += fmt.Sprintf(" and h.id > $%d", arg(c.ID))
		}
		q += " order by h.id asc"
	default: // rating, and the database half of recommended (the service re-ranks with manager tiers)
		if c != nil {
			a, n, id := arg(c.Rating), arg(c.Count), arg(c.ID)
			q += fmt.Sprintf(" and (%[1]s < $%[2]d or (%[1]s = $%[2]d and (coalesce(r.cnt, 0) < $%[3]d or (coalesce(r.cnt, 0) = $%[3]d and h.id < $%[4]d))))", rating, a, n, id)
		}
		q += " order by " + rating + " desc, coalesce(r.cnt, 0) desc, h.id desc"
	}
	q += fmt.Sprintf(" limit $%d offset $%d", arg(f.Limit), arg(f.Offset))

	rows, err := r.db.Query(ctx, q, args...)
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

// ---- room types

const roomTypeCols = `id, hotel_id, name, coalesce(description, ''), capacity, coalesce(bed, ''), coalesce(size_m2, 0),
	coalesce(view, ''), coalesce(amenities, '{}'), active`

func scanRoomType(row pgx.Row) (domain.RoomType, error) {
	var t domain.RoomType
	err := row.Scan(&t.ID, &t.HotelID, &t.Name, &t.Description, &t.Capacity, &t.Bed, &t.SizeM2, &t.View, &t.Amenities, &t.Active)
	return t, mapErr(err)
}

func (r *HotelRepository) roomTypes(ctx context.Context, hotelID int64) ([]domain.RoomType, error) {
	rows, err := r.db.Query(ctx, `select `+roomTypeCols+` from room_type where hotel_id = $1 order by id`, hotelID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.RoomType{}
	for rows.Next() {
		t, err := scanRoomType(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, mapErr(rows.Err())
}

func (r *HotelRepository) CreateRoomType(ctx context.Context, t domain.RoomType) (domain.RoomType, error) {
	return scanRoomType(r.db.QueryRow(ctx, `
		insert into room_type (hotel_id, name, description, capacity, bed, size_m2, view, amenities, active)
		values ($1, $2, nullif($3, ''), $4, nullif($5, ''), nullif($6, 0), nullif($7, ''), $8, $9)
		returning `+roomTypeCols, t.HotelID, t.Name, t.Description, t.Capacity, t.Bed, t.SizeM2, t.View, nonNil(t.Amenities), t.Active))
}

func (r *HotelRepository) UpdateRoomType(ctx context.Context, t domain.RoomType) (domain.RoomType, error) {
	return scanRoomType(r.db.QueryRow(ctx, `
		update room_type set name = $3, description = nullif($4, ''), capacity = $5, bed = nullif($6, ''),
			size_m2 = nullif($7, 0), view = nullif($8, ''), amenities = $9, active = $10
		where id = $1 and hotel_id = $2 returning `+roomTypeCols,
		t.ID, t.HotelID, t.Name, t.Description, t.Capacity, t.Bed, t.SizeM2, t.View, nonNil(t.Amenities), t.Active))
}

func (r *HotelRepository) GetRoomType(ctx context.Context, id int64) (domain.RoomType, error) {
	return scanRoomType(r.db.QueryRow(ctx, `select `+roomTypeCols+` from room_type where id = $1`, id))
}

// ---- rooms

const roomCols = `id, hotel_id, room_type_id, name, coalesce(floor, 0), is_available`

func scanRoom(row pgx.Row) (domain.Room, error) {
	var x domain.Room
	err := row.Scan(&x.ID, &x.HotelID, &x.RoomTypeID, &x.Name, &x.Floor, &x.Available)
	return x, mapErr(err)
}

func (r *HotelRepository) CreateRoom(ctx context.Context, x domain.Room) (domain.Room, error) {
	return scanRoom(r.db.QueryRow(ctx, `insert into room (hotel_id, room_type_id, name, floor, is_available)
		values ($1, $2, $3, $4, $5) returning `+roomCols, x.HotelID, x.RoomTypeID, x.Name, x.Floor, x.Available))
}

func (r *HotelRepository) UpdateRoom(ctx context.Context, x domain.Room) (domain.Room, error) {
	return scanRoom(r.db.QueryRow(ctx, `update room set room_type_id = $3, name = $4, floor = $5, is_available = $6
		where id = $1 and hotel_id = $2 returning `+roomCols, x.ID, x.HotelID, x.RoomTypeID, x.Name, x.Floor, x.Available))
}

func (r *HotelRepository) GetRoom(ctx context.Context, id int64) (domain.Room, error) {
	return scanRoom(r.db.QueryRow(ctx, `select `+roomCols+` from room where id = $1`, id))
}

func (r *HotelRepository) Rooms(ctx context.Context, hotelID, roomTypeID int64) ([]domain.Room, error) {
	rows, err := r.db.Query(ctx, `select `+roomCols+` from room where hotel_id = $1 and ($2 = 0 or room_type_id = $2)
		order by floor, name`, hotelID, roomTypeID)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Room{}
	for rows.Next() {
		x, err := scanRoom(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, x)
	}
	return out, mapErr(rows.Err())
}

// ---- inventory and rates

func (r *HotelRepository) SetInventory(ctx context.Context, p outbound.SetInventoryParams) error {
	nights := int(p.To.Sub(p.From).Hours() / 24)
	rateSet, rate := p.Rate != nil, ""
	if rateSet {
		rate = *p.Rate
	}
	var written int
	err := r.db.QueryRow(ctx, `
		with up as (
			insert into room_type_inventory (hotel_id, room_type_id, date, total_inventory, rate)
			select $1, $2, d::date,
				coalesce($5::int, (select count(*) from room where room_type_id = $2 and is_available)),
				case when $6 then nullif($7, '')::numeric else null end
			from generate_series($3::date, $4::date - 1, interval '1 day') d
			on conflict (room_type_id, date) do update
				set total_inventory = excluded.total_inventory,
					rate = case when $6 then excluded.rate else room_type_inventory.rate end
				where excluded.total_inventory >= room_type_inventory.total_reserved
			returning 1
		) select count(*) from up`,
		p.HotelID, p.RoomTypeID, p.From, p.To, p.Total, rateSet, rate).Scan(&written)
	if err != nil {
		return mapErr(err)
	}
	if written != nights {
		return fmt.Errorf("%w: some nights already have more rooms reserved than that total", domain.ErrConflict)
	}
	return nil
}

func (r *HotelRepository) Inventory(ctx context.Context, hotelID, roomTypeID int64, from, to time.Time) ([]domain.InventoryDay, error) {
	rows, err := r.db.Query(ctx, `
		select hotel_id, room_type_id, date, total_inventory, total_reserved, coalesce(rate::text, '')
		from room_type_inventory
		where hotel_id = $1 and room_type_id = $2 and date >= $3 and date < $4 order by date`, hotelID, roomTypeID, from, to)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.InventoryDay{}
	for rows.Next() {
		var d domain.InventoryDay
		if err := rows.Scan(&d.HotelID, &d.RoomTypeID, &d.Date, &d.Total, &d.Reserved, &d.Rate); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, mapErr(rows.Err())
}

func (r *HotelRepository) Quotes(ctx context.Context, hotelID int64, from, to time.Time, rooms int) ([]domain.Quote, error) {
	nights := int(to.Sub(from).Hours() / 24)
	rows, err := r.db.Query(ctx, `
		select rt.id, rt.hotel_id, rt.name, coalesce(rt.description, ''), rt.capacity, coalesce(rt.bed, ''), coalesce(rt.size_m2, 0),
			coalesce(rt.view, ''), coalesce(rt.amenities, '{}'), rt.active,
			coalesce(x.priced, 0), coalesce(x.left_min, 0), (coalesce(x.total, 0) * $4)::text
		from room_type rt
		left join lateral (
			select count(*) priced, min(total_inventory - total_reserved) left_min, sum(rate) total
			from room_type_inventory i
			where i.room_type_id = rt.id and i.date >= $2 and i.date < $3 and i.rate is not null
		) x on true
		where rt.hotel_id = $1 and rt.active order by rt.id`, hotelID, from, to, rooms)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Quote{}
	for rows.Next() {
		var q domain.Quote
		var priced, left int
		var total string
		t := &q.RoomType
		if err := rows.Scan(&t.ID, &t.HotelID, &t.Name, &t.Description, &t.Capacity, &t.Bed, &t.SizeM2, &t.View, &t.Amenities, &t.Active,
			&priced, &left, &total); err != nil {
			return nil, err
		}
		q.Nights = nights
		if priced == nights { // every night is for sale
			q.Available, q.Total = left, total
		}
		out = append(out, q)
	}
	return out, mapErr(rows.Err())
}

// ---- extra services

const serviceCols = `id, hotel_id, name, coalesce(description, ''), price::text, unit, active`

func scanService(row pgx.Row) (domain.HotelService, error) {
	var s domain.HotelService
	err := row.Scan(&s.ID, &s.HotelID, &s.Name, &s.Description, &s.Price, &s.Unit, &s.Active)
	return s, mapErr(err)
}

func (r *HotelRepository) SaveService(ctx context.Context, s domain.HotelService) (domain.HotelService, error) {
	if s.ID == 0 {
		return scanService(r.db.QueryRow(ctx, `insert into hotel_service (hotel_id, name, description, price, unit, active)
			values ($1, $2, nullif($3, ''), $4::numeric, $5, $6) returning `+serviceCols,
			s.HotelID, s.Name, s.Description, s.Price, string(s.Unit), s.Active))
	}
	return scanService(r.db.QueryRow(ctx, `update hotel_service set name = $3, description = nullif($4, ''), price = $5::numeric,
		unit = $6, active = $7 where id = $1 and hotel_id = $2 returning `+serviceCols,
		s.ID, s.HotelID, s.Name, s.Description, s.Price, string(s.Unit), s.Active))
}

func (r *HotelRepository) Services(ctx context.Context, hotelID int64, activeOnly bool) ([]domain.HotelService, error) {
	rows, err := r.db.Query(ctx, `select `+serviceCols+` from hotel_service where hotel_id = $1 and (not $2 or active) order by id`, hotelID, activeOnly)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.HotelService{}
	for rows.Next() {
		s, err := scanService(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, mapErr(rows.Err())
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
