package facebook

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/JIeeiroSst/bot-service/config"
)

const graphAPIBase = "https://graph.facebook.com/v19.0"

type Client struct {
	cfg config.FacebookConfig
	hc  *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{cfg: cfg.Facebook, hc: http.DefaultClient}
}

func (c *Client) SendText(ctx context.Context, recipientPSID, text string) error {
	endpoint := fmt.Sprintf("%s/me/messages?access_token=%s", graphAPIBase, c.cfg.PageAccessToken)

	body, err := json.Marshal(map[string]any{
		"recipient": map[string]string{"id": recipientPSID},
		"message":   map[string]string{"text": text},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("facebook: send message failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) PostToFeed(ctx context.Context, message, imageURL string) (string, error) {
	if c.cfg.PageID == "" {
		return "", fmt.Errorf("facebook: PageID not configured")
	}

	var endpoint string
	var body []byte
	var err error
	if imageURL != "" {
		endpoint = fmt.Sprintf("%s/%s/photos?access_token=%s", graphAPIBase, c.cfg.PageID, c.cfg.PageAccessToken)
		body, err = json.Marshal(map[string]string{"url": imageURL, "caption": message})
	} else {
		endpoint = fmt.Sprintf("%s/%s/feed?access_token=%s", graphAPIBase, c.cfg.PageID, c.cfg.PageAccessToken)
		body, err = json.Marshal(map[string]string{"message": message})
	}
	if err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("facebook: post to feed failed with status %d", resp.StatusCode)
	}

	var out struct {
		ID     string `json:"id"`
		PostID string `json:"post_id"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	if out.PostID != "" {
		return out.PostID, nil
	}
	return out.ID, nil
}
