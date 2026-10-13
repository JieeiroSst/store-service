package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) RentApproveHomestay(ctx context.Context, id int) (*model.RentHomestay, error) {
	return restCall[*model.RentHomestay](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/homestays/%v/approve", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) RentRejectHomestay(ctx context.Context, id int, reason string) (*model.RentHomestay, error) {
	return restCall[*model.RentHomestay](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/homestays/%v/reject", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) RentCreateHomestay(ctx context.Context, input model.RentHomestayInput) (*model.RentHomestay, error) {
	return restCall[*model.RentHomestay](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: "/api/v1/homestays", Body: input}, "")
}

func (r *mutationResolver) RentUpdateHomestay(ctx context.Context, id int, input model.RentHomestayInput) (*model.RentHomestay, error) {
	return restCall[*model.RentHomestay](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/homestays/%v", id), Body: input}, "")
}

func (r *mutationResolver) RentDeactivateHomestay(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/homestays/%v", id)}, "")
}

func (r *mutationResolver) RentSetAvailability(ctx context.Context, id int, input model.RentAvailabilityInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/homestays/%v/availability", id), Body: input}, "")
}

func (r *mutationResolver) RentCreateAmenity(ctx context.Context, name string, icon *string) (*model.RentAmenity, error) {
	return restCall[*model.RentAmenity](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: "/api/v1/amenities", Body: map[string]any{"name": name, "icon": icon}}, "")
}

func (r *mutationResolver) RentBook(ctx context.Context, input model.RentBookingInput) (*model.RentBooking, error) {
	return restCall[*model.RentBooking](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: "/api/v1/bookings", Body: input}, "")
}

func (r *mutationResolver) RentCancelBooking(ctx context.Context, id int) (*model.RentBooking, error) {
	return restCall[*model.RentBooking](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/bookings/%v/cancel", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) RentPayBooking(ctx context.Context, id int, method string, provider *string) (*model.RentBooking, error) {
	return restCall[*model.RentBooking](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/bookings/%v/pay", id), Body: map[string]any{"method": method, "provider": provider}}, "")
}

func (r *mutationResolver) RentSetRate(ctx context.Context, id int, rateModel string, input model.RentRateInput) (*model.RentRate, error) {
	return restCall[*model.RentRate](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/homestays/%v/rates/%v", id, rateModel), Body: input}, "")
}

func (r *mutationResolver) RentDeleteRate(ctx context.Context, id int, rateModel string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/homestays/%v/rates/%v", id, rateModel)}, "")
}

func (r *mutationResolver) RentStartLease(ctx context.Context, input model.RentLeaseInput) (*model.RentLease, error) {
	return restCall[*model.RentLease](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: "/api/v1/leases", Body: input}, "")
}

func (r *mutationResolver) RentCancelLease(ctx context.Context, id int) (*model.RentLease, error) {
	return restCall[*model.RentLease](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/leases/%v/cancel", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) RentTerminateLease(ctx context.Context, id int) (*model.RentLease, error) {
	return restCall[*model.RentLease](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/leases/%v/terminate", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) RentPayLeaseInvoice(ctx context.Context, id int, period string, method string, provider *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/leases/%v/invoices/%v/pay", id, period), Body: map[string]any{"method": method, "provider": provider}}, "")
}

func (r *mutationResolver) RentCreateReview(ctx context.Context, id int, input model.RentReviewInput) (*model.RentReview, error) {
	return restCall[*model.RentReview](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/homestays/%v/reviews", id), Body: input}, "")
}

func (r *mutationResolver) RentAddWishlist(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/wishlist/%v", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) RentRemoveWishlist(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/wishlist/%v", id)}, "")
}

func (r *mutationResolver) RentCreateWallet(ctx context.Context, currency string) (*model.RentWallet, error) {
	return restCall[*model.RentWallet](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "POST", Path: "/api/v1/wallet", Body: map[string]any{"currency": currency}}, "")
}

