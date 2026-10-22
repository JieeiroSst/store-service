package userclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"

	"github.com/JIeeiroSst/doordash-service/config"
	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
	"github.com/JIeeiroSst/doordash-service/pkg/authn"
)

type httpUserClient struct {
	client  *http.Client
	baseURL string
}

func NewUserClient(cfg *config.Config) port.UserClient {
	return &httpUserClient{
		client:  &http.Client{Timeout: cfg.UserService.TimeoutDuration()},
		baseURL: cfg.UserService.BaseURL,
	}
}

type profileResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"`
	Name     string `json:"name"`
	Active   bool   `json:"active"`
}

func (c *httpUserClient) GetCustomer(ctx context.Context, customerID string) (*model.Customer, error) {
	var body profileResponse
	if err := c.get(ctx, fmt.Sprintf("%s/api/v1/users/%s/profile", c.baseURL, url.PathEscape(customerID)), &body); err != nil {
		return nil, err
	}
	name := body.Name
	if name == "" {
		name = body.Username
	}
	return &model.Customer{ID: strconv.FormatInt(body.ID, 10), Name: name}, nil
}

type driverResponse struct {
	ID          string `json:"driver_id"`
	FirstName   string `json:"first_name"`
	LastName    string `json:"last_name"`
	PhoneNumber string `json:"phone_number"`
	VehicleType string `json:"vehicle_type"`
	IsActive    bool   `json:"is_active"`
}

func (c *httpUserClient) GetDriver(ctx context.Context, driverID string) (*model.Driver, error) {
	var body driverResponse
	if err := c.get(ctx, fmt.Sprintf("%s/v1/drivers/%s", c.baseURL, driverID), &body); err != nil {
		return nil, err
	}
	return &model.Driver{
		ID: body.ID, Name: fullName(body.FirstName, body.LastName), Phone: body.PhoneNumber,
		VehicleType: body.VehicleType, IsActive: body.IsActive,
	}, nil
}

func (c *httpUserClient) get(ctx context.Context, url string, out any) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	if p, ok := authn.FromContext(ctx); ok && p.Token != "" {
		req.Header.Set("Authorization", "Bearer "+p.Token)
	} else {
		authn.SetInternalToken(req.Header)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("call user_service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return port.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("user_service returned status %d for %s", resp.StatusCode, url)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("decode user_service response: %w", err)
	}
	return nil
}

func fullName(first, last string) string {
	if first == "" {
		return last
	}
	if last == "" {
		return first
	}
	return first + " " + last
}
