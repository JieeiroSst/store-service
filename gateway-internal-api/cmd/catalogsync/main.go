package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/vektah/gqlparser/v2"
	gqlast "github.com/vektah/gqlparser/v2/ast"
)

type call struct {
	Transport string `json:"transport"`
	Service   string `json:"service"`
	Method    string `json:"method"`
	Path      string `json:"path,omitempty"`
}

type manifest struct {
	Source   string            `json:"source"`
	SyncedAt string            `json:"synced_at"`
	Calls    map[string][]call `json:"calls"`
}

var rootKinds = map[string]string{
	"queryResolver":        "Query",
	"mutationResolver":     "Mutation",
	"subscriptionResolver": "Subscription",
}

func main() {
	src := flag.String("src", "../gateway-api/graph", "gateway-api graph directory")
	dst := flag.String("dst", "catalog/schema", "output directory")
	flag.Parse()

	files, err := filepath.Glob(filepath.Join(*src, "*.graphqls"))
	if err != nil || len(files) == 0 {
		log.Fatalf("no schema files in %s: %v", *src, err)
	}

	if err := os.MkdirAll(*dst, 0o755); err != nil {
		log.Fatal(err)
	}
	changed := false
	wanted := map[string]bool{}
	var sources []*gqlast.Source
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		name := filepath.Base(f)
		wanted[name] = true
		if writeIfChanged(filepath.Join(*dst, name), b) {
			changed = true
		}
		sources = append(sources, &gqlast.Source{Name: name, Input: string(b)})
	}
	existing, _ := filepath.Glob(filepath.Join(*dst, "*.graphqls"))
	for _, f := range existing {
		if wanted[filepath.Base(f)] {
			continue
		}
		if err := os.Remove(f); err != nil {
			log.Fatal(err)
		}
		changed = true
	}

	schema, gerr := gqlparser.LoadSchema(sources...)
	if gerr != nil {
		log.Fatalf("parse schema: %v", gerr)
	}

	goNames := map[string]string{}
	restFields := map[string]bool{}
	for _, root := range []*gqlast.Definition{schema.Query, schema.Mutation, schema.Subscription} {
		if root == nil {
			continue
		}
		for _, f := range root.Fields {
			if strings.HasPrefix(f.Name, "__") {
				continue
			}
			goNames[root.Name+"."+templates.ToGo(f.Name)] = root.Name + "." + f.Name
			if f.Directives.ForName("rest") != nil {
				restFields[root.Name+"."+f.Name] = true
			}
		}
	}

	goFiles, _ := filepath.Glob(filepath.Join(*src, "*.go"))
	w := &walker{methods: map[string]*ast.FuncDecl{}, funcs: map[string]*ast.FuncDecl{}, consts: map[string]string{}}
	var roots []*ast.FuncDecl
	fset := token.NewFileSet()
	for _, f := range goFiles {
		if filepath.Base(f) == "generated.go" || strings.HasSuffix(f, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fset, f, nil, 0)
		if err != nil {
			log.Fatalf("parse %s: %v", f, err)
		}
		for _, decl := range file.Decls {
			if gd, ok := decl.(*ast.GenDecl); ok && gd.Tok == token.CONST {
				for _, spec := range gd.Specs {
					vs := spec.(*ast.ValueSpec)
					for i, n := range vs.Names {
						if i < len(vs.Values) {
							if v := stringLit(vs.Values[i]); v != "" {
								w.consts[n.Name] = v
							}
						}
					}
				}
			}
			fn, ok := decl.(*ast.FuncDecl)
			if !ok || fn.Body == nil {
				continue
			}
			if fn.Recv == nil || len(fn.Recv.List) == 0 {
				w.funcs[fn.Name.Name] = fn
				continue
			}
			recv := receiverName(fn.Recv.List[0].Type)
			if _, ok := rootKinds[recv]; ok {
				roots = append(roots, fn)
			}
			if _, exists := w.methods[fn.Name.Name]; !exists || recv == "Resolver" {
				w.methods[fn.Name.Name] = fn
			}
		}
	}

	calls := map[string][]call{}
	for _, fn := range roots {
		key, ok := goNames[rootKinds[receiverName(fn.Recv.List[0].Type)]+"."+fn.Name.Name]
		if !ok || restFields[key] {
			continue
		}
		if found := w.collect(fn); len(found) > 0 {
			calls[key] = found
		}
	}

	abs, _ := filepath.Abs(*src)
	out := manifest{
		Source: filepath.ToSlash(filepath.Join(filepath.Base(filepath.Dir(abs)), filepath.Base(abs))),
		Calls:  calls,
	}
	manifestPath := filepath.Join(*dst, "manifest.json")
	var prev manifest
	if raw, err := os.ReadFile(manifestPath); err == nil && json.Unmarshal(raw, &prev) == nil {
		out.SyncedAt = prev.SyncedAt
		if !reflect.DeepEqual(prev.Calls, out.Calls) {
			changed = true
		}
	} else {
		changed = true
	}
	if changed || out.SyncedAt == "" {
		out.SyncedAt = time.Now().UTC().Format(time.RFC3339)
	}
	b, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		log.Fatal(err)
	}
	writeIfChanged(manifestPath, append(b, '\n'))
	log.Printf("synced %d schema files, %d root fields, %d resolver call maps, changed=%v", len(files), len(goNames), len(calls), changed)
}

