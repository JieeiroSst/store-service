package application

import (
	"context"

	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

type locationService struct {
	directory port.LocationDirectory
}

func NewLocationService(directory port.LocationDirectory) port.LocationUsecase {
	return &locationService{directory: directory}
}

func (s *locationService) Provinces(ctx context.Context) ([]port.Province, error) {
	return s.directory.Provinces(ctx)
}

func (s *locationService) Districts(ctx context.Context, provinceID int) ([]port.District, error) {
	if provinceID <= 0 {
		return nil, invalid("province_id is required")
	}
	return s.directory.Districts(ctx, provinceID)
}

func (s *locationService) Wards(ctx context.Context, districtID int) ([]port.Ward, error) {
	if districtID <= 0 {
		return nil, invalid("district_id is required")
	}
	return s.directory.Wards(ctx, districtID)
}

func validateVNLocation(ctx context.Context, dir port.LocationDirectory, provinceID, districtID int, wardCode string) error {
	wards, err := dir.Wards(ctx, districtID)
	if err != nil {
		return err
	}
	found := false
	for _, w := range wards {
		if w.Code == wardCode {
			found = true
			break
		}
	}
	if !found {
		return invalid("ward_code %s is not a ward of district %d in Vietnam", wardCode, districtID)
	}
	if provinceID <= 0 {
		return nil
	}
	districts, err := dir.Districts(ctx, provinceID)
	if err != nil {
		return err
	}
	for _, d := range districts {
		if d.ID == districtID {
			return nil
		}
	}
	return invalid("district %d is not in province %d", districtID, provinceID)
}
