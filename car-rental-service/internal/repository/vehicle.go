package repository

import (
	"context"
	"time"

	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type VehicleRepo struct{ db *gorm.DB }

type VehicleFilter struct {
	VehicleType *model.VehicleType
	CategoryID  *uuid.UUID
	LocationID  *uuid.UUID
	Available   bool
	From, To    time.Time
}

func (r *VehicleRepo) Create(ctx context.Context, v *model.Vehicle) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Create(v).Error; err != nil {
			return err
		}
		switch {
		case v.CarDetails != nil:
			v.CarDetails.ID, v.CarDetails.VehicleID = uuid.New(), v.ID
			return tx.Create(v.CarDetails).Error
		case v.MotorcycleDetails != nil:
			v.MotorcycleDetails.ID, v.MotorcycleDetails.VehicleID = uuid.New(), v.ID
			return tx.Create(v.MotorcycleDetails).Error
		case v.BicycleDetails != nil:
			v.BicycleDetails.ID, v.BicycleDetails.VehicleID = uuid.New(), v.ID
			return tx.Create(v.BicycleDetails).Error
		}
		return nil
	})
}

func (r *VehicleRepo) GetByID(ctx context.Context, id uuid.UUID) (*model.Vehicle, error) {
	var v model.Vehicle
	if err := r.db.WithContext(ctx).First(&v, "vehicle_id = ?", id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	vs := []model.Vehicle{v}
	if err := r.attach(ctx, vs); err != nil {
		return nil, err
	}
	return &vs[0], nil
}

func (r *VehicleRepo) GetByIDForUpdate(ctx context.Context, id uuid.UUID) (*model.Vehicle, error) {
	var v model.Vehicle
	err := r.db.WithContext(ctx).Clauses(clause.Locking{Strength: "UPDATE"}).First(&v, "vehicle_id = ?", id).Error
	if err != nil {
		return nil, wrapNotFound(err)
	}
	return &v, nil
}

func (r *VehicleRepo) Update(ctx context.Context, v *model.Vehicle) error {
	return r.db.WithContext(ctx).Save(v).Error
}

func (r *VehicleRepo) List(ctx context.Context, f VehicleFilter, offset, limit int) ([]model.Vehicle, int64, error) {
	q := r.db.WithContext(ctx).Model(&model.Vehicle{})
	if f.VehicleType != nil {
		q = q.Where("category_id IN (?)",
			r.db.Model(&model.VehicleCategory{}).Select("category_id").Where("vehicle_type = ?", *f.VehicleType))
	}
	if f.CategoryID != nil {
		q = q.Where("category_id = ?", *f.CategoryID)
	}
	if f.LocationID != nil {
		q = q.Where("location_id = ?", *f.LocationID)
	}
	if f.Available {
		q = q.Where("status NOT IN ?", []model.VehicleStatus{model.VehicleStatusMaintenance, model.VehicleStatusRetired}).
			Where("NOT EXISTS (?)", r.db.Model(&model.Reservation{}).Select("1").
				Where("reservations.vehicle_id = vehicles.vehicle_id AND status IN ? AND start_time < ? AND end_time > ?",
					[]model.ReservationStatus{model.ReservationStatusPending, model.ReservationStatusConfirmed}, f.To, f.From)).
			Where("NOT EXISTS (?)", r.db.Model(&model.Rental{}).Select("1").
				Where("rentals.vehicle_id = vehicles.vehicle_id AND status IN ?",
					[]model.RentalStatus{model.RentalStatusActive, model.RentalStatusOverdue}))
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []model.Vehicle
	if err := q.Order("created_at, vehicle_id").Offset(offset).Limit(limit).Find(&out).Error; err != nil {
		return nil, 0, err
	}
	return out, total, r.attach(ctx, out)
}

func (r *VehicleRepo) attach(ctx context.Context, vs []model.Vehicle) error {
	if len(vs) == 0 {
		return nil
	}
	ids := make([]uuid.UUID, len(vs))
	catIDs := make([]uuid.UUID, len(vs))
	for i, v := range vs {
		ids[i], catIDs[i] = v.ID, v.CategoryID
	}
	db := r.db.WithContext(ctx)

	var cats []model.VehicleCategory
	if err := db.Find(&cats, "category_id IN ?", catIDs).Error; err != nil {
		return err
	}
	catByID := make(map[uuid.UUID]*model.VehicleCategory, len(cats))
	for i := range cats {
		catByID[cats[i].ID] = &cats[i]
	}

	var cars []model.CarDetails
	var motos []model.MotorcycleDetails
	var bikes []model.BicycleDetails
	if err := db.Find(&cars, "vehicle_id IN ?", ids).Error; err != nil {
		return err
	}
	if err := db.Find(&motos, "vehicle_id IN ?", ids).Error; err != nil {
		return err
	}
	if err := db.Find(&bikes, "vehicle_id IN ?", ids).Error; err != nil {
		return err
	}
	carBy := map[uuid.UUID]*model.CarDetails{}
	motoBy := map[uuid.UUID]*model.MotorcycleDetails{}
	bikeBy := map[uuid.UUID]*model.BicycleDetails{}
	for i := range cars {
		carBy[cars[i].VehicleID] = &cars[i]
	}
	for i := range motos {
		motoBy[motos[i].VehicleID] = &motos[i]
	}
	for i := range bikes {
		bikeBy[bikes[i].VehicleID] = &bikes[i]
	}
	for i := range vs {
		vs[i].Category = catByID[vs[i].CategoryID]
		vs[i].CarDetails = carBy[vs[i].ID]
		vs[i].MotorcycleDetails = motoBy[vs[i].ID]
		vs[i].BicycleDetails = bikeBy[vs[i].ID]
	}
	return nil
}

func (r *VehicleRepo) GetCategory(ctx context.Context, id uuid.UUID) (*model.VehicleCategory, error) {
	var c model.VehicleCategory
	if err := r.db.WithContext(ctx).First(&c, "category_id = ?", id).Error; err != nil {
		return nil, wrapNotFound(err)
	}
	return &c, nil
}

func (r *VehicleRepo) RegistrationExists(ctx context.Context, reg string, excludeID uuid.UUID) (bool, error) {
	q := r.db.WithContext(ctx).Model(&model.Vehicle{}).Where("registration_number = ?", reg)
	if excludeID != uuid.Nil {
		q = q.Where("vehicle_id <> ?", excludeID)
	}
	var n int64
	err := q.Count(&n).Error
	return n > 0, err
}
