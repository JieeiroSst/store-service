package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) RentQuery() generated.RentQueryResolver { return &rentQueryResolver{r} }

type rentQueryResolver struct{ *Resolver }

func (r *rentQueryResolver) Homestay(ctx context.Context, obj *model.RentQuery, id int) (*model.RentHomestay, error) {
	return r.Clients.RentHouseService.Homestay(ctx, id)
}

func (r *rentQueryResolver) Homestays(ctx context.Context, obj *model.RentQuery, provinceID *int, districtID *int, wardID *int, typeArg *int, guests *int, qArg *string, modelArg *string, minPrice *string, maxPrice *string, sort *string, limit *int, cursor *string, amenityIds *string, checkin *string, checkout *string) (*model.RentHomestayPage, error) {
	return r.Clients.RentHouseService.Homestays(ctx, provinceID, districtID, wardID, typeArg, guests, qArg, modelArg, minPrice, maxPrice, sort, limit, cursor, amenityIds, checkin, checkout)
}

func (r *rentQueryResolver) Availability(ctx context.Context, obj *model.RentQuery, id int, from *string, to *string) ([]*model.RentSlot, error) {
	return r.Clients.RentHouseService.Availability(ctx, id, from, to)
}

func (r *rentQueryResolver) Amenities(ctx context.Context, obj *model.RentQuery) ([]*model.RentAmenity, error) {
	return r.Clients.RentHouseService.Amenities(ctx)
}

func (r *rentQueryResolver) MyHomestays(ctx context.Context, obj *model.RentQuery, limit *int, offset *int) ([]*model.RentHomestay, error) {
	return r.Clients.RentHouseService.MyHomestays(ctx, limit, offset)
}

func (r *rentQueryResolver) HomestayReviewQueue(ctx context.Context, obj *model.RentQuery, status *string, limit *int, offset *int) ([]*model.RentHomestay, error) {
	return r.Clients.RentHouseService.HomestayReviewQueue(ctx, status, limit, offset)
}

func (r *rentQueryResolver) Bookings(ctx context.Context, obj *model.RentQuery, userID *int, as *string, limit *int, offset *int) ([]*model.RentBooking, error) {
	return r.Clients.RentHouseService.Bookings(ctx, userID, as, limit, offset)
}

func (r *rentQueryResolver) Booking(ctx context.Context, obj *model.RentQuery, id int) (*model.RentBooking, error) {
	return r.Clients.RentHouseService.Booking(ctx, id)
}

func (r *rentQueryResolver) Rates(ctx context.Context, obj *model.RentQuery, id int) ([]*model.RentRate, error) {
	return r.Clients.RentHouseService.Rates(ctx, id)
}

func (r *rentQueryResolver) Leases(ctx context.Context, obj *model.RentQuery, userID *int, as *string, limit *int, offset *int) ([]*model.RentLease, error) {
	return r.Clients.RentHouseService.Leases(ctx, userID, as, limit, offset)
}

func (r *rentQueryResolver) Lease(ctx context.Context, obj *model.RentQuery, id int) (*model.RentLease, error) {
	return r.Clients.RentHouseService.Lease(ctx, id)
}

func (r *rentQueryResolver) Invoices(ctx context.Context, obj *model.RentQuery, limit *int, offset *int, status *string, dueBefore *string) ([]*model.RentInvoice, error) {
	return r.Clients.RentHouseService.Invoices(ctx, limit, offset, status, dueBefore)
}

func (r *rentQueryResolver) Reviews(ctx context.Context, obj *model.RentQuery, id int, limit *int, offset *int) ([]*model.RentReview, error) {
	return r.Clients.RentHouseService.Reviews(ctx, id, limit, offset)
}

func (r *rentQueryResolver) Wishlist(ctx context.Context, obj *model.RentQuery, limit *int, offset *int) ([]*model.RentHomestay, error) {
	return r.Clients.RentHouseService.Wishlist(ctx, limit, offset)
}

func (r *rentQueryResolver) Loyalty(ctx context.Context, obj *model.RentQuery) (*model.RentLoyalty, error) {
	return r.Clients.RentHouseService.Loyalty(ctx)
}

func (r *rentQueryResolver) Wallet(ctx context.Context, obj *model.RentQuery) (*model.RentWallet, error) {
	return r.Clients.RentHouseService.Wallet(ctx)
}

func (r *rentQueryResolver) WalletTransactions(ctx context.Context, obj *model.RentQuery, limit *int, offset *int) ([]*model.RentWalletTxn, error) {
	return r.Clients.RentHouseService.WalletTransactions(ctx, limit, offset)
}
