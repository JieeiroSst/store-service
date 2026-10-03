package main

import (
	"bytes"
	"fmt"
	"go/format"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/99designs/gqlgen/codegen/templates"
	"gopkg.in/yaml.v3"
)

const module = "github.com/JIeeiroSst/filter-service"

type Arg struct {
	Key  string
	Name string
	Type string
}

type Endpoint struct {
	Field   string    `yaml:"field"`
	Path    string    `yaml:"path"`
	Params  yaml.Node `yaml:"params"`
	Query   yaml.Node `yaml:"query"`
	Headers yaml.Node `yaml:"headers"`
	Returns string    `yaml:"returns"`
	Unwrap  *string   `yaml:"unwrap"`
	Doc     string    `yaml:"doc"`
	Raw     []string  `yaml:"rawParams"`

	pathArgs   []Arg
	queryArgs  []Arg
	headerArgs []Arg
}

func (e *Endpoint) args() []Arg {
	return append(append(append([]Arg{}, e.pathArgs...), e.queryArgs...), e.headerArgs...)
}

type Spec struct {
	Service   string     `yaml:"service"`
	Source    string     `yaml:"source"`
	Namespace string     `yaml:"namespace"`
	Prefix    string     `yaml:"prefix"`
	BaseURL   string     `yaml:"baseURL"`
	Unwrap    string     `yaml:"unwrap"`
	Forward   []string   `yaml:"forwardHeaders"`
	Endpoints []Endpoint `yaml:"endpoints"`
	Types     string     `yaml:"types"`

	file    string
	pkg     string
	objects map[string]bool
}

var (
	scalars    = map[string]bool{"String": true, "Int": true, "Float": true, "Boolean": true, "ID": true, "Map": true, "Any": true}
	identRe    = regexp.MustCompile(`^[_A-Za-z][_0-9A-Za-z]*$`)
	typeDeclRe = regexp.MustCompile(`(?m)^\s*type\s+([_A-Za-z][_0-9A-Za-z]*)`)
	pathVarRe  = regexp.MustCompile(`\{([^}]+)\}`)
	goKeywords = map[string]bool{"break": true, "case": true, "chan": true, "const": true, "continue": true, "default": true,
		"defer": true, "else": true, "fallthrough": true, "for": true, "func": true, "go": true, "goto": true, "if": true,
		"import": true, "interface": true, "map": true, "package": true, "range": true, "return": true, "select": true,
		"struct": true, "switch": true, "type": true, "var": true, "ctx": true, "obj": true, "q": true, "out": true,
		"path": true, "err": true, "url": true, "r": true, "c": true, "strconv": true, "fmt": true, "model": true}
)

func main() {
	root := "."
	if len(os.Args) > 1 {
		root = os.Args[1]
	}
	files, err := filepath.Glob(filepath.Join(root, "spec", "*.yaml"))
	check(err)
	sort.Strings(files)

	var specs []*Spec
	seenNS := map[string]string{}
	seenPrefix := map[string]string{}
	seenType := map[string]string{}
	for _, f := range files {
		s := load(f)
		if other, ok := seenNS[s.Namespace]; ok {
			fail("%s: namespace %q already used by %s", f, s.Namespace, other)
		}
		if other, ok := seenPrefix[s.Prefix]; ok {
			fail("%s: prefix %q already used by %s", f, s.Prefix, other)
		}
		seenNS[s.Namespace], seenPrefix[s.Prefix] = f, f
		for t := range s.objects {
			if other, ok := seenType[t]; ok {
				fail("%s: type %q already defined in %s", f, t, other)
			}
			seenType[t] = f
		}
		specs = append(specs, s)
	}

	for _, s := range specs {
		write(filepath.Join(root, "gql", "schema", s.pkg+".graphqls"), schemaFile(s))
		writeGo(filepath.Join(root, "adapter", s.pkg, s.pkg+".go"), clientFile(s))
		writeGo(filepath.Join(root, "gql", "resolver", s.pkg+".go"), resolverFile(s))
	}
	writeGo(filepath.Join(root, "adapter", "clients_gen.go"), registryFile(specs))
	writeGo(filepath.Join(root, "gql", "model", "namespaces_gen.go"), namespacesFile(specs))
	writeGo(filepath.Join(root, "gql", "resolver", "query_gen.go"), rootResolverFile(specs))

	n := 0
	for _, s := range specs {
		n += len(s.Endpoints)
	}
	fmt.Printf("generated %d services, %d GET endpoints\n", len(specs), n)
}

