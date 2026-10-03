package rent_house_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "rent_house_service"

const DefaultBaseURL = "http://rent-house-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Homestay(ctx context.Context, id int) (*model.RentHomestay, error) {
	path := "/api/v1/homestays/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.RentHomestay
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Homestays(ctx context.Context, provinceID *int, districtID *int, wardID *int, typeArg *int, guests *int, qArg *string, modelArg *string, minPrice *string, maxPrice *string, sort *string, limit *int, cursor *string, amenityIds *string, checkin *string, checkout *string) (*model.RentHomestayPage, error) {
	path := "/api/v1/homestays"
	q := url.Values{}
	if provinceID != nil {
		q.Set("province_id", strconv.Itoa(*provinceID))
	}
	if districtID != nil {
		q.Set("district_id", strconv.Itoa(*districtID))
	}
	if wardID != nil {
		q.Set("ward_id", strconv.Itoa(*wardID))
	}
	if typeArg != nil {
		q.Set("type", strconv.Itoa(*typeArg))
	}
	if guests != nil {
		q.Set("guests", strconv.Itoa(*guests))
	}
	if qArg != nil {
		q.Set("q", *qArg)
	}
	if modelArg != nil {
		q.Set("model", *modelArg)
	}
	if minPrice != nil {
		q.Set("min_price", *minPrice)
	}
	if maxPrice != nil {
		q.Set("max_price", *maxPrice)
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
	if amenityIds != nil {
		q.Set("amenity_ids", *amenityIds)
	}
	if checkin != nil {
		q.Set("checkin", *checkin)
	}
	if checkout != nil {
		q.Set("checkout", *checkout)
	}
	h := http.Header{}
	var out *model.RentHomestayPage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Availability(ctx context.Context, id int, from *string, to *string) ([]*model.RentSlot, error) {
	path := "/api/v1/homestays/" + url.PathEscape(strconv.Itoa(id)) + "/availability"
	q := url.Values{}
	if from != nil {
		q.Set("from", *from)
	}
	if to != nil {
		q.Set("to", *to)
	}
	h := http.Header{}
	var out []*model.RentSlot
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Amenities(ctx context.Context) ([]*model.RentAmenity, error) {
	path := "/api/v1/amenities"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RentAmenity
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) MyHomestays(ctx context.Context, limit *int, offset *int) ([]*model.RentHomestay, error) {
	path := "/api/v1/my/homestays"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.RentHomestay
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) HomestayReviewQueue(ctx context.Context, status *string, limit *int, offset *int) ([]*model.RentHomestay, error) {
	path := "/api/v1/admin/homestays"
	q := url.Values{}
	if status != nil {
		q.Set("status", *status)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.RentHomestay
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Bookings(ctx context.Context, userID *int, as *string, limit *int, offset *int) ([]*model.RentBooking, error) {
	path := "/api/v1/bookings"
	q := url.Values{}
	if userID != nil {
		q.Set("user_id", strconv.Itoa(*userID))
	}
	if as != nil {
		q.Set("as", *as)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.RentBooking
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Booking(ctx context.Context, id int) (*model.RentBooking, error) {
	path := "/api/v1/bookings/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.RentBooking
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Rates(ctx context.Context, id int) ([]*model.RentRate, error) {
	path := "/api/v1/homestays/" + url.PathEscape(strconv.Itoa(id)) + "/rates"
	q := url.Values{}
	h := http.Header{}
	var out []*model.RentRate
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Leases(ctx context.Context, userID *int, as *string, limit *int, offset *int) ([]*model.RentLease, error) {
	path := "/api/v1/leases"
	q := url.Values{}
	if userID != nil {
		q.Set("user_id", strconv.Itoa(*userID))
	}
	if as != nil {
		q.Set("as", *as)
	}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.RentLease
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Lease(ctx context.Context, id int) (*model.RentLease, error) {
	path := "/api/v1/leases/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.RentLease
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Invoices(ctx context.Context, limit *int, offset *int, status *string, dueBefore *string) ([]*model.RentInvoice, error) {
	path := "/api/v1/invoices"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	if status != nil {
		q.Set("status", *status)
	}
	if dueBefore != nil {
		q.Set("due_before", *dueBefore)
	}
	h := http.Header{}
	var out []*model.RentInvoice
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Reviews(ctx context.Context, id int, limit *int, offset *int) ([]*model.RentReview, error) {
	path := "/api/v1/homestays/" + url.PathEscape(strconv.Itoa(id)) + "/reviews"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.RentReview
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Wishlist(ctx context.Context, limit *int, offset *int) ([]*model.RentHomestay, error) {
	path := "/api/v1/wishlist"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.RentHomestay
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Loyalty(ctx context.Context) (*model.RentLoyalty, error) {
	path := "/api/v1/loyalty"
	q := url.Values{}
	h := http.Header{}
	var out *model.RentLoyalty
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Wallet(ctx context.Context) (*model.RentWallet, error) {
	path := "/api/v1/wallet"
	q := url.Values{}
	h := http.Header{}
	var out *model.RentWallet
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) WalletTransactions(ctx context.Context, limit *int, offset *int) ([]*model.RentWalletTxn, error) {
	path := "/api/v1/wallet/transactions"
	q := url.Values{}
	if limit != nil {
		q.Set("limit", strconv.Itoa(*limit))
	}
	if offset != nil {
		q.Set("offset", strconv.Itoa(*offset))
	}
	h := http.Header{}
	var out []*model.RentWalletTxn
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
