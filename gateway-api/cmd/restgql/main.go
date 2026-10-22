package main

import (
	"bytes"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/99designs/gqlgen/codegen/templates"
	"github.com/vektah/gqlparser/v2"
	gqlast "github.com/vektah/gqlparser/v2/ast"
	"golang.org/x/tools/go/ast/astutil"
)

const restImport = "github.com/JIeeiroSst/gateway-api/proxy/rest"

var placeholder = regexp.MustCompile(`\{([$*]?[A-Za-z_][A-Za-z0-9_]*)\}`)

type spec struct {
	recv        string
	field       *gqlast.FieldDefinition
	service     string
	method      string
	path        string
	unwrap      string
	body        string
	multipart   bool
	fixed       string
	userQuery   string
	userField   string
	raw         bool
	owner       string
	ownerPath   string
	ownerUnwrap string
	queryArgs   string
	keys        map[string]string
	headerArgs  map[string]string
}

func main() {
	dir := "graph"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	files, err := filepath.Glob(filepath.Join(dir, "*.graphqls"))
	if err != nil {
		log.Fatal(err)
	}
	var sources []*gqlast.Source
	for _, f := range files {
		b, err := os.ReadFile(f)
		if err != nil {
			log.Fatal(err)
		}
		sources = append(sources, &gqlast.Source{Name: f, Input: string(b)})
	}
	schema, gerr := gqlparser.LoadSchema(sources...)
	if gerr != nil {
		log.Fatal(gerr)
	}
	byFile := map[string][]spec{}
	for _, root := range []*gqlast.Definition{schema.Query, schema.Mutation} {
		if root == nil {
			continue
		}
		recv := "*queryResolver"
		if root == schema.Mutation {
			recv = "*mutationResolver"
		}
		for _, f := range root.Fields {
			d := f.Directives.ForName("rest")
			if d == nil {
				continue
			}
			s := spec{recv: recv, field: f, method: "GET"}
			for _, a := range d.Arguments {
				switch a.Name {
				case "service":
					s.service = a.Value.Raw
				case "method":
					s.method = strings.ToUpper(a.Value.Raw)
				case "path":
					s.path = a.Value.Raw
				case "unwrap":
					s.unwrap = a.Value.Raw
				case "body":
					s.body = a.Value.Raw
				case "multipart":
					s.multipart = a.Value.Raw == "true"
				case "fixed":
					s.fixed = a.Value.Raw
				case "userQuery":
					s.userQuery = a.Value.Raw
				case "userField":
					s.userField = a.Value.Raw
				case "raw":
					s.raw = a.Value.Raw == "true"
				case "owner":
					s.owner = a.Value.Raw
				case "ownerPath":
					s.ownerPath = a.Value.Raw
				case "ownerUnwrap":
					s.ownerUnwrap = a.Value.Raw
				case "queryArgs":
					s.queryArgs = a.Value.Raw
				case "headerArgs":
					s.headerArgs = map[string]string{}
					for _, kv := range strings.Split(a.Value.Raw, ",") {
						if k, v, ok := strings.Cut(strings.TrimSpace(kv), ":"); ok {
							s.headerArgs[k] = v
						}
					}
				case "keys":
					s.keys = map[string]string{}
					for _, kv := range strings.Split(a.Value.Raw, ",") {
						if k, v, ok := strings.Cut(strings.TrimSpace(kv), ":"); ok {
							s.keys[k] = v
						}
					}
				}
			}
			file := strings.TrimSuffix(f.Position.Src.Name, ".graphqls") + ".resolvers.go"
			byFile[file] = append(byFile[file], s)
		}
	}
	names := make([]string, 0, len(byFile))
	for f := range byFile {
		names = append(names, f)
	}
	sort.Strings(names)
	total := 0
	for _, f := range names {
		n, err := fill(f, byFile[f])
		if err != nil {
			log.Fatalf("%s: %v", f, err)
		}
		total += n
	}
	fmt.Printf("restgql: %d resolvers in %d files\n", total, len(names))
}

type edit struct {
	start, end int
	text       string
}

