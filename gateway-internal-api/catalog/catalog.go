package catalog

import (
	"bytes"
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/vektah/gqlparser/v2"
	"github.com/vektah/gqlparser/v2/ast"
	"github.com/vektah/gqlparser/v2/formatter"
)

//go:embed schema/*.graphqls schema/manifest.json
var files embed.FS

const (
	DefaultLimit = 50
	MaxLimit     = 5000
)

type call struct {
	Transport string `json:"transport"`
	Service   string `json:"service"`
	Method    string `json:"method"`
	Path      string `json:"path"`
}

type manifest struct {
	Source   string            `json:"source"`
	SyncedAt string            `json:"synced_at"`
	Calls    map[string][]call `json:"calls"`
}

type Catalog struct {
	SyncedAt   string
	Services   []*model.Service
	Operations []*model.Operation
	Types      []*model.TypeDef

	services   map[string]*model.Service
	operations map[string]*model.Operation
	types      map[string]*model.TypeDef
	sdl        map[string]string
	fileOrder  []string
}

func Load() (*Catalog, error) {
	var mf manifest
	raw, err := files.ReadFile("schema/manifest.json")
	if err != nil {
		return nil, err
	}
	if err := json.Unmarshal(raw, &mf); err != nil {
		return nil, fmt.Errorf("manifest: %w", err)
	}

	names, err := fs.Glob(files, "schema/*.graphqls")
	if err != nil {
		return nil, err
	}
	sort.Strings(names)

	c := &Catalog{
		SyncedAt:   mf.SyncedAt,
		services:   map[string]*model.Service{},
		operations: map[string]*model.Operation{},
		types:      map[string]*model.TypeDef{},
		sdl:        map[string]string{},
	}

	var sources []*ast.Source
	for _, n := range names {
		b, err := files.ReadFile(n)
		if err != nil {
			return nil, err
		}
		base := path.Base(n)
		c.sdl[serviceName(base)] = string(b)
		c.fileOrder = append(c.fileOrder, serviceName(base))
		sources = append(sources, &ast.Source{Name: base, Input: string(b)})
	}

	schema, gerr := gqlparser.LoadSchema(sources...)
	if gerr != nil {
		return nil, fmt.Errorf("schema: %w", gerr)
	}

	c.buildTypes(schema)
	c.buildOperations(schema, mf.Calls)
	c.buildServices()
	return c, nil
}

func serviceName(file string) string {
	return strings.TrimSuffix(file, ".graphqls")
}

func sourceService(pos *ast.Position) string {
	if pos == nil || pos.Src == nil || pos.Src.BuiltIn {
		return ""
	}
	return serviceName(pos.Src.Name)
}

func strPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

func (c *Catalog) buildTypes(schema *ast.Schema) {
	for name, def := range schema.Types {
		svc := sourceService(def.Position)
		if svc == "" || strings.HasPrefix(name, "__") || isRoot(schema, def) {
			continue
		}
		t := &model.TypeDef{
			Name:          name,
			Kind:          typeKind(def.Kind),
			Service:       svc,
			Description:   strPtr(def.Description),
			Fields:        []*model.FieldDef{},
			EnumValues:    []string{},
			PossibleTypes: append([]string{}, def.Types...),
			Sdl:           formatDefinition(def),
		}
		for _, f := range def.Fields {
			t.Fields = append(t.Fields, &model.FieldDef{
				Name:        f.Name,
				Type:        f.Type.String(),
				NamedType:   f.Type.Name(),
				Required:    f.Type.NonNull,
				Description: strPtr(f.Description),
				Args:        arguments(f.Arguments),
			})
		}
		for _, v := range def.EnumValues {
			t.EnumValues = append(t.EnumValues, v.Name)
		}
		c.types[name] = t
		c.Types = append(c.Types, t)
	}
	sort.Slice(c.Types, func(i, j int) bool { return c.Types[i].Name < c.Types[j].Name })
}

