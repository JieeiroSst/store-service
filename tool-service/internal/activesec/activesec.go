package activesec

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"
)

const (
	High   = "high"
	Medium = "medium"
	Low    = "low"
)

type Finding struct {
	ID       string `json:"id"`
	Severity string `json:"severity"`
	Title    string `json:"title"`
	Endpoint string `json:"endpoint,omitempty"`
	Detail   string `json:"detail,omitempty"`
}

type Endpoint struct {
	Method  string            `json:"method"`
	Path    string            `json:"path"`
	Params  map[string]string `json:"params,omitempty"`
	Body    map[string]any    `json:"body,omitempty"`
	Headers map[string]string `json:"headers,omitempty"`
}

type Identity struct {
	Name    string            `json:"name"`
	Headers map[string]string `json:"headers"`
}

type Resource struct {
	Method string `json:"method"`
	Path   string `json:"path"`
	Owner  string `json:"owner"`
}

type Form struct {
	Method string
	Action string
	Fields []FormField
}

type FormField struct{ Name, Type, Value string }

type Options struct {
	Endpoints  []Endpoint `json:"endpoints"`
	Identities []Identity `json:"identities"`
	Resources  []Resource `json:"resources"`
	Forms      []Form     `json:"-"`

	IncludeWrites bool `json:"include_writes"`
	MaxRequests   int  `json:"max_requests"`
}

type Result struct {
	Findings []Finding `json:"findings"`
	Requests int       `json:"requests_sent"`
	Skipped  []string  `json:"skipped,omitempty"`
}

var (
	sqlErrors = regexp.MustCompile(`(?i)(you have an error in your sql syntax|sqlstate\[|sqlstate |unterminated quoted string|syntax error at or near|ora-\d{5}|mysql_fetch|pg_query|sqlite3?[._ ]error|unclosed quotation mark|odbc (sql )?driver|psycopg2|org\.postgresql|com\.mysql\.jdbc|SQLException)`)
	traces    = regexp.MustCompile(`(?i)(traceback \(most recent|goroutine \d+ \[|panic: |at java\.|at [a-z0-9_.]+\(.*\.(java|cs|kt):\d+\)|exception in thread|stack trace:|unhandled exception|node_modules/|\bat Object\.<anonymous>|fatal error:)`)
	moneyName = regexp.MustCompile(`(?i)(amount|price|total|fee|cost|balance|qty|quantity|charge|tax|discount)`)
	pathName  = regexp.MustCompile(`(?i)(file|path|page|template|doc|dir|name|src|include|load)`)
)

type runner struct {
	c      *http.Client
	base   string
	budget int
	sent   int
	res    *Result
	seen   map[string]bool
}

type resp struct {
	status  int
	ctype   string
	body    string
	elapsed time.Duration
}

func (r *runner) do(ctx context.Context, method, rawURL string, hdr map[string]string, body []byte, ctype string) (*resp, bool) {
	if r.sent >= r.budget || ctx.Err() != nil {
		return nil, false
	}
	r.sent++
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, rawURL, rd)
	if err != nil {
		return nil, true
	}
	if body != nil && ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	for k, v := range hdr {
		req.Header.Set(k, v)
	}
	t0 := time.Now()
	rs, err := r.c.Do(req)
	if err != nil {
		return &resp{status: 0, elapsed: time.Since(t0), body: err.Error()}, true
	}
	defer rs.Body.Close()
	b, _ := io.ReadAll(io.LimitReader(rs.Body, 256<<10))
	return &resp{rs.StatusCode, rs.Header.Get("Content-Type"), string(b), time.Since(t0)}, true
}

func (r *runner) add(id, sev, title, endpoint, detail string) {
	k := id + "|" + endpoint + "|" + title
	if r.seen[k] {
		return
	}
	r.seen[k] = true
	r.res.Findings = append(r.res.Findings, Finding{id, sev, title, endpoint, detail})
}

