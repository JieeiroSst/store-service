package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/geoservice/internal/core/domain"
	"github.com/geoservice/internal/core/port"
)

type geoRepo struct {
	pool *pgxpool.Pool
}

func NewGeoRepository(pool *pgxpool.Pool) port.GeoRepository {
	return &geoRepo{pool: pool}
}

const vnLower = `lower(translate(%s,
  'ÀÁẢÃẠĂẰẮẲẴẶÂẦẤẨẪẬÈÉẺẼẸÊỀẾỂỄỆÌÍỈĨỊÒÓỎÕỌÔỒỐỔỖỘƠỜỚỞỠỢÙÚỦŨỤƯỪỨỬỮỰỲÝỶỸỴĐàáảãạăằắẳẵặâầấẩẫậèéẻẽẹêềếểễệìíỉĩịòóỏõọôồốổỗộơờớởỡợùúủũụưừứửữựỳýỷỹỵđ',
  'AAAAAAAAAAAAAAAAAEEEEEEEEEEEIIIIIOOOOOOOOOOOOOOOOOUUUUUUUUUUUYYYYYDaaaaaaaaaaaaaaaaaeeeeeeeeeeeiiiiiooooooooooooooooouuuuuuuuuuuyyyyyd'))`

func provNorm(col string) string { return fmt.Sprintf(vnLower, col) }

// nearestDistrictTolerance is the max distance (m) for the district fallback
// when a point falls in a gap between GADM polygons.
const nearestDistrictTolerance = 500.0

