package elasticsearch

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/search-service/config"
	"github.com/JIeeiroSst/search-service/internal/domain"
	esv7api "github.com/elastic/go-elasticsearch/v7/esapi"
)

func integrationSetup(t *testing.T) (*Searcher, *SchemaManager, string) {
	t.Helper()
	url := os.Getenv("SEARCH_TEST_ES_URL")
	if url == "" {
		t.Skip("SEARCH_TEST_ES_URL not set")
	}

	prefix := fmt.Sprintf("itest%d", time.Now().UnixNano())
	replicas := 0
	cfg := config.Defaults()
	cfg.Elasticsearch.DNS = url
	cfg.Elasticsearch.Indices = []string{prefix + "*"}
	cfg.Elasticsearch.Schema = config.SchemaConfig{
		Name:             prefix,
		IndexPatterns:    []string{prefix + "*"},
		Priority:         200,
		NumberOfReplicas: &replicas,
		Synonyms:         []string{"tivi, tv, television"},
	}
	client, err := NewClient(cfg)
	if err != nil {
		t.Fatal(err)
	}
	policy, _ := domain.NewSensitivePolicy("")
	searcher := NewSearcher(client, cfg)
	schema := NewSchemaManager(client, searcher, cfg, policy)

	t.Cleanup(func() {
		ctx := context.Background()
		_ = call(ctx, client, esv7api.IndicesDeleteRequest{Index: []string{prefix + "*"}}, nil)
		_ = call(ctx, client, esv7api.IndicesDeleteIndexTemplateRequest{Name: prefix}, nil)
		_ = call(ctx, client, esv7api.ClusterDeleteComponentTemplateRequest{Name: prefix + "-schema"}, nil)
	})
	return searcher, schema, prefix
}

func bulk(t *testing.T, s *Searcher, lines ...string) {
	t.Helper()
	body := strings.NewReader(strings.Join(lines, "\n") + "\n")
	var res struct {
		Errors bool `json:"errors"`
	}
	if err := call(context.Background(), s.client, esv7api.BulkRequest{Body: body, Refresh: "true"}, &res); err != nil || res.Errors {
		t.Fatalf("bulk: %v errors=%v", err, res.Errors)
	}
}

