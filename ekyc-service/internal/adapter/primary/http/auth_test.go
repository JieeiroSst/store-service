package http

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"

	"github.com/JIeeiroSst/ekyc-service/config"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/model"
	"github.com/JIeeiroSst/ekyc-service/internal/domain/port"
)

type statusOnly struct {
	port.EkycUsecase
}

func (statusOnly) GetStatus(_ context.Context, userID string) (*model.EkycStatus, error) {
	return &model.EkycStatus{UserID: userID}, nil
}

type sessions map[string]string

func (s sessions) ValidateSession(_ context.Context, token string) (string, error) {
	if token == "down" {
		return "", errors.New("connection refused")
	}
	id, ok := s[token]
	if !ok {
		return "", port.ErrUnauthenticated
	}
	return id, nil
}

func TestAuthorize(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := &config.Config{Auth: config.AuthConfig{InternalTokens: []string{"customer-info"}}}
	router := NewRouter(NewHandler(cfg, statusOnly{}, sessions{"alice-session": "7"}))

	cases := []struct {
		name, path, header, value string
		want                      int
	}{
		{"no token", "/api/v1/ekyc/7", "", "", http.StatusUnauthorized},
		{"bad session", "/api/v1/ekyc/7", "Authorization", "Bearer nope", http.StatusUnauthorized},
		{"own data", "/api/v1/ekyc/7", "Authorization", "Bearer alice-session", http.StatusOK},
		{"session header", "/api/v1/ekyc/7", "X-Session-Token", "alice-session", http.StatusOK},
		{"someone else", "/api/v1/ekyc/8", "Authorization", "Bearer alice-session", http.StatusForbidden},
		{"service token", "/api/v1/ekyc/8", "Authorization", "Bearer customer-info", http.StatusOK},
		{"user-service down", "/api/v1/ekyc/7", "Authorization", "Bearer down", http.StatusServiceUnavailable},
		{"health is public", "/health", "", "", http.StatusOK},
	}
	for _, tc := range cases {
		req := httptest.NewRequest(http.MethodGet, tc.path, nil)
		if tc.header != "" {
			req.Header.Set(tc.header, tc.value)
		}
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)
		if rec.Code != tc.want {
			t.Errorf("%s: got %d want %d (%s)", tc.name, rec.Code, tc.want, rec.Body.String())
		}
	}
}