func Run(ctx context.Context, c *http.Client, baseURL string, opt Options) Result {
	res := Result{}
	if opt.MaxRequests <= 0 || opt.MaxRequests > 1000 {
		opt.MaxRequests = 300
	}
	r := &runner{c: c, base: strings.TrimRight(baseURL, "/"), budget: opt.MaxRequests, res: &res, seen: map[string]bool{}}

	for _, ep := range opt.Endpoints {
		method := strings.ToUpper(orDefault(ep.Method, "GET"))
		if method != "GET" && !opt.IncludeWrites {
			res.Skipped = append(res.Skipped, fmt.Sprintf("%s %s skipped: writes not allowed (set include_writes)", method, ep.Path))
			continue
		}
		if len(ep.Params) > 0 {
			r.probeParams(ctx, method, ep)
		}
		if len(ep.Body) > 0 {
			r.fuzzBody(ctx, method, ep)
		}
	}
	r.authzMatrix(ctx, opt)
	r.probeForms(ctx, opt)
	res.Requests = r.sent
	if r.sent >= r.budget {
		res.Skipped = append(res.Skipped, fmt.Sprintf("request budget of %d reached; results may be incomplete", r.budget))
	}
	return res
}

func (r *runner) target(ep Endpoint, params map[string]string) string {
	u, _ := url.Parse(r.base + "/" + strings.TrimLeft(ep.Path, "/"))
	q := u.Query()
	for k, v := range params {
		q.Set(k, v)
	}
	u.RawQuery = q.Encode()
	return u.String()
}

func marker() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return "qa" + hex.EncodeToString(b)
}

func (r *runner) probeParams(ctx context.Context, method string, ep Endpoint) {
	label := method + " " + ep.Path
	baseResp, ok := r.do(ctx, method, r.target(ep, ep.Params), ep.Headers, nil, "")
	if !ok || baseResp == nil {
		return
	}
	for name := range ep.Params {
		with := func(v string) map[string]string {
			p := map[string]string{}
			for k, x := range ep.Params {
				p[k] = x
			}
			p[name] = v
			return p
		}

		payload := marker() + `<"'>`
		if rs, ok := r.do(ctx, method, r.target(ep, with(payload)), ep.Headers, nil, ""); ok && rs != nil &&
			strings.Contains(strings.ToLower(rs.ctype), "html") && strings.Contains(rs.body, payload) {
			r.add("ACT-XSS-001", High, "reflected input is not encoded (XSS)", label, fmt.Sprintf("parameter %q is echoed unescaped in an HTML response", name))
		}

		if rs, ok := r.do(ctx, method, r.target(ep, with(ep.Params[name]+"'")), ep.Headers, nil, ""); ok && rs != nil {
			switch {
			case sqlErrors.MatchString(rs.body):
				r.add("ACT-SQLI-001", High, "database error triggered by a quote (SQL injection likely)", label, fmt.Sprintf("parameter %q: %s", name, snippet(rs.body, sqlErrors)))
			case rs.status >= 500 && baseResp.status < 500:
				r.add("ACT-SQLI-002", Medium, "server error when a quote is added to a parameter", label, fmt.Sprintf("parameter %q: status %d (baseline %d)", name, rs.status, baseResp.status))
			case traces.MatchString(rs.body):
				r.add("ACT-ERR-001", Medium, "stack trace exposed", label, snippet(rs.body, traces))
			}
		}

		for _, t := range []string{"{{7*7}}", "${7*7}"} {
			if rs, ok := r.do(ctx, method, r.target(ep, with(t)), ep.Headers, nil, ""); ok && rs != nil &&
				strings.Contains(rs.body, "49") && !strings.Contains(baseResp.body, "49") && !strings.Contains(rs.body, t) {
				r.add("ACT-SSTI-001", Medium, "template expression was evaluated (server-side template injection likely)", label, fmt.Sprintf("parameter %q: %s produced 49", name, t))
			}
		}

		if pathName.MatchString(name) {
			if rs, ok := r.do(ctx, method, r.target(ep, with("../../../../../../etc/passwd")), ep.Headers, nil, ""); ok && rs != nil && strings.Contains(rs.body, "root:x:0:0") {
				r.add("ACT-TRAV-001", High, "path traversal: server returned a system file", label, fmt.Sprintf("parameter %q", name))
			}
		}
	}
}