func TestIntegrationSchemaSearchAndReindex(t *testing.T) {
	searcher, schema, prefix := integrationSetup(t)
	ctx := context.Background()
	legacy := prefix + "-legacy"
	products := prefix + "-products"

	plain := prefix + "-plain"
	body := strings.NewReader(fmt.Sprintf(`{"index_patterns":["%s"],"priority":1000,"template":{"settings":{"number_of_replicas":0}}}`, legacy))
	if err := call(ctx, searcher.client, esv7api.IndicesPutIndexTemplateRequest{Name: plain, Body: body}, nil); err != nil {
		t.Fatal(err)
	}
	bulk(t, searcher,
		fmt.Sprintf(`{"index":{"_index":"%s","_id":"1"}}`, legacy),
		`{"id":1,"name":"Nguyen Van A","created_at":"2026-01-02 10:00:00"}`,
	)
	if err := call(ctx, searcher.client, esv7api.IndicesDeleteIndexTemplateRequest{Name: plain}, nil); err != nil {
		t.Fatal(err)
	}

	if err := schema.Apply(ctx); err != nil {
		t.Fatalf("apply: %v", err)
	}
	status, err := schema.Status(ctx)
	if err != nil || !status.Applied || status.Version != SchemaVersion {
		t.Fatalf("status %+v %v", status, err)
	}
	if len(status.Indices) != 1 || status.Indices[0].Managed {
		t.Fatalf("legacy index should be unmanaged: %+v", status.Indices)
	}

	bulk(t, searcher,
		fmt.Sprintf(`{"index":{"_index":"%s","_id":"1"}}`, products),
		`{"id":1,"sku":"SKU-IP15","name":"Điện thoại iPhone 15","price":31000000,"status":"active","api_token":"tok_1","created_at":"2026-03-01 08:00:00"}`,
		fmt.Sprintf(`{"index":{"_index":"%s","_id":"2"}}`, products),
		`{"id":2,"sku":"SKU-SS24","name":"Điện thoại Samsung Galaxy","price":22000000,"status":"active","created_at":"2026-03-05 08:00:00"}`,
		fmt.Sprintf(`{"index":{"_index":"%s","_id":"3"}}`, products),
		`{"id":3,"sku":"SKU-TV55","name":"Tivi Sony 55 inch","price":15000000,"status":"inactive","created_at":"2026-02-10 08:00:00"}`,
		fmt.Sprintf(`{"index":{"_index":"%s","_id":"4"}}`, products),
		`{"id":4,"sku":"SKU-LG65","name":"Tivi LG 65 inch","price":"not-a-number","status":"active"}`,
	)

	ids := func(page domain.DocumentPage) string {
		var out []string
		for _, d := range page.Items {
			out = append(out, d.ID)
		}
		return strings.Join(out, ",")
	}
	search := func(q domain.DocumentQuery) domain.DocumentPage {
		t.Helper()
		if err := q.Normalize(domain.ModeSearch); err != nil {
			t.Fatal(err)
		}
		page, err := searcher.Query(ctx, q, domain.ModeSearch)
		if err != nil {
			t.Fatalf("search %+v: %v", q, err)
		}
		return page
	}

	if got := ids(search(domain.DocumentQuery{Keyword: "dien thoai", Indices: []string{products}})); got != "1,2" && got != "2,1" {
		t.Errorf("folding: %s", got)
	}
	if got := search(domain.DocumentQuery{Keyword: "tv", Indices: []string{products}}); got.Total != 2 {
		t.Errorf("synonym: %d", got.Total)
	}
	if got := ids(search(domain.DocumentQuery{Keyword: "galax", Indices: []string{products}})); got != "2" {
		t.Errorf("prefix: %s", got)
	}
	if got := ids(search(domain.DocumentQuery{Keyword: "samsumg", Indices: []string{products}})); got != "2" {
		t.Errorf("fuzzy: %s", got)
	}
	if got := ids(search(domain.DocumentQuery{Keyword: "22000000", Indices: []string{products}})); got != "2" {
		t.Errorf("number: %s", got)
	}
	if got := search(domain.DocumentQuery{Keyword: "tok_1", Indices: []string{products}}); got.Total != 0 {
		t.Errorf("sensitive field must not be searchable: %d", got.Total)
	}

	q := domain.DocumentQuery{
		Keyword: "dien thoai",
		Indices: []string{products},
		Filters: map[string][]string{"status": {"active"}},
		Ranges:  map[string]domain.Range{"created_at": {Gte: "2026-03-01"}},
		Sort:    domain.ParseSort("price:desc"),
		Facets:  []string{"status"},
		Size:    1,
	}
	first := search(q)
	if ids(first) != "1" || !first.HasMore || first.Facets["status"][0].Count != 2 {
		t.Fatalf("first page %+v", first)
	}
	q.Cursor = first.NextCursor
	if second := search(q); ids(second) != "2" || second.HasMore {
		t.Fatalf("second page %+v", second)
	}

	if _, err := searcher.Query(ctx, domain.DocumentQuery{Sort: domain.ParseSort("created_at"), Size: 10}, domain.ModeList); !domain.IsInvalid(err) {
		t.Errorf("date vs text conflict must be rejected before migration, got %v", err)
	}

	result, err := schema.Reindex(ctx, legacy)
	if err != nil || result.Copied != 1 || !strings.HasPrefix(result.To, legacy+"__v") {
		t.Fatalf("reindex %+v %v", result, err)
	}
	status, _ = schema.Status(ctx)
	for _, idx := range status.Indices {
		if !idx.Managed || idx.WriteBlocked {
			t.Errorf("after reindex: %+v", idx)
		}
	}
	page, err := searcher.Query(ctx, domain.DocumentQuery{Sort: domain.ParseSort("created_at:desc"), Size: 10}, domain.ModeList)
	if err != nil || page.Total != 5 || page.Items[0].ID != "2" {
		t.Fatalf("sort after migration: %+v %v", page, err)
	}
	doc, err := searcher.Get(ctx, domain.DocumentRef{Index: legacy, ID: "1"})
	if err != nil || doc.Index != legacy {
		t.Fatalf("get through alias: %+v %v", doc, err)
	}
}
