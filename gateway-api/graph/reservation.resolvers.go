package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) ResvRegisterHotel(ctx context.Context, input model.ResvHotelInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: "/api/v1/hotels", Body: input}, "")
}

func (r *mutationResolver) ResvApproveHotel(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/hotels/%v/approve", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ResvRejectHotel(ctx context.Context, id int, reason string) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/hotels/%v/reject", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) ResvUpdateHotel(ctx context.Context, id int, input model.ResvHotelInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/hotels/%v", id), Body: input}, "")
}

func (r *mutationResolver) ResvDeactivateHotel(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "reservation-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/hotels/%v", id)}, "")
}

func (r *mutationResolver) ResvCreateRoomType(ctx context.Context, id int, input model.ResvRoomTypeInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/hotels/%v/room-types", id), Body: input}, "")
}

func (r *mutationResolver) ResvUpdateRoomType(ctx context.Context, id int, typeArg string, input model.ResvRoomTypeInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/hotels/%v/room-types/%v", id, typeArg), Body: input}, "")
}

func (r *mutationResolver) ResvSetInventory(ctx context.Context, id int, typeArg string, input model.ResvInventoryInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/hotels/%v/room-types/%v/inventory", id, typeArg), Body: input}, "")
}

func (r *mutationResolver) ResvCreateRoom(ctx context.Context, id int, input model.ResvRoomInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/hotels/%v/rooms", id), Body: input}, "")
}

func (r *mutationResolver) ResvUpdateRoom(ctx context.Context, id int, room string, input model.ResvRoomInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/hotels/%v/rooms/%v", id, room), Body: input}, "")
}

func (r *mutationResolver) ResvCreatePromotion(ctx context.Context, id int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/hotels/%v/promotions", id), Body: input}, "")
}

func (r *mutationResolver) ResvUpdatePromotion(ctx context.Context, id int, promo string, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/hotels/%v/promotions/%v", id, promo), Body: input}, "")
}

func (r *mutationResolver) ResvCreateService(ctx context.Context, id int, input model.ResvServiceInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/hotels/%v/services", id), Body: input}, "")
}

func (r *mutationResolver) ResvUpdateService(ctx context.Context, id int, service string, input model.ResvServiceInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/hotels/%v/services/%v", id, service), Body: input}, "")
}

func (r *mutationResolver) ResvReserve(ctx context.Context, input model.ResvReservationInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: "/api/v1/reservations", Body: input}, "")
}

func (r *mutationResolver) ResvPayReservation(ctx context.Context, id int, method string, provider *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/reservations/%v/pay", id), Body: map[string]any{"method": method, "provider": provider}}, "")
}

func (r *mutationResolver) ResvCancelReservation(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/reservations/%v/cancel", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) ResvRejectReservation(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/reservations/%v/reject", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) ResvCheckInReservation(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/reservations/%v/check-in", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) ResvCheckOutReservation(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/reservations/%v/check-out", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) ResvNoShowReservation(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/reservations/%v/no-show", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) ResvCreateReview(ctx context.Context, id int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/hotels/%v/reviews", id), Body: input}, "")
}

func (r *mutationResolver) ResvJoinWaitlist(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: "/api/v1/waitlist", Body: input}, "")
}

func (r *mutationResolver) ResvLeaveWaitlist(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "reservation-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/waitlist/%v", id)}, "")
}

func (r *mutationResolver) ResvAddFavourite(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "reservation-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/wishlist/%v", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ResvRemoveFavourite(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "reservation-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/wishlist/%v", id)}, "")
}

func (r *mutationResolver) ResvMarkAllNotificationsRead(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: "/api/v1/notifications/read-all", Body: map[string]any{}}, "")
}

func (r *mutationResolver) ResvMarkNotificationRead(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/notifications/%v/read", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) ResvCreateWallet(ctx context.Context, currency string) (any, error) {
	return restCall[any](ctx, r.Proxy, "reservation-service", rest.Request{Method: "POST", Path: "/api/v1/wallet", Body: map[string]any{"currency": currency}}, "")
}

