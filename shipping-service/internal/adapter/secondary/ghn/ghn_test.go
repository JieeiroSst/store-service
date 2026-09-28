package ghn

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
	"github.com/JIeeiroSst/shipping-service/internal/domain/port"
)

func newTestClient(t *testing.T, h http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	cfg := config.FromEnv()
	cfg.GHN.BaseURL = srv.URL + "/shiip/public-api"
	cfg.GHN.Token = "tok"
	cfg.GHN.WebhookToken = "hook"
	return NewClient(cfg)
}

func body(t *testing.T, r *http.Request) map[string]any {
	t.Helper()
	raw, _ := io.ReadAll(r.Body)
	var m map[string]any
	_ = json.Unmarshal(raw, &m)
	return m
}

func TestQuoteCallsFeeAndLeadtime(t *testing.T) {
	var paths []string
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		paths = append(paths, r.URL.Path)
		if r.Header.Get("Token") != "tok" || r.Header.Get("ShopId") != "885" {
			t.Errorf("headers: token=%q shop=%q", r.Header.Get("Token"), r.Header.Get("ShopId"))
		}
		b := body(t, r)
		switch r.URL.Path {
		case "/shiip/public-api/v2/shipping-order/fee":
			if b["service_id"] != float64(53320) || b["to_ward_code"] != "20109" || b["cod_value"] != float64(300000) {
				t.Errorf("fee body %v", b)
			}
			io.WriteString(w, `{"code":200,"message":"Success","data":{"total":36300,"service_fee":36300}}`)
		case "/shiip/public-api/v2/shipping-order/leadtime":
			io.WriteString(w, `{"code":200,"message":"Success","data":{"leadtime":1790035200,"order_date":1789862400}}`)
		}
	})
	q, err := c.Quote(context.Background(), port.CarrierQuoteRequest{
		ShopID: 885, FromDistrictID: 1454, FromWardCode: "21211", ToDistrictID: 1442, ToWardCode: "20109",
		ServiceID: 53320, CODAmount: 300000,
		Parcel: model.Parcel{WeightGram: 500, LengthCm: 10, WidthCm: 10, HeightCm: 10, Items: []model.Item{{Name: "a", Quantity: 1}}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if q.Fee != 36300 || !q.ExpectedDeliveryAt.Equal(time.Unix(1790035200, 0)) || len(paths) != 2 {
		t.Errorf("quote = %+v paths=%v", q, paths)
	}
}

func TestErrorMapping(t *testing.T) {
	status, payload := 0, ""
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
		io.WriteString(w, payload)
	})
	cases := []struct {
		status  int
		payload string
		want    error
	}{
		{400, `{"code":400,"message":"Số điện thoại không hợp lệ"}`, port.ErrCarrierRejected},
		{401, `{"code":401,"message":"Token is invalid"}`, port.ErrCarrierUnavailable},
		{502, `bad gateway`, port.ErrCarrierUnavailable},
		{200, `{"code":400,"message":"soft error"}`, port.ErrCarrierRejected},
	}
	for _, tc := range cases {
		status, payload = tc.status, tc.payload
		_, err := c.GetOrder(context.Background(), 1, "X")
		if !errors.Is(err, tc.want) {
			t.Errorf("status %d: got %v, want %v", tc.status, err, tc.want)
		}
	}
}

func TestCreateOrderAndCancel(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		b := body(t, r)
		switch r.URL.Path {
		case "/shiip/public-api/v2/shipping-order/create":
			if b["payment_type_id"] != float64(2) || b["client_order_code"] != "SHP1" || b["required_note"] != "KHONGCHOXEMHANG" {
				t.Errorf("create body %v", b)
			}
			io.WriteString(w, `{"code":200,"data":{"order_code":"GAND6E3N","total_fee":33000,"expected_delivery_time":"2026-10-01T16:59:59Z"}}`)
		case "/shiip/public-api/v2/switch-status/cancel":
			io.WriteString(w, `{"code":200,"data":[{"order_code":"GAND6E3N","result":false,"message":"Đơn hàng đã được lấy"}]}`)
		}
	})
	o, err := c.CreateOrder(context.Background(), port.CarrierOrderRequest{
		ShopID: 1, ClientOrderCode: "SHP1", PaymentType: model.PaymentByRecipient, RequiredNote: model.NoteNotAllowView,
		Parcel: model.Parcel{Items: []model.Item{{Name: "a", Quantity: 1}}},
	})
	if err != nil || o.OrderCode != "GAND6E3N" || o.Fee != 33000 || o.ExpectedDeliveryAt == nil {
		t.Fatalf("create: %+v %v", o, err)
	}
	if err := c.CancelOrder(context.Background(), 1, "GAND6E3N"); !errors.Is(err, port.ErrCarrierRejected) {
		t.Errorf("cancel refused: %v", err)
	}
}

func TestFindByClientCodeNotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		io.WriteString(w, `{"code":400,"message":"Không tìm thấy đơn hàng"}`)
	})
	if _, err := c.FindByClientOrderCode(context.Background(), 1, "SHP1"); !errors.Is(err, port.ErrNotFound) {
		t.Errorf("got %v", err)
	}
}

func TestLocationsAreCached(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if r.Header.Get("ShopId") != "" {
			t.Error("master data should not send ShopId")
		}
		io.WriteString(w, `{"code":200,"data":[{"WardCode":"20109","DistrictID":1442,"WardName":"Phường Bến Nghé"}]}`)
	})
	l := NewLocations(c, config.FromEnv())
	for i := 0; i < 3; i++ {
		wards, err := l.Wards(context.Background(), 1442)
		if err != nil || len(wards) != 1 || wards[0].Name != "Phường Bến Nghé" {
			t.Fatalf("wards %+v %v", wards, err)
		}
	}
	if calls != 1 {
		t.Errorf("calls = %d, want 1", calls)
	}
}

func TestVerifyWebhookToken(t *testing.T) {
	c := newTestClient(t, func(http.ResponseWriter, *http.Request) {})
	if !c.VerifyWebhookToken("hook") || c.VerifyWebhookToken("nope") || c.VerifyWebhookToken("") {
		t.Error("token check wrong")
	}
	c.webhookToken = ""
	if c.VerifyWebhookToken("") {
		t.Error("empty configured token must reject")
	}
}
