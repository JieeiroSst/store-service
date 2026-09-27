package asyncflow

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/JIeeiroSst/tool-service/internal/apitest"
)

func newService(dedupe bool, lag time.Duration) *httptest.Server {
	var mu sync.Mutex
	created := map[string]time.Time{}
	seen := map[string]string{}
	n := 0
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		defer mu.Unlock()
		switch {
		case r.Method == "POST" && r.URL.Path == "/orders":
			key := r.Header.Get("Idempotency-Key")
			if dedupe && key != "" {
				if id, ok := seen[key]; ok {
					w.WriteHeader(201)
					w.Write([]byte(`{"id":"` + id + `"}`))
					return
				}
			}
			n++
			id := "o" + string(rune('0'+n))
			seen[key] = id
			created[id] = time.Now()
			w.WriteHeader(201)
			w.Write([]byte(`{"id":"` + id + `"}`))
		case r.Method == "GET" && r.URL.Path == "/orders/o1":
			if t, ok := created["o1"]; ok && time.Since(t) > lag {
				w.Write([]byte(`{"status":"PAID","total":59.97}`))
				return
			}
			w.Write([]byte(`{"status":"PENDING","total":59.97}`))
		case r.URL.Path == "/count":
			w.Write([]byte(`{"orders":` + string(rune('0'+n)) + `}`))
		default:
			http.NotFound(w, r)
		}
	}))
}

func spec() Spec {
	return Spec{
		Trigger:    apitest.TestCase{Method: "POST", Path: "/orders", Body: map[string]any{"amount": 59.97}},
		Poll:       apitest.TestCase{Path: "/orders/o1", ExpectJSON: map[string]any{"status": "PAID", "total": "calc:19.99*3"}},
		TimeoutSec: 5, IntervalMs: 100,
	}
}

func TestEventuallyConsistent(t *testing.T) {
	srv := newService(true, 300*time.Millisecond)
	defer srv.Close()
	r := Run(context.Background(), srv.Client(), srv.URL, spec())
	if !r.Passed || r.SettleMs < 250 || r.Polls < 2 {
		t.Fatalf("%+v", r)
	}
}

func TestNeverSettlesFails(t *testing.T) {
	srv := newService(true, time.Hour)
	defer srv.Close()
	s := spec()
	s.TimeoutSec = 1
	r := Run(context.Background(), srv.Client(), srv.URL, s)
	if r.Passed || r.Settled || len(r.Failures) == 0 {
		t.Fatalf("must fail when the effect never appears: %+v", r)
	}
}

func TestIdempotency(t *testing.T) {
	good := newService(true, 0)
	defer good.Close()
	s := spec()
	s.Idempotency = &Idempotency{Repeat: 3}
	if r := Run(context.Background(), good.Client(), good.URL, s); !r.Passed {
		t.Fatalf("idempotent service must pass: %+v", r)
	}
	bad := newService(false, 0)
	defer bad.Close()
	r := Run(context.Background(), bad.Client(), bad.URL, s)
	if r.Passed {
		t.Fatalf("non-idempotent replay must fail: %+v", r)
	}
}

func TestTriggerRejected(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(500) }))
	defer srv.Close()
	if r := Run(context.Background(), srv.Client(), srv.URL, spec()); r.Passed || r.TriggerStatus != 500 {
		t.Fatalf("%+v", r)
	}
}
