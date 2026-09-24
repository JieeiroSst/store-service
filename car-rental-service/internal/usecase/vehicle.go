package usecase

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/JIeeiroSst/car-rental-service/internal/repository"
	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
)

type VehiclePage struct {
	Vehicles      []model.Vehicle
	Total         int64
	NextPageToken string
}

func (u *Usecase) CreateVehicle(ctx context.Context, v *model.Vehicle) (*model.Vehicle, error) {
	if strings.TrimSpace(v.Make) == "" || strings.TrimSpace(v.Model) == "" {
		return nil, invalid("make and model are required")
	}
	if v.Year < 1900 || v.Year > time.Now().Year()+1 {
		return nil, invalid("year is out of range")
	}
	if v.HourlyRate <= 0 || v.DailyRate <= 0 {
		return nil, invalid("hourly_rate and daily_rate must be positive")
	}
	if v.Mileage < 0 {
		return nil, invalid("mileage must not be negative")
	}
	if v.Status == "" {
		v.Status = model.VehicleStatusAvailable
	}

	cat, err := u.repos.Vehicles.GetCategory(ctx, v.CategoryID)
	if err != nil {
		return nil, notFoundAs(err, "category")
	}
	if _, err := u.repos.Locations.GetByID(ctx, v.LocationID); err != nil {
		return nil, notFoundAs(err, "location")
	}
	switch cat.VehicleType {
	case model.VehicleTypeCar:
		if v.CarDetails == nil || v.MotorcycleDetails != nil || v.BicycleDetails != nil {
			return nil, invalid("car_details are required for a car category")
		}
	case model.VehicleTypeMotorcycle:
		if v.MotorcycleDetails == nil || v.CarDetails != nil || v.BicycleDetails != nil {
			return nil, invalid("motorcycle_details are required for a motorcycle category")
		}
	case model.VehicleTypeBicycle:
		if v.BicycleDetails == nil || v.CarDetails != nil || v.MotorcycleDetails != nil {
			return nil, invalid("bicycle_details are required for a bicycle category")
		}
	}
	if v.RegistrationNumber != "" {
		exists, err := u.repos.Vehicles.RegistrationExists(ctx, v.RegistrationNumber, uuid.Nil)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, model.ErrAlreadyExists
		}
	}

	v.ID = uuid.New()
	if err := u.repos.Vehicles.Create(ctx, v); err != nil {
		return nil, err
	}
	return u.repos.Vehicles.GetByID(ctx, v.ID)
}

func (u *Usecase) GetVehicle(ctx context.Context, id string) (*model.Vehicle, error) {
	vid, err := parseID("vehicle_id", id)
	if err != nil {
		return nil, err
	}
	return u.repos.Vehicles.GetByID(ctx, vid)
}

type VehiclePatch struct {
	VehicleID          string
	RegistrationNumber string
	Mileage            float64
	Status             model.VehicleStatus
	HourlyRate         float64
	DailyRate          float64
	LocationID         string
	Features           map[string]string
}

func (u *Usecase) UpdateVehicle(ctx context.Context, p VehiclePatch) (*model.Vehicle, error) {
	vid, err := parseID("vehicle_id", p.VehicleID)
	if err != nil {
		return nil, err
	}
	if p.Mileage < 0 || p.HourlyRate < 0 || p.DailyRate < 0 {
		return nil, invalid("mileage and rates must not be negative")
	}

	err = u.repos.Transaction(ctx, func(tx *repository.Repositories) error {
		v, err := tx.Vehicles.GetByIDForUpdate(ctx, vid)
		if err != nil {
			return err
		}
		if p.RegistrationNumber != "" && p.RegistrationNumber != v.RegistrationNumber {
			exists, err := tx.Vehicles.RegistrationExists(ctx, p.RegistrationNumber, vid)
			if err != nil {
				return err
			}
			if exists {
				return model.ErrAlreadyExists
			}
			v.RegistrationNumber = p.RegistrationNumber
		}
		if p.Mileage > 0 {
			if p.Mileage < v.Mileage {
				return invalid("mileage cannot decrease")
			}
			v.Mileage = p.Mileage
		}
		if p.Status != "" {
			v.Status = p.Status
		}
		if p.HourlyRate > 0 {
			v.HourlyRate = p.HourlyRate
		}
		if p.DailyRate > 0 {
			v.DailyRate = p.DailyRate
		}
		if p.LocationID != "" {
			lid, err := parseID("location_id", p.LocationID)
			if err != nil {
				return err
			}
			if _, err := tx.Locations.GetByID(ctx, lid); err != nil {
				return notFoundAs(err, "location")
			}
			v.LocationID = lid
		}
		if p.Features != nil {
			v.Features = p.Features
		}
		return tx.Vehicles.Update(ctx, v)
	})
	if err != nil {
		return nil, err
	}
	return u.repos.Vehicles.GetByID(ctx, vid)
}

func (u *Usecase) ListVehicles(ctx context.Context, vt model.VehicleType, locationID string, page Page) (*VehiclePage, error) {
	f := repository.VehicleFilter{}
	if vt != "" {
		f.VehicleType = &vt
	}
	if locationID != "" {
		lid, err := parseID("location_id", locationID)
		if err != nil {
			return nil, err
		}
		f.LocationID = &lid
	}
	return u.listVehicles(ctx, f, page)
}

type SearchVehiclesInput struct {
	Start, End                   time.Time
	PickupLocationID, CategoryID string
	ReturnLocationID             string
	VehicleType                  model.VehicleType
}

func (u *Usecase) SearchAvailableVehicles(ctx context.Context, in SearchVehiclesInput, page Page) (*VehiclePage, error) {
	if in.Start.IsZero() || in.End.IsZero() || !in.End.After(in.Start) {
		return nil, invalid("start_time and end_time are required and end_time must be after start_time")
	}
	f := repository.VehicleFilter{Available: true, From: in.Start, To: in.End}
	if in.VehicleType != "" {
		f.VehicleType = &in.VehicleType
	}
	if in.PickupLocationID != "" {
		id, err := parseID("pickup_location_id", in.PickupLocationID)
		if err != nil {
			return nil, err
		}
		f.LocationID = &id
	}
	if in.CategoryID != "" {
		id, err := parseID("category_id", in.CategoryID)
		if err != nil {
			return nil, err
		}
		f.CategoryID = &id
	}
	return u.listVehicles(ctx, f, page)
}

func (u *Usecase) listVehicles(ctx context.Context, f repository.VehicleFilter, page Page) (*VehiclePage, error) {
	vs, total, err := u.repos.Vehicles.List(ctx, f, page.Offset, page.Limit)
	if err != nil {
		return nil, err
	}
	return &VehiclePage{Vehicles: vs, Total: total, NextPageToken: page.NextToken(len(vs), total)}, nil
}

func notFoundAs(err error, what string) error {
	if errors.Is(err, model.ErrNotFound) {
		return invalid("%s does not exist", what)
	}
	return err
}
