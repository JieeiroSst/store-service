package http

import (
	"context"

	"github.com/JIeeiroSst/car-rental-servcie/internal/usecase"
	pb "github.com/JIeeiroSst/lib-gateway/car-rental-servcie/gateway/car-rental-servcie"
)

type Handler struct {
	usecase *usecase.Usecase
	pb.UnimplementedVehicleRentalServiceServer
}

func NewHandler(usecase *usecase.Usecase) *Handler {
	return &Handler{
		usecase: usecase,
	}
}

func (h *Handler) RegisterUser(ctx context.Context, in *pb.RegisterUserRequest) (*pb.UserResponse, error) {

	return &pb.UserResponse{}, nil
}

func (h *Handler) GetUser(ctx context.Context, in *pb.GetUserRequest) (*pb.UserResponse, error) {

	return &pb.UserResponse{}, nil
}

func (h *Handler) UpdateUser(ctx context.Context, in *pb.UpdateUserRequest) (*pb.UserResponse, error) {

	return &pb.UserResponse{}, nil
}

func (h *Handler) CreateVehicle(ctx context.Context, in *pb.CreateVehicleRequest) (*pb.VehicleResponse, error) {

	return &pb.VehicleResponse{}, nil
}

func (h *Handler) GetVehicle(ctx context.Context, in *pb.GetVehicleRequest) (*pb.VehicleResponse, error) {

	return &pb.VehicleResponse{}, nil
}

func (h *Handler) UpdateVehicle(ctx context.Context, in *pb.UpdateVehicleRequest) (*pb.VehicleResponse, error) {
	return &pb.VehicleResponse{}, nil
}

func (h *Handler) ListVehicles(ctx context.Context, in *pb.ListVehiclesRequest) (*pb.ListVehiclesResponse, error) {

	return &pb.ListVehiclesResponse{}, nil
}

func (h *Handler) SearchAvailableVehicles(ctx context.Context, in *pb.SearchVehiclesRequest) (*pb.ListVehiclesResponse, error) {

	return &pb.ListVehiclesResponse{}, nil
}

func (h *Handler) CreateReservation(ctx context.Context, in *pb.CreateReservationRequest) (*pb.ReservationResponse, error) {

	return &pb.ReservationResponse{}, nil
}

func (h *Handler) GetReservation(ctx context.Context, in *pb.GetReservationRequest) (*pb.ReservationResponse, error) {

	return &pb.ReservationResponse{}, nil
}

func (h *Handler) UpdateReservation(ctx context.Context, in *pb.UpdateReservationRequest) (*pb.ReservationResponse, error) {

	return &pb.ReservationResponse{}, nil
}

func (h *Handler) CancelReservation(ctx context.Context, in *pb.CancelReservationRequest) (*pb.ReservationResponse, error) {

	return &pb.ReservationResponse{}, nil
}

func (h *Handler) ListUserReservations(ctx context.Context, in *pb.ListUserReservationsRequest) (*pb.ListReservationsResponse, error) {

	return &pb.ListReservationsResponse{}, nil
}

func (h *Handler) StartRental(ctx context.Context, in *pb.StartRentalRequest) (*pb.RentalResponse, error) {

	return &pb.RentalResponse{}, nil
}
func (h *Handler) CompleteRental(ctx context.Context, in *pb.CompleteRentalRequest) (*pb.RentalResponse, error) {

	return &pb.RentalResponse{}, nil
}
func (h *Handler) GetRental(ctx context.Context, in *pb.GetRentalRequest) (*pb.RentalResponse, error) {

	return &pb.RentalResponse{}, nil
}

func (h *Handler) ListUserRentals(ctx context.Context, in *pb.ListUserRentalsRequest) (*pb.ListRentalsResponse, error) {
	return &pb.ListRentalsResponse{}, nil
}

func (h *Handler) ProcessPayment(ctx context.Context, in *pb.ProcessPaymentRequest) (*pb.PaymentResponse, error) {
	return &pb.PaymentResponse{}, nil
}

func (h *Handler) SubmitReview(ctx context.Context, in *pb.SubmitReviewRequest) (*pb.ReviewResponse, error) {
	return &pb.ReviewResponse{}, nil
}

func (h *Handler) ListVehicleReviews(ctx context.Context, in *pb.ListVehicleReviewsRequest) (*pb.ListReviewsResponse, error) {
	return &pb.ListReviewsResponse{}, nil
}

func (h *Handler) ListLocations(ctx context.Context, in *pb.ListLocationsRequest) (*pb.ListLocationsResponse, error) {

	return &pb.ListLocationsResponse{}, nil
}