func (r *queryResolver) RentHomestay(ctx context.Context, id int) (*model.RentHomestay, error) {
	return restCall[*model.RentHomestay](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/homestays/%v", id)}, "")
}

func (r *queryResolver) RentHomestays(ctx context.Context, provinceID *int, districtID *int, wardID *int, typeArg *int, guests *int, q *string, modelArg *string, minPrice *string, maxPrice *string, sort *string, limit *int, cursor *string, amenityIds *string, checkin *string, checkout *string) (*model.RentHomestayPage, error) {
	return restCall[*model.RentHomestayPage](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/homestays", Query: restQuery("province_id", provinceID, "district_id", districtID, "ward_id", wardID, "type", typeArg, "guests", guests, "q", q, "model", modelArg, "min_price", minPrice, "max_price", maxPrice, "sort", sort, "limit", limit, "cursor", cursor, "amenity_ids", amenityIds, "checkin", checkin, "checkout", checkout)}, "")
}

func (r *queryResolver) RentAvailability(ctx context.Context, id int, from *string, to *string) ([]*model.RentSlot, error) {
	return restCall[[]*model.RentSlot](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/homestays/%v/availability", id), Query: restQuery("from", from, "to", to)}, "")
}

func (r *queryResolver) RentAmenities(ctx context.Context) ([]*model.RentAmenity, error) {
	return restCall[[]*model.RentAmenity](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/amenities"}, "")
}

func (r *queryResolver) RentMyHomestays(ctx context.Context, limit *int, offset *int) ([]*model.RentHomestay, error) {
	return restCall[[]*model.RentHomestay](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/my/homestays", Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) RentHomestayReviewQueue(ctx context.Context, status *string, limit *int, offset *int) ([]*model.RentHomestay, error) {
	return restCall[[]*model.RentHomestay](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/admin/homestays", Query: restQuery("status", status, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) RentBookings(ctx context.Context, userID *int, as *string, limit *int, offset *int) ([]*model.RentBooking, error) {
	return restCall[[]*model.RentBooking](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/bookings", Query: restQuery("user_id", userID, "as", as, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) RentBooking(ctx context.Context, id int) (*model.RentBooking, error) {
	return restCall[*model.RentBooking](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/bookings/%v", id)}, "")
}

func (r *queryResolver) RentRates(ctx context.Context, id int) ([]*model.RentRate, error) {
	return restCall[[]*model.RentRate](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/homestays/%v/rates", id)}, "")
}

func (r *queryResolver) RentLeases(ctx context.Context, userID *int, as *string, limit *int, offset *int) ([]*model.RentLease, error) {
	return restCall[[]*model.RentLease](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/leases", Query: restQuery("user_id", userID, "as", as, "limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) RentLease(ctx context.Context, id int) (*model.RentLease, error) {
	return restCall[*model.RentLease](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/leases/%v", id)}, "")
}

func (r *queryResolver) RentInvoices(ctx context.Context, limit *int, offset *int, status *string, dueBefore *string) ([]*model.RentInvoice, error) {
	return restCall[[]*model.RentInvoice](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/invoices", Query: restQuery("limit", limit, "offset", offset, "status", status, "due_before", dueBefore)}, "")
}

func (r *queryResolver) RentReviews(ctx context.Context, id int, limit *int, offset *int) ([]*model.RentReview, error) {
	return restCall[[]*model.RentReview](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/homestays/%v/reviews", id), Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) RentWishlist(ctx context.Context, limit *int, offset *int) ([]*model.RentHomestay, error) {
	return restCall[[]*model.RentHomestay](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/wishlist", Query: restQuery("limit", limit, "offset", offset)}, "")
}

func (r *queryResolver) RentLoyalty(ctx context.Context) (*model.RentLoyalty, error) {
	return restCall[*model.RentLoyalty](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/loyalty"}, "")
}

func (r *queryResolver) RentWallet(ctx context.Context) (*model.RentWallet, error) {
	return restCall[*model.RentWallet](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/wallet"}, "")
}

func (r *queryResolver) RentWalletTransactions(ctx context.Context, limit *int, offset *int) ([]*model.RentWalletTxn, error) {
	return restCall[[]*model.RentWalletTxn](ctx, r.Proxy, "rent-house-service", rest.Request{Method: "GET", Path: "/api/v1/wallet/transactions", Query: restQuery("limit", limit, "offset", offset)}, "")
}
