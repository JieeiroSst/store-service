package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) TicketQuery() generated.TicketQueryResolver { return &ticketQueryResolver{r} }

type ticketQueryResolver struct{ *Resolver }

func (r *ticketQueryResolver) Events(ctx context.Context, obj *model.TicketQuery, qArg *string, city *string, category *string, sort *string, featured *bool, limit *int, offset *int, from *string, to *string, minPrice *string, maxPrice *string) ([]*model.TicketEvent, error) {
	return r.Clients.TicketService.Events(ctx, qArg, city, category, sort, featured, limit, offset, from, to, minPrice, maxPrice)
}

func (r *ticketQueryResolver) Categories(ctx context.Context, obj *model.TicketQuery) ([]string, error) {
	return r.Clients.TicketService.Categories(ctx)
}

func (r *ticketQueryResolver) Event(ctx context.Context, obj *model.TicketQuery, id int) (*model.TicketEvent, error) {
	return r.Clients.TicketService.Event(ctx, id)
}

func (r *ticketQueryResolver) Seats(ctx context.Context, obj *model.TicketQuery, id int, typeArg int) (*model.TicketSeatList, error) {
	return r.Clients.TicketService.Seats(ctx, id, typeArg)
}

func (r *ticketQueryResolver) EventSeries(ctx context.Context, obj *model.TicketQuery, id int) ([]*model.TicketEvent, error) {
	return r.Clients.TicketService.EventSeries(ctx, id)
}

func (r *ticketQueryResolver) Sessions(ctx context.Context, obj *model.TicketQuery, id int) ([]*model.TicketSession, error) {
	return r.Clients.TicketService.Sessions(ctx, id)
}

func (r *ticketQueryResolver) SeatMap(ctx context.Context, obj *model.TicketQuery, id int) (*model.TicketSeatMap, error) {
	return r.Clients.TicketService.SeatMap(ctx, id)
}

func (r *ticketQueryResolver) VenueTemplates(ctx context.Context, obj *model.TicketQuery) ([]*model.TicketTemplate, error) {
	return r.Clients.TicketService.VenueTemplates(ctx)
}

func (r *ticketQueryResolver) Resale(ctx context.Context, obj *model.TicketQuery, id int, typeArg *int, sort *string, limit *int, cursor *string) (*model.TicketResalePage, error) {
	return r.Clients.TicketService.Resale(ctx, id, typeArg, sort, limit, cursor)
}

func (r *ticketQueryResolver) MyEvents(ctx context.Context, obj *model.TicketQuery, cursor *string, limit *int) (*model.TicketEventPage, error) {
	return r.Clients.TicketService.MyEvents(ctx, cursor, limit)
}

func (r *ticketQueryResolver) Venues(ctx context.Context, obj *model.TicketQuery, scope *string, cursor *string, limit *int) (*model.TicketVenuePage, error) {
	return r.Clients.TicketService.Venues(ctx, scope, cursor, limit)
}

func (r *ticketQueryResolver) Venue(ctx context.Context, obj *model.TicketQuery, id int, seats *bool) (*model.TicketVenueDetail, error) {
	return r.Clients.TicketService.Venue(ctx, id, seats)
}

func (r *ticketQueryResolver) Promotions(ctx context.Context, obj *model.TicketQuery, id int) ([]*model.TicketPromotion, error) {
	return r.Clients.TicketService.Promotions(ctx, id)
}

func (r *ticketQueryResolver) EventReport(ctx context.Context, obj *model.TicketQuery, id int) (*model.TicketReport, error) {
	return r.Clients.TicketService.EventReport(ctx, id)
}

func (r *ticketQueryResolver) EventOrders(ctx context.Context, obj *model.TicketQuery, id int, status *string, cursor *string, limit *int) (*model.TicketOrderPage, error) {
	return r.Clients.TicketService.EventOrders(ctx, id, status, cursor, limit)
}

func (r *ticketQueryResolver) Attendees(ctx context.Context, obj *model.TicketQuery, id int, cursor *string, limit *int) (*model.TicketAttendeePage, error) {
	return r.Clients.TicketService.Attendees(ctx, id, cursor, limit)
}

func (r *ticketQueryResolver) LookupTicket(ctx context.Context, obj *model.TicketQuery, id int, code string) (*model.TicketTicket, error) {
	return r.Clients.TicketService.LookupTicket(ctx, id, code)
}

