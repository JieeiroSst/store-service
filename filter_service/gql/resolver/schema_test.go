package resolver

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sort"
	"strings"
	"sync"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/extension"
	"github.com/99designs/gqlgen/graphql/handler/transport"
	"github.com/JIeeiroSst/filter-service/adapter"
	"github.com/JIeeiroSst/filter-service/adapter/rest"
	"github.com/JIeeiroSst/filter-service/gql/generated"
)

type typeRef struct {
	Kind   string   `json:"kind"`
	Name   string   `json:"name"`
	OfType *typeRef `json:"ofType"`
}

func (t typeRef) named() typeRef {
	for t.OfType != nil {
		t = *t.OfType
	}
	return t
}

type field struct {
	Name string   `json:"name"`
	Type typeRef  `json:"type"`
	Args []argDef `json:"args"`
}

type argDef struct {
	Name string  `json:"name"`
	Type typeRef `json:"type"`
}

const typeQuery = `query($n: String!) { __type(name: $n) { fields { name args { name type { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } type { kind name ofType { kind name ofType { kind name ofType { kind name } } } } } } }`

type upstream struct {
	mu    sync.Mutex
	paths []string
}

func (u *upstream) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	u.mu.Lock()
	u.paths = append(u.paths, r.URL.EscapedPath())
	u.mu.Unlock()
	w.Write([]byte("null"))
}

func newTestServer(t *testing.T, baseURL string) http.Handler {
	t.Helper()
	for _, s := range adapter.NewClients(nil).Services {
		t.Setenv(s.EnvVar, baseURL)
	}
	srv := handler.New(generated.NewExecutableSchema(generated.Config{Resolvers: NewResolver(adapter.NewClients(nil))}))
	srv.AddTransport(transport.POST{})
	srv.Use(extension.Introspection{})
	return rest.ForwardHeaders(srv)
}

func run(t *testing.T, h http.Handler, query string, vars map[string]any) (json.RawMessage, []any) {
	t.Helper()
	body, _ := json.Marshal(map[string]any{"query": query, "variables": vars})
	req := httptest.NewRequest(http.MethodPost, "/query", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	var resp struct {
		Data   json.RawMessage `json:"data"`
		Errors []any           `json:"errors"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("%s: %v: %s", query, err, rec.Body.String())
	}
	return resp.Data, resp.Errors
}

func fieldsOf(t *testing.T, h http.Handler, typeName string) []field {
	t.Helper()
	data, errs := run(t, h, typeQuery, map[string]any{"n": typeName})
	if len(errs) > 0 {
		t.Fatalf("introspect %s: %v", typeName, errs)
	}
	var out struct {
		Type struct {
			Fields []field `json:"fields"`
		} `json:"__type"`
	}
	json.Unmarshal(data, &out)
	return out.Type.Fields
}

func dummy(t typeRef) string {
	if t.Kind == "NON_NULL" {
		return dummy(*t.OfType)
	}
	if t.Kind == "LIST" {
		return "[" + dummy(*t.OfType) + "]"
	}
	switch t.Name {
	case "Int":
		return "7"
	case "Float":
		return "1.5"
	case "Boolean":
		return "true"
	}
	return `"a b/c"`
}

// TestEveryEndpointIsWired runs every query field of every service against a
// stub upstream and checks that each one resolves without error and issues
// exactly one GET with all path parameters filled in.
func TestEveryEndpointIsWired(t *testing.T) {
	up := &upstream{}
	stub := httptest.NewServer(up)
	defer stub.Close()
	h := newTestServer(t, stub.URL)

	total := 0
	for _, ns := range fieldsOf(t, h, "Query") {
		if ns.Name == "services" {
			continue
		}
		nsType := ns.Type.named().Name
		for _, f := range fieldsOf(t, h, nsType) {
			var args []string
			for _, a := range f.Args {
				args = append(args, a.Name+": "+dummy(a.Type))
			}
			call := f.Name
			if len(args) > 0 {
				call += "(" + strings.Join(args, ", ") + ")"
			}
			if f.Type.named().Kind == "OBJECT" {
				call += " { __typename }"
			}
			query := fmt.Sprintf("{ %s { %s } }", ns.Name, call)

			before := len(up.paths)
			if _, errs := run(t, h, query, nil); len(errs) > 0 {
				t.Errorf("%s.%s: %v", ns.Name, f.Name, errs)
				continue
			}
			if got := len(up.paths) - before; got != 1 {
				t.Errorf("%s.%s made %d upstream requests, want 1", ns.Name, f.Name, got)
				continue
			}
			if p := up.paths[len(up.paths)-1]; strings.ContainsAny(p, "{}") || strings.Contains(p, "%7B") {
				t.Errorf("%s.%s left a path template unfilled: %s", ns.Name, f.Name, p)
			}
			total++
		}
	}
	if total < 600 {
		t.Fatalf("only %d endpoints exercised", total)
	}
	t.Logf("%d endpoints exercised", total)
}

func TestServicesQuery(t *testing.T) {
	h := newTestServer(t, "http://upstream.test")
	data, errs := run(t, h, `{ services { name field envVar baseURL endpoints } }`, nil)
	if len(errs) > 0 {
		t.Fatal(errs)
	}
	var out struct {
		Services []struct {
			Name, Field, EnvVar, BaseURL string
			Endpoints                    int
		}
	}
	json.Unmarshal(data, &out)
	names := map[string]bool{}
	for _, s := range out.Services {
		if s.BaseURL != "http://upstream.test" || s.Endpoints == 0 || names[s.Field] {
			t.Errorf("bad service entry %+v", s)
		}
		names[s.Field] = true
	}
	if len(out.Services) < 80 {
		t.Fatalf("only %d services", len(out.Services))
	}
}

func TestPathParamsAreEscaped(t *testing.T) {
	up := &upstream{}
	stub := httptest.NewServer(up)
	defer stub.Close()
	h := newTestServer(t, stub.URL)

	cases := map[string]string{
		`{ webrtcService { room(room_id: "a/b c") { room_id } } }`:                       "/api/rooms/a%2Fb%20c",
		`{ manageService { groupByPath(realm: "r", path: "/top/sub") { __typename } } }`: "/api/v1/admin/realms/r/group-by-path/top/sub",
	}
	keys := make([]string, 0, len(cases))
	for k := range cases {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, q := range keys {
		if _, errs := run(t, h, q, nil); len(errs) > 0 {
			t.Fatalf("%s: %v", q, errs)
		}
	}
	// A normal path parameter is one escaped segment; the wildcard keeps its
	// slashes.
	got := map[string]bool{}
	for _, p := range up.paths {
		got[p] = true
	}
	for _, want := range cases {
		if !got[want] {
			t.Errorf("missing upstream path %q in %v", want, up.paths)
		}
	}
}
