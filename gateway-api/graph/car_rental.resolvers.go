package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy"
)

func (r *mutationResolver) CarRentalRegisterUser(ctx context.Context, input model.CarRentalRegisterUserRequestInput) (*model.CarRentalUserResponse, error) {
	return grpcInvoke[model.CarRentalUserResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/RegisterUser", "vehicle_rental.RegisterUserRequest", "vehicle_rental.UserResponse", input)
}

func (r *mutationResolver) CarRentalUpdateUser(ctx context.Context, input model.CarRentalUpdateUserRequestInput) (*model.CarRentalUserResponse, error) {
	return grpcInvoke[model.CarRentalUserResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/UpdateUser", "vehicle_rental.UpdateUserRequest", "vehicle_rental.UserResponse", input)
}

func (r *mutationResolver) CarRentalCreateVehicle(ctx context.Context, input model.CarRentalCreateVehicleRequestInput) (*model.CarRentalVehicleResponse, error) {
	return grpcInvoke[model.CarRentalVehicleResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/CreateVehicle", "vehicle_rental.CreateVehicleRequest", "vehicle_rental.VehicleResponse", input)
}

func (r *mutationResolver) CarRentalUpdateVehicle(ctx context.Context, input model.CarRentalUpdateVehicleRequestInput) (*model.CarRentalVehicleResponse, error) {
	return grpcInvoke[model.CarRentalVehicleResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/UpdateVehicle", "vehicle_rental.UpdateVehicleRequest", "vehicle_rental.VehicleResponse", input)
}

func (r *mutationResolver) CarRentalCreateReservation(ctx context.Context, input model.CarRentalCreateReservationRequestInput) (*model.CarRentalReservationResponse, error) {
	return grpcInvoke[model.CarRentalReservationResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/CreateReservation", "vehicle_rental.CreateReservationRequest", "vehicle_rental.ReservationResponse", input)
}

func (r *mutationResolver) CarRentalUpdateReservation(ctx context.Context, input model.CarRentalUpdateReservationRequestInput) (*model.CarRentalReservationResponse, error) {
	return grpcInvoke[model.CarRentalReservationResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/UpdateReservation", "vehicle_rental.UpdateReservationRequest", "vehicle_rental.ReservationResponse", input)
}

func (r *mutationResolver) CarRentalCancelReservation(ctx context.Context, input model.CarRentalCancelReservationRequestInput) (*model.CarRentalReservationResponse, error) {
	return grpcInvoke[model.CarRentalReservationResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/CancelReservation", "vehicle_rental.CancelReservationRequest", "vehicle_rental.ReservationResponse", input)
}

func (r *mutationResolver) CarRentalStartRental(ctx context.Context, input model.CarRentalStartRentalRequestInput) (*model.CarRentalRentalResponse, error) {
	return grpcInvoke[model.CarRentalRentalResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/StartRental", "vehicle_rental.StartRentalRequest", "vehicle_rental.RentalResponse", input)
}

func (r *mutationResolver) CarRentalCompleteRental(ctx context.Context, input model.CarRentalCompleteRentalRequestInput) (*model.CarRentalRentalResponse, error) {
	return grpcInvoke[model.CarRentalRentalResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/CompleteRental", "vehicle_rental.CompleteRentalRequest", "vehicle_rental.RentalResponse", input)
}

func (r *mutationResolver) CarRentalProcessPayment(ctx context.Context, input model.CarRentalProcessPaymentRequestInput) (*model.CarRentalPaymentResponse, error) {
	return grpcInvoke[model.CarRentalPaymentResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/ProcessPayment", "vehicle_rental.ProcessPaymentRequest", "vehicle_rental.PaymentResponse", input)
}

func (r *mutationResolver) CarRentalSubmitReview(ctx context.Context, input model.CarRentalSubmitReviewRequestInput) (*model.CarRentalReviewResponse, error) {
	return grpcInvoke[model.CarRentalReviewResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/SubmitReview", "vehicle_rental.SubmitReviewRequest", "vehicle_rental.ReviewResponse", input)
}

func (r *queryResolver) CarRentalGetUser(ctx context.Context, input model.CarRentalGetUserRequestInput) (*model.CarRentalUserResponse, error) {
	return grpcInvoke[model.CarRentalUserResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/GetUser", "vehicle_rental.GetUserRequest", "vehicle_rental.UserResponse", input)
}

func (r *queryResolver) CarRentalGetVehicle(ctx context.Context, input model.CarRentalGetVehicleRequestInput) (*model.CarRentalVehicleResponse, error) {
	return grpcInvoke[model.CarRentalVehicleResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/GetVehicle", "vehicle_rental.GetVehicleRequest", "vehicle_rental.VehicleResponse", input)
}

func (r *queryResolver) CarRentalListVehicles(ctx context.Context, input model.CarRentalListVehiclesRequestInput) (*model.CarRentalListVehiclesResponse, error) {
	return grpcInvoke[model.CarRentalListVehiclesResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/ListVehicles", "vehicle_rental.ListVehiclesRequest", "vehicle_rental.ListVehiclesResponse", input)
}

func (r *queryResolver) CarRentalSearchAvailableVehicles(ctx context.Context, input model.CarRentalSearchVehiclesRequestInput) (*model.CarRentalListVehiclesResponse, error) {
	return grpcInvoke[model.CarRentalListVehiclesResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/SearchAvailableVehicles", "vehicle_rental.SearchVehiclesRequest", "vehicle_rental.ListVehiclesResponse", input)
}

func (r *queryResolver) CarRentalGetReservation(ctx context.Context, input model.CarRentalGetReservationRequestInput) (*model.CarRentalReservationResponse, error) {
	return grpcInvoke[model.CarRentalReservationResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/GetReservation", "vehicle_rental.GetReservationRequest", "vehicle_rental.ReservationResponse", input)
}

func (r *queryResolver) CarRentalListUserReservations(ctx context.Context, input model.CarRentalListUserReservationsRequestInput) (*model.CarRentalListReservationsResponse, error) {
	return grpcInvoke[model.CarRentalListReservationsResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/ListUserReservations", "vehicle_rental.ListUserReservationsRequest", "vehicle_rental.ListReservationsResponse", input)
}

func (r *queryResolver) CarRentalGetRental(ctx context.Context, input model.CarRentalGetRentalRequestInput) (*model.CarRentalRentalResponse, error) {
	return grpcInvoke[model.CarRentalRentalResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/GetRental", "vehicle_rental.GetRentalRequest", "vehicle_rental.RentalResponse", input)
}

func (r *queryResolver) CarRentalListUserRentals(ctx context.Context, input model.CarRentalListUserRentalsRequestInput) (*model.CarRentalListRentalsResponse, error) {
	return grpcInvoke[model.CarRentalListRentalsResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/ListUserRentals", "vehicle_rental.ListUserRentalsRequest", "vehicle_rental.ListRentalsResponse", input)
}

func (r *queryResolver) CarRentalListVehicleReviews(ctx context.Context, input model.CarRentalListVehicleReviewsRequestInput) (*model.CarRentalListReviewsResponse, error) {
	return grpcInvoke[model.CarRentalListReviewsResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/ListVehicleReviews", "vehicle_rental.ListVehicleReviewsRequest", "vehicle_rental.ListReviewsResponse", input)
}

func (r *queryResolver) CarRentalListLocations(ctx context.Context, input model.CarRentalListLocationsRequestInput) (*model.CarRentalListLocationsResponse, error) {
	return grpcInvoke[model.CarRentalListLocationsResponse](ctx, r.Proxy, proxy.CarRentalService, "/vehicle_rental.VehicleRentalService/ListLocations", "vehicle_rental.ListLocationsRequest", "vehicle_rental.ListLocationsResponse", input)
}
