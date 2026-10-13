package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) TicketCreateEvent(ctx context.Context, input model.TicketEventInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: "/api/v1/events", Body: input}, "")
}

func (r *mutationResolver) TicketUpdateEvent(ctx context.Context, id int, input model.TicketEventInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/events/%v", id), Body: input}, "")
}

func (r *mutationResolver) TicketSubmitEvent(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/submit", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketApproveEvent(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/approve", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketRejectEvent(ctx context.Context, id int, reason string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/reject", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) TicketSetFeatured(ctx context.Context, id int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/events/%v/featured", id), Body: input}, "")
}

func (r *mutationResolver) TicketDuplicateEvent(ctx context.Context, id int, input model.TicketDuplicateInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/duplicate", id), Body: input}, "")
}

func (r *mutationResolver) TicketCancelEvent(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/cancel", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) TicketPreviewVenue(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: "/api/v1/venues/preview", Body: input}, "")
}

func (r *mutationResolver) TicketCreateVenue(ctx context.Context, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: "/api/v1/venues", Body: input}, "")
}

func (r *mutationResolver) TicketUpdateVenue(ctx context.Context, id int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/venues/%v", id), Body: input}, "")
}

func (r *mutationResolver) TicketDeleteVenue(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/venues/%v", id)}, "")
}

func (r *mutationResolver) TicketCopyVenue(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/venues/%v/copy", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketShareVenue(ctx context.Context, id int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/venues/%v/shared", id), Body: input}, "")
}

func (r *mutationResolver) TicketApplyVenue(ctx context.Context, id int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/seat-map/apply", id), Body: input}, "")
}

func (r *mutationResolver) TicketCreateSession(ctx context.Context, id int, input model.TicketSessionInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/sessions", id), Body: input}, "")
}

func (r *mutationResolver) TicketUpdateSession(ctx context.Context, id int, session int, input model.TicketSessionInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/events/%v/sessions/%v", id, session), Body: input}, "")
}

func (r *mutationResolver) TicketDeleteSession(ctx context.Context, id int, session int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/events/%v/sessions/%v", id, session)}, "")
}

func (r *mutationResolver) TicketCancelSession(ctx context.Context, id int, session int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/sessions/%v/cancel", id, session), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) TicketCreateTicketType(ctx context.Context, id int, input model.TicketTypeInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/ticket-types", id), Body: input}, "")
}

func (r *mutationResolver) TicketUpdateTicketType(ctx context.Context, id int, typeArg int, input model.TicketTypeInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/events/%v/ticket-types/%v", id, typeArg), Body: input}, "")
}

func (r *mutationResolver) TicketGenerateSeats(ctx context.Context, id int, typeArg int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/ticket-types/%v/seats", id, typeArg), Body: input}, "")
}

func (r *mutationResolver) TicketSetSeatMap(ctx context.Context, id int, typeArg int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/events/%v/ticket-types/%v/seat-map", id, typeArg), Body: input}, "")
}

func (r *mutationResolver) TicketJoinWaitlist(ctx context.Context, id int, typeArg int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/events/%v/ticket-types/%v/waitlist", id, typeArg), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketLeaveWaitlist(ctx context.Context, id int, typeArg int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/events/%v/ticket-types/%v/waitlist", id, typeArg)}, "")
}

func (r *mutationResolver) TicketCreatePromotion(ctx context.Context, id int, input model.TicketPromotionInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/promotions", id), Body: input}, "")
}

func (r *mutationResolver) TicketUpdatePromotion(ctx context.Context, id int, promo int, input model.TicketPromotionInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/events/%v/promotions/%v", id, promo), Body: input}, "")
}

func (r *mutationResolver) TicketCheckIn(ctx context.Context, id int, code string, scannedAt *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/check-in", id), Body: map[string]any{"code": code, "scanned_at": scannedAt}}, "")
}

func (r *mutationResolver) TicketBatchCheckIn(ctx context.Context, id int, input map[string]any) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/check-in/batch", id), Body: input}, "")
}

func (r *mutationResolver) TicketInvite(ctx context.Context, id int, input model.TicketInviteInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/invitations", id), Body: input}, "")
}

func (r *mutationResolver) TicketAddStaff(ctx context.Context, id int, userID int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/events/%v/staff", id), Body: map[string]any{"user_id": userID}}, "")
}

func (r *mutationResolver) TicketRemoveStaff(ctx context.Context, id int, user int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/events/%v/staff/%v", id, user)}, "")
}

func (r *mutationResolver) TicketReserve(ctx context.Context, input model.TicketReserveInput) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: "/api/v1/orders", Body: input}, "")
}

func (r *mutationResolver) TicketPayOrder(ctx context.Context, id int, method string, provider *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/pay", id), Body: map[string]any{"method": method, "provider": provider}}, "")
}

