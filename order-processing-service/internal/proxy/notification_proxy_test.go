package proxy

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/order-processing-service/internal/domain/entity"
)

func TestNotificationProxySend(t *testing.T) {
	var got map[string]any
	var token string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/notifications" || r.Method != http.MethodPost {
			t.Errorf("unexpected %s %s", r.Method, r.URL.Path)
		}
		token = r.Header.Get("X-Internal-Token")
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":42,"status":"pending"}`))
	}))
	defer srv.Close()

	res, err := NewNotificationProxy(srv.URL, "s3cret").Send(context.Background(), entity.NotificationRequest{CustomerID: "7", OrderID: "o-1", Type: "email", Message: "shipped"})
	if err != nil || !res.Sent || res.ID != "42" {
		t.Fatalf("res=%+v err=%v", res, err)
	}
	if token != "s3cret" || got["user_id"] != float64(7) || got["message"] != "shipped" {
		t.Fatalf("token=%q body=%v", token, got)
	}
	if _, err := NewNotificationProxy(srv.URL, "").Send(context.Background(), entity.NotificationRequest{CustomerID: "abc"}); err == nil {
		t.Fatal("non-numeric customer id accepted")
	}
}
