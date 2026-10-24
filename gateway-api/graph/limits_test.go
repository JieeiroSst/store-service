package graph

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/99designs/gqlgen/graphql/handler"
	"github.com/99designs/gqlgen/graphql/handler/transport"
)

func limitedServer() *handler.Server {
	srv := handler.New(NewExecutableSchema(Config{Resolvers: &Resolver{}}))
	srv.AddTransport(transport.POST{})
	srv.AroundOperations(LimitRootFields)
	return srv
}

func postQuery(t *testing.T, srv http.Handler, query string) string {
	t.Helper()
	body := `{"query":` + strconvQuote(query) + `}`
	req := httptest.NewRequest(http.MethodPost, "/query", strings.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	srv.ServeHTTP(w, req)
	return w.Body.String()
}

func strconvQuote(s string) string {
	return `"` + strings.ReplaceAll(strings.ReplaceAll(s, `\`, `\\`), `"`, `\"`) + `"`
}

func aliased(n int) string {
	var b strings.Builder
	b.WriteString("{")
	for i := 0; i < n; i++ {
		b.WriteString(" a")
		b.WriteString(strings.Repeat("x", i+1))
		b.WriteString(": __typename")
	}
	b.WriteString(" }")
	return b.String()
}

func TestLimitRootFields(t *testing.T) {
	srv := limitedServer()
	if out := postQuery(t, srv, aliased(MaxRootFields)); strings.Contains(out, "root fields") {
		t.Fatalf("query at the limit rejected: %s", out)
	}
	if out := postQuery(t, srv, aliased(MaxRootFields+1)); !strings.Contains(out, "root fields") {
		t.Fatalf("query over the limit accepted: %s", out)
	}
	frag := "query { ...F } fragment F on Query {" + strings.TrimSuffix(strings.TrimPrefix(aliased(MaxRootFields+1), "{"), "}") + "}"
	if out := postQuery(t, srv, frag); !strings.Contains(out, "root fields") {
		t.Fatalf("fields hidden in a fragment were not counted: %s", out)
	}
}
