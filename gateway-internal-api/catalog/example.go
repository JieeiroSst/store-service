package catalog

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/vektah/gqlparser/v2/ast"
)

const (
	maxSelectionDepth = 3
	maxInputDepth     = 4
)

func buildExample(schema *ast.Schema, kind model.OperationKind, f *ast.FieldDefinition, auth model.AuthMode) *model.Example {
	opName := strings.ToUpper(f.Name[:1]) + f.Name[1:]

	var b strings.Builder
	b.WriteString(strings.ToLower(kind.String()))
	b.WriteString(" ")
	b.WriteString(opName)
	if len(f.Arguments) > 0 {
		vars := make([]string, 0, len(f.Arguments))
		for _, a := range f.Arguments {
			vars = append(vars, "$"+a.Name+": "+a.Type.String())
		}
		b.WriteString("(" + strings.Join(vars, ", ") + ")")
	}
	b.WriteString(" {\n  ")
	b.WriteString(f.Name)
	if len(f.Arguments) > 0 {
		args := make([]string, 0, len(f.Arguments))
		for _, a := range f.Arguments {
			args = append(args, a.Name+": $"+a.Name)
		}
		b.WriteString("(" + strings.Join(args, ", ") + ")")
	}
	if sel := selection(schema, f.Type.Name(), 1, map[string]bool{}, "    "); sel != "" {
		b.WriteString(" {\n" + sel + "  }")
	}
	b.WriteString("\n}\n")
	query := b.String()

	vars := orderedObject{}
	for _, a := range f.Arguments {
		vars = append(vars, kv{a.Name, sampleValue(schema, a.Type, 1, map[string]bool{})})
	}
	variables := vars.indent("")

	return &model.Example{
		Query:     query,
		Variables: variables,
		Curl:      curl(query, vars, auth),
	}
}

func selection(schema *ast.Schema, typeName string, depth int, path map[string]bool, indent string) string {
	def := schema.Types[typeName]
	if def == nil {
		return ""
	}
	switch def.Kind {
	case ast.Scalar, ast.Enum:
		return ""
	case ast.Union:
		var b strings.Builder
		b.WriteString(indent + "__typename\n")
		for _, member := range def.Types {
			if sub := selection(schema, member, depth, path, indent+"  "); sub != "" {
				b.WriteString(indent + "... on " + member + " {\n" + sub + indent + "}\n")
			}
		}
		return b.String()
	}

	path[typeName] = true
	defer delete(path, typeName)

	var b strings.Builder
	if def.Kind == ast.Interface {
		b.WriteString(indent + "__typename\n")
	}
	for _, fd := range def.Fields {
		if strings.HasPrefix(fd.Name, "__") || hasRequiredArgs(fd) {
			continue
		}
		child := schema.Types[fd.Type.Name()]
		if child == nil || child.Kind == ast.Scalar || child.Kind == ast.Enum {
			b.WriteString(indent + fd.Name + "\n")
			continue
		}
		if depth >= maxSelectionDepth || path[child.Name] {
			continue
		}
		if sub := selection(schema, child.Name, depth+1, path, indent+"  "); sub != "" {
			b.WriteString(indent + fd.Name + " {\n" + sub + indent + "}\n")
		}
	}
	if b.Len() == 0 {
		return indent + "__typename\n"
	}
	return b.String()
}

func hasRequiredArgs(fd *ast.FieldDefinition) bool {
	for _, a := range fd.Arguments {
		if a.Type.NonNull && a.DefaultValue == nil {
			return true
		}
	}
	return false
}

type kv struct {
	key   string
	value any
}

type orderedObject []kv

func (o orderedObject) indent(prefix string) string {
	if len(o) == 0 {
		return "{}"
	}
	var b strings.Builder
	b.WriteString("{\n")
	for i, p := range o {
		k, _ := json.Marshal(p.key)
		b.WriteString(prefix + "  " + string(k) + ": " + encode(p.value, prefix+"  "))
		if i < len(o)-1 {
			b.WriteString(",")
		}
		b.WriteString("\n")
	}
	b.WriteString(prefix + "}")
	return b.String()
}

func (o orderedObject) compact() string {
	parts := make([]string, 0, len(o))
	for _, p := range o {
		k, _ := json.Marshal(p.key)
		parts = append(parts, string(k)+":"+encodeCompact(p.value))
	}
	return "{" + strings.Join(parts, ",") + "}"
}

func encode(v any, prefix string) string {
	switch t := v.(type) {
	case orderedObject:
		return t.indent(prefix)
	case []any:
		items := make([]string, 0, len(t))
		for _, it := range t {
			items = append(items, encode(it, prefix))
		}
		return "[" + strings.Join(items, ", ") + "]"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func encodeCompact(v any) string {
	switch t := v.(type) {
	case orderedObject:
		return t.compact()
	case []any:
		items := make([]string, 0, len(t))
		for _, it := range t {
			items = append(items, encodeCompact(it))
		}
		return "[" + strings.Join(items, ",") + "]"
	}
	b, _ := json.Marshal(v)
	return string(b)
}

func sampleValue(schema *ast.Schema, t *ast.Type, depth int, path map[string]bool) any {
	if t.Elem != nil {
		return []any{sampleValue(schema, t.Elem, depth, path)}
	}
	switch t.NamedType {
	case "String", "ID":
		return ""
	case "Int", "Int64", "Float":
		return 0
	case "Boolean":
		return false
	case "Upload":
		return nil
	case "Map", "Any":
		return orderedObject{}
	}
	def := schema.Types[t.NamedType]
	if def == nil {
		return nil
	}
	switch def.Kind {
	case ast.Enum:
		if len(def.EnumValues) > 0 {
			return def.EnumValues[0].Name
		}
		return ""
	case ast.InputObject:
		if depth > maxInputDepth || path[def.Name] {
			return nil
		}
		path[def.Name] = true
		defer delete(path, def.Name)
		obj := orderedObject{}
		for _, fd := range def.Fields {
			obj = append(obj, kv{fd.Name, sampleValue(schema, fd.Type, depth+1, path)})
		}
		return obj
	}
	return nil
}

func curl(query string, vars orderedObject, auth model.AuthMode) string {
	q, _ := json.Marshal(query)
	body := fmt.Sprintf(`{"query":%s,"variables":%s}`, q, vars.compact())
	var b strings.Builder
	b.WriteString("curl -X POST \"$GATEWAY_URL/query\" \\\n")
	b.WriteString("  -H 'Content-Type: application/json' \\\n")
	if auth != model.AuthModePublic {
		b.WriteString("  -H 'Authorization: Bearer <access_token>' \\\n")
	}
	b.WriteString("  --data '" + strings.ReplaceAll(body, "'", `'\''`) + "'")
	return b.String()
}