func load(file string) *Spec {
	data, err := os.ReadFile(file)
	check(err)
	s := &Spec{file: file}
	if err := yaml.Unmarshal(data, s); err != nil {
		fail("%s: %v", file, err)
	}
	if s.Service == "" || s.Namespace == "" || s.Prefix == "" || s.BaseURL == "" {
		fail("%s: service, namespace, prefix and baseURL are required", file)
	}
	s.pkg = strings.NewReplacer("-", "_", ".", "_").Replace(s.Service)
	s.objects = map[string]bool{}
	for _, m := range typeDeclRe.FindAllStringSubmatch(s.Types, -1) {
		if !strings.HasPrefix(m[1], s.Prefix) {
			fail("%s: type %s must start with prefix %s", file, m[1], s.Prefix)
		}
		s.objects[m[1]] = true
	}

	seen := map[string]bool{}
	for i := range s.Endpoints {
		e := &s.Endpoints[i]
		if !identRe.MatchString(e.Field) || seen[e.Field] {
			fail("%s: bad or duplicate field %q", file, e.Field)
		}
		seen[e.Field] = true
		if !strings.HasPrefix(e.Path, "/") || e.Returns == "" {
			fail("%s: %s needs path and returns", file, e.Field)
		}
		params := pairs(file, &e.Params)
		for _, m := range pathVarRe.FindAllStringSubmatch(e.Path, -1) {
			t, ok := params[m[1]]
			if !ok {
				t = "String!"
			}
			if !strings.HasSuffix(t, "!") {
				fail("%s: %s path param %s must be non-null", file, e.Field, m[1])
			}
			e.pathArgs = append(e.pathArgs, Arg{Key: m[1], Name: argName(m[1]), Type: t})
		}
		for _, kv := range orderedPairs(file, &e.Query) {
			e.queryArgs = append(e.queryArgs, Arg{Key: kv[0], Name: argName(kv[0]), Type: kv[1]})
		}
		for _, kv := range orderedPairs(file, &e.Headers) {
			e.headerArgs = append(e.headerArgs, Arg{Key: kv[0], Name: argName(kv[0]), Type: kv[1]})
		}
		names := map[string]bool{}
		for _, a := range e.args() {
			if names[a.Name] {
				fail("%s: %s has duplicate argument %s", file, e.Field, a.Name)
			}
			names[a.Name] = true
			checkType(s, a.Type, true)
		}
		checkType(s, e.Returns, false)
	}
	return s
}

func pairs(file string, n *yaml.Node) map[string]string {
	out := map[string]string{}
	for _, kv := range orderedPairs(file, n) {
		out[kv[0]] = kv[1]
	}
	return out
}

func orderedPairs(file string, n *yaml.Node) [][2]string {
	if n.Kind == 0 {
		return nil
	}
	if n.Kind != yaml.MappingNode {
		fail("%s: line %d: expected a map of name: Type", file, n.Line)
	}
	var out [][2]string
	for i := 0; i+1 < len(n.Content); i += 2 {
		out = append(out, [2]string{n.Content[i].Value, n.Content[i+1].Value})
	}
	return out
}

func argName(key string) string {
	if identRe.MatchString(key) {
		return key
	}
	parts := regexp.MustCompile(`[^0-9A-Za-z]+`).Split(key, -1)
	var b strings.Builder
	for i, p := range parts {
		if p == "" {
			continue
		}
		if i == 0 || b.Len() == 0 {
			b.WriteString(strings.ToLower(p[:1]) + p[1:])
		} else {
			b.WriteString(strings.ToUpper(p[:1]) + p[1:])
		}
	}
	name := b.String()
	if name == "" || !identRe.MatchString(name) {
		name = "arg_" + name
	}
	return name
}

type gqlType struct {
	list      bool
	elemNull  bool
	null      bool
	name      string
}

