package elasticsearch

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/JIeeiroSst/search-service/config"
	"github.com/JIeeiroSst/search-service/internal/domain"
)

func TestBuildSortResolvesFields(t *testing.T) {
	caps := map[string]fieldCap{
		"name":            {types: []string{"text"}},
		"name.sort":       {types: []string{"keyword"}, aggregatable: true},
		"price":           {types: []string{"double"}, aggregatable: true},
		"amount":          {types: []string{"float", "long"}, aggregatable: true},
		"mixed":           {types: []string{"keyword", "long"}, aggregatable: true},
		"created":         {types: []string{"date", "text"}},
		"created.keyword": {types: []string{"keyword"}, aggregatable: true},
		"status.raw":      {types: []string{"text"}},
	}
	sort, err := buildSort(domain.DocumentQuery{Keyword: "x", Sort: []domain.SortField{{Field: "name"}, {Field: "price", Desc: true}}}, caps)
	if err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(sort)
	want := `[{"name.sort":{"missing":"_last","order":"asc","unmapped_type":"keyword"}},{"price":{"missing":"_last","order":"desc","unmapped_type":"double"}},{"_index":"asc"},{"_id":"asc"}]`
	if string(out) != want {
		t.Errorf("sort = %s", out)
	}

	sort, _ = buildSort(domain.DocumentQuery{Sort: []domain.SortField{{Field: "amount"}}}, caps)
	if out, _ := json.Marshal(sort[0]); string(out) != `{"amount":{"missing":"_last","numeric_type":"double","order":"asc","unmapped_type":"double"}}` {
		t.Errorf("mixed numeric sort = %s", out)
	}
	sort, _ = buildSort(domain.DocumentQuery{Keyword: "x"}, caps)
	if out, _ := json.Marshal(sort[0]); string(out) != `{"_score":"desc"}` {
		t.Errorf("default search sort = %s", out)
	}
	for _, f := range []string{"mixed", "missing", "status.raw", "created"} {
		if _, err := buildSort(domain.DocumentQuery{Sort: []domain.SortField{{Field: f}}}, caps); !domain.IsInvalid(err) {
			t.Errorf("%s should be invalid, got %v", f, err)
		}
	}
}

func TestCursorRoundTripAndMismatch(t *testing.T) {
	raw, err := encodeCursor(cursor{Kind: "search", Fingerprint: "fp", SearchAfter: []any{1.5, "accounts", "42"}})
	if err != nil {
		t.Fatal(err)
	}
	c, err := decodeCursor(raw, "search", "fp")
	if err != nil || c.SearchAfter[0] != json.Number("1.5") {
		t.Fatalf("round trip: %+v %v", c, err)
	}
	for _, tc := range [][3]string{{"!!!", "search", "fp"}, {raw, "list", "fp"}, {raw, "search", "other"}} {
		if _, err := decodeCursor(tc[0], tc[1], tc[2]); !domain.IsInvalid(err) {
			t.Errorf("%v should be invalid", tc)
		}
	}
}

func TestSchemaBodies(t *testing.T) {
	replicas := 0
	policy, _ := domain.NewSensitivePolicy("")
	cfg := config.Defaults()
	cfg.Elasticsearch.Schema = config.SchemaConfig{NumberOfReplicas: &replicas, Synonyms: []string{"tv, tivi"}}
	s := NewSchemaManager(nil, nil, cfg, policy)

	settings, err := s.settings()
	if err != nil {
		t.Fatal(err)
	}
	analysis := settings["analysis"].(map[string]any)
	if analysis["filter"].(map[string]any)["synonyms"] == nil || settings["index"].(map[string]any)["number_of_replicas"] != 0 {
		t.Errorf("settings = %v", settings)
	}
	mappings, err := s.mappings()
	if err != nil {
		t.Fatal(err)
	}
	out, _ := json.Marshal(mappings)
	if strings.Contains(string(out), sensitivePlaceholder) || !strings.Contains(string(out), "password") {
		t.Error("sensitive placeholder not replaced")
	}
}

type fakeES struct {
	path   string
	query  string
	body   map[string]any
	status int
	reply  string
}

func newFakeES(t *testing.T, f *fakeES) *Searcher {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Elastic-Product", "Elasticsearch")
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/" && r.Method == http.MethodGet {
			_, _ = io.WriteString(w, `{"version":{"number":"7.17.5"},"tagline":"You Know, for Search"}`)
			return
		}
		f.path, f.query, f.body = r.URL.Path, r.URL.RawQuery, nil
		_ = json.NewDecoder(r.Body).Decode(&f.body)
		if f.status != 0 {
			w.WriteHeader(f.status)
		}
		_, _ = io.WriteString(w, f.reply)
	}))
	t.Cleanup(srv.Close)
	cfg := config.Defaults()
	cfg.Elasticsearch.DNS = srv.URL
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	return NewSearcher(client, cfg)
}