func writeIfChanged(path string, b []byte) bool {
	if old, err := os.ReadFile(path); err == nil && bytes.Equal(old, b) {
		return false
	}
	if err := os.WriteFile(path, b, 0o644); err != nil {
		log.Fatal(err)
	}
	return true
}

func receiverName(expr ast.Expr) string {
	if star, ok := expr.(*ast.StarExpr); ok {
		expr = star.X
	}
	if id, ok := expr.(*ast.Ident); ok {
		return id.Name
	}
	return ""
}

const maxFollowDepth = 4

type walker struct {
	methods map[string]*ast.FuncDecl
	funcs   map[string]*ast.FuncDecl
	consts  map[string]string
}

func (w *walker) collect(fn *ast.FuncDecl) []call {
	seen := map[call]bool{}
	visited := map[*ast.FuncDecl]bool{}
	var out []call
	add := func(c call) {
		if c.Service == "" || c.Method == "" || seen[c] {
			return
		}
		seen[c] = true
		out = append(out, c)
	}
	var walk func(node ast.Node, depth int)
	walk = func(node ast.Node, depth int) {
		ast.Inspect(node, func(n ast.Node) bool {
			ce, ok := n.(*ast.CallExpr)
			if !ok {
				return true
			}
			name := funcIdent(ce.Fun)
			if strings.HasPrefix(name, "grpcInvoke") && len(ce.Args) >= 4 {
				add(call{Transport: "grpc", Service: selectorName(ce.Args[2]), Method: stringLit(ce.Args[3])})
				return true
			}
			if strings.HasPrefix(name, "rest") && len(ce.Args) >= 4 {
				if method, path, ok := restRequest(ce.Args[3]); ok {
					add(call{Transport: "rest", Service: w.resolve(ce.Args[2]), Method: method, Path: path})
					return true
				}
			}
			if name == "grpcCall" && len(ce.Args) >= 4 {
				svc := selectorName(ce.Args[2])
				ast.Inspect(ce.Args[3], func(m ast.Node) bool {
					if inner, ok := m.(*ast.CallExpr); ok {
						if client, method, ok := clientCall(inner); ok {
							add(call{Transport: "grpc", Service: svc, Method: "/" + client + "/" + method})
						}
					}
					return true
				})
				return false
			}
			if svc, method, ok := proxyAPICall(ce); ok {
				add(call{Transport: "grpc", Service: svc, Method: method})
				return true
			}
			if depth >= maxFollowDepth {
				return true
			}
			var target *ast.FuncDecl
			switch fun := ce.Fun.(type) {
			case *ast.SelectorExpr:
				if id, ok := fun.X.(*ast.Ident); ok && id.Name == "r" {
					target = w.methods[fun.Sel.Name]
				}
			case *ast.Ident, *ast.IndexExpr, *ast.IndexListExpr:
				target = w.funcs[name]
			}
			if target != nil && !visited[target] {
				visited[target] = true
				walk(target.Body, depth+1)
			}
			return true
		})
	}
	visited[fn] = true
	walk(fn.Body, 0)
	return out
}