func isRoot(schema *ast.Schema, def *ast.Definition) bool {
	return def == schema.Query || def == schema.Mutation || def == schema.Subscription
}

func typeKind(k ast.DefinitionKind) model.TypeKind {
	switch k {
	case ast.Object:
		return model.TypeKindObject
	case ast.Interface:
		return model.TypeKindInterface
	case ast.Union:
		return model.TypeKindUnion
	case ast.Enum:
		return model.TypeKindEnum
	case ast.InputObject:
		return model.TypeKindInputObject
	}
	return model.TypeKindScalar
}

func arguments(list ast.ArgumentDefinitionList) []*model.Argument {
	out := []*model.Argument{}
	for _, a := range list {
		arg := &model.Argument{
			Name:        a.Name,
			Type:        a.Type.String(),
			NamedType:   a.Type.Name(),
			Required:    a.Type.NonNull && a.DefaultValue == nil,
			Description: strPtr(a.Description),
		}
		if a.DefaultValue != nil {
			arg.DefaultValue = strPtr(a.DefaultValue.String())
		}
		out = append(out, arg)
	}
	return out
}

func (c *Catalog) buildOperations(schema *ast.Schema, calls map[string][]call) {
	roots := []struct {
		def  *ast.Definition
		kind model.OperationKind
	}{
		{schema.Query, model.OperationKindQuery},
		{schema.Mutation, model.OperationKindMutation},
		{schema.Subscription, model.OperationKindSubscription},
	}
	for _, root := range roots {
		if root.def == nil {
			continue
		}
		for _, f := range root.def.Fields {
			svc := sourceService(f.Position)
			if svc == "" || strings.HasPrefix(f.Name, "__") {
				continue
			}
			id := root.def.Name + "." + f.Name
			op := &model.Operation{
				ID:              id,
				Name:            f.Name,
				Kind:            root.kind,
				Service:         svc,
				Description:     strPtr(f.Description),
				Args:            arguments(f.Arguments),
				ReturnType:      f.Type.String(),
				ReturnNamedType: f.Type.Name(),
				Auth:            authOf(f),
				Upstreams:       []*model.Upstream{},
				Sdl:             formatDefinition(&ast.Definition{Kind: ast.Object, Name: root.def.Name, Fields: ast.FieldList{f}}),
			}
			if rest := f.Directives.ForName("rest"); rest != nil {
				op.Transport = model.TransportRest
				op.Upstreams = append(op.Upstreams, restUpstream(rest))
			} else if cs := calls[id]; len(cs) > 0 {
				for _, cl := range cs {
					t := model.TransportGrpc
					if cl.Transport == "rest" {
						t = model.TransportRest
					}
					op.Upstreams = append(op.Upstreams, &model.Upstream{
						Transport: t,
						Service:   cl.Service,
						Method:    strPtr(cl.Method),
						Path:      strPtr(cl.Path),
						Options:   []*model.Option{},
					})
				}
				op.Transport = op.Upstreams[0].Transport
			} else {
				op.Transport = model.TransportGateway
			}
			op.Example = buildExample(schema, root.kind, f, op.Auth.Mode)
			c.operations[id] = op
			c.Operations = append(c.Operations, op)
		}
	}
	sort.SliceStable(c.Operations, func(i, j int) bool {
		a, b := c.Operations[i], c.Operations[j]
		if a.Service != b.Service {
			return a.Service < b.Service
		}
		if a.Kind != b.Kind {
			return a.Kind == model.OperationKindQuery
		}
		return a.Name < b.Name
	})
}

func authOf(f *ast.FieldDefinition) *model.Auth {
	if f.Directives.ForName("public") != nil {
		return &model.Auth{Mode: model.AuthModePublic}
	}
	if d := f.Directives.ForName("hasPermission"); d != nil {
		return &model.Auth{
			Mode:   model.AuthModePermission,
			Object: strPtr(directiveArg(d, "object")),
			Action: strPtr(directiveArg(d, "action")),
		}
	}
	return &model.Auth{Mode: model.AuthModeAuthenticated}
}

