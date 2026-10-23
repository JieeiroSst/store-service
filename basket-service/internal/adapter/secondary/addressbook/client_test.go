package addressbook

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/basket-service/common"
)

func TestShippingAddressReadsTheGHNCodes(t *testing.T) {
	var auth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/user-addresses/4" {
			t.Errorf("path %s", r.URL.Path)
		}
		auth = r.Header.Get("Authorization")
		_, _ = w.Write([]byte(`{"data":{"first_name":"Lưu Mạnh","last_name":"Quân","line1":"24 Lê Văn Hưu","line2":"Phường Bến Nghé, Quận 1","state":"Hồ Chí Minh","postcode":"ghn:202:1442:20109","phone_number":"0912345678"}}`))
	}))
	defer srv.Close()

	a, err := New(srv.URL+"/", 0).ShippingAddress(context.Background(), "user-token", 4)
	if err != nil {
		t.Fatal(err)
	}
	if auth != "Bearer user-token" {
		t.Fatalf("authorization %q", auth)
	}
	if a.Name != "Lưu Mạnh Quân" || a.Street != "24 Lê Văn Hưu" || a.Phone != "0912345678" || a.Ward != "Phường Bến Nghé" || a.District != "Quận 1" || a.Province != "Hồ Chí Minh" {
		t.Fatalf("address %+v", a)
	}
	if a.ProvinceID != 202 || a.DistrictID != 1442 || a.WardCode != "20109" {
		t.Fatalf("ghn codes %+v", a)
	}
}

func TestShippingAddressWithoutGHNCodes(t *testing.T) {
	a := toShipping(userAddress{FirstName: "A", Line1: "x", Postcode: "70000"})
	if a.WardCode != "" || a.DistrictID != 0 {
		t.Fatalf("plain postcode must not produce GHN codes: %+v", a)
	}
}

func TestShippingAddressErrors(t *testing.T) {
	for status, want := range map[int]error{http.StatusForbidden: common.ErrNotFound, http.StatusNotFound: common.ErrNotFound, http.StatusBadGateway: common.ErrUpstream} {
		srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(status) }))
		_, err := New(srv.URL, 0).ShippingAddress(context.Background(), "Bearer t", 1)
		srv.Close()
		if !errors.Is(err, want) {
			t.Fatalf("%d: got %v", status, err)
		}
	}
}
