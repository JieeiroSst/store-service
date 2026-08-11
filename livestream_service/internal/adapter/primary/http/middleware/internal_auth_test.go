package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestRequireInternalSecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.POST("/internal/action", RequireInternalSecret("shh"), func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	tests := []struct {
		name       string
		header     string
		wantStatus int
	}{
		{"missing header", "", http.StatusForbidden},
		{"wrong secret", "wrong", http.StatusForbidden},
		{"correct secret", "shh", http.StatusNoContent},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, "/internal/action", nil)
			if tt.header != "" {
				req.Header.Set("X-Internal-Token", tt.header)
			}
			w := httptest.NewRecorder()
			r.ServeHTTP(w, req)
			if w.Code != tt.wantStatus {
				t.Errorf("status = %d, want %d", w.Code, tt.wantStatus)
			}
		})
	}
}