type mutation struct {
	name string
	val  any
}

func (r *runner) fuzzBody(ctx context.Context, method string, ep Endpoint) {
	label := method + " " + ep.Path
	check := func(what string, rs *resp) {
		switch {
		case rs.status == 0:
			return
		case rs.elapsed > 5*time.Second:
			r.add("ACT-FUZZ-003", Medium, "request took over 5s on unusual input (possible DoS vector)", label, what)
		case rs.status >= 500:
			r.add("ACT-FUZZ-001", Medium, "server error (5xx) on malformed input; should be a 4xx validation error", label, what)
		}
		if traces.MatchString(rs.body) {
			r.add("ACT-ERR-001", Medium, "stack trace exposed", label, what)
		}
		if sqlErrors.MatchString(rs.body) {
			r.add("ACT-SQLI-001", High, "database error exposed", label, what)
		}
	}

	for _, raw := range []string{``, `{`, `null`, `[]`, `"str"`, `{"a":` + strings.Repeat("[", 500) + strings.Repeat("]", 500) + `}`} {
		rs, ok := r.do(ctx, method, r.target(ep, nil), ep.Headers, []byte(raw), "application/json")
		if !ok {
			return
		}
		if rs != nil {
			check(fmt.Sprintf("body %.30q", raw), rs)
		}
	}

	deep := any("x")
	for i := 0; i < 100; i++ {
		deep = []any{deep}
	}
	for field, orig := range ep.Body {
		muts := []mutation{
			{"null", nil}, {"empty string", ""}, {"string instead of number", "abc"}, {"number instead of string", 12345},
			{"object instead of scalar", map[string]any{"a": 1}}, {"100KB string", strings.Repeat("A", 100<<10)},
			{"quote and markup", `'"><b>qa</b>`}, {"unicode", "🔥مرحبا‮"}, {"deep nesting", deep},
			{"integer overflow", json.Number("9223372036854775808")}, {"huge exponent", json.Number("1e999")},
		}
		if isNumber(orig) || moneyName.MatchString(field) {
			muts = append(muts, mutation{"negative", -1}, mutation{"zero", 0})
		}
		for _, m := range muts {
			body := map[string]any{}
			for k, v := range ep.Body {
				body[k] = v
			}
			body[field] = m.val
			raw, err := json.Marshal(body)
			if err != nil {
				continue
			}
			rs, ok := r.do(ctx, method, r.target(ep, nil), ep.Headers, raw, "application/json")
			if !ok {
				return
			}
			if rs == nil {
				continue
			}
			what := fmt.Sprintf("field %q = %s", field, m.name)
			check(what, rs)
			if moneyName.MatchString(field) && rs.status >= 200 && rs.status < 300 {
				switch m.name {
				case "negative":
					r.add("ACT-MONEY-001", High, "negative money value accepted", label, what+" returned "+fmt.Sprint(rs.status))
				case "zero":
					r.add("ACT-MONEY-002", Medium, "zero money value accepted", label, what+" returned "+fmt.Sprint(rs.status))
				case "integer overflow", "huge exponent":
					r.add("ACT-MONEY-003", Medium, "out-of-range money value accepted", label, what+" returned "+fmt.Sprint(rs.status))
				case "string instead of number":
					r.add("ACT-MONEY-004", Medium, "non-numeric money value accepted", label, what+" returned "+fmt.Sprint(rs.status))
				}
			}
		}
	}
}

func isNumber(v any) bool {
	switch v.(type) {
	case float64, float32, int, int64, json.Number:
		return true
	}
	return false
}

