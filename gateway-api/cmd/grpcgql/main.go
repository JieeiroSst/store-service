package main

import (
	"bytes"
	"fmt"
	"go/format"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/99designs/gqlgen/codegen/templates"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	_ "github.com/JIeeiroSst/gateway-api/proxy/pb/accounttxpb"
	_ "github.com/JIeeiroSst/gateway-api/proxy/pb/carrentalpb"
	_ "github.com/JIeeiroSst/gateway-api/proxy/pb/hospitalpb"
)

type target struct {
	File              string
	Proxy             string
	Prefix            string
	Object            string
	Services          []protoreflect.FullName
	DefaultPermission bool
	Public            []string
	Permission        []string
}

var targets = []target{
	{
		File:              "account_transaction",
		Proxy:             "AccountTransactionService",
		Prefix:            "AccountTx",
		Object:            "/graphql/account-transaction",
		Services:          []protoreflect.FullName{"banking.BankingService"},
		DefaultPermission: true,
	},
	{
		File:     "car_rental",
		Proxy:    "CarRentalService",
		Prefix:   "CarRental",
		Object:   "/graphql/car-rental",
		Services: []protoreflect.FullName{"vehicle_rental.VehicleRentalService"},
		Public: []string{
			"GetVehicle", "ListVehicles", "SearchAvailableVehicles", "ListVehicleReviews", "ListLocations",
		},
		Permission: []string{"CreateVehicle", "UpdateVehicle"},
	},
	{
		File:              "hospital",
		Proxy:             "HospitalService",
		Prefix:            "Hospital",
		Object:            "/graphql/hospital",
		Services:          []protoreflect.FullName{"hospital.HospotalService", "hospital.AppointmentService"},
		DefaultPermission: true,
	},
}

var queryPrefix = regexp.MustCompile(`^(Get|List|Search|Find|Check|Count|Fetch|Lookup|Retrieve|Query)([A-Z]|$)`)

type gen struct {
	t        target
	pkg      protoreflect.FullName
	outTypes map[protoreflect.FullName]protoreflect.MessageDescriptor
	inTypes  map[protoreflect.FullName]protoreflect.MessageDescriptor
	enums    map[protoreflect.FullName]protoreflect.EnumDescriptor
}

type method struct {
	field    string
	query    bool
	md       protoreflect.MethodDescriptor
	hasInput bool
	empty    bool
}

func main() {
	dir := "graph"
	if len(os.Args) > 1 {
		dir = os.Args[1]
	}
	for _, t := range targets {
		if err := run(dir, t); err != nil {
			log.Fatalf("%s: %v", t.File, err)
		}
	}
}

func run(dir string, t target) error {
	g := &gen{
		t:        t,
		outTypes: map[protoreflect.FullName]protoreflect.MessageDescriptor{},
		inTypes:  map[protoreflect.FullName]protoreflect.MessageDescriptor{},
		enums:    map[protoreflect.FullName]protoreflect.EnumDescriptor{},
	}
	var methods []method
	used := map[string]bool{}
	for _, name := range t.Services {
		d, err := protoregistry.GlobalFiles.FindDescriptorByName(name)
		if err != nil {
			return err
		}
		sd := d.(protoreflect.ServiceDescriptor)
		g.pkg = sd.ParentFile().Package()
		for i := 0; i < sd.Methods().Len(); i++ {
			md := sd.Methods().Get(i)
			if md.IsStreamingClient() || md.IsStreamingServer() {
				continue
			}
			field := lowerFirst(t.Prefix) + string(md.Name())
			if used[field] {
				field = lowerFirst(t.Prefix) + string(sd.Name()) + string(md.Name())
			}
			used[field] = true
			m := method{
				field:    field,
				query:    queryPrefix.MatchString(string(md.Name())),
				md:       md,
				hasInput: md.Input().FullName() != "google.protobuf.Empty" && md.Input().Fields().Len() > 0,
				empty:    md.Output().FullName() == "google.protobuf.Empty",
			}
			if m.hasInput {
				g.collect(md.Input(), true)
			}
			if !m.empty {
				if isWellKnown(md.Output()) {
					return fmt.Errorf("%s returns %s", md.FullName(), md.Output().FullName())
				}
				g.collect(md.Output(), false)
			}
			methods = append(methods, m)
		}
	}
	schema := g.schema(methods)
	if err := os.WriteFile(filepath.Join(dir, t.File+".graphqls"), schema, 0o644); err != nil {
		return err
	}
	src, err := g.resolvers(methods)
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(dir, t.File+".resolvers.go"), src, 0o644)
}