func (r *mutationResolver) TicketCancelOrder(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/orders/%v/cancel", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) TicketSetHolder(ctx context.Context, id int, name string, email *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/tickets/%v/holder", id), Body: map[string]any{"name": name, "email": email}}, "")
}

func (r *mutationResolver) TicketReissue(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/tickets/%v/reissue", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketListForResale(ctx context.Context, id int, price int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/tickets/%v/resale", id), Body: map[string]any{"price": price}}, "")
}

func (r *mutationResolver) TicketCancelResale(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/resale/%v", id)}, "")
}

func (r *mutationResolver) TicketBuyResale(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/resale/%v/buy", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketOfferTransfer(ctx context.Context, id int, toUserID *int, toEmail *string, message *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/tickets/%v/transfer", id), Body: map[string]any{"to_user_id": toUserID, "to_email": toEmail, "message": message}}, "")
}

func (r *mutationResolver) TicketAcceptTransfer(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/transfers/%v/accept", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketDeclineTransfer(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/transfers/%v/decline", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketCancelTransfer(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/transfers/%v/cancel", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketVoid(ctx context.Context, id int, reason *string) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/tickets/%v/void", id), Body: map[string]any{"reason": reason}}, "")
}

func (r *mutationResolver) TicketRevertCheckIn(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/tickets/%v/revert-check-in", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketAddWishlist(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/wishlist/%v", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketRemoveWishlist(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/wishlist/%v", id)}, "")
}

func (r *mutationResolver) TicketMarkAllNotificationsRead(ctx context.Context) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: "/api/v1/notifications/read-all", Body: map[string]any{}}, "")
}

func (r *mutationResolver) TicketMarkNotificationRead(ctx context.Context, id int) (any, error) {
	return restCall[any](ctx, r.Proxy, "ticket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/notifications/%v/read", id), Body: map[string]any{}}, "")
}