func directiveArg(d *ast.Directive, name string) string {
	a := d.Arguments.ForName(name)
	if a == nil || a.Value == nil {
		return ""
	}
	return a.Value.Raw
}

func restUpstream(d *ast.Directive) *model.Upstream {
	method := strings.ToUpper(directiveArg(d, "method"))
	if method == "" {
		method = "GET"
	}
	if directiveArg(d, "multipart") == "true" && method == "GET" {
		method = "POST"
	}
	up := &model.Upstream{
		Transport: model.TransportRest,
		Service:   directiveArg(d, "service"),
		Method:    &method,
		Path:      strPtr(directiveArg(d, "path")),
		Options:   []*model.Option{},
	}
	for _, a := range d.Arguments {
		switch a.Name {
		case "service", "method", "path":
			continue
		}
		up.Options = append(up.Options, &model.Option{Key: a.Name, Value: a.Value.Raw})
	}
	return up
}

func formatDefinition(def *ast.Definition) string {
	var buf bytes.Buffer
	formatter.NewFormatter(&buf, formatter.WithIndent("  ")).FormatSchemaDocument(&ast.SchemaDocument{
		Definitions: ast.DefinitionList{def},
	})
	return strings.TrimSpace(buf.String())
}

func (c *Catalog) buildServices() {
	for _, name := range c.fileOrder {
		c.services[name] = &model.Service{
			Name:       name,
			SchemaFile: name + ".graphqls",
			Upstreams:  []string{},
			Operations: []*model.Operation{},
			Types:      []*model.TypeDef{},
		}
	}
	seen := map[string]map[string]bool{}
	for _, op := range c.Operations {
		s := c.services[op.Service]
		s.Operations = append(s.Operations, op)
		switch op.Kind {
		case model.OperationKindQuery:
			s.QueryCount++
		case model.OperationKindMutation:
			s.MutationCount++
		}
		if seen[s.Name] == nil {
			seen[s.Name] = map[string]bool{}
		}
		for _, up := range op.Upstreams {
			if !seen[s.Name][up.Service] {
				seen[s.Name][up.Service] = true
				s.Upstreams = append(s.Upstreams, up.Service)
			}
		}
	}
	for _, t := range c.Types {
		if s := c.services[t.Service]; s != nil {
			s.Types = append(s.Types, t)
		}
	}
	for _, name := range c.fileOrder {
		s := c.services[name]
		if len(s.Operations) == 0 {
			continue
		}
		sort.Strings(s.Upstreams)
		c.Services = append(c.Services, s)
	}
}

func (c *Catalog) Service(name string) *model.Service {
	s := c.services[name]
	if s == nil || len(s.Operations) == 0 {
		return nil
	}
	return s
}

func (c *Catalog) Operation(name string, kind *model.OperationKind) *model.Operation {
	if op := c.operations[name]; op != nil {
		return op
	}
	for _, root := range []model.OperationKind{model.OperationKindQuery, model.OperationKindMutation, model.OperationKindSubscription} {
		if kind != nil && *kind != root {
			continue
		}
		if op := c.operations[rootName(root)+"."+name]; op != nil {
			return op
		}
	}
	return nil
}

func rootName(k model.OperationKind) string {
	switch k {
	case model.OperationKindMutation:
		return "Mutation"
	case model.OperationKindSubscription:
		return "Subscription"
	}
	return "Query"
}

func (c *Catalog) Type(name string) *model.TypeDef {
	return c.types[name]
}

func (c *Catalog) SDL(service string) (string, bool) {
	if service == "" {
		parts := make([]string, 0, len(c.fileOrder))
		for _, name := range c.fileOrder {
			parts = append(parts, "# "+name+".graphqls\n"+strings.TrimSpace(c.sdl[name]))
		}
		return strings.Join(parts, "\n\n") + "\n", true
	}
	s, ok := c.sdl[service]
	return s, ok
}