func (r *ticketQueryResolver) Staff(ctx context.Context, obj *model.TicketQuery, id int) ([]*model.TicketStaff, error) {
	return r.Clients.TicketService.Staff(ctx, id)
}

func (r *ticketQueryResolver) StaffEvents(ctx context.Context, obj *model.TicketQuery) ([]*model.TicketEvent, error) {
	return r.Clients.TicketService.StaffEvents(ctx)
}

func (r *ticketQueryResolver) EventReviewQueue(ctx context.Context, obj *model.TicketQuery, cursor *string, limit *int) (*model.TicketEventPage, error) {
	return r.Clients.TicketService.EventReviewQueue(ctx, cursor, limit)
}

func (r *ticketQueryResolver) Orders(ctx context.Context, obj *model.TicketQuery, status *string, eventID *int, cursor *string, limit *int) (*model.TicketOrderPage, error) {
	return r.Clients.TicketService.Orders(ctx, status, eventID, cursor, limit)
}

func (r *ticketQueryResolver) Order(ctx context.Context, obj *model.TicketQuery, id int) (*model.TicketOrder, error) {
	return r.Clients.TicketService.Order(ctx, id)
}

func (r *ticketQueryResolver) Invoice(ctx context.Context, obj *model.TicketQuery, id int) (*model.TicketInvoice, error) {
	return r.Clients.TicketService.Invoice(ctx, id)
}

func (r *ticketQueryResolver) OrderDocuments(ctx context.Context, obj *model.TicketQuery, id int) ([]*model.TicketDocument, error) {
	return r.Clients.TicketService.OrderDocuments(ctx, id)
}

func (r *ticketQueryResolver) Tickets(ctx context.Context, obj *model.TicketQuery, status *string, eventID *int, cursor *string, limit *int) (*model.TicketTicketPage, error) {
	return r.Clients.TicketService.Tickets(ctx, status, eventID, cursor, limit)
}

func (r *ticketQueryResolver) MyTickets(ctx context.Context, obj *model.TicketQuery, status *string, eventID *int, cursor *string, limit *int) (*model.TicketTicketPage, error) {
	return r.Clients.TicketService.MyTickets(ctx, status, eventID, cursor, limit)
}

func (r *ticketQueryResolver) Ticket(ctx context.Context, obj *model.TicketQuery, id int) (*model.TicketTicket, error) {
	return r.Clients.TicketService.Ticket(ctx, id)
}

func (r *ticketQueryResolver) TicketHistory(ctx context.Context, obj *model.TicketQuery, id int) ([]*model.TicketTicketEvent, error) {
	return r.Clients.TicketService.TicketHistory(ctx, id)
}

func (r *ticketQueryResolver) MyResale(ctx context.Context, obj *model.TicketQuery, status *string, cursor *string, limit *int) (*model.TicketResalePage, error) {
	return r.Clients.TicketService.MyResale(ctx, status, cursor, limit)
}

func (r *ticketQueryResolver) Transfers(ctx context.Context, obj *model.TicketQuery, direction *string, cursor *string, limit *int) (*model.TicketTransferPage, error) {
	return r.Clients.TicketService.Transfers(ctx, direction, cursor, limit)
}

func (r *ticketQueryResolver) Waitlist(ctx context.Context, obj *model.TicketQuery) ([]*model.TicketWaitlist, error) {
	return r.Clients.TicketService.Waitlist(ctx)
}

func (r *ticketQueryResolver) Wishlist(ctx context.Context, obj *model.TicketQuery, limit *int, offset *int) ([]*model.TicketEvent, error) {
	return r.Clients.TicketService.Wishlist(ctx, limit, offset)
}

func (r *ticketQueryResolver) Notifications(ctx context.Context, obj *model.TicketQuery, unread *bool, cursor *string, limit *int) (*model.TicketNotificationPage, error) {
	return r.Clients.TicketService.Notifications(ctx, unread, cursor, limit)
}

func (r *ticketQueryResolver) UnreadNotifications(ctx context.Context, obj *model.TicketQuery) (*model.TicketUnreadCount, error) {
	return r.Clients.TicketService.UnreadNotifications(ctx)
}
