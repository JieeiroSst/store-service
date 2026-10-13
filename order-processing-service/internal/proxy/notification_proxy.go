package proxy

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
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
	userID, err := strconv.ParseUint(req.CustomerID, 10, 64)
	if err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", fmt.Errorf("customer id %q is not a user id: %w", req.CustomerID, err))
	}
	body, err := json.Marshal(map[string]any{
		"user_id": userID,
		"type":    req.Type,
		"title":   "Order " + req.OrderID,
		"message": req.Message,
		"data":    map[string]string{"order_id": req.OrderID},
	})
	if err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, p.baseURL+"/api/v1/notifications", bytes.NewReader(body))
	if err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if p.apiKey != "" {
		httpReq.Header.Set("X-Internal-Token", p.apiKey)
	}

	resp, err := p.httpClient.Do(httpReq)
	if err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 != 2 {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, apperrors.NewProxyError("notification", "Send",
			fmt.Errorf("status %d: %s", resp.StatusCode, string(respBody)))
	}

	var created struct {
		ID     uint   `json:"id"`
		Status string `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&created); err != nil {
		return nil, apperrors.NewProxyError("notification", "Send", err)
	}
	return &entity.NotificationResponse{Sent: true, ID: strconv.FormatUint(uint64(created.ID), 10)}, nil
}