func parseType(t string) gqlType {
	t = strings.TrimSpace(t)
	var g gqlType
	g.null = !strings.HasSuffix(t, "!")
	t = strings.TrimSuffix(t, "!")
	if strings.HasPrefix(t, "[") {
		g.list = true
		t = strings.TrimSuffix(strings.TrimPrefix(t, "["), "]")
		g.elemNull = !strings.HasSuffix(t, "!")
		t = strings.TrimSuffix(t, "!")
	}
	g.name = t
	return g
}

func checkType(s *Spec, t string, input bool) {
	g := parseType(t)
	if scalars[g.name] {
		return
	}
	if input {
		fail("%s: argument type %s must be a scalar", s.file, t)
	}
	if !s.objects[g.name] {
		fail("%s: unknown type %s", s.file, g.name)
	}
}

func goScalar(name string) string {
	switch name {
	case "String", "ID":
		return "string"
	case "Int":
		return "int"
	case "Float":
		return "float64"
	case "Boolean":
		return "bool"
	case "Map":
		return "map[string]interface{}"
	case "Any":
		return "interface{}"
	}
	return ""
}

func goType(t string) string {
	g := parseType(t)
	if s := goScalar(g.name); s != "" {
		elem := s
		if g.list {
			if g.elemNull && g.name != "Map" && g.name != "Any" {
				elem = "*" + s
			}
			return "[]" + elem
		}
		if g.null && g.name != "Map" && g.name != "Any" {
			return "*" + s
		}
		return s
	}
	if g.list {
		return "[]*model." + templates.ToGo(g.name)
	}
	return "*model." + templates.ToGo(g.name)
}

func goParam(name string) string {
	p := templates.ToGoPrivate(name)
	if goKeywords[p] {
		p += "Arg"
	}
	return p
}

func gqlArgs(e *Endpoint) string {
	var args []string
	for _, a := range e.args() {
		args = append(args, a.Name+": "+a.Type)
	}
	if len(args) == 0 {
		return ""
	}
	return "(" + strings.Join(args, ", ") + ")"
}

func schemaFile(s *Spec) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "extend type Query {\n  \"\"\"GET endpoints of %s\"\"\"\n  %s: %sQuery!\n}\n\n", s.Service, s.Namespace, s.Prefix)
	fmt.Fprintf(&b, "type %sQuery {\n", s.Prefix)
	for i, e := range s.Endpoints {
		if i > 0 {
			b.WriteString("\n")
		}
		desc := "GET " + e.Path
		if e.Doc != "" {
			desc += "\n\n" + strings.TrimSpace(e.Doc)
		}
		fmt.Fprintf(&b, "  \"\"\"\n  %s\n  \"\"\"\n", strings.ReplaceAll(desc, "\n", "\n  "))
		fmt.Fprintf(&b, "  %s%s: %s\n", e.Field, gqlArgs(&e), e.Returns)
	}
	b.WriteString("}\n\n")
	b.WriteString(strings.TrimSpace(s.Types))
	b.WriteString("\n")
	return b.Bytes()
}

func unwrapOf(s *Spec, e *Endpoint) string {
	if e.Unwrap != nil {
		return *e.Unwrap
	}
	return s.Unwrap
}

func clientFile(s *Spec) []byte {
	var b bytes.Buffer
	fmt.Fprintf(&b, "package %s\n\nimport (\n\"context\"\n\"net/http\"\n\"net/url\"\n\"strconv\"\n\n\"%s/adapter/rest\"\n\"%s/gql/model\"\n)\n\n", s.pkg, module, module)
	b.WriteString("var _ = strconv.Itoa\n\n")
	fmt.Fprintf(&b, "const Service = %q\n\nconst DefaultBaseURL = %q\n\n", s.Service, s.BaseURL)
	b.WriteString("type Client struct{ rest *rest.Client }\n\nfunc New(c *rest.Client) *Client { return &Client{rest: c} }\n\n")
	for _, e := range s.Endpoints {
		var params []string
		for _, a := range e.args() {
			params = append(params, goParam(a.Name)+" "+goType(a.Type))
		}
		ret := goType(e.Returns)
		fmt.Fprintf(&b, "func (c *Client) %s(ctx context.Context%s) (%s, error) {\n", templates.ToGo(e.Field), prefixComma(params), ret)
		b.WriteString("path := " + pathExpr(&e) + "\n")
		b.WriteString("q := url.Values{}\n")
		for _, a := range e.queryArgs {
			b.WriteString(queryStmt(a))
		}
		b.WriteString("h := http.Header{}\n")
		for _, a := range e.headerArgs {
			b.WriteString(strings.ReplaceAll(strings.ReplaceAll(queryStmt(a), "q.Set(", "h.Set("), "q.Add(", "h.Add("))
		}
		fmt.Fprintf(&b, "var out %s\nerr := c.rest.Get(ctx, path, q, h, %q, &out)\nreturn out, err\n}\n\n", ret, unwrapOf(s, &e))
	}
	return b.Bytes()
}

