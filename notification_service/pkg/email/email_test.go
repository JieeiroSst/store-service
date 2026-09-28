package email

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSendBatch(t *testing.T) {
	var got []map[string]any
	var key string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		key = r.Header.Get("Idempotency-Key")
		if r.Header.Get("Authorization") != "Bearer re_test" {
			t.Errorf("auth header %q", r.Header.Get("Authorization"))
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		_, _ = w.Write([]byte(`{"data":[{"id":"1"},{"id":"2"}]}`))
	}))
	defer srv.Close()

	c := NewClient("re_test", "shop@example.vn").WithEndpoints(srv.URL, srv.URL)
	err := c.SendBatch(context.Background(), []Message{{To: "a@x.vn", Subject: "s", HTML: "<p>1</p>"}, {To: "b@x.vn", Subject: "s", HTML: "<p>1</p>"}}, "campaign-1-batch-2")
	if err != nil {
		t.Fatal(err)
	}
	if key != "campaign-1-batch-2" || len(got) != 2 || got[1]["from"] != "shop@example.vn" || got[1]["to"].([]any)[0] != "b@x.vn" {
		t.Errorf("key=%q body=%v", key, got)
	}
}

func TestSendBatchErrors(t *testing.T) {
	status := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(status)
	}))
	defer srv.Close()
	c := NewClient("k", "f@x.vn").WithEndpoints(srv.URL, srv.URL)
	one := []Message{{To: "a@x.vn"}}

	for code, retryable := range map[int]bool{422: false, 429: true, 503: true} {
		status = code
		var apiErr *APIError
		if err := c.SendBatch(context.Background(), one, ""); !errors.As(err, &apiErr) || apiErr.Retryable() != retryable {
			t.Errorf("status %d: err=%v", code, err)
		}
	}
	if err := c.SendBatch(context.Background(), make([]Message, MaxBatchSize+1), ""); err == nil {
		t.Error("batch above the Resend limit accepted")
	}
}
