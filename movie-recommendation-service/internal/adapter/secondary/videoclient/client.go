package videoclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
	"github.com/JIeeiroSst/movie-recommendation-service/internal/domain/model"
)

const pageSize = 100

type Client struct {
	baseURL string
	http    *http.Client
}

func New(cfg *config.Config) *Client {
	return &Client{baseURL: cfg.VideoService.BaseURL, http: &http.Client{Timeout: cfg.VideoService.Timeout}}
}

type wireVideo struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Status      string    `json:"status"`
	Duration    float64   `json:"duration"`
	Views       int64     `json:"views"`
	CreatedAt   time.Time `json:"created_at"`
}

type wirePage struct {
	Items []wireVideo `json:"items"`
	Total int         `json:"total"`
}

func (c *Client) Fetch(ctx context.Context) ([]model.Video, error) {
	if c.baseURL == "" {
		return nil, fmt.Errorf("VIDEO_SERVICE_BASE_URL is not set")
	}
	var out []model.Video
	for page := 1; ; page++ {
		p, err := c.page(ctx, page)
		if err != nil {
			return nil, err
		}
		for _, v := range p.Items {
			out = append(out, model.Video{
				ID: v.ID, Title: v.Title, Description: v.Description, Status: v.Status,
				Duration: v.Duration, Views: v.Views, CreatedAt: v.CreatedAt,
			})
		}
		if len(p.Items) == 0 || len(out) >= p.Total {
			return out, nil
		}
	}
}

func (c *Client) page(ctx context.Context, page int) (*wirePage, error) {
	url := fmt.Sprintf("%s/api/videos?page=%d&page_size=%d", c.baseURL, page, pageSize)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.http.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("video-service returned %s", resp.Status)
	}
	var p wirePage
	if err := json.NewDecoder(resp.Body).Decode(&p); err != nil {
		return nil, fmt.Errorf("decode video-service response: %w", err)
	}
	return &p, nil
}