func prefixComma(params []string) string {
	if len(params) == 0 {
		return ""
	}
	return ", " + strings.Join(params, ", ")
}

func pathExpr(e *Endpoint) string {
	var parts []string
	rest := e.Path
	for _, a := range e.pathArgs {
		i := strings.Index(rest, "{"+a.Key+"}")
		if i > 0 {
			parts = append(parts, fmt.Sprintf("%q", rest[:i]))
		}
		esc := "url.PathEscape("
		for _, r := range e.Raw {
			if r == a.Key {
				esc = "rest.EscapePath("
			}
		}
		parts = append(parts, esc+toString(goParam(a.Name), parseType(a.Type).name)+")")
		rest = rest[i+len(a.Key)+2:]
	}
	if rest != "" || len(parts) == 0 {
		parts = append(parts, fmt.Sprintf("%q", rest))
	}
	return strings.Join(parts, " + ")
}

func toString(v, scalar string) string {
	switch scalar {
	case "Int":
		return "strconv.Itoa(" + v + ")"
	case "Float":
		return "strconv.FormatFloat(" + v + ", 'f', -1, 64)"
	case "Boolean":
		return "strconv.FormatBool(" + v + ")"
	}
	return v
}

func queryStmt(a Arg) string {
	g := parseType(a.Type)
	v := goParam(a.Name)
	if g.list {
		elem := "e"
		if g.elemNull {
			return fmt.Sprintf("for _, e := range %s {\nif e != nil {\nq.Add(%q, %s)\n}\n}\n", v, a.Key, toString("*e", g.name))
		}
		return fmt.Sprintf("for _, e := range %s {\nq.Add(%q, %s)\n}\n", v, a.Key, toString(elem, g.name))
	}
	if g.null {
		return fmt.Sprintf("if %s != nil {\nq.Set(%q, %s)\n}\n", v, a.Key, toString("*"+v, g.name))
	}
	return fmt.Sprintf("q.Set(%q, %s)\n", a.Key, toString(v, g.name))
}

func resolverFile(s *Spec) []byte {
	var b bytes.Buffer
	recv := templates.ToGoPrivate(s.Prefix) + "QueryResolver"
	fmt.Fprintf(&b, "package resolver\n\nimport (\n\"context\"\n\n\"%s/gql/generated\"\n\"%s/gql/model\"\n)\n\n", module, module)
	fmt.Fprintf(&b, "func (r *Resolver) %sQuery() generated.%sQueryResolver { return &%s{r} }\n\n", s.Prefix, s.Prefix, recv)
	fmt.Fprintf(&b, "type %s struct{ *Resolver }\n\n", recv)
	for _, e := range s.Endpoints {
		var params, names []string
		for _, a := range e.args() {
			params = append(params, goParam(a.Name)+" "+goType(a.Type))
			names = append(names, goParam(a.Name))
		}
		fmt.Fprintf(&b, "func (r *%s) %s(ctx context.Context, obj *model.%s%s) (%s, error) {\n", recv, templates.ToGo(e.Field), templates.ToGo(s.Prefix+"Query"), prefixComma(params), goType(e.Returns))
		fmt.Fprintf(&b, "return r.Clients.%s.%s(ctx%s)\n}\n\n", templates.ToGo(s.pkg), templates.ToGo(e.Field), prefixComma(names))
	}
	return b.Bytes()
}

