package http

import (
	stdhttp "net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/voucher-service/internal/adapters/inbound/http/middleware"
	authapp "github.com/JIeeiroSst/voucher-service/internal/application/auth"
	"github.com/JIeeiroSst/voucher-service/internal/domain/order"
	"github.com/JIeeiroSst/voucher-service/internal/domain/voucher"
	"github.com/gin-gonic/gin"
)

func ctxAs(userID, role string) *gin.Context {
	gin.SetMode(gin.TestMode)
	c, _ := gin.CreateTestContext(httptest.NewRecorder())
	c.Set("auth_claims", &authapp.Claims{UserID: userID, Role: role})
	return c
}

func TestCanAccessOrder(t *testing.T) {
	mine := &order.Order{BuyerType: order.BuyerTypeRetail, BuyerID: "u1"}
	corp := &order.Order{BuyerType: order.BuyerTypeCorporate, BuyerID: "u1"}
	if !canAccessOrder(ctxAs("u1", "user"), mine) {
		t.Error("buyer denied own order")
	}
	if canAccessOrder(ctxAs("u2", "user"), mine) {
		t.Error("other user allowed")
	}
	if canAccessOrder(ctxAs("u1", "user"), corp) {
		t.Error("user allowed on corporate order")
	}
	if !canAccessOrder(ctxAs("x", "admin"), mine) {
		t.Error("admin denied")
	}
}

func TestCanAccessVoucher(t *testing.T) {
	owner := "u1"
	v := &voucher.Voucher{OwnerType: voucher.OwnerTypeUser, OwnerID: &owner}
	if !canAccessVoucher(ctxAs("u1", "user"), v) || canAccessVoucher(ctxAs("u2", "user"), v) {
		t.Error("voucher ownership check wrong")
	}
	if canAccessVoucher(ctxAs("u1", "user"), &voucher.Voucher{OwnerType: voucher.OwnerTypeUser}) {
		t.Error("unowned voucher readable by a user")
	}
	if canAccessVoucher(ctxAs("", "user"), &voucher.Voucher{OwnerType: voucher.OwnerTypeUser, OwnerID: new(string)}) {
		t.Error("empty caller matched empty owner")
	}
}

func TestRequireAdmin(t *testing.T) {
	gin.SetMode(gin.TestMode)
	for role, want := range map[string]int{"user": stdhttp.StatusForbidden, "admin": stdhttp.StatusOK, "super_admin": stdhttp.StatusOK} {
		r := gin.New()
		r.POST("/x", func(c *gin.Context) { c.Set("auth_claims", &authapp.Claims{UserID: "u", Role: role}) }, middleware.RequireAdmin(), func(c *gin.Context) { c.Status(stdhttp.StatusOK) })
		w := httptest.NewRecorder()
		r.ServeHTTP(w, httptest.NewRequest(stdhttp.MethodPost, "/x", nil))
		if w.Code != want {
			t.Errorf("role %s: status %d, want %d", role, w.Code, want)
		}
	}
}