func (g *gen) collect(md protoreflect.MessageDescriptor, input bool) {
	if isWellKnown(md) {
		return
	}
	set := g.outTypes
	if input {
		set = g.inTypes
	}
	if _, ok := set[md.FullName()]; ok {
		return
	}
	set[md.FullName()] = md
	fields := md.Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		if fd.IsMap() {
			continue
		}
		switch fd.Kind() {
		case protoreflect.EnumKind:
			g.enums[fd.Enum().FullName()] = fd.Enum()
		case protoreflect.MessageKind, protoreflect.GroupKind:
			g.collect(fd.Message(), input)
		}
	}
}

func (g *gen) typeName(full protoreflect.FullName) string {
	rel := strings.TrimPrefix(string(full), string(g.pkg)+".")
	return g.t.Prefix + strings.ReplaceAll(rel, ".", "")
}

func (g *gen) fieldType(fd protoreflect.FieldDescriptor, input bool) string {
	if fd.IsMap() {
		if input {
			return "Map"
		}
		return "Map!"
	}
	base, nullable := g.baseType(fd, input)
	if fd.IsList() {
		if input {
			return "[" + base + "!]"
		}
		return "[" + base + "!]!"
	}
	if input || nullable || fd.HasPresence() {
		return base
	}
	return base + "!"
}

func (g *gen) baseType(fd protoreflect.FieldDescriptor, input bool) (string, bool) {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return "Boolean", false
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind:
		return "Int", false
	case protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind,
		protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return "Int64", false
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		return "Float", true
	case protoreflect.StringKind, protoreflect.BytesKind:
		return "String", false
	case protoreflect.EnumKind:
		return g.typeName(fd.Enum().FullName()), true
	}
	md := fd.Message()
	switch md.FullName() {
	case "google.protobuf.Timestamp", "google.protobuf.Duration", "google.protobuf.FieldMask",
		"google.protobuf.StringValue", "google.protobuf.BytesValue":
		return "String", true
	case "google.protobuf.Struct", "google.protobuf.Any", "google.protobuf.Empty":
		return "Map", true
	case "google.protobuf.Value", "google.protobuf.ListValue":
		return "Any", true
	case "google.protobuf.BoolValue":
		return "Boolean", true
	case "google.protobuf.Int32Value":
		return "Int", true
	case "google.protobuf.Int64Value", "google.protobuf.UInt64Value", "google.protobuf.UInt32Value":
		return "Int64", true
	case "google.protobuf.DoubleValue", "google.protobuf.FloatValue":
		return "Float", true
	}
	name := g.typeName(md.FullName())
	if input {
		name += "Input"
	}
	return name, true
}

func (g *gen) schema(methods []method) []byte {
	var b bytes.Buffer
	for _, full := range sortedKeys(g.enums) {
		ed := g.enums[full]
		fmt.Fprintf(&b, "enum %s {\n", g.typeName(full))
		for i := 0; i < ed.Values().Len(); i++ {
			fmt.Fprintf(&b, "  %s\n", ed.Values().Get(i).Name())
		}
		b.WriteString("}\n\n")
	}
	for _, full := range sortedKeys(g.outTypes) {
		g.writeObject(&b, "type", g.typeName(full), g.outTypes[full], false)
	}
	for _, full := range sortedKeys(g.inTypes) {
		g.writeObject(&b, "input", g.typeName(full)+"Input", g.inTypes[full], true)
	}
	for _, query := range []bool{true, false} {
		root := "Mutation"
		if query {
			root = "Query"
		}
		var lines []string
		for _, m := range methods {
			if m.query != query {
				continue
			}
			args := ""
			if m.hasInput {
				args = "(input: " + g.typeName(m.md.Input().FullName()) + "Input!)"
			}
			ret := "Boolean!"
			if !m.empty {
				ret = g.typeName(m.md.Output().FullName())
			}
			lines = append(lines, fmt.Sprintf("  %s%s: %s%s", m.field, args, ret, g.directive(m)))
		}
		if len(lines) == 0 {
			continue
		}
		fmt.Fprintf(&b, "extend type %s {\n%s\n}\n\n", root, strings.Join(lines, "\n"))
	}
	return bytes.TrimRight(b.Bytes(), "\n")
}