func (w *walker) resolve(expr ast.Expr) string {
	if id, ok := expr.(*ast.Ident); ok {
		if v, ok := w.consts[id.Name]; ok {
			return v
		}
	}
	return selectorName(expr)
}

func restRequest(expr ast.Expr) (string, string, bool) {
	if u, ok := expr.(*ast.UnaryExpr); ok {
		expr = u.X
	}
	lit, ok := expr.(*ast.CompositeLit)
	if !ok {
		return "", "", false
	}
	method, path := "GET", ""
	for _, el := range lit.Elts {
		kv, ok := el.(*ast.KeyValueExpr)
		if !ok {
			continue
		}
		key, _ := kv.Key.(*ast.Ident)
		if key == nil {
			continue
		}
		switch key.Name {
		case "Method":
			if m := stringLit(kv.Value); m != "" {
				method = m
			}
		case "Path":
			path = pathTemplate(kv.Value)
		}
	}
	return method, path, path != ""
}

func pathTemplate(expr ast.Expr) string {
	if s := stringLit(expr); s != "" {
		return s
	}
	ce, ok := expr.(*ast.CallExpr)
	if !ok || len(ce.Args) == 0 {
		return ""
	}
	format := stringLit(ce.Args[0])
	for _, arg := range ce.Args[1:] {
		name := "param"
		if id, ok := arg.(*ast.Ident); ok {
			name = id.Name
		}
		i := strings.Index(format, "%")
		if i < 0 || i+1 >= len(format) {
			break
		}
		format = format[:i] + "{" + name + "}" + format[i+2:]
	}
	return format
}

func clientCall(ce *ast.CallExpr) (string, string, bool) {
	method, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}
	ctor, ok := method.X.(*ast.CallExpr)
	if !ok {
		return "", "", false
	}
	name := funcIdent(ctor.Fun)
	if sel, ok := ctor.Fun.(*ast.SelectorExpr); ok {
		name = sel.Sel.Name
	}
	if !strings.HasPrefix(name, "New") || !strings.HasSuffix(name, "Client") {
		return "", "", false
	}
	return strings.TrimSuffix(strings.TrimPrefix(name, "New"), "Client"), method.Sel.Name, true
}

func funcIdent(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.Ident:
		return e.Name
	case *ast.IndexExpr:
		return funcIdent(e.X)
	case *ast.IndexListExpr:
		return funcIdent(e.X)
	}
	return ""
}

func selectorName(expr ast.Expr) string {
	switch e := expr.(type) {
	case *ast.SelectorExpr:
		return e.Sel.Name
	case *ast.Ident:
		return e.Name
	case *ast.BasicLit:
		return stringLit(e)
	}
	return ""
}

func stringLit(expr ast.Expr) string {
	lit, ok := expr.(*ast.BasicLit)
	if !ok || lit.Kind != token.STRING {
		return ""
	}
	s, err := strconv.Unquote(lit.Value)
	if err != nil {
		return ""
	}
	return s
}

func proxyAPICall(ce *ast.CallExpr) (string, string, bool) {
	method, ok := ce.Fun.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}
	apiCall, ok := method.X.(*ast.CallExpr)
	if !ok {
		return "", "", false
	}
	api, ok := apiCall.Fun.(*ast.SelectorExpr)
	if !ok || api.Sel.Name != "API" {
		return "", "", false
	}
	svc, ok := api.X.(*ast.SelectorExpr)
	if !ok {
		return "", "", false
	}
	px, ok := svc.X.(*ast.SelectorExpr)
	if !ok || px.Sel.Name != "Proxy" {
		return "", "", false
	}
	return svc.Sel.Name, method.Sel.Name, true
}
