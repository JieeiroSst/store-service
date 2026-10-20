package http

import (
	"net/http/httptest"
	"testing"
)

func TestClientInfo(t *testing.T) {
	r := httptest.NewRequest("POST", "/", nil)
	r.RemoteAddr = "10.1.2.3:5555"
	r.Header.Set("X-Forwarded-For", "203.0.113.9, 10.0.0.1")
	r.Header.Set("X-Device-Id", "dev-1")
	r.Header.Set("User-Agent", "LuckyApp/2")
	if c := clientInfo(r); c.IP != "203.0.113.9" || c.DeviceID != "dev-1" || c.UserAgent != "LuckyApp/2" {
		t.Fatalf("forwarded: %+v", c)
	}
	r.Header.Set("X-Real-IP", "198.51.100.1")
	if c := clientInfo(r); c.IP != "198.51.100.1" {
		t.Fatalf("gateway X-Real-IP wins: %+v", c)
	}
	r.Header.Del("X-Real-IP")
	r.Header.Del("X-Forwarded-For")
	if c := clientInfo(r); c.IP != "10.1.2.3" {
		t.Fatalf("remote addr fallback: %+v", c)
	}
}
