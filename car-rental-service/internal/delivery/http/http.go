package http

import (
	"context"

	"github.com/JIeeiroSst/car-rental-service/internal/auth"
	"github.com/JIeeiroSst/car-rental-service/internal/usecase"
	"github.com/JIeeiroSst/car-rental-service/model"
	pb "github.com/JIeeiroSst/lib-gateway/car-rental-servcie/gateway/car-rental-servcie"
)

type Handler struct {
	usecase *usecase.Usecase
	auth    *auth.Authenticator
	pb.UnimplementedVehicleRentalServiceServer
}

func NewHandler(usecase *usecase.Usecase, authn *auth.Authenticator) *Handler {
	return &Handler{
		usecase: usecase,
		auth:    authn,
	}
}

func (h *Handler) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest) (*pb.UserResponse, error) {
	u, err := h.usecase.RegisterUser(ctx, usecase.RegisterUserInput{
		Email: in.Email, Password: in.Password, FirstName: in.FirstName, LastName: in.LastName,
		PhoneNumber: in.PhoneNumber, Address: in.Address, DrivingLicense: in.DrivingLicense,
		UserType: userTypes.model(in.UserType), AllowPrivileged: h.auth.IsAdmin(auth.FromContext(ctx)),
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return userResponse(u), nil
}

func (h *Handler) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.UserResponse, error) {
	u, err := h.usecase.GetUser(ctx, in.UserId)
	if err != nil {
		return nil, grpcError(err)
	}
	return userResponse(u), nil
}

func (h *Handler) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UserResponse, error) {
	u, err := h.usecase.UpdateUser(ctx, usecase.UpdateUserInput{
		UserID: in.UserId, Email: in.Email, FirstName: in.FirstName, LastName: in.LastName,
		PhoneNumber: in.PhoneNumber, Address: in.Address, DrivingLicense: in.DrivingLicense,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return userResponse(u), nil
}

func (h *Handler) CreateVehicle(ctx context.Context, in *pb.CreateVehicleRequest) (*pb.VehicleResponse, error) {
	catID, err := uuidOf("category_id", in.CategoryId)
	if err != nil {
		return nil, grpcError(err)
	}
	locID, err := uuidOf("location_id", in.LocationId)
	if err != nil {
		return nil, grpcError(err)
	}
	v := &model.Vehicle{
		CategoryID:         catID,
		RegistrationNumber: in.RegistrationNumber,
		Make:               in.Make,
		Model:              in.Model,
		Year:               int(in.Year),
		Color:              in.Color,
		Mileage:            in.Mileage,
		Status:             vehicleStatuses.model(in.Status),
		HourlyRate:         in.HourlyRate,
		DailyRate:          in.DailyRate,
		LocationID:         locID,
		Features:           in.Features,
	}
	switch d := in.VehicleDetails.(type) {
	case *pb.CreateVehicleRequest_CarDetails:
		if c := d.CarDetails; c != nil {
			v.CarDetails = &model.CarDetails{FuelType: c.FuelType, Transmission: c.Transmission,
				SeatingCapacity: int(c.SeatingCapacity), TrunkCapacity: c.TrunkCapacity, AirConditioning: c.AirConditioning}
		}
	case *pb.CreateVehicleRequest_MotorcycleDetails:
		if m := d.MotorcycleDetails; m != nil {
			v.MotorcycleDetails = &model.MotorcycleDetails{EngineCapacity: int(m.EngineCapacity),
				MotorcycleType: m.MotorcycleType, HelmetIncluded: m.HelmetIncluded}
		}
	case *pb.CreateVehicleRequest_BicycleDetails:
		if b := d.BicycleDetails; b != nil {
			v.BicycleDetails = &model.BicycleDetails{BicycleType: b.BicycleType, FrameSize: b.FrameSize,
				GearCount: int(b.GearCount), HasBasket: b.HasBasket}
		}
	}
	created, err := h.usecase.CreateVehicle(ctx, v)
	if err != nil {
		return nil, grpcError(err)
	}
	return vehicleResponse(created), nil
}

func (h *Handler) GetVehicle(ctx context.Context, in *pb.GetVehicleRequest) (*pb.VehicleResponse, error) {
	v, err := h.usecase.GetVehicle(ctx, in.VehicleId)
	if err != nil {
		return nil, grpcError(err)
	}
	return vehicleResponse(v), nil
}

func (h *Handler) UpdateVehicle(ctx context.Context, in *pb.UpdateVehicleRequest) (*pb.VehicleResponse, error) {
	v, err := h.usecase.UpdateVehicle(ctx, usecase.VehiclePatch{
		VehicleID: in.VehicleId, RegistrationNumber: in.RegistrationNumber, Mileage: in.Mileage,
		Status: vehicleStatuses.model(in.Status), HourlyRate: in.HourlyRate, DailyRate: in.DailyRate,
		LocationID: in.LocationId, Features: in.Features,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return vehicleResponse(v), nil
}

func (h *Handler) ListVehicles(ctx context.Context, in *pb.ListVehiclesRequest) (*pb.ListVehiclesResponse, error) {
	page, err := usecase.NewPage(in.PageSize, in.PageToken)
	if err != nil {
		return nil, grpcError(err)
	}
	res, err := h.usecase.ListVehicles(ctx, vehicleTypes.model(in.VehicleType), in.LocationId, page)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pb.ListVehiclesResponse{Vehicles: vehicleResponses(res.Vehicles), NextPageToken: res.NextPageToken, TotalCount: int32(res.Total)}, nil
}

func (h *Handler) SearchAvailableVehicles(ctx context.Context, in *pb.SearchVehiclesRequest) (*pb.ListVehiclesResponse, error) {
	page, err := usecase.NewPage(in.PageSize, in.PageToken)
	if err != nil {
		return nil, grpcError(err)
	}
	res, err := h.usecase.SearchAvailableVehicles(ctx, usecase.SearchVehiclesInput{
		Start: toTime(in.StartTime), End: toTime(in.EndTime),
		PickupLocationID: in.PickupLocationId, ReturnLocationID: in.ReturnLocationId,
		VehicleType: vehicleTypes.model(in.VehicleType), CategoryID: in.CategoryId,
	}, page)
	if err != nil {
		return nil, grpcError(err)
	}
	return &pb.ListVehiclesResponse{Vehicles: vehicleResponses(res.Vehicles), NextPageToken: res.NextPageToken, TotalCount: int32(res.Total)}, nil
}

func (h *Handler) CreateReservation(ctx context.Context, in *pb.CreateReservationRequest) (*pb.ReservationResponse, error) {
	r, err := h.usecase.CreateReservation(ctx, usecase.CreateReservationInput{
		UserID: in.UserId, VehicleID: in.VehicleId,
		PickupLocationID: in.PickupLocationId, ReturnLocationID: in.ReturnLocationId,
		Start: toTime(in.StartTime), End: toTime(in.EndTime),
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return reservationResponse(r), nil
}

func (h *Handler) GetReservation(ctx context.Context, in *pb.GetReservationRequest) (*pb.ReservationResponse, error) {
	r, err := h.usecase.GetReservation(ctx, in.ReservationId)
	if err != nil {
		return nil, grpcError(err)
	}
	return reservationResponse(r), nil
}

func (h *Handler) UpdateReservation(ctx context.Context, in *pb.UpdateReservationRequest) (*pb.ReservationResponse, error) {
	r, err := h.usecase.UpdateReservation(ctx, usecase.UpdateReservationInput{
		ReservationID: in.ReservationId, PickupLocationID: in.PickupLocationId, ReturnLocationID: in.ReturnLocationId,
		Start: toTime(in.StartTime), End: toTime(in.EndTime),
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return reservationResponse(r), nil
}

func (h *Handler) CancelReservation(ctx context.Context, in *pb.CancelReservationRequest) (*pb.ReservationResponse, error) {
	r, err := h.usecase.CancelReservation(ctx, in.ReservationId, in.CancellationReason)
	if err != nil {
		return nil, grpcError(err)
	}
	return reservationResponse(r), nil
}

func (h *Handler) ListUserReservations(ctx context.Context, in *pb.ListUserReservationsRequest) (*pb.ListReservationsResponse, error) {
	page, err := usecase.NewPage(in.PageSize, in.PageToken)
	if err != nil {
		return nil, grpcError(err)
	}
	res, err := h.usecase.ListUserReservations(ctx, in.UserId, reservationStatuses.model(in.Status), page)
	if err != nil {
		return nil, grpcError(err)
	}
	out := make([]*pb.ReservationResponse, len(res.Reservations))
	for i := range res.Reservations {
		out[i] = reservationResponse(&res.Reservations[i])
	}
	return &pb.ListReservationsResponse{Reservations: out, NextPageToken: res.NextPageToken, TotalCount: int32(res.Total)}, nil
}

func (h *Handler) StartRental(ctx context.Context, in *pb.StartRentalRequest) (*pb.RentalResponse, error) {
	r, err := h.usecase.StartRental(ctx, usecase.StartRentalInput{
		ReservationID: in.ReservationId, PickupMileage: in.PickupMileage,
		StaffID: in.StaffId, PickupLocationID: in.PickupLocationId,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return rentalResponse(r), nil
}

func (h *Handler) CompleteRental(ctx context.Context, in *pb.CompleteRentalRequest) (*pb.RentalResponse, error) {
	r, err := h.usecase.CompleteRental(ctx, usecase.CompleteRentalInput{
		RentalID: in.RentalId, ReturnMileage: in.ReturnMileage, StaffID: in.StaffId,
		ReturnLocationID: in.ReturnLocationId, AdditionalFees: in.AdditionalFees, Notes: in.Notes,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return rentalResponse(r), nil
}

func (h *Handler) GetRental(ctx context.Context, in *pb.GetRentalRequest) (*pb.RentalResponse, error) {
	r, err := h.usecase.GetRental(ctx, in.RentalId)
	if err != nil {
		return nil, grpcError(err)
	}
	return rentalResponse(r), nil
}

func (h *Handler) ListUserRentals(ctx context.Context, in *pb.ListUserRentalsRequest) (*pb.ListRentalsResponse, error) {
	page, err := usecase.NewPage(in.PageSize, in.PageToken)
	if err != nil {
		return nil, grpcError(err)
	}
	res, err := h.usecase.ListUserRentals(ctx, in.UserId, rentalStatuses.model(in.Status), page)
	if err != nil {
		return nil, grpcError(err)
	}
	out := make([]*pb.RentalResponse, len(res.Rentals))
	for i := range res.Rentals {
		out[i] = rentalResponse(&res.Rentals[i])
	}
	return &pb.ListRentalsResponse{Rentals: out, NextPageToken: res.NextPageToken, TotalCount: int32(res.Total)}, nil
}

func (h *Handler) ProcessPayment(ctx context.Context, in *pb.ProcessPaymentRequest) (*pb.PaymentResponse, error) {
	p, err := h.usecase.ProcessPayment(ctx, usecase.ProcessPaymentInput{
		RentalID: in.RentalId, UserID: in.UserId, Amount: in.Amount,
		Method: paymentMethods.model(in.PaymentMethod), TransactionID: in.TransactionId,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return paymentResponse(p), nil
}

func (h *Handler) SubmitReview(ctx context.Context, in *pb.SubmitReviewRequest) (*pb.ReviewResponse, error) {
	r, err := h.usecase.SubmitReview(ctx, usecase.SubmitReviewInput{
		RentalID: in.RentalId, UserID: in.UserId, VehicleID: in.VehicleId, Rating: int(in.Rating), Comment: in.Comment,
	})
	if err != nil {
		return nil, grpcError(err)
	}
	return reviewResponse(r), nil
}

func (h *Handler) ListVehicleReviews(ctx context.Context, in *pb.ListVehicleReviewsRequest) (*pb.ListReviewsResponse, error) {
	page, err := usecase.NewPage(in.PageSize, in.PageToken)
	if err != nil {
		return nil, grpcError(err)
	}
	res, err := h.usecase.ListVehicleReviews(ctx, in.VehicleId, page)
	if err != nil {
		return nil, grpcError(err)
	}
	out := make([]*pb.ReviewResponse, len(res.Reviews))
	for i := range res.Reviews {
		out[i] = reviewResponse(&res.Reviews[i])
	}
	return &pb.ListReviewsResponse{Reviews: out, NextPageToken: res.NextPageToken, TotalCount: int32(res.Total)}, nil
}

func (h *Handler) ListLocations(ctx context.Context, in *pb.ListLocationsRequest) (*pb.ListLocationsResponse, error) {
	page, err := usecase.NewPage(in.PageSize, in.PageToken)
	if err != nil {
		return nil, grpcError(err)
	}
	res, err := h.usecase.ListLocations(ctx, in.City, in.State, in.Country, page)
	if err != nil {
		return nil, grpcError(err)
	}
	out := make([]*pb.LocationResponse, len(res.Locations))
	for i := range res.Locations {
		out[i] = locationResponse(&res.Locations[i])
	}
	return &pb.ListLocationsResponse{Locations: out, NextPageToken: res.NextPageToken, TotalCount: int32(res.Total)}, nil
}
