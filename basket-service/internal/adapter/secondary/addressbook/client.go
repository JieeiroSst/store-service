package addressbook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/basket-service/common"
	"github.com/JIeeiroSst/basket-service/config"
	"github.com/JIeeiroSst/basket-service/internal/domain/port"
	"go.uber.org/fx"
)

type Client struct {
	baseURL string
	http    *http.Client
}

func New(baseURL string, timeout time.Duration) *Client {
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	return &Client{baseURL: strings.TrimRight(baseURL, "/"), http: &http.Client{Timeout: timeout}}
}

var Module = fx.Options(fx.Provide(func(cfg *config.Config) port.AddressBook {
	if cfg.AddressService.BaseURL == "" {
		return nil
	}
	return New(cfg.AddressService.BaseURL, time.Duration(cfg.AddressService.TimeoutSeconds)*time.Second)
}))

type userAddress struct {
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	Line1       string `json:"line1"`
	Line2       string `json:"line2"`
	State       string `json:"state"`
	Postcode    string `json:"postcode"`
	PhoneNumber string `json:"phone_number"`
}

func (c *Client) ShippingAddress(ctx context.Context, accessToken string, id int) (*port.ShippingAddress, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, fmt.Sprintf("%s/user-addresses/%d", c.baseURL, id), nil)
	if err != nil {
		return nil, err
	}
	token := strings.TrimSpace(accessToken)
	if token != "" && !strings.HasPrefix(token, "Bearer ") {
		token = "Bearer " + token
	}
	req.Header.Set("Authorization", token)
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%w: address service: %v", common.ErrUpstream, err)
	}
	defer resp.Body.Close()
	switch {
	case resp.StatusCode == http.StatusNotFound || resp.StatusCode == http.StatusForbidden:
		return nil, common.ErrNotFound
	case resp.StatusCode != http.StatusOK:
		return nil, fmt.Errorf("%w: address service returned %d", common.ErrUpstream, resp.StatusCode)
	}
	var body struct {
		Data userAddress `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("%w: address service: %v", common.ErrUpstream, err)
	}
	return toShipping(body.Data), nil
}

func toShipping(a userAddress) *port.ShippingAddress {
	out := &port.ShippingAddress{
		Name:     strings.TrimSpace(strings.TrimSpace(a.FirstName) + " " + strings.TrimSpace(a.LastName)),
		Phone:    strings.TrimSpace(a.PhoneNumber),
		Street:   strings.TrimSpace(a.Line1),
		Province: strings.TrimSpace(a.State),
	}
	if parts := strings.SplitN(strings.TrimSpace(a.Line2), ",", 2); len(parts) == 2 {
		out.Ward, out.District = strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	if p := strings.Split(strings.TrimSpace(a.Postcode), ":"); len(p) == 4 && p[0] == "ghn" {
		province, perr := strconv.Atoi(p[1])
		district, derr := strconv.Atoi(p[2])
		if perr == nil && derr == nil && p[3] != "" {
			out.ProvinceID, out.DistrictID, out.WardCode = province, district, p[3]
		}
	}
	return out
}
