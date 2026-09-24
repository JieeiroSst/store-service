package usecase

import (
	"context"
	"strings"
	"time"

	"github.com/JIeeiroSst/car-rental-service/internal/repository"
	"github.com/JIeeiroSst/car-rental-service/model"
	"github.com/google/uuid"
)

type ProcessPaymentInput struct {
	RentalID, UserID string
	Amount           float64
	Method           model.PaymentMethod
	TransactionID    string
}

func (u *Usecase) ProcessPayment(ctx context.Context, in ProcessPaymentInput) (*model.Payment, error) {
	rentalID, err := parseID("rental_id", in.RentalID)
	if err != nil {
		return nil, err
	}
	userID, err := parseID("user_id", in.UserID)
	if err != nil {
		return nil, err
	}
	if in.Amount <= 0 {
		return nil, invalid("amount must be positive")
	}
	if in.Method == "" {
		return nil, invalid("payment_method is required")
	}

	rt0, err := u.repos.Rentals.GetByID(ctx, rentalID)
	if err != nil {
		return nil, err
	}
	if rt0.UserID != userID {
		return nil, model.ErrPermissionDenied
	}

	status, err := u.gateway.Charge(ctx, ChargeRequest{Method: in.Method, Amount: in.Amount, TransactionID: in.TransactionID})
	if err != nil {
		return nil, err
	}

	p := &model.Payment{
		ID:            uuid.New(),
		RentalID:      rentalID,
		UserID:        userID,
		Amount:        in.Amount,
		PaymentMethod: in.Method,
		TransactionID: in.TransactionID,
		PaymentStatus: status,
		PaymentDate:   time.Now(),
	}
	err = u.repos.Transaction(ctx, func(tx *repository.Repositories) error {
		rt, err := tx.Rentals.GetByIDForUpdate(ctx, rentalID)
		if err != nil {
			return err
		}
		if rt.UserID != userID {
			return model.ErrPermissionDenied
		}
		if err := tx.Payments.Create(ctx, p); err != nil {
			return err
		}
		paid, err := tx.Payments.SumCompleted(ctx, rentalID)
		if err != nil {
			return err
		}
		if rt.Status == model.RentalStatusCompleted {
			if st := paymentStatusFor(paid, rt.BaseFee+rt.AdditionalFees); st != rt.PaymentStatus {
				rt.PaymentStatus = st
				return tx.Rentals.Update(ctx, rt)
			}
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return p, nil
}

type SubmitReviewInput struct {
	RentalID, UserID, VehicleID string
	Rating                      int
	Comment                     string
}

func (u *Usecase) SubmitReview(ctx context.Context, in SubmitReviewInput) (*model.Review, error) {
	rentalID, err := parseID("rental_id", in.RentalID)
	if err != nil {
		return nil, err
	}
	userID, err := parseID("user_id", in.UserID)
	if err != nil {
		return nil, err
	}
	vehicleID, err := parseID("vehicle_id", in.VehicleID)
	if err != nil {
		return nil, err
	}
	if in.Rating < 1 || in.Rating > 5 {
		return nil, invalid("rating must be between 1 and 5")
	}

	rt, err := u.repos.Rentals.GetByID(ctx, rentalID)
	if err != nil {
		return nil, err
	}
	if rt.UserID != userID {
		return nil, model.ErrPermissionDenied
	}
	if rt.VehicleID != vehicleID {
		return nil, invalid("vehicle_id does not match the rental")
	}
	if rt.Status != model.RentalStatusCompleted {
		return nil, precondition("only completed rentals can be reviewed")
	}
	if done, err := u.repos.Reviews.ExistsForRental(ctx, rentalID); err != nil {
		return nil, err
	} else if done {
		return nil, model.ErrAlreadyExists
	}
	rv := &model.Review{
		ID:        uuid.New(),
		RentalID:  rentalID,
		UserID:    userID,
		VehicleID: vehicleID,
		Rating:    in.Rating,
		Comment:   strings.TrimSpace(in.Comment),
	}
	if err := u.repos.Reviews.Create(ctx, rv); err != nil {
		return nil, err
	}
	return rv, nil
}

type ReviewPage struct {
	Reviews       []model.Review
	Total         int64
	NextPageToken string
}

func (u *Usecase) ListVehicleReviews(ctx context.Context, vehicleID string, page Page) (*ReviewPage, error) {
	vid, err := parseID("vehicle_id", vehicleID)
	if err != nil {
		return nil, err
	}
	rs, total, err := u.repos.Reviews.ListByVehicle(ctx, vid, page.Offset, page.Limit)
	if err != nil {
		return nil, err
	}
	return &ReviewPage{Reviews: rs, Total: total, NextPageToken: page.NextToken(len(rs), total)}, nil
}
