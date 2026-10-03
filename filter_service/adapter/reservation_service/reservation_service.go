package reservation_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "reservation-service"

const DefaultBaseURL = "http://reservation-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Hotels(ctx context.Context, city *string, qArg *string, rooms *int, guests *int, sort *string, limit *int, minRate *string, maxRate *string, amenities *string, cursor *string, checkIn *string, checkOut *string) (*model.ResvHotelPage, error) {
	path := "/api/v1/hotels"
	q := url.Values{}
	if city != nil {
		q.Set("city", *city)
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if rooms != nil {
		q.Set("rooms", strconv.Itoa(*rooms))
	}
	if guests != nil {
		q.Set("guests", strconv.Itoa(*guests))
	}
	if sort != nil {
		q.Set("sort", *sort)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if minRate != nil {
		q.Set("min_rate", *minRate)
	}
	if maxRate != nil {
		q.Set("max_rate", *maxRate)
	}
	if amenities != nil {
		q.Set("amenities", *amenities)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if checkIn != nil {
		q.Set("check_in", *checkIn)
	}
	if checkOut != nil {
		q.Set("check_out", *checkOut)
	}
	h := http.Header{}
	var out *model.ResvHotelPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Hotel(ctx context.Context, id int) (*model.ResvHotel, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.ResvHotel
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) RoomTypes(ctx context.Context, id int) ([]*model.ResvRoomType, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id)) + "/room-types"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ResvRoomType
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Inventory(ctx context.Context, id int, typeArg int, from *string, to *string) ([]*model.ResvInventory, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id)) + "/room-types/" + url.PathEscape(strconv.Itoa(typeArg)) + "/inventory"
	q := url.Values{}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	h := http.Header{}
	var out []*model.ResvInventory
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Quotes(ctx context.Context, id int, checkIn string, checkOut string, rooms *int) ([]*model.ResvQuote, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id)) + "/quotes"
	q := url.Values{}
	q.Set("check_in", checkIn)
	q.Set("check_out", checkOut)
	if rooms != nil {
		q.Set("rooms", strconv.Itoa(*rooms))
	}
	h := http.Header{}
	var out []*model.ResvQuote
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Services(ctx context.Context, id int) ([]*model.ResvService, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id)) + "/services"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ResvService
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Reviews(ctx context.Context, id int, cursor *string, limit *int) (*model.ResvReviewPage, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id)) + "/reviews"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ResvReviewPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MyHotels(ctx context.Context, limit *int) (*model.ResvHotelPage, error) {
	path := "/api/v1/my/hotels"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ResvHotelPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) HotelReviewQueue(ctx context.Context, status *string, limit *int) (*model.ResvHotelPage, error) {
	path := "/api/v1/admin/hotels"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ResvHotelPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Rooms(ctx context.Context, id int, roomTypeID *int) ([]*model.ResvRoom, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id)) + "/rooms"
	q := url.Values{}
	if roomTypeID != nil {
		q.Set("room_type_id", strconv.Itoa(*roomTypeID))
	}
	h := http.Header{}
	var out []*model.ResvRoom
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Promotions(ctx context.Context, id int) ([]*model.ResvPromotion, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id)) + "/promotions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ResvPromotion
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) HotelReport(ctx context.Context, id int, from *string, to *string) (*model.ResvReport, error) {
	path := "/api/v1/hotels/" + url.PathEscape(strconv.Itoa(id)) + "/report"
	q := url.Values{}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	h := http.Header{}
	var out *model.ResvReport
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Reservations(ctx context.Context, status *string, cursor *string, as *string, hotelID *int, limit *int) (*model.ResvReservationPage, error) {
	path := "/api/v1/reservations"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if as != nil {
		q.Set("as", *as)
	}
	if hotelID != nil {
		q.Set("hotel_id", strconv.Itoa(*hotelID))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ResvReservationPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Reservation(ctx context.Context, id int) (*model.ResvReservation, error) {
	path := "/api/v1/reservations/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.ResvReservation
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CancellationQuote(ctx context.Context, id int) (*model.ResvCancellationQuote, error) {
	path := "/api/v1/reservations/" + url.PathEscape(strconv.Itoa(id)) + "/cancellation-quote"
	q := url.Values{}
	h := http.Header{}
	var out *model.ResvCancellationQuote
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) ReservationHistory(ctx context.Context, id int) ([]*model.ResvHistoryEntry, error) {
	path := "/api/v1/reservations/" + url.PathEscape(strconv.Itoa(id)) + "/history"
	q := url.Values{}
	h := http.Header{}
	var out []*model.ResvHistoryEntry
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Waitlist(ctx context.Context, cursor *string, limit *int) (*model.ResvWaitPage, error) {
	path := "/api/v1/waitlist"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ResvWaitPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Wishlist(ctx context.Context, cursor *string, limit *int) (*model.ResvHotelPage, error) {
	path := "/api/v1/wishlist"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ResvHotelPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Notifications(ctx context.Context, cursor *string, unread *bool, limit *int) (*model.ResvNotificationPage, error) {
	path := "/api/v1/notifications"
	q := url.Values{}
	if cursor != nil {
		q.Set("cursor", *cursor)
	}
	if unread != nil {
		q.Set("unread", strconv.FormatBool(*unread))
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	h := http.Header{}
	var out *model.ResvNotificationPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UnreadNotifications(ctx context.Context) (*model.ResvUnreadCount, error) {
	path := "/api/v1/notifications/unread-count"
	q := url.Values{}
	h := http.Header{}
	var out *model.ResvUnreadCount
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Loyalty(ctx context.Context) (*model.ResvLoyalty, error) {
	path := "/api/v1/loyalty"
	q := url.Values{}
	h := http.Header{}
	var out *model.ResvLoyalty
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Wallet(ctx context.Context) (*model.ResvWallet, error) {
	path := "/api/v1/wallet"
	q := url.Values{}
	h := http.Header{}
	var out *model.ResvWallet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) WalletTransactions(ctx context.Context, limit *int, offset *int) ([]*model.ResvWalletTxn, error) {
	path := "/api/v1/wallet/transactions"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.ResvWalletTxn
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
