package car_rental_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "car-rental-service"

const DefaultBaseURL = "http://car-rental-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) GetUser(ctx context.Context, userID string) (*model.CarRentalUserResponse, error) {
	path := "/api/v1/users/" + url.PathEscape(userID)
	q := url.Values{}
	h := http.Header{}
	var out *model.CarRentalUserResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetVehicle(ctx context.Context, vehicleID string) (*model.CarRentalVehicleResponse, error) {
	path := "/api/v1/vehicles/" + url.PathEscape(vehicleID)
	q := url.Values{}
	h := http.Header{}
	var out *model.CarRentalVehicleResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListVehicles(ctx context.Context, pageSize *int, pageToken *string, vehicleType *string, locationID *string) (*model.CarRentalListVehiclesResponse, error) {
	path := "/api/v1/vehicles"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	if vehicleType != nil {
		q.Set("vehicle_type", *vehicleType)
	}
	if locationID != nil {
		q.Set("location_id", *locationID)
	}
	h := http.Header{}
	var out *model.CarRentalListVehiclesResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) SearchAvailableVehicles(ctx context.Context, startTime *string, endTime *string, pickupLocationID *string, returnLocationID *string, vehicleType *string, categoryID *string, pageSize *int, pageToken *string) (*model.CarRentalListVehiclesResponse, error) {
	path := "/api/v1/vehicles/available"
	q := url.Values{}
	if startTime != nil {
		q.Set("start_time", *startTime)
	}
	if endTime != nil {
		q.Set("end_time", *endTime)
	}
	if pickupLocationID != nil {
		q.Set("pickup_location_id", *pickupLocationID)
	}
	if returnLocationID != nil {
		q.Set("return_location_id", *returnLocationID)
	}
	if vehicleType != nil {
		q.Set("vehicle_type", *vehicleType)
	}
	if categoryID != nil {
		q.Set("category_id", *categoryID)
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	h := http.Header{}
	var out *model.CarRentalListVehiclesResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetReservation(ctx context.Context, reservationID string) (*model.CarRentalReservationResponse, error) {
	path := "/api/v1/reservations/" + url.PathEscape(reservationID)
	q := url.Values{}
	h := http.Header{}
	var out *model.CarRentalReservationResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListUserReservations(ctx context.Context, userID string, status *string, pageSize *int, pageToken *string) (*model.CarRentalListReservationsResponse, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/reservations"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	h := http.Header{}
	var out *model.CarRentalListReservationsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) GetRental(ctx context.Context, rentalID string) (*model.CarRentalRentalResponse, error) {
	path := "/api/v1/rentals/" + url.PathEscape(rentalID)
	q := url.Values{}
	h := http.Header{}
	var out *model.CarRentalRentalResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListUserRentals(ctx context.Context, userID string, status *string, pageSize *int, pageToken *string) (*model.CarRentalListRentalsResponse, error) {
	path := "/api/v1/users/" + url.PathEscape(userID) + "/rentals"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	h := http.Header{}
	var out *model.CarRentalListRentalsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListVehicleReviews(ctx context.Context, vehicleID string, pageSize *int, pageToken *string) (*model.CarRentalListReviewsResponse, error) {
	path := "/api/v1/vehicles/" + url.PathEscape(vehicleID) + "/reviews"
	q := url.Values{}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	h := http.Header{}
	var out *model.CarRentalListReviewsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ListLocations(ctx context.Context, city *string, state *string, country *string, pageSize *int, pageToken *string) (*model.CarRentalListLocationsResponse, error) {
	path := "/api/v1/locations"
	q := url.Values{}
	if city != nil {
		q.Set("city", *city)
	}
	if state != nil {
		q.Set("state", *state)
	}
	if country != nil {
		q.Set("country", *country)
	}
	if pageSize != nil {
		q.Set("page_size", strconv.Itoa(*pageSize))
	}
	if pageToken != nil {
		q.Set("page_token", *pageToken)
	}
	h := http.Header{}
	var out *model.CarRentalListLocationsResponse
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