func fill(path string, specs []spec) (int, error) {
	src, err := os.ReadFile(path)
	if err != nil {
		return 0, err
	}
	fset := token.NewFileSet()
	file, err := parser.ParseFile(fset, path, src, parser.ParseComments)
	if err != nil {
		return 0, err
	}
	funcs := map[string]*ast.FuncDecl{}
	for _, decl := range file.Decls {
		fd, ok := decl.(*ast.FuncDecl)
		if !ok || fd.Recv == nil || len(fd.Recv.List) != 1 {
			continue
		}
		recv := exprString(fset, src, fd.Recv.List[0].Type)
		if recv == "*queryResolver" || recv == "*mutationResolver" {
			funcs[recv+"."+fd.Name.Name] = fd
		}
	}
	var edits []edit
	for _, s := range specs {
		fd, ok := funcs[s.recv+"."+templates.ToGo(s.field.Name)]
		if !ok {
			return 0, fmt.Errorf("resolver for %s not found; run gqlgen first", s.field.Name)
		}
		body, err := build(fset, src, fd, s)
		if err != nil {
			return 0, fmt.Errorf("%s: %w", s.field.Name, err)
		}
		edits = append(edits, edit{
			start: fset.Position(fd.Body.Lbrace).Offset,
			end:   fset.Position(fd.Body.Rbrace).Offset + 1,
			text:  body,
		})
	}
	sort.Slice(edits, func(i, j int) bool { return edits[i].start > edits[j].start })
	out := append([]byte(nil), src...)
	for _, e := range edits {
		out = append(out[:e.start], append([]byte(e.text), out[e.end:]...)...)
	}
	fset = token.NewFileSet()
	file, err = parser.ParseFile(fset, path, out, parser.ParseComments)
	if err != nil {
		return 0, fmt.Errorf("generated code does not parse: %w", err)
	}
	astutil.AddImport(fset, file, restImport)
	if !bytes.Contains(out, []byte("fmt.")) {
		astutil.DeleteImport(fset, file, "fmt")
	}
	var buf bytes.Buffer
	if err := format.Node(&buf, fset, file); err != nil {
		return 0, err
	}
	formatted, err := format.Source(buf.Bytes())
	if err != nil {
		return 0, err
	}
	return len(edits), os.WriteFile(path, formatted, 0o644)
}

type param struct {
	name    string
	typ     string
	pointer bool
}