func (r *geoRepo) LocateChain(ctx context.Context, pt domain.Point) (*domain.LocationResult, error) {
	res := &domain.LocationResult{Point: pt}

	const qProv = `
		SELECT code, name, name_full, name_en, type, region, slug, COALESCE(gid,''), COALESCE(area_km2,0)
		FROM provinces
		WHERE geom IS NOT NULL
		  AND ST_Contains(geom, ST_SetSRID(ST_MakePoint($1,$2),4326))
		LIMIT 1`
	var p domain.Province
	err := r.pool.QueryRow(ctx, qProv, pt.Lng, pt.Lat).
		Scan(&p.Code, &p.Name, &p.NameFull, &p.NameEn, &p.Type, &p.Region, &p.Slug, &p.Gid, &p.AreaKm2)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, nil
		}
		return nil, fmt.Errorf("locate province: %w", err)
	}
	res.Found = true
	res.Province = &p

	// District: try exact ST_Contains first, fall back to nearest within tolerance.
	const qDistExact = `
		SELECT code, province_code, name, name_full, name_en, type, slug, COALESCE(gid,''), COALESCE(area_km2,0)
		FROM districts
		WHERE province_code=$1 AND geom IS NOT NULL
		  AND ST_Contains(geom, ST_SetSRID(ST_MakePoint($2,$3),4326))
		LIMIT 1`
	const qDistNearest = `
		SELECT code, province_code, name, name_full, name_en, type, slug, COALESCE(gid,''), COALESCE(area_km2,0)
		FROM districts
		WHERE province_code=$1 AND geom IS NOT NULL
		  AND ST_DWithin(geom::geography, ST_SetSRID(ST_MakePoint($2,$3),4326)::geography, $4)
		ORDER BY geom <-> ST_SetSRID(ST_MakePoint($2,$3),4326)
		LIMIT 1`

	var d domain.District
	err = r.pool.QueryRow(ctx, qDistExact, p.Code, pt.Lng, pt.Lat).
		Scan(&d.Code, &d.ProvinceCode, &d.Name, &d.NameFull, &d.NameEn, &d.Type, &d.Slug, &d.Gid, &d.AreaKm2)
	if errors.Is(err, pgx.ErrNoRows) {
		err = r.pool.QueryRow(ctx, qDistNearest, p.Code, pt.Lng, pt.Lat, nearestDistrictTolerance).
			Scan(&d.Code, &d.ProvinceCode, &d.Name, &d.NameFull, &d.NameEn, &d.Type, &d.Slug, &d.Gid, &d.AreaKm2)
	}
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, fmt.Errorf("locate district: %w", err)
	}
	if err == nil {
		res.District = &d

		// Ward: exact ST_Contains, fall back to nearest ward in the district (no distance cap —
		// ensures every point inside a district always resolves to a ward even when some wards
		// have no geometry in the source data).
		const qWardExact = `
			SELECT code, district_code, province_code, name, name_full, name_en, type, slug, COALESCE(gid,''), COALESCE(area_km2,0)
			FROM wards
			WHERE district_code=$1 AND geom IS NOT NULL
			  AND ST_Contains(geom, ST_SetSRID(ST_MakePoint($2,$3),4326))
			LIMIT 1`
		const qWardNearest = `
			SELECT code, district_code, province_code, name, name_full, name_en, type, slug, COALESCE(gid,''), COALESCE(area_km2,0)
			FROM wards
			WHERE district_code=$1 AND geom IS NOT NULL
			ORDER BY geom <-> ST_SetSRID(ST_MakePoint($2,$3),4326)
			LIMIT 1`

		var w domain.Ward
		werr := r.pool.QueryRow(ctx, qWardExact, d.Code, pt.Lng, pt.Lat).
			Scan(&w.Code, &w.DistrictCode, &w.ProvinceCode, &w.Name, &w.NameFull, &w.NameEn, &w.Type, &w.Slug, &w.Gid, &w.AreaKm2)
		if errors.Is(werr, pgx.ErrNoRows) {
			werr = r.pool.QueryRow(ctx, qWardNearest, d.Code, pt.Lng, pt.Lat).
				Scan(&w.Code, &w.DistrictCode, &w.ProvinceCode, &w.Name, &w.NameFull, &w.NameEn, &w.Type, &w.Slug, &w.Gid, &w.AreaKm2)
		}
		if werr == nil {
			res.Ward = &w
		} else if !errors.Is(werr, pgx.ErrNoRows) {
			return nil, fmt.Errorf("locate ward: %w", werr)
		}
	}

	// Nearest street (province-scoped, within 500 m of the point).
	const qStreet = `
		SELECT id, province_code, COALESCE(district_code,''), name, COALESCE(name_full,''), COALESCE(name_en,''),
		       COALESCE(slug,''), COALESCE(gid,''), COALESCE(length_m,0),
		       ST_Distance(geom::geography, ST_SetSRID(ST_MakePoint($2,$3),4326)::geography) AS dist_m
		FROM streets
		WHERE province_code=$1 AND geom IS NOT NULL
		  AND ST_DWithin(geom::geography, ST_SetSRID(ST_MakePoint($2,$3),4326)::geography, 500)
		ORDER BY geom <-> ST_SetSRID(ST_MakePoint($2,$3),4326)
		LIMIT 1`
	var st domain.Street
	serr := r.pool.QueryRow(ctx, qStreet, p.Code, pt.Lng, pt.Lat).
		Scan(&st.ID, &st.ProvinceCode, &st.DistrictCode, &st.Name, &st.NameFull, &st.NameEn,
			&st.Slug, &st.Gid, &st.LengthM, &st.DistanceM)
	if serr == nil {
		res.Street = &st
	} else if !errors.Is(serr, pgx.ErrNoRows) {
		return nil, fmt.Errorf("locate street: %w", serr)
	}

	return res, nil
}

