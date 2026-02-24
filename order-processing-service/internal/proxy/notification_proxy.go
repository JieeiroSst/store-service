package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
	apperrors "github.com/JIeeiroSst/order-processing-service/pkg/errors"
)

type notificationProxy struct {
	baseURL    string
	httpClient *http.Client
	apiKey     string
}

func NewNotificationProxy(baseURL, apiKey string) NotificationProxy {
	return &notificationProxy{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (p *notificationProxy) Send(ctx context.Context, req entity.NotificationRequest) (*entity.NotificationResponse, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/notifications/send", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", "Bearer "+p.apiKey)

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, apperrors.NewProxyError("notification", "Send",
			fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)))
	}

	var result entity.NotificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", err)
	}
	return &result, nil
}