func registryFile(specs []*Spec) []byte {
	var b bytes.Buffer
	b.WriteString("package adapter\n\nimport (\n\"net/http\"\n\"os\"\n\n")
	fmt.Fprintf(&b, "\"%s/adapter/rest\"\n", module)
	for _, s := range specs {
		fmt.Fprintf(&b, "\"%s/adapter/%s\"\n", module, s.pkg)
	}
	b.WriteString(")\n\ntype ServiceInfo struct {\nName string\nField string\nEnvVar string\nBaseURL string\nEndpoints int\n}\n\n")
	b.WriteString("type Clients struct {\nServices []ServiceInfo\n")
	for _, s := range specs {
		fmt.Fprintf(&b, "%s *%s.Client\n", templates.ToGo(s.pkg), s.pkg)
	}
	b.WriteString("}\n\n")
	b.WriteString("func baseURL(env, fallback string) string {\nif v := os.Getenv(env); v != \"\" {\nreturn v\n}\nreturn fallback\n}\n\n")
	b.WriteString("func NewClients(httpClient *http.Client) *Clients {\nc := &Clients{}\n")
	for _, s := range specs {
		env := strings.ToUpper(s.pkg) + "_URL"
		fwd := ""
		for _, h := range s.Forward {
			fwd += fmt.Sprintf(", %q", h)
		}
		fmt.Fprintf(&b, "{\nu := baseURL(%q, %s.DefaultBaseURL)\nc.%s = %s.New(rest.New(%s.Service, u, httpClient%s))\n", env, s.pkg, templates.ToGo(s.pkg), s.pkg, s.pkg, fwd)
		fmt.Fprintf(&b, "c.Services = append(c.Services, ServiceInfo{Name: %s.Service, Field: %q, EnvVar: %q, BaseURL: u, Endpoints: %d})\n}\n", s.pkg, s.Namespace, env, len(s.Endpoints))
	}
	b.WriteString("return c\n}\n")
	return b.Bytes()
}

func namespacesFile(specs []*Spec) []byte {
	var b bytes.Buffer
	b.WriteString("package model\n\n")
	for _, s := range specs {
		fmt.Fprintf(&b, "type %s struct{}\n\n", templates.ToGo(s.Prefix+"Query"))
	}
	return b.Bytes()
}

func rootResolverFile(specs []*Spec) []byte {
	var b bytes.Buffer
	b.WriteString("package resolver\n\nimport (\n\"context\"\n\n")
	fmt.Fprintf(&b, "\"%s/gql/generated\"\n\"%s/gql/model\"\n)\n\n", module, module)
	b.WriteString("func (r *Resolver) Query() generated.QueryResolver { return &queryResolver{r} }\n\ntype queryResolver struct{ *Resolver }\n\n")
	b.WriteString("func (r *queryResolver) Services(ctx context.Context) ([]*model.GatewayService, error) {\nout := make([]*model.GatewayService, 0, len(r.Clients.Services))\n")
	b.WriteString("for _, s := range r.Clients.Services {\nout = append(out, &model.GatewayService{Name: s.Name, Field: s.Field, EnvVar: s.EnvVar, BaseURL: s.BaseURL, Endpoints: s.Endpoints})\n}\nreturn out, nil\n}\n\n")
	for _, s := range specs {
		ns := templates.ToGo(s.Prefix + "Query")
		fmt.Fprintf(&b, "func (r *queryResolver) %s(ctx context.Context) (*model.%s, error) {\nreturn &model.%s{}, nil\n}\n\n", templates.ToGo(s.Namespace), ns, ns)
	}
	return b.Bytes()
}

func write(path string, data []byte) {
	check(os.MkdirAll(filepath.Dir(path), 0o755))
	check(os.WriteFile(path, data, 0o644))
}

func writeGo(path string, src []byte) {
	out, err := format.Source(src)
	if err != nil {
		fail("%s: %v\n%s", path, err, src)
	}
	write(path, out)
}

func check(err error) {
	if err != nil {
		fail("%v", err)
	}
}

func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "gen: "+format+"\n", args...)
	os.Exit(1)
}
