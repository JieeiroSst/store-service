package postgres

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

type HomestayRepository struct{ db *pgxpool.Pool }

var _ outbound.HomestayRepository = (*HomestayRepository)(nil)

func NewHomestayRepository(db *pgxpool.Pool) *HomestayRepository { return &HomestayRepository{db: db} }

const homestayCols = `h.id, h.name, coalesce(h.description, ''), coalesce(h.type, 0), coalesce(h.status, 0),
	coalesce(h.phone_number, ''), coalesce(h.address, ''),
	coalesce(h.ward_id, 0), coalesce(h.district_id, 0), coalesce(h.province_id, 0),
	coalesce(h.images, '{}'), coalesce(h.guests, 0), coalesce(h.bedrooms, 0), coalesce(h.bathrooms, 0),
	array(select amenity_id from homestay_amenity where homestay_id = h.id order by amenity_id),
	coalesce(h.version, 0), coalesce(h.created_at, now()), coalesce(h.updated_at, h.created_at, now()),
	coalesce(h.created_by, 0), coalesce(r.avg, 0)::float8, coalesce(r.cnt, 0)::int, coalesce(h.wallet_id, ''), coalesce(h.review_note, '')`

// homestayFrom joins each homestay's review summary.
const homestayFrom = ` from homestay h left join (
	select homestay_id, avg(rating) avg, count(*) cnt from review group by homestay_id
) r on r.homestay_id = h.id`

func scanHomestay(row pgx.Row) (domain.Homestay, error) {
	var h domain.Homestay
	var amenities []int32
	err := row.Scan(&h.ID, &h.Name, &h.Description, &h.Type, &h.Status, &h.PhoneNumber, &h.Address,
		&h.WardID, &h.DistrictID, &h.ProvinceID, &h.Images, &h.Guests, &h.Bedrooms, &h.Bathrooms,
		&amenities, &h.Version, &h.CreatedAt, &h.UpdatedAt, &h.HostID, &h.Rating, &h.ReviewCount, &h.WalletID, &h.ReviewNote)
	if err != nil {
		return domain.Homestay{}, mapErr(err)
	}
	h.AmenityIDs = make([]int, len(amenities))
	for i, a := range amenities {
		h.AmenityIDs[i] = int(a)
	}
	return h, nil
}

func (r *HomestayRepository) Get(ctx context.Context, id int64) (domain.Homestay, error) {
	h, err := scanHomestay(r.db.QueryRow(ctx, `select `+homestayCols+homestayFrom+` where h.id = $1`, id))
	if err != nil {
		return h, err
	}
	h.Rates, err = r.Rates(ctx, id)
	return h, err
}

func (r *HomestayRepository) Create(ctx context.Context, h domain.Homestay, actor int64) (domain.Homestay, error) {
	var id int64
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		if err := tx.QueryRow(ctx, `
			insert into homestay (name, description, type, status, phone_number, address,
				ward_id, district_id, province_id, images, guests, bedrooms, bathrooms,
				wallet_id, version, created_at, created_by, updated_at, updated_by)
			values ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, nullif($15, ''), 1, now(), $14, now(), $14)
			returning id`,
			h.Name, h.Description, h.Type, h.Status, h.PhoneNumber, h.Address,
			h.WardID, h.DistrictID, h.ProvinceID, nonNil(h.Images), h.Guests, h.Bedrooms, h.Bathrooms, actor, h.WalletID,
		).Scan(&id); err != nil {
			return err
		}
		return setAmenities(ctx, tx, id, h.AmenityIDs)
	})
	if err != nil {
		return domain.Homestay{}, mapErr(err)
	}
	return r.Get(ctx, id)
}

func (r *HomestayRepository) Update(ctx context.Context, h domain.Homestay, actor int64) (domain.Homestay, error) {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `
			update homestay set name = $2, description = $3, type = $4, status = $5, phone_number = $6, address = $7,
				ward_id = $8, district_id = $9, province_id = $10, images = $11,
				guests = $12, bedrooms = $13, bathrooms = $14, wallet_id = nullif($16, ''),
				review_note = case when $5 = 4 then review_note else null end,
				version = coalesce(version, 0) + 1, updated_at = now(), updated_by = $15
			where id = $1`,
			h.ID, h.Name, h.Description, h.Type, h.Status, h.PhoneNumber, h.Address,
			h.WardID, h.DistrictID, h.ProvinceID, nonNil(h.Images), h.Guests, h.Bedrooms, h.Bathrooms, actor, h.WalletID)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		return setAmenities(ctx, tx, h.ID, h.AmenityIDs)
	})
	if err != nil {
		return domain.Homestay{}, mapErr(err)
	}
	return r.Get(ctx, h.ID)
}

