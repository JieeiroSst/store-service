package webqa

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

const goodPage = `<!doctype html><html lang="en"><head><title>Home</title><meta name="viewport" content="width=device-width"></head>
<body><h1>Shop</h1><a href="/about">About</a><a href="/missing">Gone</a><a href="https://other.example/x">ext</a><a href="mailto:a@b.c">mail</a>
<form><label for="e">Email</label><input id="e" name="email"><label>Name <input name="n"></label></form><img src="/a.png" alt="logo"></body></html>`

const badPage = `<html><body><h3>x</h3><img src="/a.png"><a href="/"><img src="/i.png" alt="home"></a><button></button>
<input name="q"><div id="d"></div><div id="d"></div><a href="/x" target="_blank">out</a></body></html>`

func rules(p Page) map[string]bool {
	m := map[string]bool{}
	for _, i := range p.Issues {
		m[i.Rule] = true
	}
	return m
}

func TestCrawl(t *testing.T) {
	mux := http.NewServeMux()
	page := func(body string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "text/html")
			w.Write([]byte(body))
		}
	}
	home := page(goodPage)
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		home(w, r)
	})
	mux.HandleFunc("/about", page(badPage))
	srv := httptest.NewServer(mux)
	defer srv.Close()

	rep := Crawl(context.Background(), srv.Client(), srv.URL, 10, 2)
	if len(rep.Pages) != 2 {
		t.Fatalf("want home+about crawled (external/mailto skipped), got %d pages", len(rep.Pages))
	}
	if len(rep.BrokenLinks) != 2 || rep.BrokenLinks[0].Status != 404 || rep.BrokenLinks[0].From != srv.URL+"/" || rep.BrokenLinks[1].URL != srv.URL+"/x" {
		t.Fatalf("broken links: %+v", rep.BrokenLinks)
	}
	if got := rules(rep.Pages[0]); len(got) != 0 {
		t.Errorf("well-formed page must have no issues, got %v", got)
	}
	bad := rules(rep.Pages[1])
	for _, want := range []string{"html-lang", "title", "viewport", "img-alt", "button-name", "form-label", "duplicate-id", "noopener", "h1"} {
		if !bad[want] {
			t.Errorf("bad page: missing rule %s (got %v)", want, bad)
		}
	}
	if bad["link-name"] {
		t.Error("link containing an img with alt must not be flagged")
	}
	if rep.Errors < 3 {
		t.Errorf("errors = %d", rep.Errors)
	}
}

func TestCrawlUnreachableStart(t *testing.T) {
	rep := Crawl(context.Background(), http.DefaultClient, "http://127.0.0.1:1", 5, 1)
	if len(rep.Pages) != 1 || rep.Errors != 1 {
		t.Fatalf("%+v", rep)
	}
}
