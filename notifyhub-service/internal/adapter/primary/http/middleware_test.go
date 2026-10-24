package http

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

func TestRateLimiterWindowResetsUnderSteadyTraffic(t *testing.T) {
	rl := NewRateLimiter(2)
	defer rl.Close()
	start := time.Unix(0, 0)

	if !rl.Allow("ip", start) || !rl.Allow("ip", start.Add(100*time.Millisecond)) {
		t.Fatal("first two requests must pass")
	}
	if rl.Allow("ip", start.Add(900*time.Millisecond)) {
		t.Fatal("third request in the same window must be limited")
	}
	// A client that keeps sending must be allowed again in the next window.
	if !rl.Allow("ip", start.Add(1100*time.Millisecond)) {
		t.Fatal("request in the next window must pass")
	}
	if !rl.Allow("other", start) {
		t.Fatal("limits are per client")
	}
}

func TestAPIKeyAuthEmptySecret(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for _, tc := range []struct {
		dev  bool
		want int
	}{{dev: false, want: http.StatusServiceUnavailable}, {dev: true, want: http.StatusOK}} {
		r := gin.New()
		r.GET("/x", APIKeyAuth("", tc.dev), func(c *gin.Context) { c.Status(http.StatusOK) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/x", nil))
		if w.Code != tc.want {
			t.Errorf("dev=%v: status = %d, want %d", tc.dev, w.Code, tc.want)
		}
	}
}
