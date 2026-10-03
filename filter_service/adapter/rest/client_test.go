package rest

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func TestGetUnwrapsAndForwardsHeaders(t *testing.T) {
	var got *http.Request
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r
		w.Write([]byte(`{"data":{"items":[{"id":1},{"id":2}]}}`))
	}))
	defer srv.Close()

	incoming := http.Header{}
	incoming.Set("Authorization", "Bearer abc")
	incoming.Set("X-Api-Key", "k")
	incoming.Set("X-Other", "dropped")
	ctx := WithIncomingHeaders(context.Background(), incoming)

	c := New("svc", srv.URL+"/", nil, "X-Api-Key")
	var out []struct{ ID int }
	err := c.Get(ctx, "/things", url.Values{"page": {"2"}}, http.Header{"X-User-Id": {"u1"}}, "data.items", &out)
	if err != nil {
		t.Fatal(err)
	}
	if len(out) != 2 || out[1].ID != 2 {
		t.Fatalf("out = %+v", out)
	}
	if got.URL.Path != "/things" || got.URL.Query().Get("page") != "2" {
		t.Fatalf("request = %s", got.URL)
	}
	if got.Header.Get("Authorization") != "Bearer abc" || got.Header.Get("X-Api-Key") != "k" || got.Header.Get("X-User-Id") != "u1" {
		t.Fatalf("headers = %v", got.Header)
	}
	if got.Header.Get("X-Other") != "" {
		t.Fatal("unlisted header was forwarded")
	}
}

func TestServiceSpecificHeaderNotSentElsewhere(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header
		w.Write([]byte(`{}`))
	}))
	defer srv.Close()

	incoming := http.Header{}
	incoming.Set("X-Api-Key", "secret")
	ctx := WithIncomingHeaders(context.Background(), incoming)

	var out map[string]any
	if err := New("other", srv.URL, nil).Get(ctx, "/", nil, nil, "", &out); err != nil {
		t.Fatal(err)
	}
	if got.Get("X-Api-Key") != "" {
		t.Fatal("X-Api-Key leaked to a service that did not ask for it")
	}
}

func TestGetErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/missing":
			w.WriteHeader(http.StatusNotFound)
			w.Write([]byte(`{"error":"room not found"}`))
		case "/badtype":
			w.Write([]byte(`{"id":"not-a-number"}`))
		case "/empty":
			w.WriteHeader(http.StatusNoContent)
		}
	}))
	defer srv.Close()
	c := New("svc", srv.URL, nil)

	var out struct {
		ID *int `json:"id"`
	}
	err := c.Get(context.Background(), "/missing", nil, nil, "", &out)
	var apiErr *Error
	if !errors.As(err, &apiErr) || apiErr.Status != 404 || apiErr.Body != "room not found" {
		t.Fatalf("err = %v", err)
	}

	err = c.Get(context.Background(), "/badtype", nil, nil, "", &out)
	if err == nil || !strings.Contains(err.Error(), `"id"`) {
		t.Fatalf("type mismatch err = %v", err)
	}

	var list []int
	if err := c.Get(context.Background(), "/empty", nil, nil, "", &list); err != nil || list != nil {
		t.Fatalf("204 = %v, %v", list, err)
	}

}

func TestDecodeUnwrapMissingKey(t *testing.T) {
	var out any
	if err := Decode([]byte(`{"items":[]}`), "data", &out); err == nil {
		t.Fatal("expected error for missing envelope key")
	}
}

func TestEscapePath(t *testing.T) {
	if got := EscapePath("/a b/c%d/"); got != "a%20b/c%25d" {
		t.Fatalf("EscapePath = %q", got)
	}
}

func TestDecodeNullBodyWithEnvelope(t *testing.T) {
	out := []int{1}
	if err := Decode([]byte("null"), "data.items", &out); err != nil {
		t.Fatal(err)
	}
}
