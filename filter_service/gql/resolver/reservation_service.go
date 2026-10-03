package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) ResvQuery() generated.ResvQueryResolver { return &resvQueryResolver{r} }

type resvQueryResolver struct{ *Resolver }

func (r *resvQueryResolver) Hotels(ctx context.Context, obj *model.ResvQuery, city *string, qArg *string, rooms *int, guests *int, sort *string, limit *int, minRate *string, maxRate *string, amenities *string, cursor *string, checkIn *string, checkOut *string) (*model.ResvHotelPage, error) {
	return r.Clients.ReservationService.Hotels(ctx, city, qArg, rooms, guests, sort, limit, minRate, maxRate, amenities, cursor, checkIn, checkOut)
}

func (r *resvQueryResolver) Hotel(ctx context.Context, obj *model.ResvQuery, id int) (*model.ResvHotel, error) {
	return r.Clients.ReservationService.Hotel(ctx, id)
}

func (r *resvQueryResolver) RoomTypes(ctx context.Context, obj *model.ResvQuery, id int) ([]*model.ResvRoomType, error) {
	return r.Clients.ReservationService.RoomTypes(ctx, id)
}

func (r *resvQueryResolver) Inventory(ctx context.Context, obj *model.ResvQuery, id int, typeArg int, from *string, to *string) ([]*model.ResvInventory, error) {
	return r.Clients.ReservationService.Inventory(ctx, id, typeArg, from, to)
}

func (r *resvQueryResolver) Quotes(ctx context.Context, obj *model.ResvQuery, id int, checkIn string, checkOut string, rooms *int) ([]*model.ResvQuote, error) {
	return r.Clients.ReservationService.Quotes(ctx, id, checkIn, checkOut, rooms)
}

func (r *resvQueryResolver) Services(ctx context.Context, obj *model.ResvQuery, id int) ([]*model.ResvService, error) {
	return r.Clients.ReservationService.Services(ctx, id)
}

func (r *resvQueryResolver) Reviews(ctx context.Context, obj *model.ResvQuery, id int, cursor *string, limit *int) (*model.ResvReviewPage, error) {
	return r.Clients.ReservationService.Reviews(ctx, id, cursor, limit)
}

func (r *resvQueryResolver) MyHotels(ctx context.Context, obj *model.ResvQuery, limit *int) (*model.ResvHotelPage, error) {
	return r.Clients.ReservationService.MyHotels(ctx, limit)
}

func (r *resvQueryResolver) HotelReviewQueue(ctx context.Context, obj *model.ResvQuery, status *string, limit *int) (*model.ResvHotelPage, error) {
	return r.Clients.ReservationService.HotelReviewQueue(ctx, status, limit)
}

func (r *resvQueryResolver) Rooms(ctx context.Context, obj *model.ResvQuery, id int, roomTypeID *int) ([]*model.ResvRoom, error) {
	return r.Clients.ReservationService.Rooms(ctx, id, roomTypeID)
}

func (r *resvQueryResolver) Promotions(ctx context.Context, obj *model.ResvQuery, id int) ([]*model.ResvPromotion, error) {
	return r.Clients.ReservationService.Promotions(ctx, id)
}

func (r *resvQueryResolver) HotelReport(ctx context.Context, obj *model.ResvQuery, id int, from *string, to *string) (*model.ResvReport, error) {
	return r.Clients.ReservationService.HotelReport(ctx, id, from, to)
}

func (r *resvQueryResolver) Reservations(ctx context.Context, obj *model.ResvQuery, status *string, cursor *string, as *string, hotelID *int, limit *int) (*model.ResvReservationPage, error) {
	return r.Clients.ReservationService.Reservations(ctx, status, cursor, as, hotelID, limit)
}

func (r *resvQueryResolver) Reservation(ctx context.Context, obj *model.ResvQuery, id int) (*model.ResvReservation, error) {
	return r.Clients.ReservationService.Reservation(ctx, id)
}

func (r *resvQueryResolver) CancellationQuote(ctx context.Context, obj *model.ResvQuery, id int) (*model.ResvCancellationQuote, error) {
	return r.Clients.ReservationService.CancellationQuote(ctx, id)
}

func (r *resvQueryResolver) ReservationHistory(ctx context.Context, obj *model.ResvQuery, id int) ([]*model.ResvHistoryEntry, error) {
	return r.Clients.ReservationService.ReservationHistory(ctx, id)
}

func (r *resvQueryResolver) Waitlist(ctx context.Context, obj *model.ResvQuery, cursor *string, limit *int) (*model.ResvWaitPage, error) {
	return r.Clients.ReservationService.Waitlist(ctx, cursor, limit)
}

func (r *resvQueryResolver) Wishlist(ctx context.Context, obj *model.ResvQuery, cursor *string, limit *int) (*model.ResvHotelPage, error) {
	return r.Clients.ReservationService.Wishlist(ctx, cursor, limit)
}

func (r *resvQueryResolver) Notifications(ctx context.Context, obj *model.ResvQuery, cursor *string, unread *bool, limit *int) (*model.ResvNotificationPage, error) {
	return r.Clients.ReservationService.Notifications(ctx, cursor, unread, limit)
}

func (r *resvQueryResolver) UnreadNotifications(ctx context.Context, obj *model.ResvQuery) (*model.ResvUnreadCount, error) {
	return r.Clients.ReservationService.UnreadNotifications(ctx)
}

func (r *resvQueryResolver) Loyalty(ctx context.Context, obj *model.ResvQuery) (*model.ResvLoyalty, error) {
	return r.Clients.ReservationService.Loyalty(ctx)
}

func (r *resvQueryResolver) Wallet(ctx context.Context, obj *model.ResvQuery) (*model.ResvWallet, error) {
	return r.Clients.ReservationService.Wallet(ctx)
}

func (r *resvQueryResolver) WalletTransactions(ctx context.Context, obj *model.ResvQuery, limit *int, offset *int) ([]*model.ResvWalletTxn, error) {
	return r.Clients.ReservationService.WalletTransactions(ctx, limit, offset)
}