func build(fset *token.FileSet, src []byte, fd *ast.FuncDecl, s spec) (string, error) {
	var params []param
	for _, f := range fd.Type.Params.List {
		typ := exprString(fset, src, f.Type)
		for _, n := range f.Names {
			params = append(params, param{name: n.Name, typ: typ, pointer: strings.HasPrefix(typ, "*")})
		}
	}
	if len(params) == 0 || params[0].typ != "context.Context" {
		return "", fmt.Errorf("unexpected signature")
	}
	params = params[1:]
	for _, p := range params {
		switch p.name {
		case "model", "rest", "r", "ctx", "graphql", "proxy", "context":
			return "", fmt.Errorf("argument %q shadows a Go identifier; rename it in the schema", p.name)
		}
	}
	if len(params) != len(s.field.Arguments) {
		return "", fmt.Errorf("argument count mismatch")
	}
	if fd.Type.Results == nil || len(fd.Type.Results.List) != 2 {
		return "", fmt.Errorf("unexpected results")
	}
	ret := exprString(fset, src, fd.Type.Results.List[0].Type)

	wire := func(name string) string {
		if w, ok := s.keys[name]; ok {
			return w
		}
		return name
	}
	byArg := map[string]param{}
	for i, a := range s.field.Arguments {
		byArg[a.Name] = params[i]
	}
	used := map[string]bool{}

	format := s.path
	var pathArgs []string
	for _, m := range placeholder.FindAllStringSubmatch(s.path, -1) {
		if m[1] == "$userId" {
			format = strings.Replace(format, m[0], "%v", 1)
			pathArgs = append(pathArgs, "userIDOf(ctx)")
			continue
		}
		if strings.HasPrefix(m[1], "*") {
			name := strings.TrimPrefix(m[1], "*")
			p, ok := byArg[name]
			if !ok {
				return "", fmt.Errorf("path placeholder {%s} has no argument", m[1])
			}
			used[name] = true
			format = strings.Replace(format, m[0], "%s", 1)
			v := p.name
			if p.pointer {
				v = "deref(" + p.name + ")"
			}
			pathArgs = append(pathArgs, "rest.Raw("+v+")")
			continue
		}
		p, ok := byArg[m[1]]
		if !ok {
			return "", fmt.Errorf("path placeholder {%s} has no argument", m[1])
		}
		used[m[1]] = true
		format = strings.Replace(format, m[0], "%v", 1)
		if p.pointer {
			pathArgs = append(pathArgs, "deref("+p.name+")")
		} else {
			pathArgs = append(pathArgs, p.name)
		}
	}
	pathExpr := fmt.Sprintf("%q", format)
	if len(pathArgs) > 0 {
		pathExpr = fmt.Sprintf("rest.P(%q, %s)", format, strings.Join(pathArgs, ", "))
	}

	if s.raw {
		var upload string
		var query []string
		for _, a := range s.field.Arguments {
			if used[a.Name] {
				continue
			}
			p := byArg[a.Name]
			if strings.Contains(p.typ, "graphql.Upload") && upload == "" {
				upload = p.name
				if p.pointer {
					upload = "deref(" + p.name + ")"
				}
				continue
			}
			query = append(query, fmt.Sprintf("%q, %s", wire(a.Name), p.name))
		}
		if upload == "" && s.body != "" {
			p, ok := byArg[s.body]
			if !ok {
				return "", fmt.Errorf("raw body argument %s not found", s.body)
			}
			var q2 []string
			for _, a := range s.field.Arguments {
				if used[a.Name] || a.Name == s.body {
					continue
				}
				q2 = append(q2, fmt.Sprintf("%q, %s", wire(a.Name), byArg[a.Name].name))
			}
			qe := ""
			if len(q2) > 0 {
				qe = ", Query: restQuery(" + strings.Join(q2, ", ") + ")"
			}
			v := p.name
			if p.pointer {
				v = "deref(" + p.name + ")"
			}
			return fmt.Sprintf("{\n\treturn restText[%s](ctx, r.Proxy, %q, rest.Request{Method: %q, Path: %s%s}, %s, %q)\n}",
				ret, s.service, s.method, pathExpr, qe, v, s.unwrap), nil
		}
		if upload == "" {
			return "", fmt.Errorf("raw requires an Upload argument")
		}
		queryExpr := ""
		if len(query) > 0 {
			queryExpr = ", Query: restQuery(" + strings.Join(query, ", ") + ")"
		}
		return fmt.Sprintf("{\n\treturn restRaw[%s](ctx, r.Proxy, %q, rest.Request{Method: %q, Path: %s%s}, %s, %q)\n}",
			ret, s.service, s.method, pathExpr, queryExpr, upload, s.unwrap), nil
	}

	if s.multipart {
		var fields, files, mquery []string
		for _, a := range s.field.Arguments {
			if used[a.Name] {
				continue
			}
			p := byArg[a.Name]
			switch {
			case strings.Contains(p.typ, "graphql.Upload"):
				files = append(files, fmt.Sprintf("%q, %s", wire(a.Name), p.name))
			case s.body == "query":
				mquery = append(mquery, fmt.Sprintf("%q, %s", wire(a.Name), p.name))
			default:
				fields = append(fields, fmt.Sprintf("%q: %s", wire(a.Name), p.name))
			}
		}
		if s.userField != "" {
			field, getter := s.userField, "userIDOf(ctx)"
			if strings.HasSuffix(field, ":int") {
				field, getter = strings.TrimSuffix(field, ":int"), "userIDInt(ctx)"
			}
			fields = append(fields, fmt.Sprintf("%q: %s", field, getter))
		}
		method := s.method
		if method == "GET" {
			method = "POST"
		}
		mq := ""
		if len(mquery) > 0 {
			mq = ", Query: restQuery(" + strings.Join(mquery, ", ") + ")"
		}
		return fmt.Sprintf("{\n\treturn restMultipart[%s](ctx, r.Proxy, %q, rest.Request{Method: %q, Path: %s%s}, map[string]any{%s}, uploadFiles(%s), %q)\n}",
			ret, s.service, method, pathExpr, mq, strings.Join(fields, ", "), strings.Join(files, ", "), s.unwrap), nil
	}

	headerExpr := ""
	if len(s.headerArgs) > 0 {
		var hs []string
		for _, a := range s.field.Arguments {
			h, ok := s.headerArgs[a.Name]
			if !ok {
				continue
			}
			hs = append(hs, fmt.Sprintf("%q, %s", h, byArg[a.Name].name))
		}
		headerExpr = ", Header: restHeader(" + strings.Join(hs, ", ") + ")"
	}
	var forcedQuery []string
	if s.queryArgs != "" {
		for _, n := range strings.Split(s.queryArgs, ",") {
			n = strings.TrimSpace(n)
			p, ok := byArg[n]
			if !ok {
				return "", fmt.Errorf("queryArgs %s not found", n)
			}
			used[n] = true
			forcedQuery = append(forcedQuery, fmt.Sprintf("%q, %s", wire(n), p.name))
		}
	}
	bodyArg := s.body
	argsAsBody := false
	if bodyArg == "" && s.method != "GET" && s.method != "DELETE" {
		if _, ok := byArg["input"]; ok {
			bodyArg = "input"
		} else {
			argsAsBody = true
		}
	}
	if bodyArg == "query" {
		bodyArg = ""
	}
	bodyExpr := ""
	if argsAsBody {
		var kv []string
		for _, a := range s.field.Arguments {
			if used[a.Name] {
				continue
			}
			used[a.Name] = true
			kv = append(kv, fmt.Sprintf("%q: %s", wire(a.Name), byArg[a.Name].name))
		}
		bodyExpr = ", Body: map[string]any{" + strings.Join(kv, ", ") + "}"
	}
	if bodyArg != "" {
		p, ok := byArg[bodyArg]
		if !ok {
			return "", fmt.Errorf("body argument %s not found", bodyArg)
		}
		used[bodyArg] = true
		bodyExpr = ", Body: " + p.name
	}
	queryFixed := s.fixed != "" && (s.method == "GET" || s.method == "DELETE")
	if s.fixed != "" && !queryFixed {
		inner := "nil"
		if bodyExpr != "" {
			inner = strings.TrimPrefix(bodyExpr, ", Body: ")
		}
		bodyExpr = fmt.Sprintf(", Body: mergeBody(%s, %q)", inner, s.fixed)
	}
	query := forcedQuery
	for _, a := range s.field.Arguments {
		if used[a.Name] {
			continue
		}
		query = append(query, fmt.Sprintf("%q, %s", wire(a.Name), byArg[a.Name].name))
	}
	if s.userQuery != "" {
		query = append(query, fmt.Sprintf("%q, userIDOf(ctx)", s.userQuery))
	}
	if s.userField != "" {
		inner := "nil"
		if bodyExpr != "" {
			inner = strings.TrimPrefix(bodyExpr, ", Body: ")
		}
		field, getter := s.userField, "userIDOf(ctx)"
		if strings.HasSuffix(field, ":int") {
			field, getter = strings.TrimSuffix(field, ":int"), "userIDInt(ctx)"
		}
		bodyExpr = fmt.Sprintf(", Body: withField(%s, %q, %s)", inner, field, getter)
	}
	queryExpr := ""
	if len(query) > 0 {
		queryExpr = ", Query: restQuery(" + strings.Join(query, ", ") + ")"
	}
	if queryFixed {
		inner := "nil"
		if queryExpr != "" {
			inner = strings.TrimPrefix(queryExpr, ", Query: ")
		}
		queryExpr = fmt.Sprintf(", Query: mergeQuery(%s, %q)", inner, s.fixed)
	}
	if s.owner != "" {
		check := "nil"
		if s.ownerPath != "" {
			cf := s.ownerPath
			var cargs []string
			for _, m := range placeholder.FindAllStringSubmatch(s.ownerPath, -1) {
				cf = strings.Replace(cf, m[0], "%v", 1)
				if m[1] == "$userId" {
					cargs = append(cargs, "userIDOf(ctx)")
					continue
				}
				p, ok := byArg[m[1]]
				if !ok {
					return "", fmt.Errorf("ownerPath placeholder {%s} has no argument", m[1])
				}
				if p.pointer {
					cargs = append(cargs, "deref("+p.name+")")
				} else {
					cargs = append(cargs, p.name)
				}
			}
			cpath := fmt.Sprintf("%q", cf)
			if len(cargs) > 0 {
				cpath = fmt.Sprintf("rest.P(%q, %s)", cf, strings.Join(cargs, ", "))
			}
			check = fmt.Sprintf("&rest.Request{Method: \"GET\", Path: %s}", cpath)
		}
		return fmt.Sprintf("{\n\treturn restOwned[%s](ctx, r.Proxy, %q, rest.Request{Method: %q, Path: %s%s%s%s}, %q, %q, %s, %q)\n}",
			ret, s.service, s.method, pathExpr, queryExpr, bodyExpr, headerExpr, s.unwrap, s.owner, check, s.ownerUnwrap), nil
	}
	return fmt.Sprintf("{\n\treturn restCall[%s](ctx, r.Proxy, %q, rest.Request{Method: %q, Path: %s%s%s%s}, %q)\n}",
		ret, s.service, s.method, pathExpr, queryExpr, bodyExpr, headerExpr, s.unwrap), nil
}

func exprString(fset *token.FileSet, src []byte, e ast.Expr) string {
	return string(src[fset.Position(e.Pos()).Offset:fset.Position(e.End()).Offset])
}
