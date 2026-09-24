package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/config"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

func TestNotifyChoosesEmailOrPush(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/notifications" {
			t.Errorf("path = %s", r.URL.Path)
		}
		_ = json.NewDecoder(r.Body).Decode(&got)
		w.WriteHeader(http.StatusCreated)
	}))
	defer srv.Close()
	c := New(&config.Config{Notification: config.UpstreamConfig{BaseURL: srv.URL, Timeout: time.Second}})

	_ = c.Notify(context.Background(), port.Notification{PatientID: 4, Email: "a@b.vn", Title: "t", Message: "m"})
	if got["type"] != "email" || got["recipient"] != "a@b.vn" || got["user_id"] != float64(4) {
		t.Errorf("email notification = %v", got)
	}
	_ = c.Notify(context.Background(), port.Notification{PatientID: 4, Title: "t", Message: "m"})
	if got["type"] != "push" || got["recipient"] != "4" {
		t.Errorf("push notification = %v", got)
	}
}

func TestNotifyIsANoOpWhenUnconfigured(t *testing.T) {
	if err := New(&config.Config{}).Notify(context.Background(), port.Notification{PatientID: 1}); err != nil {
		t.Errorf("err = %v", err)
	}
}
