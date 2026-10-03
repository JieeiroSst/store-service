package ticket_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "ticket-service"

const DefaultBaseURL = "http://ticket-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Events(ctx context.Context, qArg *string, city *string, category *string, sort *string, featured *bool, limit *int, offset *int, from *string, to *string, minPrice *string, maxPrice *string) ([]*model.TicketEvent, error) {
	path := "/api/v1/events"
	q := url.Values{}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if city != nil {
		q.Set("city", *city)
	}
	if category != nil {
		q.Set("category", *category)
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	if featured != nil {
		q.Set("featured", strconv.FormatBool(*featured))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	if minPrice != nil {
		q.Set("min_price", *minPrice)
	}
	if maxPrice != nil {
		q.Set("max_price", *maxPrice)
	}
	h := http.Header{}
	var out []*model.TicketEvent
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Categories(ctx context.Context) ([]string, error) {
	path := "/api/v1/categories"
	q := url.Values{}
	h := http.Header{}
	var out []string
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Event(ctx context.Context, id int) (*model.TicketEvent, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketEvent
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Seats(ctx context.Context, id int, typeArg int) (*model.TicketSeatList, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/ticket-types/" + url.PathEscape(strconv.Itoa(typeArg)) + "/seats"
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketSeatList
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) EventSeries(ctx context.Context, id int) ([]*model.TicketEvent, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/series"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketEvent
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Sessions(ctx context.Context, id int) ([]*model.TicketSession, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/sessions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketSession
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) SeatMap(ctx context.Context, id int) (*model.TicketSeatMap, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/seat-map"
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketSeatMap
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) VenueTemplates(ctx context.Context) ([]*model.TicketTemplate, error) {
	path := "/api/v1/venue-templates"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketTemplate
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Resale(ctx context.Context, id int, typeArg *int, sort *string, limit *int, cursor *string) (*model.TicketResalePage, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/resale"
	q := url.Values{}
	if typeArg != nil {
		q.Set("type", strconv.Itoa(*typeArg))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	h := http.Header{}
	var out *model.TicketResalePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MyEvents(ctx context.Context, cursor *string, limit *int) (*model.TicketEventPage, error) {
	path := "/api/v1/my/events"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketEventPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Venues(ctx context.Context, scope *string, cursor *string, limit *int) (*model.TicketVenuePage, error) {
	path := "/api/v1/venues"
	q := url.Values{}
	if scope != nil {
		q.Set("scope", *scope)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketVenuePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Venue(ctx context.Context, id int, seats *bool) (*model.TicketVenueDetail, error) {
	path := "/api/v1/venues/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	if seats != nil {
		q.Set("seats", strconv.FormatBool(*seats))
	}
	h := http.Header{}
	var out *model.TicketVenueDetail
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Promotions(ctx context.Context, id int) ([]*model.TicketPromotion, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/promotions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketPromotion
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) EventReport(ctx context.Context, id int) (*model.TicketReport, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/report"
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketReport
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) EventOrders(ctx context.Context, id int, status *string, cursor *string, limit *int) (*model.TicketOrderPage, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/orders"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketOrderPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Attendees(ctx context.Context, id int, cursor *string, limit *int) (*model.TicketAttendeePage, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/attendees"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketAttendeePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) LookupTicket(ctx context.Context, id int, code string) (*model.TicketTicket, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/tickets/" + url.PathEscape(code)
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketTicket
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Staff(ctx context.Context, id int) ([]*model.TicketStaff, error) {
	path := "/api/v1/events/" + url.PathEscape(strconv.Itoa(id)) + "/staff"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketStaff
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) StaffEvents(ctx context.Context) ([]*model.TicketEvent, error) {
	path := "/api/v1/my/staff-events"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketEvent
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) EventReviewQueue(ctx context.Context, cursor *string, limit *int) (*model.TicketEventPage, error) {
	path := "/api/v1/admin/events"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketEventPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Orders(ctx context.Context, status *string, eventID *int, cursor *string, limit *int) (*model.TicketOrderPage, error) {
	path := "/api/v1/orders"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if eventID != nil {
		q.Set("event_id", strconv.Itoa(*eventID))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketOrderPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Order(ctx context.Context, id int) (*model.TicketOrder, error) {
	path := "/api/v1/orders/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketOrder
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Invoice(ctx context.Context, id int) (*model.TicketInvoice, error) {
	path := "/api/v1/orders/" + url.PathEscape(strconv.Itoa(id)) + "/invoice"
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketInvoice
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) OrderDocuments(ctx context.Context, id int) ([]*model.TicketDocument, error) {
	path := "/api/v1/orders/" + url.PathEscape(strconv.Itoa(id)) + "/documents"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketDocument
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Tickets(ctx context.Context, status *string, eventID *int, cursor *string, limit *int) (*model.TicketTicketPage, error) {
	path := "/api/v1/tickets"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if eventID != nil {
		q.Set("event_id", strconv.Itoa(*eventID))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketTicketPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MyTickets(ctx context.Context, status *string, eventID *int, cursor *string, limit *int) (*model.TicketTicketPage, error) {
	path := "/api/v1/my/tickets"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if eventID != nil {
		q.Set("event_id", strconv.Itoa(*eventID))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketTicketPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Ticket(ctx context.Context, id int) (*model.TicketTicket, error) {
	path := "/api/v1/tickets/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketTicket
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) TicketHistory(ctx context.Context, id int) ([]*model.TicketTicketEvent, error) {
	path := "/api/v1/tickets/" + url.PathEscape(strconv.Itoa(id)) + "/history"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketTicketEvent
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) MyResale(ctx context.Context, status *string, cursor *string, limit *int) (*model.TicketResalePage, error) {
	path := "/api/v1/resale"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketResalePage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Transfers(ctx context.Context, direction *string, cursor *string, limit *int) (*model.TicketTransferPage, error) {
	path := "/api/v1/transfers"
	q := url.Values{}
	if direction != nil {
		q.Set("direction", *direction)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketTransferPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Waitlist(ctx context.Context) ([]*model.TicketWaitlist, error) {
	path := "/api/v1/waitlist"
	q := url.Values{}
	h := http.Header{}
	var out []*model.TicketWaitlist
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Wishlist(ctx context.Context, limit *int, offset *int) ([]*model.TicketEvent, error) {
	path := "/api/v1/wishlist"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.TicketEvent
	err := c.rest.Get(ctx, path, q, h, "items", &out)
	return out, err
}

func (c *Client) Notifications(ctx context.Context, unread *bool, cursor *string, limit *int) (*model.TicketNotificationPage, error) {
	path := "/api/v1/notifications"
	q := url.Values{}
	if unread != nil {
		q.Set("unread", strconv.FormatBool(*unread))
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.TicketNotificationPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UnreadNotifications(ctx context.Context) (*model.TicketUnreadCount, error) {
	path := "/api/v1/notifications/unread-count"
	q := url.Values{}
	h := http.Header{}
	var out *model.TicketUnreadCount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
