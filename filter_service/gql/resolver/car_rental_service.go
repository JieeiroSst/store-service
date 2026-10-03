package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) CarRentalQuery() generated.CarRentalQueryResolver {
	return &carRentalQueryResolver{r}
}

type carRentalQueryResolver struct{ *Resolver }

func (r *carRentalQueryResolver) GetUser(ctx context.Context, obj *model.CarRentalQuery, userID string) (*model.CarRentalUserResponse, error) {
	return r.Clients.CarRentalService.GetUser(ctx, userID)
}

func (r *carRentalQueryResolver) GetVehicle(ctx context.Context, obj *model.CarRentalQuery, vehicleID string) (*model.CarRentalVehicleResponse, error) {
	return r.Clients.CarRentalService.GetVehicle(ctx, vehicleID)
}

func (r *carRentalQueryResolver) ListVehicles(ctx context.Context, obj *model.CarRentalQuery, pageSize *int, pageToken *string, vehicleType *string, locationID *string) (*model.CarRentalListVehiclesResponse, error) {
	return r.Clients.CarRentalService.ListVehicles(ctx, pageSize, pageToken, vehicleType, locationID)
}

func (r *carRentalQueryResolver) SearchAvailableVehicles(ctx context.Context, obj *model.CarRentalQuery, startTime *string, endTime *string, pickupLocationID *string, returnLocationID *string, vehicleType *string, categoryID *string, pageSize *int, pageToken *string) (*model.CarRentalListVehiclesResponse, error) {
	return r.Clients.CarRentalService.SearchAvailableVehicles(ctx, startTime, endTime, pickupLocationID, returnLocationID, vehicleType, categoryID, pageSize, pageToken)
}

func (r *carRentalQueryResolver) GetReservation(ctx context.Context, obj *model.CarRentalQuery, reservationID string) (*model.CarRentalReservationResponse, error) {
	return r.Clients.CarRentalService.GetReservation(ctx, reservationID)
}

func (r *carRentalQueryResolver) ListUserReservations(ctx context.Context, obj *model.CarRentalQuery, userID string, status *string, pageSize *int, pageToken *string) (*model.CarRentalListReservationsResponse, error) {
	return r.Clients.CarRentalService.ListUserReservations(ctx, userID, status, pageSize, pageToken)
}

func (r *carRentalQueryResolver) GetRental(ctx context.Context, obj *model.CarRentalQuery, rentalID string) (*model.CarRentalRentalResponse, error) {
	return r.Clients.CarRentalService.GetRental(ctx, rentalID)
}

func (r *carRentalQueryResolver) ListUserRentals(ctx context.Context, obj *model.CarRentalQuery, userID string, status *string, pageSize *int, pageToken *string) (*model.CarRentalListRentalsResponse, error) {
	return r.Clients.CarRentalService.ListUserRentals(ctx, userID, status, pageSize, pageToken)
}

func (r *carRentalQueryResolver) ListVehicleReviews(ctx context.Context, obj *model.CarRentalQuery, vehicleID string, pageSize *int, pageToken *string) (*model.CarRentalListReviewsResponse, error) {
	return r.Clients.CarRentalService.ListVehicleReviews(ctx, vehicleID, pageSize, pageToken)
}

func (r *carRentalQueryResolver) ListLocations(ctx context.Context, obj *model.CarRentalQuery, city *string, state *string, country *string, pageSize *int, pageToken *string) (*model.CarRentalListLocationsResponse, error) {
	return r.Clients.CarRentalService.ListLocations(ctx, city, state, country, pageSize, pageToken)
}
