package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/search-service/internal/domain"
)

type fakeSearcher struct {
	query  domain.DocumentQuery
	calls  int
	fields []domain.FieldInfo
}

func (f *fakeSearcher) Query(_ context.Context, q domain.DocumentQuery, _ domain.QueryMode) (domain.DocumentPage, error) {
	f.query = q
	f.calls++
	return domain.DocumentPage{Items: []domain.Document{{Index: "users", ID: "1", Source: map[string]any{"name": "a", "password": "x"}}}}, nil
}

func (f *fakeSearcher) Autocomplete(context.Context, domain.AutocompleteQuery) ([]domain.Document, error) {
	return []domain.Document{{Source: map[string]any{"api_token": "t", "name": "b"}}}, nil
}

func (f *fakeSearcher) Get(context.Context, domain.DocumentRef) (domain.Document, error) {
	return domain.Document{Source: map[string]any{"secret": "s", "id": 1}}, nil
}

func (f *fakeSearcher) Similar(context.Context, domain.DocumentRef, int) ([]domain.Document, error) {
	return nil, nil
}

func (f *fakeSearcher) Indices(context.Context) ([]domain.IndexInfo, error) { return nil, nil }

func (f *fakeSearcher) Fields(context.Context, string) ([]domain.FieldInfo, error) {
	return f.fields, nil
}

func newDocuments(t *testing.T, s *fakeSearcher) *DocumentService {
	t.Helper()
	p, err := domain.NewSensitivePolicy("")
	if err != nil {
		t.Fatal(err)
	}
	return NewDocumentService(s, p)
}

func TestSearchNormalizesAndRedacts(t *testing.T) {
	s := &fakeSearcher{}
	svc := newDocuments(t, s)

	page, err := svc.Search(context.Background(), domain.DocumentQuery{Keyword: " iphone "})
	if err != nil {
		t.Fatal(err)
	}
	if s.query.Keyword != "iphone" || s.query.Size != domain.DefaultPageSize {
		t.Errorf("query not normalized: %+v", s.query)
	}
	if _, ok := page.Items[0].Source["password"]; ok {
		t.Error("password must be redacted")
	}

	_, err = svc.Search(context.Background(), domain.DocumentQuery{Keyword: "x", Sort: domain.ParseSort("password")})
	if !domain.IsInvalid(err) || s.calls != 1 {
		t.Errorf("sorting on a sensitive field must be rejected before reaching the port, err=%v calls=%d", err, s.calls)
	}
	if _, err := svc.Search(context.Background(), domain.DocumentQuery{}); !domain.IsInvalid(err) {
		t.Errorf("empty keyword: %v", err)
	}
	if _, err := svc.List(context.Background(), domain.DocumentQuery{}); err != nil {
		t.Errorf("list without keyword: %v", err)
	}

	docs, _ := svc.Autocomplete(context.Background(), domain.AutocompleteQuery{Keyword: "b"})
	if _, ok := docs[0].Source["api_token"]; ok {
		t.Error("autocomplete must redact")
	}
	doc, _ := svc.Get(context.Background(), domain.DocumentRef{Index: "users", ID: "1"})
	if _, ok := doc.Source["secret"]; ok {
		t.Error("get must redact")
	}
}

func TestFieldsHidesSensitive(t *testing.T) {
	s := &fakeSearcher{fields: []domain.FieldInfo{{Name: "name"}, {Name: "password_hash"}, {Name: "profile.api_key"}}}
	fields, err := newDocuments(t, s).Fields(context.Background(), "users")
	if err != nil {
		t.Fatal(err)
	}
	if len(fields) != 1 || fields[0].Name != "name" {
		t.Errorf("fields = %+v", fields)
	}
}

type fakeSchema struct {
	reindexed []string
	failApply int
}

func (f *fakeSchema) Apply(context.Context) error {
	if f.failApply > 0 {
		f.failApply--
		return errors.New("es down")
	}
	return nil
}

func (f *fakeSchema) Status(context.Context) (domain.SchemaStatus, error) {
	return domain.SchemaStatus{Applied: true, Indices: []domain.IndexInfo{
		{Name: "accounts", Managed: false},
		{Name: "products", Managed: true},
		{Name: "baskets", Managed: true, WriteBlocked: true},
	}}, nil
}

func (f *fakeSchema) Reindex(_ context.Context, name string) (domain.ReindexResult, error) {
	f.reindexed = append(f.reindexed, name)
	if name == "accounts" {
		return domain.ReindexResult{}, errors.New("boom")
	}
	return domain.ReindexResult{Name: name}, nil
}

func TestReindexUnmanagedAndBlocked(t *testing.T) {
	f := &fakeSchema{}
	out, err := NewSchemaService(f).Reindex(context.Background(), nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if len(f.reindexed) != 2 || f.reindexed[0] != "accounts" || f.reindexed[1] != "baskets" {
		t.Errorf("reindexed = %v", f.reindexed)
	}
	if out[0].Error != "boom" || out[0].Name != "accounts" || out[1].Error != "" {
		t.Errorf("results = %+v", out)
	}
	if _, err := NewSchemaService(f).Reindex(context.Background(), nil, false); !domain.IsInvalid(err) {
		t.Errorf("no index: %v", err)
	}
	if _, err := NewSchemaService(f).Reindex(context.Background(), []string{"acc*"}, false); !domain.IsInvalid(err) {
		t.Errorf("wildcard: %v", err)
	}
}

func TestApplyWithRetryStopsOnCancel(t *testing.T) {
	f := &fakeSchema{failApply: 100}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	NewSchemaService(f).ApplyWithRetry(ctx)
}
