package callback

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/JIeeiroSst/shipping-service/config"
	"github.com/JIeeiroSst/shipping-service/internal/domain/model"
)

func TestAllowed(t *testing.T) {
	cfg := config.FromEnv()
	cfg.Callback.AllowedHosts = ".svc.cluster.local, 127.0.0.1"
	s := New(cfg)
	for _, ok := range []string{"http://order-svc.default.svc.cluster.local/x", "http://127.0.0.1:9000/cb"} {
		if err := s.Allowed(ok); err != nil {
			t.Errorf("%s: %v", ok, err)
		}
	}
	for _, bad := range []string{"http://evil.com/x", "http://svc.cluster.local.evil.com/", "ftp://127.0.0.1/", "/relative"} {
		if s.Allowed(bad) == nil {
			t.Errorf("%s accepted", bad)
		}
	}
	if New(config.FromEnv()).Allowed("http://127.0.0.1/") == nil {
		t.Error("empty allowlist must reject")
	}
}

func TestSendSignsAndDoesNotFollowRedirects(t *testing.T) {
	var gotSig, gotID string
	var gotBody []byte
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/redirect" {
			http.Redirect(w, r, "http://169.254.169.254/latest", http.StatusFound)
			return
		}
		gotSig, gotID = r.Header.Get(HeaderSignature), r.Header.Get(HeaderEventID)
		gotBody, _ = io.ReadAll(r.Body)
	}))
	defer srv.Close()
	u, _ := url.Parse(srv.URL)

	cfg := config.FromEnv()
	cfg.Callback.AllowedHosts = u.Hostname()
	cfg.Callback.Secret = "s3cret"
	s := New(cfg)

	if err := s.Send(context.Background(), srv.URL+"/cb", model.CallbackEvent{EventID: "evt-1", Status: model.StatusDelivered}); err != nil {
		t.Fatal(err)
	}
	if gotID != "evt-1" || gotSig != Sign([]byte("s3cret"), gotBody) {
		t.Errorf("id=%q sig=%q", gotID, gotSig)
	}
	if err := s.Send(context.Background(), srv.URL+"/redirect", model.CallbackEvent{EventID: "evt-2"}); err == nil {
		t.Error("redirect response must count as a failed delivery, not be followed")
	}
}
