package proxy

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/order-service/config"
	"github.com/JIeeiroSst/order-service/internal/domain/port"
)

type kitchenFoodPricer struct {
	baseURL string
	client  *http.Client
}

func NewKitchenFoodPricer(cfg *config.Config) port.FoodPricer {
	return &kitchenFoodPricer{
		baseURL: cfg.Kitchen.URL,
		client:  &http.Client{Timeout: 5 * time.Second},
	}
}

type foodResponse struct {
	Data []struct {
		ID    int     `json:"id"`
		Name  string  `json:"name"`
		Price float64 `json:"price"`
	} `json:"data"`
}

func (p *kitchenFoodPricer) GetPrices(ctx context.Context, foodIDs []int) (map[int]port.FoodPrice, error) {
	prices := make(map[int]port.FoodPrice, len(foodIDs))
	if len(foodIDs) == 0 {
		return prices, nil
	}

	ids := make([]string, len(foodIDs))
	for i, id := range foodIDs {
		ids[i] = strconv.Itoa(id)
	}

	url := fmt.Sprintf("%s/api/v1/food/batch?ids=%s", strings.TrimRight(p.baseURL, "/"), strings.Join(ids, ","))

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	resp, err := p.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("kitchen-service food lookup failed: status %d", resp.StatusCode)
	}

	var parsed foodResponse
	if err := json.NewDecoder(resp.Body).Decode(&parsed); err != nil {
		return nil, err
	}

	for _, f := range parsed.Data {
		prices[f.ID] = port.FoodPrice{ID: f.ID, Name: f.Name, Price: f.Price}
	}

	return prices, nil
}
