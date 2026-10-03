package coupon_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "coupon-service"

const DefaultBaseURL = "http://coupon-service-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) Coupons(ctx context.Context) ([]*model.CouponCoupon, error) {
	path := "/api/v1/coupons"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CouponCoupon
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CouponByCode(ctx context.Context, code string) (*model.CouponCoupon, error) {
	path := "/api/v1/coupons/code/" + url.PathEscape(code)
	q := url.Values{}
	h := http.Header{}
	var out *model.CouponCoupon
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) Coupon(ctx context.Context, id int) (*model.CouponCoupon, error) {
	path := "/api/v1/coupons/" + url.PathEscape(strconv.Itoa(id))
	q := url.Values{}
	h := http.Header{}
	var out *model.CouponCoupon
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CouponRestrictions(ctx context.Context, id int) ([]*model.CouponCouponRestriction, error) {
	path := "/api/v1/coupons/" + url.PathEscape(strconv.Itoa(id)) + "/restrictions"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CouponCouponRestriction
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CouponUsages(ctx context.Context, id int) ([]*model.CouponCouponUsage, error) {
	path := "/api/v1/coupons/" + url.PathEscape(strconv.Itoa(id)) + "/usages"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CouponCouponUsage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) CouponUserCoupons(ctx context.Context, id int) ([]*model.CouponUserCoupon, error) {
	path := "/api/v1/coupons/" + url.PathEscape(strconv.Itoa(id)) + "/user-coupons"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CouponUserCoupon
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserCoupons(ctx context.Context, userID int) ([]*model.CouponUserCoupon, error) {
	path := "/api/v1/user-coupons"
	q := url.Values{}
	q.Set("user_id", strconv.Itoa(userID))
	h := http.Header{}
	var out []*model.CouponUserCoupon
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}

func (c *Client) UserCouponUsages(ctx context.Context, userID int) ([]*model.CouponCouponUsage, error) {
	path := "/api/v1/users/" + url.PathEscape(strconv.Itoa(userID)) + "/coupon-usages"
	q := url.Values{}
	h := http.Header{}
	var out []*model.CouponCouponUsage
	err := c.rest.Get(ctx, path, q, h, "", &out)
	return out, err
}
