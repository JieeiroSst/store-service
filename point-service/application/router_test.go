package application

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestAllControllersRegisterOnOneEngine(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()

	registerRewardDiscountRoutes(r, &RewardDiscountController{})
	registerConvertedRewardPointRoutes(r, &ConvertedRewardPointContrller{})
	registerRewardPointRoutes(r, &RewardPointController{})

	want := map[string]bool{
		"GET /api/v1/reward-discounts":            true,
		"GET /api/v1/reward-discounts/:id":        true,
		"GET /api/v1/converted-reward-points":     true,
		"GET /api/v1/converted-reward-points/:id": true,
		"GET /api/v1/reward-points":               true,
		"GET /api/v1/reward-points/:id":           true,
	}
	for _, rt := range r.Routes() {
		delete(want, rt.Method+" "+rt.Path)
	}
	if len(want) > 0 {
		t.Fatalf("routes not registered: %v", want)
	}

	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/", nil))
	if rec.Code != http.StatusNotFound {
		t.Fatalf("GET / = %d, want 404 now that nothing is mounted at the root", rec.Code)
	}
}