const threeHits = `{"hits":{"total":{"value":7},"hits":[
{"_index":"accounts__v17","_id":"1","_score":2.5,"_source":{"name":"phone a"},"highlight":{"name":["<em>phone</em> a"],"all_text":["x"]},"sort":[2.5,"accounts__v17","1"]},
{"_index":"book-store.public.book_store_book","_id":"9","_score":1.25,"_source":{"title":"phone b"},"sort":[1.25,"book-store.public.book_store_book","9"]},
{"_index":"baskets","_id":"3","_score":1.0,"_source":{"note":"phone c"},"sort":[1.0,"baskets","3"]}]},
"aggregations":{"_index":{"buckets":[{"key":"accounts__v17","doc_count":1}]}},
"suggest":{"did_you_mean":[{"options":[{"text":"phones"}]}]}}`

func TestQueryPaginates(t *testing.T) {
	f := &fakeES{reply: threeHits}
	s := newFakeES(t, f)

	q := domain.DocumentQuery{Keyword: "phone", Size: 2}
	page, err := s.Query(context.Background(), q, domain.ModeSearch)
	if err != nil {
		t.Fatal(err)
	}
	if f.path != "/*,-.*/_search" || !strings.Contains(f.query, "track_total_hits=true") || f.body["size"] != float64(3) {
		t.Errorf("request %s?%s size=%v", f.path, f.query, f.body["size"])
	}
	if _, ok := f.body["search_after"]; ok {
		t.Error("first page must not send search_after")
	}
	if len(page.Items) != 2 || !page.HasMore || page.NextCursor == "" || page.Total != 7 || page.Suggestion != "phones" {
		t.Fatalf("page = %+v", page)
	}
	if page.Items[0].Index != "accounts" || page.Facets["_index"][0].Value != "accounts" || len(page.Items[0].Highlight) != 1 {
		t.Errorf("display index / highlight: %+v %+v", page.Items[0], page.Facets)
	}

	q.Cursor = page.NextCursor
	if _, err := s.Query(context.Background(), q, domain.ModeSearch); err != nil {
		t.Fatal(err)
	}
	after, _ := json.Marshal(f.body["search_after"])
	if string(after) != `[1.25,"book-store.public.book_store_book","9"]` || f.body["aggs"] != nil {
		t.Errorf("second page body: search_after=%s aggs=%v", after, f.body["aggs"])
	}

	q.Keyword = "other"
	if _, err := s.Query(context.Background(), q, domain.ModeSearch); !domain.IsInvalid(err) {
		t.Errorf("cursor from another query must be rejected, got %v", err)
	}

	f.reply = threeHits
	if _, err := s.Query(context.Background(), domain.DocumentQuery{Indices: []string{"accounts", "book-store.*"}, Size: 5}, domain.ModeList); err != nil {
		t.Fatal(err)
	}
	if f.path != "/accounts,book-store.*,-.*/_search" {
		t.Errorf("path = %s", f.path)
	}
}

func TestErrorMapping(t *testing.T) {
	cases := []struct {
		status int
		reply  string
		check  func(error) bool
	}{
		{404, `{"_index":"x","found":false}`, func(err error) bool { return errors.Is(err, domain.ErrNotFound) }},
		{400, `{"error":{"root_cause":[{"reason":"failed to parse date field [abc]"}]}}`, func(err error) bool {
			return domain.IsInvalid(err) && strings.Contains(err.Error(), "abc")
		}},
		{503, `{"error":{"reason":"all shards failed"}}`, func(err error) bool { return errors.Is(err, domain.ErrUnavailable) }},
		{500, `{"error":{"type":"x","reason":"boom"}}`, func(err error) bool {
			return err != nil && !domain.IsInvalid(err) && !errors.Is(err, domain.ErrUnavailable)
		}},
	}
	for _, c := range cases {
		s := newFakeES(t, &fakeES{status: c.status, reply: c.reply})
		_, err := s.Get(context.Background(), domain.DocumentRef{Index: "x", ID: "1"})
		if !c.check(err) {
			t.Errorf("status %d mapped to %v", c.status, err)
		}
	}

	cfg := config.Defaults()
	cfg.Elasticsearch.DNS = "http://127.0.0.1:1"
	client, _ := NewClient(cfg)
	if err := NewHealth(client).Ping(context.Background()); !errors.Is(err, domain.ErrUnavailable) {
		t.Errorf("unreachable cluster: %v", err)
	}
}