func (r *geoRepo) NearestProvince(ctx context.Context, pt domain.Point, toleranceM float64) (*domain.LocationResult, error) {
	const q = `
		SELECT code, name, name_full, name_en, type, region, slug, COALESCE(gid,''), COALESCE(area_km2,0),
		       ST_Distance(geom::geography, ST_SetSRID(ST_MakePoint($1,$2),4326)::geography) AS dist_m
		FROM provinces
		WHERE geom IS NOT NULL
		ORDER BY geom <-> ST_SetSRID(ST_MakePoint($1,$2),4326)
		LIMIT 1`
	var p domain.Province
	var dist float64
	err := r.pool.QueryRow(ctx, q, pt.Lng, pt.Lat).
		Scan(&p.Code, &p.Name, &p.NameFull, &p.NameEn, &p.Type, &p.Region, &p.Slug, &p.Gid, &p.AreaKm2, &dist)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return &domain.LocationResult{Found: false, Point: pt}, nil
		}
		return nil, fmt.Errorf("nearest province: %w", err)
	}
	if dist > toleranceM {
		return &domain.LocationResult{Found: false, Point: pt, DistanceM: dist}, nil
	}
	res := &domain.LocationResult{Found: true, Province: &p, Point: pt, DistanceM: dist}

	const qDist = `
		SELECT code, province_code, name, name_full, name_en, type, slug, COALESCE(gid,''), COALESCE(area_km2,0)
		FROM districts
		WHERE province_code=$1 AND geom IS NOT NULL
		ORDER BY geom <-> ST_SetSRID(ST_MakePoint($2,$3),4326)
		LIMIT 1`
	var d domain.District
	err = r.pool.QueryRow(ctx, qDist, p.Code, pt.Lng, pt.Lat).
		Scan(&d.Code, &d.ProvinceCode, &d.Name, &d.NameFull, &d.NameEn, &d.Type, &d.Slug, &d.Gid, &d.AreaKm2)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return res, nil
		}
		return nil, fmt.Errorf("nearest district: %w", err)
	}
	res.District = &d

	const qWard = `
		SELECT code, district_code, province_code, name, name_full, name_en, type, slug, COALESCE(gid,''), COALESCE(area_km2,0)
		FROM wards
		WHERE district_code=$1 AND geom IS NOT NULL
		ORDER BY geom <-> ST_SetSRID(ST_MakePoint($2,$3),4326)
		LIMIT 1`
	var w domain.Ward
	werr := r.pool.QueryRow(ctx, qWard, d.Code, pt.Lng, pt.Lat).
		Scan(&w.Code, &w.DistrictCode, &w.ProvinceCode, &w.Name, &w.NameFull, &w.NameEn, &w.Type, &w.Slug, &w.Gid, &w.AreaKm2)
	if werr == nil {
		res.Ward = &w
	} else if !errors.Is(werr, pgx.ErrNoRows) {
		return nil, fmt.Errorf("nearest ward: %w", werr)
	}
	return res, nil
}

func (r *geoRepo) AttachGeometryByName(ctx context.Context, f domain.GeoFeature) (bool, error) {
	geomExpr := `ST_Multi(ST_MakeValid(ST_SetSRID(ST_GeomFromGeoJSON($1),4326)))`

	var q string
	var args []any
	switch f.Level {
	case domain.LevelProvince:
		q = `UPDATE provinces SET geom=` + geomExpr + `, gid=$2,
		        area_km2=ST_Area(` + geomExpr + `::geography)/1e6
		     WHERE ` + provNorm("$3") + ` IN (` + provNorm("name") + `, ` + provNorm("name_full") + `)`
		args = []any{f.GeoJSON, f.Gid, f.Name}

	case domain.LevelDistrict:
		q = `UPDATE districts d SET geom=` + geomExpr + `, gid=$2,
		        area_km2=ST_Area(` + geomExpr + `::geography)/1e6
		     FROM provinces p
		     WHERE d.province_code=p.code AND p.gid=$3
		       AND ` + provNorm("$4") + ` IN (` + provNorm("d.name") + `, ` + provNorm("d.name_full") + `)`
		args = []any{f.GeoJSON, f.Gid, f.ProvinceGid, f.Name}

	case domain.LevelWard:
		q = `UPDATE wards w SET geom=` + geomExpr + `, gid=$2,
		        area_km2=ST_Area(` + geomExpr + `::geography)/1e6
		     FROM districts d
		     WHERE w.district_code=d.code AND d.gid=$3
		       AND ` + provNorm("$4") + ` IN (` + provNorm("w.name") + `, ` + provNorm("w.name_full") + `)`
		args = []any{f.GeoJSON, f.Gid, f.DistrictGid, f.Name}

	default:
		return false, fmt.Errorf("unknown level %d", f.Level)
	}

	tag, err := r.pool.Exec(ctx, q, args...)
	if err != nil {
		return false, fmt.Errorf("attach L%d %q: %w", f.Level, f.Name, err)
	}
	return tag.RowsAffected() > 0, nil
}

func (r *geoRepo) GetProvince(ctx context.Context, code string) (*domain.Province, error) {
	const q = `SELECT code, name, name_full, name_en, type, region, slug, COALESCE(gid,''), COALESCE(area_km2,0)
	           FROM provinces WHERE code=$1`
	var p domain.Province
	err := r.pool.QueryRow(ctx, q, code).
		Scan(&p.Code, &p.Name, &p.NameFull, &p.NameEn, &p.Type, &p.Region, &p.Slug, &p.Gid, &p.AreaKm2)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &p, nil
}

