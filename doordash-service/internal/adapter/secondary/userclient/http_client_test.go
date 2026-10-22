package userclient

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/JIeeiroSst/doordash-service/config"
	"github.com/JIeeiroSst/doordash-service/pkg/authn"
)

func TestGetCustomerForwardsCallerToken(t *testing.T) {
	var gotAuth, gotPath string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth, gotPath = r.Header.Get("Authorization"), r.URL.Path
		_ = json.NewEncoder(w).Encode(map[string]any{"id": 7, "username": "quan", "name": "Quan", "active": true})
	}))
	defer srv.Close()
	c := NewUserClient(&config.Config{UserService: config.ExternalServiceConfig{BaseURL: srv.URL, Timeout: "2s"}})

	ctx := authn.WithPrincipal(context.Background(), authn.Principal{UserID: "7", Token: "session-token"})
	customer, err := c.GetCustomer(ctx, "7")
	if err != nil {
		t.Fatal(err)
	}
	if gotPath != "/api/v1/users/7/profile" || gotAuth != "Bearer session-token" {
		t.Fatalf("unexpected call: %s %q", gotPath, gotAuth)
	}
	if customer.ID != "7" || customer.Name != "Quan" {
		t.Fatalf("unexpected customer: %+v", customer)
	}
}
