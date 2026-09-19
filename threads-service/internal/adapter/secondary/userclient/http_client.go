package userclient

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	"github.com/JIeeiroSst/threads-service/config"
	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"github.com/JIeeiroSst/threads-service/internal/domain/port"
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

type userResponse struct {
	ID              string `json:"id"`
	Username        string `json:"username"`
	ProfileImageURL string `json:"profile_image_url"`
}

func (c *httpUserClient) GetUser(ctx context.Context, userID string) (*model.Author, error) {
	url := fmt.Sprintf("%s/v1/users/%s", c.baseURL, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("build request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("call user_service: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("user_service returned status %d for user %s", resp.StatusCode, userID)
	}

	var body userResponse
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		return nil, fmt.Errorf("decode user_service response: %w", err)
	}
	return &model.Author{ID: body.ID, Username: body.Username, ProfileImageURL: body.ProfileImageURL}, nil
}

func (c *httpUserClient) GetUsers(ctx context.Context, userIDs []string) (map[string]*model.Author, error) {
	result := make(map[string]*model.Author, len(userIDs))
	var mu sync.Mutex
	var wg sync.WaitGroup

	for _, id := range userIDs {
		wg.Add(1)
		go func(id string) {
			defer wg.Done()
			author, err := c.GetUser(ctx, id)
			if err != nil {
				return
			}
			mu.Lock()
			result[id] = author
			mu.Unlock()
		}(id)
	}
	wg.Wait()

	return result, nil
}
