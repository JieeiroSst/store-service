package postgres

import (
	"context"
	"encoding/json"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type VenueRepository struct{ db *pgxpool.Pool }

var _ outbound.VenueRepository = (*VenueRepository)(nil)

func NewVenueRepository(db *pgxpool.Pool) *VenueRepository { return &VenueRepository{db: db} }

const venueCols = `id, owner_id, name, coalesce(city, ''), coalesce(address, ''), coalesce(description, ''), shared, layout, seats, sections, created_at, updated_at`

func scanVenue(row pgx.Row) (domain.Venue, error) {
	var v domain.Venue
	var raw []byte
	if err := row.Scan(&v.ID, &v.OwnerID, &v.Name, &v.City, &v.Address, &v.Description, &v.Shared, &raw, &v.Seats, &v.Sections, &v.CreatedAt, &v.UpdatedAt); err != nil {
		return v, mapErr(err)
	}
	return v, json.Unmarshal(raw, &v.Map)
}

func (r *VenueRepository) Create(ctx context.Context, v domain.Venue) (domain.Venue, error) {
	raw, err := json.Marshal(v.Map)
	if err != nil {
		return domain.Venue{}, err
	}
	return scanVenue(r.db.QueryRow(ctx, `insert into venue (owner_id, name, city, address, description, layout, seats, sections)
		values ($1, $2, nullif($3, ''), nullif($4, ''), nullif($5, ''), $6, $7, $8) returning `+venueCols,
		v.OwnerID, v.Name, v.City, v.Address, v.Description, raw, v.Seats, v.Sections))
}

func (r *VenueRepository) Get(ctx context.Context, id int64) (domain.Venue, error) {
	return scanVenue(r.db.QueryRow(ctx, `select `+venueCols+` from venue where id = $1`, id))
}

func (r *VenueRepository) Update(ctx context.Context, v domain.Venue) (domain.Venue, error) {
	raw, err := json.Marshal(v.Map)
	if err != nil {
		return domain.Venue{}, err
	}
	return scanVenue(r.db.QueryRow(ctx, `update venue set name = $2, city = nullif($3, ''), address = nullif($4, ''), description = nullif($5, ''),
		layout = $6, seats = $7, sections = $8, updated_at = now() where id = $1 returning `+venueCols,
		v.ID, v.Name, v.City, v.Address, v.Description, raw, v.Seats, v.Sections))
}

func (r *VenueRepository) Delete(ctx context.Context, id int64) error {
	err := pgx.BeginFunc(ctx, r.db, func(tx pgx.Tx) error {
		tag, err := tx.Exec(ctx, `delete from venue where id = $1`, id)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			return pgx.ErrNoRows
		}
		// events keep their seats and drawing; only the pointer goes
		_, err = tx.Exec(ctx, `update event set venue_id = null where venue_id = $1`, id)
		return err
	})
	return mapErr(err)
}

func (r *VenueRepository) SetShared(ctx context.Context, id int64, shared bool) error {
	tag, err := r.db.Exec(ctx, `update venue set shared = $2, updated_at = now() where id = $1`, id, shared)
	if err != nil {
		return mapErr(err)
	}
	if tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

func (r *VenueRepository) List(ctx context.Context, f outbound.VenueFilter) ([]domain.Venue, error) {
	rows, err := r.db.Query(ctx, `select `+venueCols+` from venue where ($1::bigint = 0 or owner_id = $1) and (not $2 or shared)
		and ($3::bigint = 0 or id < $3) order by id desc limit $4`, f.OwnerID, f.Shared, f.AfterID, f.Limit)
	if err != nil {
		return nil, mapErr(err)
	}
	defer rows.Close()
	out := []domain.Venue{}
	for rows.Next() {
		v, err := scanVenue(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, v)
	}
	return out, mapErr(rows.Err())
}
