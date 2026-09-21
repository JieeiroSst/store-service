package restaurantclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/JIeeiroSst/doordash-service/config"
	"github.com/JIeeiroSst/doordash-service/internal/domain/model"
	"github.com/JIeeiroSst/doordash-service/internal/domain/port"
)

type httpRestaurantClient struct {
	client  *http.Client
	baseURL string
}

func NewRestaurantClient(cfg *config.Config) port.RestaurantClient {
	return &httpRestaurantClient{
		client:  &http.Client{Timeout: cfg.RestaurantService.TimeoutDuration()},
		baseURL: cfg.RestaurantService.BaseURL,
	}
}

type restaurantResponse struct {
	ID             string  `json:"restaurant_id"`
	Name           string  `json:"name"`
	IsActive       bool    `json:"is_active"`
	CommissionRate float64 `json:"commission_rate"`
}

func (c *httpRestaurantClient) GetRestaurant(ctx context.Context, restaurantID string) (*model.Restaurant, error) {
	url := fmt.Sprintf("%s/v1/restaurants/%s", c.baseURL, restaurantID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call restaurant service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, port.ErrNotFound
	}
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("restaurant service returned status %d for restaurant %s", resp.StatusCode, restaurantID)
	}

	var body restaurantResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode restaurant service response: %w", err)
	}
	return &model.Restaurant{ID: body.ID, Name: body.Name, IsActive: body.IsActive, CommissionRate: body.CommissionRate}, nil
}

type menuItemResponse struct {
	ID           string  `json:"item_id"`
	RestaurantID string  `json:"restaurant_id"`
	Name         string  `json:"name"`
	Price        float64 `json:"price"`
	IsActive     bool    `json:"is_active"`
}

// GetMenuItems fetches each ID concurrently and silently omits any that
// fail or don't exist - the same "missing means unresolved" contract as
// port.RestaurantClient documents, mirroring userclient.GetUsers in
// threads-service.
func (c *httpRestaurantClient) GetMenuItems(ctx context.Context, menuItemIDs []string) (map[string]*model.MenuItem, error) {
	result := make(map[string]*model.MenuItem, len(menuItemIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range menuItemIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			item, err := c.getMenuItem(ctx, id)
			if err != nil {
				return
			}
			mu.Lock()
			result[id] = item
			mu.Unlock()
		}(id)
	}
	wg.Wait()

	return result, nil
}

func (c *httpRestaurantClient) getMenuItem(ctx context.Context, id string) (*model.MenuItem, error) {
	url := fmt.Sprintf("%s/v1/menu-items/%s", c.baseURL, id)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("restaurant service returned status %d for menu item %s", resp.StatusCode, id)
	}

	var body menuItemResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	return &model.MenuItem{ID: body.ID, RestaurantID: body.RestaurantID, Name: body.Name, Price: body.Price, IsActive: body.IsActive}, nil
}
