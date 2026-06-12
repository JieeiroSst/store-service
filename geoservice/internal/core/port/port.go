package port

import (
	"context"

	"github.com/geoservice/internal/core/domain"
)

// GeoRepository is the driven (outbound) port for persistence + spatial queries.
type GeoRepository interface {
	// LocateChain resolves a point to its full province→district→ward chain
	// using point-in-polygon at each level (needs geometry to be attached).
	LocateChain(ctx context.Context, p domain.Point) (*domain.LocationResult, error)
	// NearestProvince finds the closest province within toleranceMeters when a
	// point falls outside every polygon (coastal/border points).
	NearestProvince(ctx context.Context, p domain.Point, toleranceMeters float64) (*domain.LocationResult, error)

	// AttachGeometryByName attaches a GADM polygon to a seeded row, matching by
	// normalized name within parent scope. Returns true if a row was updated.
	AttachGeometryByName(ctx context.Context, f domain.GeoFeature) (bool, error)

	// Read helpers.
	GetProvince(ctx context.Context, code string) (*domain.Province, error)
	ListProvinces(ctx context.Context) ([]domain.Province, error)
	ListDistricts(ctx context.Context, provinceCode string) ([]domain.District, error)
	ListWards(ctx context.Context, districtCode string) ([]domain.Ward, error)
	// CountSeeded returns row counts per level (for diagnostics / health).
	CountSeeded(ctx context.Context) (provinces, districts, wards int64, err error)

	ListStreets(ctx context.Context, provinceCode string) ([]domain.Street, error)

	// CheckWhitelist resolves the full location chain for a point and reports
	// whether the province belongs to the priority whitelist.
	CheckWhitelist(ctx context.Context, p domain.Point, toleranceMeters float64) (*domain.WhitelistResult, error)

	// ListWhitelistCities returns all cities in the priority whitelist.
	ListWhitelistCities(ctx context.Context) ([]domain.WhitelistCity, error)
}

// LocatorService is the driving (inbound) port used by handlers.
type LocatorService interface {
	Locate(ctx context.Context, p domain.Point, toleranceMeters float64) (*domain.LocationResult, error)
	GetProvince(ctx context.Context, code string) (*domain.Province, error)
	ListProvinces(ctx context.Context) ([]domain.Province, error)
	ListDistricts(ctx context.Context, provinceCode string) ([]domain.District, error)
	ListWards(ctx context.Context, districtCode string) ([]domain.Ward, error)
	ListStreets(ctx context.Context, provinceCode string) ([]domain.Street, error)
	CountSeeded(ctx context.Context) (provinces, districts, wards int64, err error)

	// CheckWhitelist resolves the full location chain for a point and reports
	// whether the province belongs to the priority whitelist.
	CheckWhitelist(ctx context.Context, p domain.Point, toleranceMeters float64) (*domain.WhitelistResult, error)

	// ListWhitelistCities returns all cities in the priority whitelist.
	ListWhitelistCities(ctx context.Context) ([]domain.WhitelistCity, error)
}
