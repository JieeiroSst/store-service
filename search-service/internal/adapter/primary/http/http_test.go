package http

import (
	"context"
	"encoding/base64"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/JIeeiroSst/search-service/config"
	"github.com/JIeeiroSst/search-service/internal/application"
	"github.com/JIeeiroSst/search-service/internal/domain"
)

type fakeHealth struct{ err error }

func (f fakeHealth) Ping(context.Context) error { return f.err }

type fakeSearcher struct{ err error }

func (f fakeSearcher) Query(context.Context, domain.DocumentQuery, domain.QueryMode) (domain.DocumentPage, error) {
	return domain.DocumentPage{Items: []domain.Document{}}, f.err
}
func (f fakeSearcher) Autocomplete(context.Context, domain.AutocompleteQuery) ([]domain.Document, error) {
	return nil, f.err
}
func (f fakeSearcher) Get(context.Context, domain.DocumentRef) (domain.Document, error) {
	return domain.Document{}, f.err
}
func (f fakeSearcher) Similar(context.Context, domain.DocumentRef, int) ([]domain.Document, error) {
	return nil, f.err
}
func (f fakeSearcher) Indices(context.Context) ([]domain.IndexInfo, error) { return nil, f.err }
func (f fakeSearcher) Fields(context.Context, string) ([]domain.FieldInfo, error) {
	return nil, f.err
}

func newTestRouter(t *testing.T, searchErr, healthErr error) http.Handler {
	t.Helper()
	cfg := config.Defaults()
	cfg.Secret.AuthorizeKey = "no_name"
	policy, _ := domain.NewSensitivePolicy("")
	docs := application.NewDocumentService(fakeSearcher{err: searchErr}, policy)
	return NewRouter(NewHandler(cfg, docs, application.NewSchemaService(nil), fakeHealth{err: healthErr}))
}

func do(h http.Handler, method, target, auth string) int {
	req := httptest.NewRequest(method, target, nil)
	if auth != "" {
		req.Header.Set("Authorization", auth)
	}
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	return rec.Code
}

func TestAuthorization(t *testing.T) {
	h := newTestRouter(t, nil, nil)
	legacy := base64.StdEncoding.EncodeToString([]byte("no_name"))
	cases := map[string]int{
		legacy:           http.StatusOK,
		"Bearer no_name": http.StatusOK,
		"":               http.StatusUnauthorized,
		"Bearer NO_NAME": http.StatusUnauthorized,
		base64.StdEncoding.EncodeToString([]byte("NO_NAME")): http.StatusUnauthorized,
		"no_name": http.StatusUnauthorized,
	}
	for auth, want := range cases {
		if got := do(h, http.MethodGet, "/api/v1/search?q=x", auth); got != want {
			t.Errorf("auth %q: got %d want %d", auth, got, want)
		}
	}
	if got := do(h, http.MethodGet, "/api/v1/admin/schema", ""); got != http.StatusUnauthorized {
		t.Errorf("admin without auth: %d", got)
	}
}

func TestErrorStatus(t *testing.T) {
	auth := "Bearer no_name"
	cases := []struct {
		err    error
		target string
		want   int
	}{
		{nil, "/api/v1/search", http.StatusBadRequest},
		{nil, "/api/v1/search?q=x&size=abc", http.StatusBadRequest},
		{nil, "/api/v1/documents?index=.security", http.StatusBadRequest},
		{domain.ErrNotFound, "/api/v1/documents/products/1", http.StatusNotFound},
		{domain.ErrUnavailable, "/api/v1/documents", http.StatusServiceUnavailable},
		{errors.New("boom"), "/api/v1/indices", http.StatusInternalServerError},
		{nil, "/api/v1/nope", http.StatusNotFound},
	}
	for _, c := range cases {
		if got := do(newTestRouter(t, c.err, nil), http.MethodGet, c.target, auth); got != c.want {
			t.Errorf("%s with %v: got %d want %d", c.target, c.err, got, c.want)
		}
	}
}

func TestHealth(t *testing.T) {
	if got := do(newTestRouter(t, nil, nil), http.MethodGet, "/health", ""); got != http.StatusOK {
		t.Errorf("healthy: %d", got)
	}
	if got := do(newTestRouter(t, nil, domain.ErrUnavailable), http.MethodGet, "/health", ""); got != http.StatusServiceUnavailable {
		t.Errorf("unhealthy: %d", got)
	}
	if got := do(newTestRouter(t, nil, domain.ErrUnavailable), http.MethodGet, "/livez", ""); got != http.StatusOK {
		t.Errorf("livez: %d", got)
	}
}

func TestParseDocumentQuery(t *testing.T) {
	values, _ := url.ParseQuery("q=tv&index=a&index=b&filter[status]=active&filter[status]=new&gte[price]=10&lt[price]=20&sort=price:desc&facets=status&size=5&cursor=c&bogus[x]=1&filter[]=y")
	q, err := parseDocumentQuery(values)
	if err != nil {
		t.Fatal(err)
	}
	if q.Keyword != "tv" || len(q.Indices) != 2 || len(q.Filters["status"]) != 2 || len(q.Filters) != 1 {
		t.Errorf("query = %+v", q)
	}
	if q.Ranges["price"] != (domain.Range{Gte: "10", Lt: "20"}) || len(q.Ranges) != 1 {
		t.Errorf("ranges = %+v", q.Ranges)
	}
	if q.Size != 5 || q.Cursor != "c" || !q.Sort[0].Desc || q.Facets[0] != "status" {
		t.Errorf("query = %+v", q)
	}
}