func (r *queryResolver) TicketEvents(ctx context.Context, q *string, city *string, category *string, sort *string, featured *bool, limit *int, offset *int, from *string, to *string, minPrice *string, maxPrice *string) ([]*model.TicketEvent, error) {
	return restCall[[]*model.TicketEvent](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/events", Query: restQuery("q", q, "city", city, "category", category, "sort", sort, "featured", featured, "limit", limit, "offset", offset, "from", from, "to", to, "min_price", minPrice, "max_price", maxPrice)}, "items")
}

func (r *queryResolver) TicketCategories(ctx context.Context) ([]string, error) {
	return restCall[[]string](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/categories"}, "items")
}

func (r *queryResolver) TicketEvent(ctx context.Context, id int) (*model.TicketEvent, error) {
	return restCall[*model.TicketEvent](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v", id)}, "")
}

func (r *queryResolver) TicketSeats(ctx context.Context, id int, typeArg int) (*model.TicketSeatList, error) {
	return restCall[*model.TicketSeatList](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/ticket-types/%v/seats", id, typeArg)}, "")
}

func (r *queryResolver) TicketEventSeries(ctx context.Context, id int) ([]*model.TicketEvent, error) {
	return restCall[[]*model.TicketEvent](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/series", id)}, "items")
}

func (r *queryResolver) TicketSessions(ctx context.Context, id int) ([]*model.TicketSession, error) {
	return restCall[[]*model.TicketSession](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/sessions", id)}, "items")
}

func (r *queryResolver) TicketSeatMap(ctx context.Context, id int) (*model.TicketSeatMap, error) {
	return restCall[*model.TicketSeatMap](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/seat-map", id)}, "")
}

func (r *queryResolver) TicketVenueTemplates(ctx context.Context) ([]*model.TicketTemplate, error) {
	return restCall[[]*model.TicketTemplate](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/venue-templates"}, "items")
}

func (r *queryResolver) TicketResale(ctx context.Context, id int, typeArg *int, sort *string, limit *int, cursor *string) (*model.TicketResalePage, error) {
	return restCall[*model.TicketResalePage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/resale", id), Query: restQuery("type", typeArg, "sort", sort, "limit", limit, "cursor", cursor)}, "")
}

func (r *queryResolver) TicketMyEvents(ctx context.Context, cursor *string, limit *int) (*model.TicketEventPage, error) {
	return restCall[*model.TicketEventPage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/my/events", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketVenues(ctx context.Context, scope *string, cursor *string, limit *int) (*model.TicketVenuePage, error) {
	return restCall[*model.TicketVenuePage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/venues", Query: restQuery("scope", scope, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketVenue(ctx context.Context, id int, seats *bool) (*model.TicketVenueDetail, error) {
	return restCall[*model.TicketVenueDetail](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/venues/%v", id), Query: restQuery("seats", seats)}, "")
}

func (r *queryResolver) TicketPromotions(ctx context.Context, id int) ([]*model.TicketPromotion, error) {
	return restCall[[]*model.TicketPromotion](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/promotions", id)}, "items")
}

func (r *queryResolver) TicketEventReport(ctx context.Context, id int) (*model.TicketReport, error) {
	return restCall[*model.TicketReport](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/report", id)}, "")
}

func (r *queryResolver) TicketEventOrders(ctx context.Context, id int, status *string, cursor *string, limit *int) (*model.TicketOrderPage, error) {
	return restCall[*model.TicketOrderPage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/orders", id), Query: restQuery("status", status, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketAttendees(ctx context.Context, id int, cursor *string, limit *int) (*model.TicketAttendeePage, error) {
	return restCall[*model.TicketAttendeePage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/attendees", id), Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketLookupTicket(ctx context.Context, id int, code string) (*model.TicketTicket, error) {
	return restCall[*model.TicketTicket](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/tickets/%v", id, code)}, "")
}

func (r *queryResolver) TicketStaff(ctx context.Context, id int) ([]*model.TicketStaff, error) {
	return restCall[[]*model.TicketStaff](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/events/%v/staff", id)}, "items")
}

func (r *queryResolver) TicketStaffEvents(ctx context.Context) ([]*model.TicketEvent, error) {
	return restCall[[]*model.TicketEvent](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/my/staff-events"}, "items")
}

func (r *queryResolver) TicketEventReviewQueue(ctx context.Context, cursor *string, limit *int) (*model.TicketEventPage, error) {
	return restCall[*model.TicketEventPage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/admin/events", Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketOrders(ctx context.Context, status *string, eventID *int, cursor *string, limit *int) (*model.TicketOrderPage, error) {
	return restCall[*model.TicketOrderPage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/orders", Query: restQuery("status", status, "event_id", eventID, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketOrder(ctx context.Context, id int) (*model.TicketOrder, error) {
	return restCall[*model.TicketOrder](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "")
}

func (r *queryResolver) TicketInvoice(ctx context.Context, id int) (*model.TicketInvoice, error) {
	return restCall[*model.TicketInvoice](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v/invoice", id)}, "")
}

func (r *queryResolver) TicketOrderDocuments(ctx context.Context, id int) ([]*model.TicketDocument, error) {
	return restCall[[]*model.TicketDocument](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v/documents", id)}, "items")
}

func (r *queryResolver) Tickets(ctx context.Context, status *string, eventID *int, cursor *string, limit *int) (*model.TicketTicketPage, error) {
	return restCall[*model.TicketTicketPage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/tickets", Query: restQuery("status", status, "event_id", eventID, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketMyTickets(ctx context.Context, status *string, eventID *int, cursor *string, limit *int) (*model.TicketTicketPage, error) {
	return restCall[*model.TicketTicketPage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/my/tickets", Query: restQuery("status", status, "event_id", eventID, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) Ticket(ctx context.Context, id int) (*model.TicketTicket, error) {
	return restCall[*model.TicketTicket](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/tickets/%v", id)}, "")
}

func (r *queryResolver) TicketHistory(ctx context.Context, id int) ([]*model.TicketTicketEvent, error) {
	return restCall[[]*model.TicketTicketEvent](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/tickets/%v/history", id)}, "items")
}

func (r *queryResolver) TicketMyResale(ctx context.Context, status *string, cursor *string, limit *int) (*model.TicketResalePage, error) {
	return restCall[*model.TicketResalePage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/resale", Query: restQuery("status", status, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketTransfers(ctx context.Context, direction *string, cursor *string, limit *int) (*model.TicketTransferPage, error) {
	return restCall[*model.TicketTransferPage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/transfers", Query: restQuery("direction", direction, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketWaitlist(ctx context.Context) ([]*model.TicketWaitlist, error) {
	return restCall[[]*model.TicketWaitlist](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/waitlist"}, "items")
}

func (r *queryResolver) TicketWishlist(ctx context.Context, limit *int, offset *int) ([]*model.TicketEvent, error) {
	return restCall[[]*model.TicketEvent](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/wishlist", Query: restQuery("limit", limit, "offset", offset)}, "items")
}

func (r *queryResolver) TicketNotifications(ctx context.Context, unread *bool, cursor *string, limit *int) (*model.TicketNotificationPage, error) {
	return restCall[*model.TicketNotificationPage](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/notifications", Query: restQuery("unread", unread, "cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) TicketUnreadNotifications(ctx context.Context) (*model.TicketUnreadCount, error) {
	return restCall[*model.TicketUnreadCount](ctx, r.Proxy, "ticket-service", rest.Request{Method: "GET", Path: "/api/v1/notifications/unread-count"}, "")
}