func contains(haystack, needle string) bool {
	return strings.Contains(strings.ToLower(haystack), needle)
}

func (c *Catalog) SearchServices(search string) []*model.Service {
	q := strings.ToLower(strings.TrimSpace(search))
	if q == "" {
		return c.Services
	}
	out := []*model.Service{}
	for _, s := range c.Services {
		if contains(s.Name, q) || contains(strings.Join(s.Upstreams, " "), q) {
			out = append(out, s)
		}
	}
	return out
}

func (c *Catalog) FilterOperations(f *model.OperationFilter) []*model.Operation {
	if f == nil {
		return c.Operations
	}
	q := strings.ToLower(strings.TrimSpace(deref(f.Search)))
	upstream := strings.ToLower(strings.TrimSpace(deref(f.Upstream)))
	out := []*model.Operation{}
	for _, op := range c.Operations {
		if f.Service != nil && *f.Service != "" && op.Service != *f.Service {
			continue
		}
		if f.Kind != nil && op.Kind != *f.Kind {
			continue
		}
		if f.Auth != nil && op.Auth.Mode != *f.Auth {
			continue
		}
		if f.Transport != nil && op.Transport != *f.Transport {
			continue
		}
		if upstream != "" && !matchUpstream(op, upstream) {
			continue
		}
		if q != "" && !matchOperation(op, q) {
			continue
		}
		out = append(out, op)
	}
	return out
}

func matchUpstream(op *model.Operation, q string) bool {
	for _, up := range op.Upstreams {
		if contains(up.Service, q) {
			return true
		}
	}
	return false
}

func matchOperation(op *model.Operation, q string) bool {
	if contains(op.Name, q) || contains(op.Service, q) || contains(op.ReturnNamedType, q) {
		return true
	}
	if op.Description != nil && contains(*op.Description, q) {
		return true
	}
	for _, up := range op.Upstreams {
		if contains(up.Service, q) || contains(deref(up.Path), q) || contains(deref(up.Method), q) {
			return true
		}
	}
	return false
}

func (c *Catalog) FilterTypes(f *model.TypeFilter) []*model.TypeDef {
	if f == nil {
		return c.Types
	}
	q := strings.ToLower(strings.TrimSpace(deref(f.Search)))
	out := []*model.TypeDef{}
	for _, t := range c.Types {
		if f.Service != nil && *f.Service != "" && t.Service != *f.Service {
			continue
		}
		if f.Kind != nil && t.Kind != *f.Kind {
			continue
		}
		if q != "" && !contains(t.Name, q) {
			continue
		}
		out = append(out, t)
	}
	return out
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func Page(total int, limit, offset *int) (int, int, int, int) {
	l := DefaultLimit
	if limit != nil && *limit > 0 {
		l = *limit
	}
	if l > MaxLimit {
		l = MaxLimit
	}
	o := 0
	if offset != nil && *offset > 0 {
		o = *offset
	}
	start := o
	if start > total {
		start = total
	}
	end := start + l
	if end > total {
		end = total
	}
	return start, end, l, o
}

func (c *Catalog) Stats() *model.CatalogStats {
	st := &model.CatalogStats{
		Services:   len(c.Services),
		Operations: len(c.Operations),
		Types:      len(c.Types),
		SyncedAt:   c.SyncedAt,
	}
	auth := map[string]int{}
	transport := map[string]int{}
	for _, op := range c.Operations {
		switch op.Kind {
		case model.OperationKindQuery:
			st.Queries++
		case model.OperationKindMutation:
			st.Mutations++
		}
		auth[op.Auth.Mode.String()]++
		transport[op.Transport.String()]++
	}
	st.ByAuth = counts(auth)
	st.ByTransport = counts(transport)
	return st
}

func counts(m map[string]int) []*model.CountByKey {
	out := make([]*model.CountByKey, 0, len(m))
	for k, v := range m {
		out = append(out, &model.CountByKey{Key: k, Count: v})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Count > out[j].Count })
	return out
}
