package http

import (
	"errors"
	"fmt"
	"time"

	"github.com/JIeeiroSst/car-rental-service/model"
	pb "github.com/JIeeiroSst/lib-gateway/car-rental-servcie/gateway/car-rental-servcie"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type enumMap[P comparable, M ~string] struct {
	toModel map[P]M
	toPb    map[M]P
}

func newEnumMap[P comparable, M ~string](pairs map[P]M) enumMap[P, M] {
	e := enumMap[P, M]{toModel: pairs, toPb: make(map[M]P, len(pairs))}
	for p, m := range pairs {
		e.toPb[m] = p
	}
	return e
}

func (e enumMap[P, M]) model(p P) M { return e.toModel[p] }
func (e enumMap[P, M]) pb(m M) P    { return e.toPb[m] }

var (
	userTypes = newEnumMap(map[pb.UserType]model.UserType{
		pb.UserType_USER_TYPE_CUSTOMER: model.UserTypeCustomer,
		pb.UserType_USER_TYPE_STAFF:    model.UserTypeStaff,
		pb.UserType_USER_TYPE_ADMIN:    model.UserTypeAdmin,
	})
	vehicleTypes = newEnumMap(map[pb.VehicleType]model.VehicleType{
		pb.VehicleType_VEHICLE_TYPE_CAR:        model.VehicleTypeCar,
		pb.VehicleType_VEHICLE_TYPE_MOTORCYCLE: model.VehicleTypeMotorcycle,
		pb.VehicleType_VEHICLE_TYPE_BICYCLE:    model.VehicleTypeBicycle,
	})
	vehicleStatuses = newEnumMap(map[pb.VehicleStatus]model.VehicleStatus{
		pb.VehicleStatus_VEHICLE_STATUS_AVAILABLE:   model.VehicleStatusAvailable,
		pb.VehicleStatus_VEHICLE_STATUS_RENTED:      model.VehicleStatusRented,
		pb.VehicleStatus_VEHICLE_STATUS_MAINTENANCE: model.VehicleStatusMaintenance,
		pb.VehicleStatus_VEHICLE_STATUS_RETIRED:     model.VehicleStatusRetired,
	})
	reservationStatuses = newEnumMap(map[pb.ReservationStatus]model.ReservationStatus{
		pb.ReservationStatus_RESERVATION_STATUS_PENDING:   model.ReservationStatusPending,
		pb.ReservationStatus_RESERVATION_STATUS_CONFIRMED: model.ReservationStatusConfirmed,
		pb.ReservationStatus_RESERVATION_STATUS_CANCELLED: model.ReservationStatusCancelled,
		pb.ReservationStatus_RESERVATION_STATUS_COMPLETED: model.ReservationStatusCompleted,
	})
	rentalStatuses = newEnumMap(map[pb.RentalStatus]model.RentalStatus{
		pb.RentalStatus_RENTAL_STATUS_ACTIVE:    model.RentalStatusActive,
		pb.RentalStatus_RENTAL_STATUS_COMPLETED: model.RentalStatusCompleted,
		pb.RentalStatus_RENTAL_STATUS_OVERDUE:   model.RentalStatusOverdue,
	})
	paymentMethods = newEnumMap(map[pb.PaymentMethod]model.PaymentMethod{
		pb.PaymentMethod_PAYMENT_METHOD_CREDIT_CARD: model.PaymentMethodCreditCard,
		pb.PaymentMethod_PAYMENT_METHOD_DEBIT_CARD:  model.PaymentMethodDebitCard,
		pb.PaymentMethod_PAYMENT_METHOD_CASH:        model.PaymentMethodCash,
		pb.PaymentMethod_PAYMENT_METHOD_ONLINE:      model.PaymentMethodOnline,
	})
	paymentStatuses = newEnumMap(map[pb.PaymentStatus]model.PaymentStatus{
		pb.PaymentStatus_PAYMENT_STATUS_PENDING:   model.PaymentStatusPending,
		pb.PaymentStatus_PAYMENT_STATUS_COMPLETED: model.PaymentStatusCompleted,
		pb.PaymentStatus_PAYMENT_STATUS_FAILED:    model.PaymentStatusFailed,
		pb.PaymentStatus_PAYMENT_STATUS_REFUNDED:  model.PaymentStatusRefunded,
	})
)

func toTime(t *timestamppb.Timestamp) time.Time {
	if t == nil {
		return time.Time{}
	}
	return t.AsTime()
}

func ts(t time.Time) *timestamppb.Timestamp {
	if t.IsZero() {
		return nil
	}
	return timestamppb.New(t)
}

func tsPtr(t *time.Time) *timestamppb.Timestamp {
	if t == nil {
		return nil
	}
	return ts(*t)
}

func grpcError(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, model.ErrNotFound):
		return status.Error(codes.NotFound, err.Error())
	case errors.Is(err, model.ErrInvalidArgument):
		return status.Error(codes.InvalidArgument, err.Error())
	case errors.Is(err, model.ErrAlreadyExists):
		return status.Error(codes.AlreadyExists, err.Error())
	case errors.Is(err, model.ErrPermissionDenied):
		return status.Error(codes.PermissionDenied, err.Error())
	case errors.Is(err, model.ErrFailedPrecond):
		return status.Error(codes.FailedPrecondition, err.Error())
	}
	if s, ok := status.FromError(err); ok {
		return s.Err()
	}
	return status.Error(codes.Internal, "internal error")
}

