package twitter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"

	"github.com/JIeeiroSst/bot-service/config"
)

const (
	apiBase        = "https://api.twitter.com/2"
	mediaUploadURL = "https://upload.twitter.com/1.1/media/upload.json"
)

type Client struct {
	cfg config.TwitterConfig
	hc  *http.Client
}

func NewClient(cfg *config.Config) *Client {
	return &Client{cfg: cfg.Twitter, hc: http.DefaultClient}
}

func (c *Client) ReplyToTweet(ctx context.Context, inReplyToTweetID, text string) error {
	endpoint := apiBase + "/tweets"
	body, err := json.Marshal(map[string]any{
		"text":  text,
		"reply": map[string]string{"in_reply_to_tweet_id": inReplyToTweetID},
	})
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return err
	}
	authHeader, err := signOAuth1(http.MethodPost, endpoint, c.cfg.ConsumerKey, c.cfg.ConsumerSecret, c.cfg.AccessToken, c.cfg.AccessTokenSecret)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("twitter: reply failed with status %d", resp.StatusCode)
	}
	return nil
}

func (c *Client) CreateTweet(ctx context.Context, text, mediaID string) (string, error) {
	payload := map[string]any{"text": text}
	if mediaID != "" {
		payload["media"] = map[string]any{"media_ids": []string{mediaID}}
	}
	body, err := json.Marshal(payload)
	if err != nil {
		return "", err
	}

	endpoint := apiBase + "/tweets"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
	if err != nil {
		return "", err
	}
	authHeader, err := signOAuth1(http.MethodPost, endpoint, c.cfg.ConsumerKey, c.cfg.ConsumerSecret, c.cfg.AccessToken, c.cfg.AccessTokenSecret)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("twitter: create tweet failed with status %d", resp.StatusCode)
	}

	var out struct {
		Data struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.Data.ID, nil
}

func (c *Client) UploadMedia(ctx context.Context, imageURL string) (string, error) {
	getReq, err := http.NewRequestWithContext(ctx, http.MethodGet, imageURL, nil)
	if err != nil {
		return "", err
	}
	imgResp, err := c.hc.Do(getReq)
	if err != nil {
		return "", fmt.Errorf("twitter: download media: %w", err)
	}
	defer imgResp.Body.Close()
	if imgResp.StatusCode >= 300 {
		return "", fmt.Errorf("twitter: download media failed with status %d", imgResp.StatusCode)
	}
	imgBytes, err := io.ReadAll(imgResp.Body)
	if err != nil {
		return "", err
	}

	var buf bytes.Buffer
	w := multipart.NewWriter(&buf)
	part, err := w.CreateFormFile("media", "upload")
	if err != nil {
		return "", err
	}
	if _, err := part.Write(imgBytes); err != nil {
		return "", err
	}
	if err := w.Close(); err != nil {
		return "", err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, mediaUploadURL, &buf)
	if err != nil {
		return "", err
	}

	authHeader, err := signOAuth1(http.MethodPost, mediaUploadURL, c.cfg.ConsumerKey, c.cfg.ConsumerSecret, c.cfg.AccessToken, c.cfg.AccessTokenSecret)
	if err != nil {
		return "", err
	}
	req.Header.Set("Authorization", authHeader)
	req.Header.Set("Content-Type", w.FormDataContentType())

	resp, err := c.hc.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return "", fmt.Errorf("twitter: media upload failed with status %d", resp.StatusCode)
	}

	var out struct {
		MediaIDString string `json:"media_id_string"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return "", err
	}
	return out.MediaIDString, nil
}

type Mention struct {
	TweetID  string
	AuthorID string
	Text     string
}

type mentionsResponse struct {
	Data []struct {
		ID       string `json:"id"`
		Text     string `json:"text"`
		AuthorID string `json:"author_id"`
	} `json:"data"`
	Meta struct {
		NewestID string `json:"newest_id"`
	} `json:"meta"`
}

func (c *Client) FetchMentions(ctx context.Context, userID, sinceID string) ([]Mention, string, error) {
	endpoint := fmt.Sprintf("%s/users/%s/mentions", apiBase, userID)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, sinceID, err
	}
	q := req.URL.Query()
	q.Set("tweet.fields", "author_id")
	if sinceID != "" {
		q.Set("since_id", sinceID)
	}
	req.URL.RawQuery = q.Encode()
	req.Header.Set("Authorization", "Bearer "+c.cfg.BearerToken)

	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, sinceID, err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return nil, sinceID, fmt.Errorf("twitter: fetch mentions failed with status %d", resp.StatusCode)
	}

	var out mentionsResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return nil, sinceID, err
	}

	mentions := make([]Mention, 0, len(out.Data))
	for _, d := range out.Data {
		mentions = append(mentions, Mention{TweetID: d.ID, AuthorID: d.AuthorID, Text: d.Text})
	}
	newSinceID := sinceID
	if out.Meta.NewestID != "" {
		newSinceID = out.Meta.NewestID
	}
	return mentions, newSinceID, nil
}
