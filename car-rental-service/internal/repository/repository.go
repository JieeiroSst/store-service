package repository

import (
	"context"
	"errors"

	"github.com/JIeeiroSst/car-rental-service/model"
	"gorm.io/gorm"
)

type Repositories struct {
	db           *gorm.DB
	Users        *UserRepo
	Locations    *LocationRepo
	Vehicles     *VehicleRepo
	Reservations *ReservationRepo
	Rentals      *RentalRepo
	Payments     *PaymentRepo
	Reviews      *ReviewRepo
}

func NewRepositories(db *gorm.DB) *Repositories {
	return &Repositories{
		db:           db,
		Users:        &UserRepo{db},
		Locations:    &LocationRepo{db},
		Vehicles:     &VehicleRepo{db},
		Reservations: &ReservationRepo{db},
		Rentals:      &RentalRepo{db},
		Payments:     &PaymentRepo{db},
		Reviews:      &ReviewRepo{db},
	}
}

func (r *Repositories) Transaction(ctx context.Context, fn func(tx *Repositories) error) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return fn(NewRepositories(tx))
	})
}

func wrapNotFound(err error) error {
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrNotFound
	}
	return err
}
