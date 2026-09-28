package notification

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

func TestSlackPostsFormattedAlert(t *testing.T) {
	var got map[string]any
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/notifications/slack" || r.Header.Get("X-Requested-By") != "vending" {
			t.Errorf("unexpected %s %s", r.URL.Path, r.Header.Get("X-Requested-By"))
		}
		json.NewDecoder(r.Body).Decode(&got)
		rw.WriteHeader(http.StatusAccepted)
	}))
	defer srv.Close()

	s := NewSlack(httpx.New(srv.URL, time.Second, "vending"))
	err := s.Notify(context.Background(), domain.Event{
		ID: "e1", EventType: domain.EventInventoryEmpty, RelatedEntity: "inventory", EntityID: "i1",
		MachineID: "m1", Data: map[string]any{"slot": "A1", "product": "Cola"}, OccurredAt: time.Now(),
	})
	if err != nil {
		t.Fatal(err)
	}
	text, _ := got["text"].(string)
	if got["title"] != "Vending slot is empty" || !strings.Contains(text, "*slot*: A1") || !strings.Contains(text, "*machine*: m1") {
		t.Fatalf("payload %v", got)
	}
}

func TestSlackSurfacesFailure(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		rw.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()
	if err := NewSlack(httpx.New(srv.URL, time.Second, "")).Notify(context.Background(), domain.Event{}); err == nil {
		t.Fatal("expected error so the alert is retried")
	}
}