func (r *geoRepo) ListProvinces(ctx context.Context) ([]domain.Province, error) {
	const q = `SELECT code, name, name_full, name_en, type, region, slug, COALESCE(gid,''), COALESCE(area_km2,0)
	           FROM provinces ORDER BY code`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Province
	for rows.Next() {
		var p domain.Province
		if err := rows.Scan(&p.Code, &p.Name, &p.NameFull, &p.NameEn, &p.Type, &p.Region, &p.Slug, &p.Gid, &p.AreaKm2); err != nil {
			return nil, err
		}
		out = append(out, p)
	}
	return out, rows.Err()
}

func (r *geoRepo) ListDistricts(ctx context.Context, provinceCode string) ([]domain.District, error) {
	const q = `SELECT code, province_code, name, name_full, name_en, type, slug, COALESCE(gid,''), COALESCE(area_km2,0)
	           FROM districts WHERE province_code=$1 ORDER BY code`
	rows, err := r.pool.Query(ctx, q, provinceCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.District
	for rows.Next() {
		var d domain.District
		if err := rows.Scan(&d.Code, &d.ProvinceCode, &d.Name, &d.NameFull, &d.NameEn, &d.Type, &d.Slug, &d.Gid, &d.AreaKm2); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}

func (r *geoRepo) ListWards(ctx context.Context, districtCode string) ([]domain.Ward, error) {
	const q = `SELECT code, district_code, province_code, name, name_full, name_en, type, slug, COALESCE(gid,''), COALESCE(area_km2,0)
	           FROM wards WHERE district_code=$1 ORDER BY code`
	rows, err := r.pool.Query(ctx, q, districtCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Ward
	for rows.Next() {
		var w domain.Ward
		if err := rows.Scan(&w.Code, &w.DistrictCode, &w.ProvinceCode, &w.Name, &w.NameFull, &w.NameEn, &w.Type, &w.Slug, &w.Gid, &w.AreaKm2); err != nil {
			return nil, err
		}
		out = append(out, w)
	}
	return out, rows.Err()
}

func (r *geoRepo) ListStreets(ctx context.Context, provinceCode string) ([]domain.Street, error) {
	const q = `SELECT id, province_code, COALESCE(district_code,''), name, COALESCE(name_full,''),
	                  COALESCE(name_en,''), COALESCE(slug,''), COALESCE(gid,''), COALESCE(length_m,0)
	           FROM streets WHERE province_code=$1 ORDER BY name`
	rows, err := r.pool.Query(ctx, q, provinceCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.Street
	for rows.Next() {
		var s domain.Street
		if err := rows.Scan(&s.ID, &s.ProvinceCode, &s.DistrictCode, &s.Name, &s.NameFull,
			&s.NameEn, &s.Slug, &s.Gid, &s.LengthM); err != nil {
			return nil, err
		}
		out = append(out, s)
	}
	return out, rows.Err()
}

func (r *geoRepo) CountSeeded(ctx context.Context) (p, d, w int64, err error) {
	err = r.pool.QueryRow(ctx, `
		SELECT (SELECT count(*) FROM provinces),
		       (SELECT count(*) FROM districts),
		       (SELECT count(*) FROM wards)`).Scan(&p, &d, &w)
	return
}

func (r *geoRepo) CheckWhitelist(ctx context.Context, pt domain.Point, toleranceM float64) (*domain.WhitelistResult, error) {
	loc, err := r.LocateChain(ctx, pt)
	if err != nil {
		return nil, err
	}
	if !loc.Found && toleranceM > 0 {
		loc, err = r.NearestProvince(ctx, pt, toleranceM)
		if err != nil {
			return nil, err
		}
	}

	res := &domain.WhitelistResult{
		Province: loc.Province,
		District: loc.District,
		Ward:     loc.Ward,
		Street:   loc.Street,
		Point:    pt,
	}

	if loc.Province != nil {
		var exists bool
		err = r.pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM whitelist_cities WHERE province_code=$1)`,
			loc.Province.Code,
		).Scan(&exists)
		if err != nil {
			return nil, fmt.Errorf("check whitelist: %w", err)
		}
		res.Whitelisted = exists
	}

	return res, nil
}

func (r *geoRepo) ListWhitelistCities(ctx context.Context) ([]domain.WhitelistCity, error) {
	const q = `SELECT province_code, COALESCE(note,'') FROM whitelist_cities ORDER BY province_code`
	rows, err := r.pool.Query(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []domain.WhitelistCity
	for rows.Next() {
		var c domain.WhitelistCity
		if err := rows.Scan(&c.ProvinceCode, &c.Note); err != nil {
			return nil, err
		}
		out = append(out, c)
	}
	return out, rows.Err()
}
