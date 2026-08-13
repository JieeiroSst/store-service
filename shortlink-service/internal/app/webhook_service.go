package app

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/shortlink-service/internal/adapters/repo"
	"github.com/JIeeiroSst/shortlink-service/internal/domain"
	"github.com/JIeeiroSst/shortlink-service/internal/ports"
)

var ErrWebhookNotFound = errors.New("webhook not found")
var ErrNoFieldsToUpdate = errors.New("no fields to update")

type WebhookService struct {
	webhooks ports.WebhookRepository
	sender   ports.WebhookSender
}

func NewWebhookService(webhooks ports.WebhookRepository, sender ports.WebhookSender) *WebhookService {
	return &WebhookService{webhooks, sender}
}

func (s *WebhookService) List(ctx context.Context, userID *string) ([]*domain.Webhook, error) {
	return s.webhooks.List(ctx, userID)
}

func (s *WebhookService) Get(ctx context.Context, id string, userID *string) (*domain.Webhook, error) {
	wh, err := s.webhooks.GetByID(ctx, id, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return nil, ErrWebhookNotFound
	}
	return wh, err
}

type CreateWebhookInput struct {
	UserID     *string
	Name       string
	URL        string
	Events     []domain.WebhookEvent
	Headers    map[string]string
	RetryCount int
	TimeoutMs  int
}

func (s *WebhookService) Create(ctx context.Context, in CreateWebhookInput) (*domain.Webhook, error) {
	secret, err := domain.GenerateWebhookSecret()
	if err != nil {
		return nil, err
	}
	retryCount := in.RetryCount
	if retryCount == 0 {
		retryCount = 3
	}
	timeoutMs := in.TimeoutMs
	if timeoutMs == 0 {
		timeoutMs = 10000
	}
	headers := in.Headers
	if headers == nil {
		headers = map[string]string{}
	}

	wh := &domain.Webhook{
		UserID: in.UserID, Name: in.Name, URL: in.URL, Secret: secret, Events: in.Events,
		IsActive: true, RetryCount: retryCount, TimeoutMs: timeoutMs, Headers: headers,
	}
	if err := s.webhooks.Create(ctx, wh); err != nil {
		return nil, err
	}
	return wh, nil
}

type UpdateWebhookInput struct {
	Name       *string
	URL        *string
	Events     []domain.WebhookEvent
	IsActive   *bool
	Headers    map[string]string
	RetryCount *int
	TimeoutMs  *int
}

func (s *WebhookService) Update(ctx context.Context, id string, userID *string, in UpdateWebhookInput) (*domain.Webhook, error) {
	patch := map[string]interface{}{}
	if in.Name != nil {
		patch["name"] = *in.Name
	}
	if in.URL != nil {
		patch["url"] = *in.URL
	}
	if in.Events != nil {
		events := make([]string, len(in.Events))
		for i, e := range in.Events {
			events[i] = string(e)
		}
		patch["events"] = events
	}
	if in.IsActive != nil {
		patch["is_active"] = *in.IsActive
	}
	if in.Headers != nil {
		patch["headers"] = mustStringMapJSON(in.Headers)
	}
	if in.RetryCount != nil {
		patch["retry_count"] = *in.RetryCount
	}
	if in.TimeoutMs != nil {
		patch["timeout_ms"] = *in.TimeoutMs
	}
	if len(patch) == 0 {
		return nil, ErrNoFieldsToUpdate
	}

	wh, err := s.webhooks.Update(ctx, id, userID, patch)
	if err != nil {
		if errors.Is(err, repo.ErrNotFound) {
			return nil, ErrWebhookNotFound
		}
		return nil, err
	}
	return wh, nil
}

func (s *WebhookService) Delete(ctx context.Context, id string, userID *string) error {
	err := s.webhooks.Delete(ctx, id, userID)
	if errors.Is(err, repo.ErrNotFound) {
		return ErrWebhookNotFound
	}
	return err
}

type TestWebhookResult struct {
	Success      bool
	StatusCode   int
	ResponseBody string
	Error        string
}

func (s *WebhookService) Test(ctx context.Context, id string, userID *string) (*TestWebhookResult, error) {
	wh, err := s.Get(ctx, id, userID)
	if err != nil {
		return nil, err
	}

	payload := domain.WebhookPayload{
		Event:     domain.WebhookEventClick,
		EventID:   "00000000-0000-0000-0000-000000000000",
		Timestamp: time.Now().UTC().Format(time.RFC3339Nano),
		Data: map[string]interface{}{
			"id":          "00000000-0000-0000-0000-000000000000",
			"linkId":      "00000000-0000-0000-0000-000000000000",
			"clickedAt":   time.Now().UTC().Format(time.RFC3339Nano),
			"ipAddress":   "127.0.0.1",
			"userAgent":   "LinkForty-Test/1.0",
			"deviceType":  "web",
			"platform":    "test",
			"countryCode": "US",
		},
	}

	result := s.sender.Send(ctx, wh, payload)
	return &TestWebhookResult{
		Success: result.Success, StatusCode: result.ResponseStatus,
		ResponseBody: result.ResponseBody, Error: result.ErrorMessage,
	}, nil
}

func mustStringMapJSON(m map[string]string) interface{} {
	generic := make(map[string]interface{}, len(m))
	for k, v := range m {
		generic[k] = v
	}
	return mustMapJSON(generic)
}