func userResponse(u *model.User) *pb.UserResponse {
	return &pb.UserResponse{
		UserId:         u.ID.String(),
		Email:          u.Email,
		FirstName:      u.FirstName,
		LastName:       u.LastName,
		PhoneNumber:    u.PhoneNumber,
		Address:        u.Address,
		UserType:       userTypes.pb(u.UserType),
		DrivingLicense: u.DrivingLicense,
		CreatedAt:      ts(u.CreatedAt),
		UpdatedAt:      ts(u.UpdatedAt),
	}
}

func vehicleResponse(v *model.Vehicle) *pb.VehicleResponse {
	r := &pb.VehicleResponse{
		VehicleId:          v.ID.String(),
		CategoryId:         v.CategoryID.String(),
		RegistrationNumber: v.RegistrationNumber,
		Make:               v.Make,
		Model:              v.Model,
		Year:               int32(v.Year),
		Color:              v.Color,
		Mileage:            v.Mileage,
		Status:             vehicleStatuses.pb(v.Status),
		HourlyRate:         v.HourlyRate,
		DailyRate:          v.DailyRate,
		LocationId:         v.LocationID.String(),
		Features:           v.Features,
		CreatedAt:          ts(v.CreatedAt),
		UpdatedAt:          ts(v.UpdatedAt),
	}
	if v.Category != nil {
		r.VehicleType = vehicleTypes.pb(v.Category.VehicleType)
	}
	switch {
	case v.CarDetails != nil:
		d := v.CarDetails
		r.VehicleDetails = &pb.VehicleResponse_CarDetails{CarDetails: &pb.CarDetails{
			FuelType: d.FuelType, Transmission: d.Transmission, SeatingCapacity: int32(d.SeatingCapacity),
			TrunkCapacity: d.TrunkCapacity, AirConditioning: d.AirConditioning,
		}}
	case v.MotorcycleDetails != nil:
		d := v.MotorcycleDetails
		r.VehicleDetails = &pb.VehicleResponse_MotorcycleDetails{MotorcycleDetails: &pb.MotorcycleDetails{
			EngineCapacity: int32(d.EngineCapacity), MotorcycleType: d.MotorcycleType, HelmetIncluded: d.HelmetIncluded,
		}}
	case v.BicycleDetails != nil:
		d := v.BicycleDetails
		r.VehicleDetails = &pb.VehicleResponse_BicycleDetails{BicycleDetails: &pb.BicycleDetails{
			BicycleType: d.BicycleType, FrameSize: d.FrameSize, GearCount: int32(d.GearCount), HasBasket: d.HasBasket,
		}}
	}
	return r
}

