package videoclient

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/JIeeiroSst/movie-recommendation-service/config"
)

func TestFetchPagesThroughCatalog(t *testing.T) {
	const total = 150
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		page := 1
		fmt.Sscan(r.URL.Query().Get("page"), &page)
		start, end := (page-1)*pageSize, min(page*pageSize, total)
		fmt.Fprintf(w, `{"total":%d,"items":[`, total)
		for i := start; i < end; i++ {
			if i > start {
				fmt.Fprint(w, ",")
			}
			fmt.Fprintf(w, `{"id":"v%d","title":"t%d","status":"ready","views":%d}`, i, i, i)
		}
		fmt.Fprint(w, "]}")
	}))
	defer srv.Close()

	cfg := &config.Config{}
	cfg.VideoService.BaseURL = srv.URL
	cfg.VideoService.Timeout = 5 * time.Second
	got, err := New(cfg).Fetch(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != total || got[149].ID != "v149" || got[149].Views != 149 {
		t.Fatalf("got %d videos, last %+v", len(got), got[len(got)-1])
	}
}

func TestFetchErrors(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	}))
	defer srv.Close()
	cfg := &config.Config{}
	cfg.VideoService.BaseURL = srv.URL
	cfg.VideoService.Timeout = time.Second
	if _, err := New(cfg).Fetch(t.Context()); err == nil {
		t.Fatal("expected error on 500")
	}
	if _, err := New(&config.Config{}).Fetch(t.Context()); err == nil {
		t.Fatal("expected error with no base URL")
	}
}
