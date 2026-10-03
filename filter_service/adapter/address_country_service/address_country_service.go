package address_country_service

import (
	"context"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

var _ = strconv.Itoa

const Service = "address-country-service"

const DefaultBaseURL = "http://address-country-svc"

type Client struct{ rest *rest.Client }

func New(c *rest.Client) *Client { return &Client{rest: c} }

func (c *Client) AddressCountries(ctx context.Context) ([]*model.AddrAddressCountry, error) {
	path := "/address-countries"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AddrAddressCountry
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) AddressCountry(ctx context.Context, id string) (*model.AddrAddressCountry, error) {
	path := "/address-countries/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AddrAddressCountry
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) OrderBillingAddresses(ctx context.Context) ([]*model.AddrOrderBillingaddress, error) {
	path := "/order-billing-addresses"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AddrOrderBillingaddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) OrderBillingAddress(ctx context.Context, id string) (*model.AddrOrderBillingaddress, error) {
	path := "/order-billing-addresses/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AddrOrderBillingaddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) OrderShippingAddresses(ctx context.Context) ([]*model.AddrOrderShippingaddress, error) {
	path := "/order-shipping-addresses"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AddrOrderShippingaddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) OrderShippingAddress(ctx context.Context, id string) (*model.AddrOrderShippingaddress, error) {
	path := "/order-shipping-addresses/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AddrOrderShippingaddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) PartnerAddresses(ctx context.Context) ([]*model.AddrPartneraddress, error) {
	path := "/partner-addresses"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AddrPartneraddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) PartnerAddress(ctx context.Context, id string) (*model.AddrPartneraddress, error) {
	path := "/partner-addresses/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AddrPartneraddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) ShippingOrderAndItemChangesCountries(ctx context.Context) ([]*model.AddrShippingOrderanditemchangesCountry, error) {
	path := "/shipping-orderanditemchanges-countries"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AddrShippingOrderanditemchangesCountry
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) ShippingOrderAndItemChangesCountry(ctx context.Context, id string) (*model.AddrShippingOrderanditemchangesCountry, error) {
	path := "/shipping-orderanditemchanges-countries/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AddrShippingOrderanditemchangesCountry
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) ShippingWeightBasedCountries(ctx context.Context) ([]*model.AddrShippingWeightbasedCountry, error) {
	path := "/shipping-weightbased-countries"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AddrShippingWeightbasedCountry
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) ShippingWeightBasedCountry(ctx context.Context, id string) (*model.AddrShippingWeightbasedCountry, error) {
	path := "/shipping-weightbased-countries/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AddrShippingWeightbasedCountry
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) UserAddresses(ctx context.Context) ([]*model.AddrUserAddress, error) {
	path := "/user-addresses"
	q := url.Values{}
	h := http.Header{}
	var out []*model.AddrUserAddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}

func (c *Client) UserAddress(ctx context.Context, id string) (*model.AddrUserAddress, error) {
	path := "/user-addresses/" + url.PathEscape(id)
	q := url.Values{}
	h := http.Header{}
	var out *model.AddrUserAddress
	err := c.rest.Get(ctx, path, q, h, "data", &out)
	return out, err
}