func vehicleResponses(vs []model.Vehicle) []*pb.VehicleResponse {
	out := make([]*pb.VehicleResponse, len(vs))
	for i := range vs {
		out[i] = vehicleResponse(&vs[i])
	}
	return out
}

func reservationResponse(r *model.Reservation) *pb.ReservationResponse {
	resp := &pb.ReservationResponse{
		ReservationId:    r.ID.String(),
		UserId:           r.UserID.String(),
		VehicleId:        r.VehicleID.String(),
		PickupLocationId: r.PickupLocationID.String(),
		ReturnLocationId: r.ReturnLocationID.String(),
		StartTime:        ts(r.StartTime),
		EndTime:          ts(r.EndTime),
		Status:           reservationStatuses.pb(r.Status),
		CreatedAt:        ts(r.CreatedAt),
		UpdatedAt:        ts(r.UpdatedAt),
	}
	if r.Vehicle != nil {
		resp.Vehicle = vehicleResponse(r.Vehicle)
	}
	return resp
}

func rentalResponse(r *model.Rental) *pb.RentalResponse {
	resp := &pb.RentalResponse{
		RentalId:         r.ID.String(),
		VehicleId:        r.VehicleID.String(),
		UserId:           r.UserID.String(),
		PickupTime:       ts(r.PickupTime),
		ActualReturnTime: tsPtr(r.ActualReturnTime),
		PickupLocationId: r.PickupLocationID.String(),
		PickupMileage:    r.PickupMileage,
		Status:           rentalStatuses.pb(r.Status),
		BaseFee:          r.BaseFee,
		AdditionalFees:   r.AdditionalFees,
		PaymentStatus:    paymentStatuses.pb(r.PaymentStatus),
		CreatedAt:        ts(r.CreatedAt),
	}
	if r.ReservationID != nil {
		resp.ReservationId = r.ReservationID.String()
	}
	if r.ReturnLocationID != nil {
		resp.ReturnLocationId = r.ReturnLocationID.String()
	}
	if r.ReturnMileage != nil {
		resp.ReturnMileage = *r.ReturnMileage
	}
	if r.Vehicle != nil {
		resp.Vehicle = vehicleResponse(r.Vehicle)
	}
	return resp
}

func paymentResponse(p *model.Payment) *pb.PaymentResponse {
	return &pb.PaymentResponse{
		PaymentId:     p.ID.String(),
		RentalId:      p.RentalID.String(),
		UserId:        p.UserID.String(),
		Amount:        p.Amount,
		PaymentMethod: paymentMethods.pb(p.PaymentMethod),
		TransactionId: p.TransactionID,
		PaymentStatus: paymentStatuses.pb(p.PaymentStatus),
		PaymentDate:   ts(p.PaymentDate),
	}
}

func reviewResponse(r *model.Review) *pb.ReviewResponse {
	resp := &pb.ReviewResponse{
		ReviewId:  r.ID.String(),
		RentalId:  r.RentalID.String(),
		UserId:    r.UserID.String(),
		VehicleId: r.VehicleID.String(),
		Rating:    int32(r.Rating),
		Comment:   r.Comment,
		CreatedAt: ts(r.CreatedAt),
	}
	if r.User != nil {
		resp.User = userResponse(r.User)
	}
	return resp
}

func locationResponse(l *model.Location) *pb.LocationResponse {
	return &pb.LocationResponse{
		LocationId:   l.ID.String(),
		Name:         l.Name,
		Address:      l.Address,
		City:         l.City,
		State:        l.State,
		Country:      l.Country,
		ZipCode:      l.ZipCode,
		Latitude:     l.Latitude,
		Longitude:    l.Longitude,
		ContactPhone: l.ContactPhone,
		OpeningHours: l.OpeningHours,
	}
}

func uuidOf(field, s string) (uuid.UUID, error) {
	id, err := uuid.Parse(s)
	if err != nil {
		return uuid.Nil, fmt.Errorf("%w: %s must be a valid UUID", model.ErrInvalidArgument, field)
	}
	return id, nil
}
