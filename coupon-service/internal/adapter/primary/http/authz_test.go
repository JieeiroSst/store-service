package http

import (
	"context"
	"testing"

	"github.com/JIeeiroSst/coupon-service/internal/domain/model"
	"github.com/JIeeiroSst/coupon-service/internal/domain/port"
)

type fakeCoupons struct {
	port.CouponUsecase
	validatedFor int64
}

func (f *fakeCoupons) ValidateCoupon(_ context.Context, in port.ValidateCouponInput) (*model.Coupon, float64, error) {
	f.validatedFor = in.UserID
	return &model.Coupon{}, 1, nil
}
func (f *fakeCoupons) ListCoupons(context.Context) ([]model.Coupon, error) { return nil, nil }

type fakeUsages struct{ port.CouponUsageUsecase }

func (fakeUsages) ListUsagesByUser(context.Context, int64) ([]model.CouponUsage, error) {
	return nil, nil
}

func TestCouponAuthorization(t *testing.T) {
	fakeUserService(t)
	coupons := &fakeCoupons{}
	router := NewRouter(NewHandler(coupons, nil, fakeUsages{}, nil))
	user := testToken("7", "user")
	admin := testToken("1", "admin")
	cases := []struct {
		name string
		call authCall
		want int
	}{
		{"anonymous", authCall{method: "POST", path: "/api/v1/coupons/validate", body: `{}`}, 401},
		{"user creates coupon", authCall{method: "POST", path: "/api/v1/coupons", body: `{}`, bearer: user}, 403},
		{"user lists coupons", authCall{method: "GET", path: "/api/v1/coupons", bearer: user}, 403},
		{"admin lists coupons", authCall{method: "GET", path: "/api/v1/coupons", bearer: admin}, 200},
		{"user uses coupon", authCall{method: "POST", path: "/api/v1/user-coupons/1/use", bearer: user}, 403},
		{"foreign usages", authCall{method: "GET", path: "/api/v1/users/8/coupon-usages", bearer: user}, 403},
		{"own usages", authCall{method: "GET", path: "/api/v1/users/7/coupon-usages", bearer: user}, 200},
		{"foreign user coupons", authCall{method: "GET", path: "/api/v1/user-coupons?user_id=8", bearer: user}, 403},
	}
	for _, c := range cases {
		if got, body := doAuth(router, c.call); got != c.want {
			t.Errorf("%s: got %d want %d (%s)", c.name, got, c.want, body)
		}
	}
	if code, body := doAuth(router, authCall{method: "POST", path: "/api/v1/coupons/validate", body: `{"code":"X","user_id":8,"purchase_amount":10}`, bearer: user}); code != 200 || coupons.validatedFor != 7 {
		t.Fatalf("validate forced owner: %d %s user=%d", code, body, coupons.validatedFor)
	}
	if code, _ := doAuth(router, authCall{method: "POST", path: "/api/v1/coupons/validate", body: `{"code":"X","user_id":8,"purchase_amount":10}`, internal: testInternalToken}); code != 200 || coupons.validatedFor != 8 {
		t.Fatalf("internal validate kept user: %d user=%d", code, coupons.validatedFor)
	}
}
