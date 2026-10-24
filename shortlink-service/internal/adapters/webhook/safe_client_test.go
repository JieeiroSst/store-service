package webhook

import (
	"context"
	"net/http"
	"net/http/httptest"
	"net/netip"
	"testing"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
)

func TestIsPublicAddr(t *testing.T) {
	cases := map[string]bool{
		"8.8.8.8":          true,
		"2606:4700::1111":  true,
		"127.0.0.1":        false,
		"10.1.2.3":         false,
		"172.16.0.1":       false,
		"192.168.1.1":      false,
		"169.254.169.254":  false,
		"100.64.0.1":       false,
		"0.0.0.0":          false,
		"::1":              false,
		"fd00::1":          false,
		"fe80::1":          false,
		"::ffff:127.0.0.1": false,
	}
	for in, want := range cases {
		if got := isPublicAddr(netip.MustParseAddr(in)); got != want {
			t.Errorf("isPublicAddr(%s) = %v, want %v", in, got, want)
		}
	}
}

func TestSendRefusesLoopbackTarget(t *testing.T) {
	hit := false
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hit = true
		_, _ = w.Write([]byte("internal secret"))
	}))
	defer srv.Close()

	res := NewHTTPSender().Send(context.Background(), &domain.Webhook{ID: "w1", URL: srv.URL, TimeoutMs: 2000}, domain.WebhookPayload{EventID: "e1"})
	if res.Success || hit {
		t.Fatalf("Send to loopback succeeded (hit=%v, result=%+v)", hit, res)
	}
	if res.ResponseBody != "" {
		t.Fatalf("ResponseBody = %q, want empty", res.ResponseBody)
	}
}
