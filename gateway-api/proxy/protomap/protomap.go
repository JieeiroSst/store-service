package protomap

import (
	"encoding/base64"
	"math"
	"time"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/durationpb"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func ToMap(m proto.Message) map[string]any {
	if m == nil {
		return nil
	}
	return message(m.ProtoReflect())
}

func message(m protoreflect.Message) map[string]any {
	out := map[string]any{}
	fields := m.Descriptor().Fields()
	for i := 0; i < fields.Len(); i++ {
		fd := fields.Get(i)
		name := fd.JSONName()
		switch {
		case fd.IsList():
			list := m.Get(fd).List()
			items := make([]any, 0, list.Len())
			for j := 0; j < list.Len(); j++ {
				items = append(items, value(fd, list.Get(j)))
			}
			out[name] = items
		case fd.IsMap():
			mp := map[string]any{}
			m.Get(fd).Map().Range(func(k protoreflect.MapKey, v protoreflect.Value) bool {
				mp[k.String()] = value(fd.MapValue(), v)
				return true
			})
			out[name] = mp
		case fd.HasPresence() && !m.Has(fd):
			out[name] = nil
		default:
			out[name] = value(fd, m.Get(fd))
		}
	}
	return out
}

func value(fd protoreflect.FieldDescriptor, v protoreflect.Value) any {
	switch fd.Kind() {
	case protoreflect.BoolKind:
		return v.Bool()
	case protoreflect.Int32Kind, protoreflect.Sint32Kind, protoreflect.Sfixed32Kind,
		protoreflect.Int64Kind, protoreflect.Sint64Kind, protoreflect.Sfixed64Kind:
		return v.Int()
	case protoreflect.Uint32Kind, protoreflect.Fixed32Kind, protoreflect.Uint64Kind, protoreflect.Fixed64Kind:
		return v.Uint()
	case protoreflect.FloatKind, protoreflect.DoubleKind:
		f := v.Float()
		if math.IsNaN(f) || math.IsInf(f, 0) {
			return nil
		}
		return f
	case protoreflect.StringKind:
		return v.String()
	case protoreflect.BytesKind:
		return base64.StdEncoding.EncodeToString(v.Bytes())
	case protoreflect.EnumKind:
		if ev := fd.Enum().Values().ByNumber(v.Enum()); ev != nil {
			return string(ev.Name())
		}
		return nil
	case protoreflect.MessageKind, protoreflect.GroupKind:
		return wellKnown(v.Message())
	}
	return nil
}

func wellKnown(m protoreflect.Message) any {
	switch m.Descriptor().FullName() {
	case "google.protobuf.Timestamp":
		ts := &timestamppb.Timestamp{}
		proto.Merge(ts, m.Interface())
		return ts.AsTime().UTC().Format(time.RFC3339Nano)
	case "google.protobuf.Duration":
		d := &durationpb.Duration{}
		proto.Merge(d, m.Interface())
		return d.AsDuration().String()
	case "google.protobuf.Struct":
		s := &structpb.Struct{}
		proto.Merge(s, m.Interface())
		return s.AsMap()
	case "google.protobuf.Value":
		s := &structpb.Value{}
		proto.Merge(s, m.Interface())
		return s.AsInterface()
	case "google.protobuf.ListValue":
		s := &structpb.ListValue{}
		proto.Merge(s, m.Interface())
		return s.AsSlice()
	case "google.protobuf.DoubleValue", "google.protobuf.FloatValue", "google.protobuf.Int64Value",
		"google.protobuf.UInt64Value", "google.protobuf.Int32Value", "google.protobuf.UInt32Value",
		"google.protobuf.BoolValue", "google.protobuf.StringValue", "google.protobuf.BytesValue":
		fd := m.Descriptor().Fields().ByName("value")
		return value(fd, m.Get(fd))
	case "google.protobuf.Any":
		fd := m.Descriptor().Fields().ByName("type_url")
		return map[string]any{"@type": m.Get(fd).String()}
	}
	return message(m)
}
