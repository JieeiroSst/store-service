package noaa

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
)

const timeLayout = "2006-01-02 15:04"

type client struct {
	baseURL    string
	httpClient *http.Client
}

// baseURL="https://api.tidesandcurrents.noaa.gov/api/prod/datagetter".
func NewClient(baseURL string, httpClient *http.Client) port.TideProvider {
	return &client{baseURL: baseURL, httpClient: httpClient}
}

type predictionsResponse struct {
	Predictions []struct {
		Time  string `json:"t"`
		Value string `json:"v"`
		Type  string `json:"type"`
	} `json:"predictions"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (c *client) GetTidePredictions(ctx context.Context, stationID string) (*model.TidePrediction, error) {
	q := url.Values{}
	q.Set("station", stationID)
	q.Set("product", "predictions")
	q.Set("datum", "MLLW")
	q.Set("time_zone", "gmt")
	q.Set("units", "metric")
	q.Set("format", "json")
	q.Set("date", "today")
	q.Set("interval", "hilo")

	reqURL := c.baseURL + "?" + q.Encode()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("noaa: unexpected status %d", resp.StatusCode)
	}

	var body predictionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}
	if body.Error != nil {
		return nil, fmt.Errorf("noaa: %s", body.Error.Message)
	}

	events := make([]model.TideEvent, 0, len(body.Predictions))
	for _, p := range body.Predictions {
		t, err := time.Parse(timeLayout, p.Time)
		if err != nil {
			continue
		}
		height, err := strconv.ParseFloat(p.Value, 64)
		if err != nil {
			continue
		}
		eventType := model.TideLow
		if p.Type == "H" {
			eventType = model.TideHigh
		}
		events = append(events, model.TideEvent{
			Time:    t,
			Type:    eventType,
			HeightM: height,
		})
	}

	return &model.TidePrediction{
		StationID: stationID,
		Source:    "noaa-tides-and-currents",
		Events:    events,
	}, nil
}
