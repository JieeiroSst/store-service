package usecase

import (
	"context"
	"math"
	"time"

	"github.com/JIeeiroSst/car-rental-service/internal/repository"
	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
)

type RentalPage struct {
	Rentals       []model.Rental
	Total         int64
	NextPageToken string
}

type StartRentalInput struct {
	ReservationID    string
	PickupMileage    float64
	StaffID          string
	PickupLocationID string
}

func (u *Usecase) StartRental(ctx context.Context, in StartRentalInput) (*model.Rental, error) {
	resID, err := parseID("reservation_id", in.ReservationID)
	if err != nil {
		return nil, err
	}
	if _, err := u.requireStaff(ctx, in.StaffID); err != nil {
		return nil, err
	}
	var out *model.Rental
	err = u.repos.Transaction(ctx, func(tx *repository.Repositories) error {
		res, err := tx.Reservations.GetByID(ctx, resID)
		if err != nil {
			return err
		}
		v, err := tx.Vehicles.GetByIDForUpdate(ctx, res.VehicleID)
		if err != nil {
			return err
		}
		// Re-read now that the vehicle is locked.
		if res, err = tx.Reservations.GetByID(ctx, resID); err != nil {
			return err
		}
		if res.Status != model.ReservationStatusConfirmed {
			return precondition("reservation is %s, only confirmed reservations can be started", res.Status)
		}
		if started, err := tx.Rentals.ExistsForReservation(ctx, resID); err != nil {
			return err
		} else if started {
			return precondition("reservation has already been started")
		}
		if v.Status != model.VehicleStatusAvailable {
			return precondition("vehicle is %s", v.Status)
		}

		pickupLoc := res.PickupLocationID
		if in.PickupLocationID != "" {
			if pickupLoc, err = parseID("pickup_location_id", in.PickupLocationID); err != nil {
				return err
			}
			if err := checkLocations(ctx, tx, pickupLoc); err != nil {
				return err
			}
		}
		mileage := in.PickupMileage
		if mileage <= 0 {
			mileage = v.Mileage
		}
		if mileage < v.Mileage {
			return invalid("pickup_mileage is below the vehicle's recorded mileage")
		}

		now := time.Now()
		rt := &model.Rental{
			ID:               uuid.New(),
			ReservationID:    &res.ID,
			VehicleID:        v.ID,
			UserID:           res.UserID,
			PickupTime:       now,
			PickupLocationID: pickupLoc,
			PickupMileage:    mileage,
			Status:           model.RentalStatusActive,
			BaseFee:          rentalFee(res.StartTime, res.EndTime, v.HourlyRate, v.DailyRate),
			PaymentStatus:    model.PaymentStatusPending,
		}
		if err := tx.Rentals.Create(ctx, rt); err != nil {
			return err
		}
		v.Status, v.Mileage = model.VehicleStatusRented, mileage
		out = rt
		return tx.Vehicles.Update(ctx, v)
	})
	return out, err
}

type CompleteRentalInput struct {
	RentalID         string
	ReturnMileage    float64
	StaffID          string
	ReturnLocationID string
	AdditionalFees   map[string]float64
	Notes            string
}

func (u *Usecase) CompleteRental(ctx context.Context, in CompleteRentalInput) (*model.Rental, error) {
	rentalID, err := parseID("rental_id", in.RentalID)
	if err != nil {
		return nil, err
	}
	if _, err := u.requireStaff(ctx, in.StaffID); err != nil {
		return nil, err
	}
	var extra float64
	for name, fee := range in.AdditionalFees {
		if fee < 0 {
			return nil, invalid("additional fee %q must not be negative", name)
		}
		extra += fee
	}

	var out *model.Rental
	err = u.repos.Transaction(ctx, func(tx *repository.Repositories) error {
		rt, err := tx.Rentals.GetByIDForUpdate(ctx, rentalID)
		if err != nil {
			return err
		}
		if rt.Status == model.RentalStatusCompleted {
			return precondition("rental is already completed")
		}
		if in.ReturnMileage < rt.PickupMileage {
			return invalid("return_mileage must not be below the pickup mileage")
		}
		v, err := tx.Vehicles.GetByIDForUpdate(ctx, rt.VehicleID)
		if err != nil {
			return err
		}
		var res *model.Reservation
		if rt.ReservationID != nil {
			if res, err = tx.Reservations.GetByID(ctx, *rt.ReservationID); err != nil {
				return err
			}
		}
		returnLoc := rt.PickupLocationID
		switch {
		case in.ReturnLocationID != "":
			if returnLoc, err = parseID("return_location_id", in.ReturnLocationID); err != nil {
				return err
			}
		case res != nil:
			returnLoc = res.ReturnLocationID
		}
		if err := checkLocations(ctx, tx, returnLoc); err != nil {
			return err
		}

		now := time.Now()
		rt.ActualReturnTime, rt.ReturnLocationID, rt.ReturnMileage = &now, &returnLoc, &in.ReturnMileage
		rt.AdditionalFees = extra
		rt.Notes = in.Notes
		if res != nil && now.After(res.EndTime) {
			rt.AdditionalFees += u.pricing.lateFee(now.Sub(res.EndTime), v.HourlyRate)
		}
		rt.Status = model.RentalStatusCompleted
		paid, err := tx.Payments.SumCompleted(ctx, rt.ID)
		if err != nil {
			return err
		}
		rt.PaymentStatus = paymentStatusFor(paid, rt.BaseFee+rt.AdditionalFees)
		if err := tx.Rentals.Update(ctx, rt); err != nil {
			return err
		}

		v.Status, v.Mileage, v.LocationID = model.VehicleStatusAvailable, in.ReturnMileage, returnLoc
		if err := tx.Vehicles.Update(ctx, v); err != nil {
			return err
		}
		if res != nil {
			res.Status = model.ReservationStatusCompleted
			if err := tx.Reservations.Update(ctx, res); err != nil {
				return err
			}
		}
		out = rt
		return nil
	})
	return out, err
}

func (u *Usecase) GetRental(ctx context.Context, id string) (*model.Rental, error) {
	rid, err := parseID("rental_id", id)
	if err != nil {
		return nil, err
	}
	rt, err := u.repos.Rentals.GetByID(ctx, rid)
	if err != nil {
		return nil, err
	}
	rt.Vehicle, err = u.repos.Vehicles.GetByID(ctx, rt.VehicleID)
	return rt, err
}

func (u *Usecase) ListUserRentals(ctx context.Context, userID string, status model.RentalStatus, page Page) (*RentalPage, error) {
	uid, err := parseID("user_id", userID)
	if err != nil {
		return nil, err
	}
	var st *model.RentalStatus
	if status != "" {
		st = &status
	}
	rs, total, err := u.repos.Rentals.ListByUser(ctx, uid, st, page.Offset, page.Limit)
	if err != nil {
		return nil, err
	}
	return &RentalPage{Rentals: rs, Total: total, NextPageToken: page.NextToken(len(rs), total)}, nil
}

func rentalFee(start, end time.Time, hourly, daily float64) float64 {
	hours := int(math.Ceil(end.Sub(start).Hours()))
	if hours < 1 {
		hours = 1
	}
	days, rem := hours/24, hours%24
	return float64(days)*daily + math.Min(float64(rem)*hourly, daily)
}

func paymentStatusFor(paid, due float64) model.PaymentStatus {
	if paid >= due && due > 0 {
		return model.PaymentStatusCompleted
	}
	return model.PaymentStatusPending
}
