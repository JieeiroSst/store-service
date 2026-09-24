package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/JIeeiroSst/car-rental-service/internal/repository"
	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
)

type ReservationPage struct {
	Reservations  []model.Reservation
	Total         int64
	NextPageToken string
}

type CreateReservationInput struct {
	UserID, VehicleID                  string
	PickupLocationID, ReturnLocationID string
	Start, End                         time.Time
}

func (u *Usecase) CreateReservation(ctx context.Context, in CreateReservationInput) (*model.Reservation, error) {
	userID, err := parseID("user_id", in.UserID)
	if err != nil {
		return nil, err
	}
	vehicleID, err := parseID("vehicle_id", in.VehicleID)
	if err != nil {
		return nil, err
	}
	pickupID, err := parseID("pickup_location_id", in.PickupLocationID)
	if err != nil {
		return nil, err
	}
	returnID := pickupID
	if in.ReturnLocationID != "" {
		if returnID, err = parseID("return_location_id", in.ReturnLocationID); err != nil {
			return nil, err
		}
	}
	if err := validateWindow(in.Start, in.End); err != nil {
		return nil, err
	}

	res := &model.Reservation{
		ID:               uuid.New(),
		UserID:           userID,
		VehicleID:        vehicleID,
		PickupLocationID: pickupID,
		ReturnLocationID: returnID,
		StartTime:        in.Start,
		EndTime:          in.End,
		Status:           model.ReservationStatusConfirmed,
	}
	err = u.repos.Transaction(ctx, func(tx *repository.Repositories) error {
		if _, err := tx.Users.GetByID(ctx, userID); err != nil {
			return notFoundAs(err, "user")
		}
		if err := checkLocations(ctx, tx, pickupID, returnID); err != nil {
			return err
		}
		v, err := tx.Vehicles.GetByIDForUpdate(ctx, vehicleID)
		if err != nil {
			return notFoundAs(err, "vehicle")
		}
		if v.Status == model.VehicleStatusMaintenance || v.Status == model.VehicleStatusRetired {
			return precondition("vehicle is not available for rent")
		}
		return u.saveIfFree(ctx, tx, res, uuid.Nil, true)
	})
	if err != nil {
		return nil, err
	}
	return res, nil
}

type UpdateReservationInput struct {
	ReservationID                      string
	PickupLocationID, ReturnLocationID string
	Start, End                         time.Time
}

func (u *Usecase) UpdateReservation(ctx context.Context, in UpdateReservationInput) (*model.Reservation, error) {
	id, err := parseID("reservation_id", in.ReservationID)
	if err != nil {
		return nil, err
	}
	var out *model.Reservation
	err = u.repos.Transaction(ctx, func(tx *repository.Repositories) error {
		res, err := tx.Reservations.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if _, err := tx.Vehicles.GetByIDForUpdate(ctx, res.VehicleID); err != nil {
			return err
		}
		if res, err = tx.Reservations.GetByID(ctx, id); err != nil {
			return err
		}
		if !modifiable(res) {
			return precondition("reservation is %s and can no longer be changed", res.Status)
		}
		if in.PickupLocationID != "" {
			if res.PickupLocationID, err = parseID("pickup_location_id", in.PickupLocationID); err != nil {
				return err
			}
		}
		if in.ReturnLocationID != "" {
			if res.ReturnLocationID, err = parseID("return_location_id", in.ReturnLocationID); err != nil {
				return err
			}
		}
		if !in.Start.IsZero() {
			res.StartTime = in.Start
		}
		if !in.End.IsZero() {
			res.EndTime = in.End
		}
		if err := validateWindow(res.StartTime, res.EndTime); err != nil {
			return err
		}
		if err := checkLocations(ctx, tx, res.PickupLocationID, res.ReturnLocationID); err != nil {
			return err
		}
		if err := u.saveIfFree(ctx, tx, res, res.ID, false); err != nil {
			return err
		}
		out = res
		return nil
	})
	return out, err
}

func (u *Usecase) CancelReservation(ctx context.Context, reservationID, reason string) (*model.Reservation, error) {
	id, err := parseID("reservation_id", reservationID)
	if err != nil {
		return nil, err
	}
	var out *model.Reservation
	err = u.repos.Transaction(ctx, func(tx *repository.Repositories) error {
		res, err := tx.Reservations.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if !modifiable(res) {
			return precondition("reservation is %s and cannot be cancelled", res.Status)
		}
		res.Status = model.ReservationStatusCancelled
		res.CancellationReason = strings.TrimSpace(reason)
		out = res
		return tx.Reservations.Update(ctx, res)
	})
	return out, err
}

func (u *Usecase) GetReservation(ctx context.Context, id string) (*model.Reservation, error) {
	rid, err := parseID("reservation_id", id)
	if err != nil {
		return nil, err
	}
	res, err := u.repos.Reservations.GetByID(ctx, rid)
	if err != nil {
		return nil, err
	}
	res.Vehicle, err = u.repos.Vehicles.GetByID(ctx, res.VehicleID)
	return res, err
}

func (u *Usecase) ListUserReservations(ctx context.Context, userID string, status model.ReservationStatus, page Page) (*ReservationPage, error) {
	uid, err := parseID("user_id", userID)
	if err != nil {
		return nil, err
	}
	var st *model.ReservationStatus
	if status != "" {
		st = &status
	}
	rs, total, err := u.repos.Reservations.ListByUser(ctx, uid, st, page.Offset, page.Limit)
	if err != nil {
		return nil, err
	}
	return &ReservationPage{Reservations: rs, Total: total, NextPageToken: page.NextToken(len(rs), total)}, nil
}

func (u *Usecase) saveIfFree(ctx context.Context, tx *repository.Repositories, res *model.Reservation, exclude uuid.UUID, create bool) error {
	busy, err := tx.Reservations.HasOverlap(ctx, res.VehicleID, res.StartTime, res.EndTime, exclude)
	if err != nil {
		return err
	}
	if busy {
		return precondition("vehicle is already reserved for the requested period")
	}
	if create {
		return tx.Reservations.Create(ctx, res)
	}
	return tx.Reservations.Update(ctx, res)
}

func modifiable(r *model.Reservation) bool {
	return r.Status == model.ReservationStatusPending || r.Status == model.ReservationStatusConfirmed
}

func validateWindow(start, end time.Time) error {
	if start.IsZero() || end.IsZero() {
		return invalid("start_time and end_time are required")
	}
	if !end.After(start) {
		return invalid("end_time must be after start_time")
	}
	if start.Before(time.Now().Add(-time.Minute)) {
		return invalid("start_time must not be in the past")
	}
	return nil
}

func checkLocations(ctx context.Context, tx *repository.Repositories, ids ...uuid.UUID) error {
	for _, id := range ids {
		if _, err := tx.Locations.GetByID(ctx, id); err != nil {
			return notFoundAs(err, "location")
		}
	}
	return nil
}