func (r *runner) authzMatrix(ctx context.Context, opt Options) {
	byName := map[string]Identity{}
	for _, id := range opt.Identities {
		byName[id.Name] = id
	}
	for _, res := range opt.Resources {
		method := strings.ToUpper(orDefault(res.Method, "GET"))
		if method != "GET" && !opt.IncludeWrites {
			r.res.Skipped = append(r.res.Skipped, fmt.Sprintf("%s %s skipped: writes not allowed", method, res.Path))
			continue
		}
		owner, ok := byName[res.Owner]
		if !ok {
			r.res.Skipped = append(r.res.Skipped, fmt.Sprintf("%s: unknown owner %q", res.Path, res.Owner))
			continue
		}
		u := r.base + "/" + strings.TrimLeft(res.Path, "/")
		label := method + " " + res.Path
		own, ok := r.do(ctx, method, u, owner.Headers, nil, "")
		if !ok {
			return
		}
		if own == nil || own.status < 200 || own.status >= 300 {
			st := 0
			if own != nil {
				st = own.status
			}
			r.res.Skipped = append(r.res.Skipped, fmt.Sprintf("%s: owner %q could not read it (status %d), matrix skipped", res.Path, res.Owner, st))
			continue
		}
		others := []Identity{{Name: "anonymous"}}
		for _, id := range opt.Identities {
			if id.Name != res.Owner {
				others = append(others, id)
			}
		}
		for _, id := range others {
			rs, ok := r.do(ctx, method, u, id.Headers, nil, "")
			if !ok {
				return
			}
			if rs != nil && rs.status >= 200 && rs.status < 300 {
				title := fmt.Sprintf("broken access control: %q can access a resource owned by %q", id.Name, res.Owner)
				if id.Name == "anonymous" {
					title = fmt.Sprintf("resource owned by %q is readable without authentication", res.Owner)
				}
				r.add("ACT-AUTHZ-001", High, title, label, fmt.Sprintf("status %d", rs.status))
			}
		}
	}
}

var probeValues = []struct{ name, val string }{
	{"empty", ""}, {"long text", strings.Repeat("a", 500)}, {"markup and quotes", `'"><i>qa</i>`},
	{"negative number", "-1"}, {"overflow number", "99999999999999999999"}, {"unicode", "🔥مرحبا‮"},
}

func (r *runner) probeForms(ctx context.Context, opt Options) {
	for _, f := range opt.Forms {
		if f.Method != "GET" && !opt.IncludeWrites {
			r.res.Skipped = append(r.res.Skipped, fmt.Sprintf("%s form %s skipped: writes not allowed", f.Method, f.Action))
			continue
		}
		label := f.Method + " form " + f.Action
		for _, pv := range probeValues {
			vals := url.Values{}
			for _, fld := range f.Fields {
				v := fld.Value
				if fld.Type != "hidden" && fld.Type != "checkbox" && fld.Type != "radio" && fld.Type != "select" {
					v = pv.val
				}
				vals.Set(fld.Name, v)
			}
			var rs *resp
			var ok bool
			if f.Method == "GET" {
				u, err := url.Parse(f.Action)
				if err != nil {
					break
				}
				u.RawQuery = vals.Encode()
				rs, ok = r.do(ctx, "GET", u.String(), nil, nil, "")
			} else {
				rs, ok = r.do(ctx, f.Method, f.Action, nil, []byte(vals.Encode()), "application/x-www-form-urlencoded")
			}
			if !ok {
				return
			}
			if rs == nil {
				continue
			}
			what := "value: " + pv.name
			switch {
			case sqlErrors.MatchString(rs.body):
				r.add("ACT-SQLI-001", High, "database error exposed by form input", label, what)
			case rs.status >= 500:
				r.add("ACT-FORM-001", Medium, "server error (5xx) from form input", label, fmt.Sprintf("%s -> %d", what, rs.status))
			case traces.MatchString(rs.body):
				r.add("ACT-ERR-001", Medium, "stack trace exposed", label, what)
			case pv.name == "markup and quotes" && strings.Contains(strings.ToLower(rs.ctype), "html") && strings.Contains(rs.body, `<i>qa</i>`):
				r.add("ACT-XSS-001", High, "form input is reflected without encoding (XSS)", label, what)
			}
		}
	}
}

func snippet(body string, re *regexp.Regexp) string {
	m := re.FindString(body)
	if len(m) > 120 {
		m = m[:120]
	}
	return m
}

func orDefault(s, d string) string {
	if s == "" {
		return d
	}
	return s
}