func setAmenities(ctx context.Context, tx pgx.Tx, homestayID int64, ids []int) error {
	if _, err := tx.Exec(ctx, `delete from homestay_amenity where homestay_id = $1`, homestayID); err != nil {
		return err
	}
	if len(ids) == 0 {
		return nil
	}
	_, err := tx.Exec(ctx, `
		insert into homestay_amenity (homestay_id, amenity_id)
		select $1, a from unnest($2::int[]) a
		on conflict do nothing`, homestayID, ids)
	return err
}

func (r *HomestayRepository) SetStatus(ctx context.Context, id int64, s domain.HomestayStatus, actor int64) error {
	tag, err := r.db.Exec(ctx, `
		update homestay set status = $2, version = coalesce(version, 0) + 1, updated_at = now(), updated_by = $3
		where id = $1`, id, s, actor)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *HomestayRepository) SetReview(ctx context.Context, id int64, s domain.HomestayStatus, note string, actor int64) error {
	tag, err := r.db.Exec(ctx, `
		update homestay set status = $2, review_note = nullif($3, ''), version = coalesce(version, 0) + 1,
			updated_at = now(), updated_by = $4
		where id = $1`, id, s, note, actor)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *HomestayRepository) List(ctx context.Context, f domain.HomestayFilter) ([]domain.Homestay, error) {
	q := `select ` + homestayCols + homestayFrom + ` where true`
	var args []any
	arg := func(v any) int {
		args = append(args, v)
		return len(args)
	}
	switch {
	case f.Status != 0:
		q += fmt.Sprintf(" and h.status = $%d", arg(f.Status))
	case !f.IncludeInactive:
		q += fmt.Sprintf(" and h.status = $%d", arg(domain.HomestayActive))
	}
	if f.HostID > 0 {
		q += fmt.Sprintf(" and h.created_by = $%d", arg(f.HostID))
	}
	for _, c := range []struct {
		col string
		v   int
	}{{"province_id", f.ProvinceID}, {"district_id", f.DistrictID}, {"ward_id", f.WardID}, {"type", f.Type}} {
		if c.v > 0 {
			q += fmt.Sprintf(" and h.%s = $%d", c.col, arg(c.v))
		}
	}
	if f.MinGuests > 0 {
		q += fmt.Sprintf(" and h.guests >= $%d", arg(f.MinGuests))
	}
	if f.Query != "" {
		n := arg("%" + likeEscaper.Replace(f.Query) + "%")
		q += fmt.Sprintf(" and (h.name ilike $%d or h.address ilike $%d)", n, n)
	}
	if len(f.AmenityIDs) > 0 {
		q += fmt.Sprintf(` and (select count(distinct amenity_id) from homestay_amenity
			where homestay_id = h.id and amenity_id = any($%d::int[])) = %d`, arg(f.AmenityIDs), len(uniqueInts(f.AmenityIDs)))
	}
	model := f.Model
	if model == "" {
		model = domain.ModelDay
	}
	if f.MinPrice != "" || f.MaxPrice != "" {
		q += fmt.Sprintf(" and exists (select 1 from homestay_rate hr where hr.homestay_id = h.id and hr.active and hr.model = $%d", arg(string(model)))
		if f.MinPrice != "" {
			q += fmt.Sprintf(" and hr.price >= $%d::numeric", arg(f.MinPrice))
		}
		if f.MaxPrice != "" {
			q += fmt.Sprintf(" and hr.price <= $%d::numeric", arg(f.MaxPrice))
		}
		q += ")"
	}
	if !f.CheckIn.IsZero() && f.CheckOut.After(f.CheckIn) {
		nights := int(f.CheckOut.Sub(f.CheckIn).Hours() / 24)
		in, out := arg(f.CheckIn), arg(f.CheckOut)
		q += fmt.Sprintf(` and (select count(*) from homestay_availability a where a.homestay_id = h.id
			and a.date >= $%d and a.date < $%d and a.status = $%d and a.price is not null) = %d`,
			in, out, arg(domain.SlotAvailable), nights)
	}

	c := f.After
	rating := "coalesce(r.avg, 0)::float8"
	switch f.Sort {
	case "price_asc", "price_desc":
		asc := f.Sort == "price_asc"
		sentinel, cmp, dir := "-1", "<", "desc"
		if asc {
			sentinel, cmp, dir = "1000000000000000", ">", "asc"
		}
		key := fmt.Sprintf("coalesce((select price from homestay_rate where homestay_id = h.id and model = $%d and active), %s::numeric)", arg(string(model)), sentinel)
		if c != nil {
			p := c.Price
			if p == "" {
				p = sentinel
			}
			k, id := arg(p), arg(c.ID)
			q += fmt.Sprintf(" and (%s %s $%d::numeric or (%s = $%d::numeric and h.id < $%d))", key, cmp, k, key, k, id)
		}
		q += fmt.Sprintf(" order by %s %s, h.id desc", key, dir)
	case "newest":
		if c != nil {
			q += fmt.Sprintf(" and h.id < $%d", arg(c.ID))
		}
		q += " order by h.id desc"
	case "oldest":
		q += " order by h.id asc"
	default:
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
	return out, r.attachRates(ctx, out)
}

var likeEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func uniqueInts(in []int) []int {
	seen := map[int]bool{}
	out := in[:0:0]
	for _, v := range in {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func (r *HomestayRepository) attachRates(ctx context.Context, hs []domain.Homestay) error {
	if len(hs) == 0 {
		return nil
	}
	ids := make([]int64, len(hs))
	idx := make(map[int64]int, len(hs))
	for i, h := range hs {
		ids[i], idx[h.ID] = h.ID, i
	}
	rows, err := r.db.Query(ctx, rateSelect+` where homestay_id = any($1) and active order by homestay_id, model`, ids)
	if err != nil {
		return mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		rt, err := scanRate(rows)
		if err != nil {
			return err
		}
		hs[idx[rt.HomestayID]].Rates = append(hs[idx[rt.HomestayID]].Rates, rt)
	}
	return mapErr(rows.Err())
}

const rateSelect = `select homestay_id, model, price::text, currency, min_periods, active from homestay_rate`

func scanRate(row pgx.Row) (domain.Rate, error) {
	var rt domain.Rate
	err := row.Scan(&rt.HomestayID, &rt.Model, &rt.Price, &rt.Currency, &rt.MinPeriods, &rt.Active)
	return rt, err
}

func (r *HomestayRepository) SetRate(ctx context.Context, rt domain.Rate, actor int64) error {
	_, err := r.db.Exec(ctx, `
		insert into homestay_rate (homestay_id, model, price, currency, min_periods, active, updated_by, updated_at)
		values ($1, $2, $3::numeric, $4, $5, $6, $7, now())
		on conflict (homestay_id, model) do update set price = excluded.price, currency = excluded.currency,
			min_periods = excluded.min_periods, active = excluded.active, updated_by = excluded.updated_by, updated_at = now()`,
		rt.HomestayID, string(rt.Model), rt.Price, rt.Currency, rt.MinPeriods, rt.Active, actor)
	return mapErr(err)
}

func (r *HomestayRepository) DeleteRate(ctx context.Context, id int64, model domain.RentalModel) error {
	tag, err := r.db.Exec(ctx, `delete from homestay_rate where homestay_id = $1 and model = $2`, id, string(model))
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *HomestayRepository) Rates(ctx context.Context, id int64) ([]domain.Rate, error) {
	rows, err := r.db.Query(ctx, rateSelect+` where homestay_id = $1 order by model`, id)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Rate{}
	for rows.Next() {
		rt, err := scanRate(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, rt)
	}
	return out, mapErr(rows.Err())
}

func (r *HomestayRepository) SetAvailability(ctx context.Context, id int64, from, to time.Time, price string, status domain.AvailabilityStatus) error {
	// Booked nights belong to a booking and are never overwritten here.
	_, err := r.db.Exec(ctx, `
		insert into homestay_availability (homestay_id, date, price, status)
		select $1, d::date, nullif($4, '')::numeric, $5
		from generate_series($2::date, $3::date - 1, interval '1 day') d
		on conflict (homestay_id, date) do update
			set price = excluded.price, status = excluded.status
			where homestay_availability.status <> $6`,
		id, from, to, price, status, domain.SlotBooked)
	return mapErr(err)
}

func (r *HomestayRepository) Availability(ctx context.Context, id int64, from, to time.Time) ([]domain.Slot, error) {
	rows, err := r.db.Query(ctx, `
		select homestay_id, date, coalesce(price::text, ''), coalesce(status, 0)
		from homestay_availability
		where homestay_id = $1 and date >= $2 and date < $3
		order by date`, id, from, to)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Slot{}
	for rows.Next() {
		var s domain.Slot
		if err := rows.Scan(&s.HomestayID, &s.Date, &s.Price, &s.Status); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, mapErr(rows.Err())
}

func (r *HomestayRepository) ListAmenities(ctx context.Context) ([]domain.Amenity, error) {
	rows, err := r.db.Query(ctx, `select id, name, icon from amenity order by name`)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Amenity{}
	for rows.Next() {
		var a domain.Amenity
		if err := rows.Scan(&a.ID, &a.Name, &a.Icon); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, mapErr(rows.Err())
}

func (r *HomestayRepository) CreateAmenity(ctx context.Context, a domain.Amenity) (domain.Amenity, error) {
	err := r.db.QueryRow(ctx, `insert into amenity (name, icon) values ($1, $2) returning id`, a.Name, a.Icon).Scan(&a.ID)
	return a, mapErr(err)
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}