func (r *queryResolver) ResvHotels(ctx context.Context, city *string, q *string, rooms *int, guests *int, sort *string, limit *int, minRate *string, maxRate *string, amenities *string, cursor *string, checkIn *string, checkOut *string) (*model.ResvHotelPage, error) {
	return restCall[*model.ResvHotelPage](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/hotels", Query: restQuery("city", city, "q", q, "rooms", rooms, "guests", guests, "sort", sort, "limit", limit, "min_rate", minRate, "max_rate", maxRate, "amenities", amenities, "cursor", cursor, "check_in", checkIn, "check_out", checkOut)}, "")
}

func (r *queryResolver) ResvHotel(ctx context.Context, id int) (*model.ResvHotel, error) {
	return restCall[*model.ResvHotel](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v", id)}, "")
}

func (r *queryResolver) ResvRoomTypes(ctx context.Context, id int) ([]*model.ResvRoomType, error) {
	return restCall[[]*model.ResvRoomType](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v/room-types", id)}, "")
}

func (r *queryResolver) ResvInventory(ctx context.Context, id int, typeArg int, from *string, to *string) ([]*model.ResvInventory, error) {
	return restCall[[]*model.ResvInventory](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v/room-types/%v/inventory", id, typeArg), Query: restQuery("from", from, "to", to)}, "")
}

func (r *queryResolver) ResvQuotes(ctx context.Context, id int, checkIn string, checkOut string, rooms *int) ([]*model.ResvQuote, error) {
	return restCall[[]*model.ResvQuote](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v/quotes", id), Query: restQuery("check_in", checkIn, "check_out", checkOut, "rooms", rooms)}, "")
}

func (r *queryResolver) ResvServices(ctx context.Context, id int) ([]*model.ResvService, error) {
	return restCall[[]*model.ResvService](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v/services", id)}, "")
}

func (r *queryResolver) ResvReviews(ctx context.Context, id int, cursor *string, limit *int) (*model.ResvReviewPage, error) {
	return restCall[*model.ResvReviewPage](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v/reviews", id), Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ResvMyHotels(ctx context.Context, limit *int) (*model.ResvHotelPage, error) {
	return restCall[*model.ResvHotelPage](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/my/hotels", Query: restQuery("limit", limit)}, "")
}

func (r *queryResolver) ResvHotelReviewQueue(ctx context.Context, status *string, limit *int) (*model.ResvHotelPage, error) {
	return restCall[*model.ResvHotelPage](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/admin/hotels", Query: restQuery("status", status, "limit", limit)}, "")
}

func (r *queryResolver) ResvRooms(ctx context.Context, id int, roomTypeID *int) ([]*model.ResvRoom, error) {
	return restCall[[]*model.ResvRoom](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v/rooms", id), Query: restQuery("room_type_id", roomTypeID)}, "")
}

func (r *queryResolver) ResvPromotions(ctx context.Context, id int) ([]*model.ResvPromotion, error) {
	return restCall[[]*model.ResvPromotion](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v/promotions", id)}, "")
}

func (r *queryResolver) ResvHotelReport(ctx context.Context, id int, from *string, to *string) (*model.ResvReport, error) {
	return restCall[*model.ResvReport](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/hotels/%v/report", id), Query: restQuery("from", from, "to", to)}, "")
}

func (r *queryResolver) ResvReservations(ctx context.Context, status *string, cursor *string, as *string, hotelID *int, limit *int) (*model.ResvReservationPage, error) {
	return restCall[*model.ResvReservationPage](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/reservations", Query: restQuery("status", status, "cursor", cursor, "as", as, "hotel_id", hotelID, "limit", limit)}, "")
}

func (r *queryResolver) ResvReservation(ctx context.Context, id int) (*model.ResvReservation, error) {
	return restCall[*model.ResvReservation](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/reservations/%v", id)}, "")
}

func (r *queryResolver) ResvCancellationQuote(ctx context.Context, id int) (*model.ResvCancellationQuote, error) {
	return restCall[*model.ResvCancellationQuote](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/reservations/%v/cancellation-quote", id)}, "")
}

func (r *queryResolver) ResvReservationHistory(ctx context.Context, id int) ([]*model.ResvHistoryEntry, error) {
	return restCall[[]*model.ResvHistoryEntry](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/reservations/%v/history", id)}, "")
}

func (r *queryResolver) ResvWaitlist(ctx context.Context, cursor *string, limit *int) (*model.ResvWaitPage, error) {
	return restCall[*model.ResvWaitPage](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/waitlist", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ResvWishlist(ctx context.Context, cursor *string, limit *int) (*model.ResvHotelPage, error) {
	return restCall[*model.ResvHotelPage](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/wishlist", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) ResvNotifications(ctx context.Context, cursor *string, unread *bool, limit *int) (*model.ResvNotificationPage, error) {
	return restCall[*model.ResvNotificationPage](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/notifications", Query: restQuery("cursor", cursor, "unread", unread, "limit", limit)}, "")
}

func (r *queryResolver) ResvUnreadNotifications(ctx context.Context) (*model.ResvUnreadCount, error) {
	return restCall[*model.ResvUnreadCount](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/notifications/unread-count"}, "")
}

func (r *queryResolver) ResvLoyalty(ctx context.Context) (*model.ResvLoyalty, error) {
	return restCall[*model.ResvLoyalty](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/loyalty"}, "")
}

func (r *queryResolver) ResvWallet(ctx context.Context) (*model.ResvWallet, error) {
	return restCall[*model.ResvWallet](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/wallet"}, "")
}

func (r *queryResolver) ResvWalletTransactions(ctx context.Context, limit *int, offset *int) ([]*model.ResvWalletTxn, error) {
	return restCall[[]*model.ResvWalletTxn](ctx, r.Proxy, "reservation-service", rest.Request{Method: "GET", Path: "/api/v1/wallet/transactions", Query: restQuery("limit", limit, "offset", offset)}, "")
}
