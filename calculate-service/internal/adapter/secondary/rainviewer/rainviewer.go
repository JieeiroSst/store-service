package rainviewer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/calculate-service/internal/domain/model"
	"github.com/JIeeiroSst/calculate-service/internal/domain/port"
)

const tileTemplate = "%s%s/256/{z}/{x}/{y}/2/1_1.png"

type client struct {
	baseURL    string
	httpClient *http.Client
}

// e.g. baseURL="https://api.rainviewer.com/public/weather-maps.json".
func NewClient(baseURL string, httpClient *http.Client) port.RadarProvider {
	return &client{baseURL: baseURL, httpClient: httpClient}
}

type weatherMapsResponse struct {
	Generated int64  `json:"generated"`
	Host      string `json:"host"`
	Radar     struct {
		Past []struct {
			Time int64  `json:"time"`
			Path string `json:"path"`
		} `json:"past"`
		Nowcast []struct {
			Time int64  `json:"time"`
			Path string `json:"path"`
		} `json:"nowcast"`
	} `json:"radar"`
}

func (c *client) GetLatestRadar(ctx context.Context) (*model.RainRadar, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.baseURL, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("rainviewer: unexpected status %d", resp.StatusCode)
	}

	var body weatherMapsResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, err
	}

	toFrames := func(frames []struct {
		Time int64  `json:"time"`
		Path string `json:"path"`
	}) []model.RadarFrame {
		out := make([]model.RadarFrame, 0, len(frames))
		for _, f := range frames {
			out = append(out, model.RadarFrame{
				Time:    time.Unix(f.Time, 0).UTC(),
				TileURL: fmt.Sprintf(tileTemplate, body.Host, f.Path),
			})
		}
		return out
	}

	return &model.RainRadar{
		GeneratedAt: time.Unix(body.Generated, 0).UTC(),
		Host:        body.Host,
		Past:        toFrames(body.Radar.Past),
		Nowcast:     toFrames(body.Radar.Nowcast),
		Source:      "rainviewer",
	}, nil
}