func (g *gen) writeObject(b *bytes.Buffer, kind, name string, md protoreflect.MessageDescriptor, input bool) {
	fmt.Fprintf(b, "%s %s {\n", kind, name)
	fields := md.Fields()
	if fields.Len() == 0 {
		if input {
			b.WriteString("  _empty: Boolean\n")
		} else {
			b.WriteString("  _empty: Boolean\n")
		}
	}
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		fmt.Fprintf(b, "  %s: %s\n", fd.JSONName(), g.fieldType(fd, input))
	}
	b.WriteString("}\n\n")
}

func (g *gen) directive(m method) string {
	name := string(m.md.Name())
	action := "MUTATION"
	if m.query {
		action = "QUERY"
	}
	switch {
	case contains(g.t.Public, name):
		return " @public"
	case g.t.DefaultPermission || contains(g.t.Permission, name):
		return fmt.Sprintf(" @hasPermission(object: %q, action: %q)", g.t.Object, action)
	}
	return ""
}

func (g *gen) resolvers(methods []method) ([]byte, error) {
	var b bytes.Buffer
	b.WriteString("package graph\n\nimport (\n\t\"context\"\n\n\t\"github.com/JIeeiroSst/gateway-api/graph/model\"\n\t\"github.com/JIeeiroSst/gateway-api/proxy\"\n)\n\n")
	sorted := append([]method(nil), methods...)
	sort.SliceStable(sorted, func(i, j int) bool {
		if sorted[i].query != sorted[j].query {
			return !sorted[i].query
		}
		return sorted[i].field < sorted[j].field
	})
	usesModel := false
	for _, m := range sorted {
		recv := "mutationResolver"
		if m.query {
			recv = "queryResolver"
		}
		params, arg := "", "nil"
		if m.hasInput {
			params = ", input model." + templates.ToGo(g.typeName(m.md.Input().FullName())+"Input")
			arg = "input"
			usesModel = true
		}
		methodName := fmt.Sprintf("/%s/%s", m.md.Parent().FullName(), m.md.Name())
		call := fmt.Sprintf("r.Proxy, proxy.%s, %q, %q, %q, %s", g.t.Proxy, methodName, m.md.Input().FullName(), m.md.Output().FullName(), arg)
		if m.empty {
			fmt.Fprintf(&b, "func (r *%s) %s(ctx context.Context%s) (bool, error) {\n\treturn grpcInvokeEmpty(ctx, %s)\n}\n\n", recv, templates.ToGo(m.field), params, call)
			continue
		}
		out := templates.ToGo(g.typeName(m.md.Output().FullName()))
		usesModel = true
		fmt.Fprintf(&b, "func (r *%s) %s(ctx context.Context%s) (*model.%s, error) {\n\treturn grpcInvoke[model.%s](ctx, %s)\n}\n\n", recv, templates.ToGo(m.field), params, out, out, call)
	}
	src := b.Bytes()
	if !usesModel {
		src = bytes.Replace(src, []byte("\t\"github.com/JIeeiroSst/gateway-api/graph/model\"\n"), nil, 1)
	}
	return format.Source(src)
}

func isWellKnown(md protoreflect.MessageDescriptor) bool {
	return md.ParentFile().Package() == "google.protobuf"
}

func lowerFirst(s string) string {
	if s == "" {
		return s
	}
	return strings.ToLower(s[:1]) + s[1:]
}

func contains(list []string, s string) bool {
	for _, v := range list {
		if v == s {
			return true
		}
	}
	return false
}

func sortedKeys[V any](m map[protoreflect.FullName]V) []protoreflect.FullName {
	keys := make([]protoreflect.FullName, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return keys[i] < keys[j] })
	return keys
}
