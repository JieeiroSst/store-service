package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/geoservice/internal/core/domain"
	"github.com/geoservice/internal/core/port"
)

var ErrInvalidCoordinate = errors.New("invalid coordinate")

// Vietnam bounding box (approx) used as a cheap sanity gate.
const (
	vnMinLng, vnMaxLng = 102.0, 110.0
	vnMinLat, vnMaxLat = 8.0, 24.0
)

type locatorService struct {
	repo port.GeoRepository
}

func NewLocatorService(repo port.GeoRepository) port.LocatorService {
	return &locatorService{repo: repo}
}

func (s *locatorService) Locate(ctx context.Context, p domain.Point, toleranceMeters float64) (*domain.LocationResult, error) {
	if p.Lng < vnMinLng || p.Lng > vnMaxLng || p.Lat < vnMinLat || p.Lat > vnMaxLat {
		return nil, fmt.Errorf("%w: point (%f,%f) outside Vietnam bbox", ErrInvalidCoordinate, p.Lng, p.Lat)
	}

	res, err := s.repo.LocateChain(ctx, p)
	if err != nil {
		return nil, err
	}
	if res.Found {
		return res, nil
	}
	if toleranceMeters > 0 {
		return s.repo.NearestProvince(ctx, p, toleranceMeters)
	}
	return res, nil
}

func (s *locatorService) GetProvince(ctx context.Context, code string) (*domain.Province, error) {
	return s.repo.GetProvince(ctx, code)
}

func (s *locatorService) ListProvinces(ctx context.Context) ([]domain.Province, error) {
	return s.repo.ListProvinces(ctx)
}

func (s *locatorService) ListDistricts(ctx context.Context, provinceCode string) ([]domain.District, error) {
	return s.repo.ListDistricts(ctx, provinceCode)
}

func (s *locatorService) ListWards(ctx context.Context, districtCode string) ([]domain.Ward, error) {
	return s.repo.ListWards(ctx, districtCode)
}

func (s *locatorService) ListStreets(ctx context.Context, provinceCode string) ([]domain.Street, error) {
	return s.repo.ListStreets(ctx, provinceCode)
}

func (s *locatorService) CountSeeded(ctx context.Context) (int64, int64, int64, error) {
	return s.repo.CountSeeded(ctx)
}

func (s *locatorService) CheckWhitelist(ctx context.Context, p domain.Point, toleranceMeters float64) (*domain.WhitelistResult, error) {
	if p.Lng < vnMinLng || p.Lng > vnMaxLng || p.Lat < vnMinLat || p.Lat > vnMaxLat {
		return nil, fmt.Errorf("%w: point (%f,%f) outside Vietnam bbox", ErrInvalidCoordinate, p.Lng, p.Lat)
	}
	return s.repo.CheckWhitelist(ctx, p, toleranceMeters)
}

func (s *locatorService) ListWhitelistCities(ctx context.Context) ([]domain.WhitelistCity, error) {
	return s.repo.ListWhitelistCities(ctx)
}
