package graph

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"reflect"
	"strconv"
	"strings"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/proxy"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
	"github.com/go-viper/mapstructure/v2"
)

func restCall[T any](ctx context.Context, p *proxy.Proxy, service string, req rest.Request, unwrap string) (T, error) {
	var raw any
	if err := p.REST(service).Do(ctx, req, &raw); err != nil {
		var zero T
		return zero, err
	}
	return restDecode[T](ctx, service, req.Path, dig(raw, unwrap))
}

func restOwned[T any](ctx context.Context, p *proxy.Proxy, service string, req rest.Request, unwrap string, owner string, check *rest.Request, checkUnwrap string) (T, error) {
	var zero T
	uid := userIDOf(ctx)
	if uid == "" {
		return zero, notOwned(ctx)
	}
	if check != nil {
		var existing any
		if err := p.REST(service).Do(ctx, *check, &existing); err != nil {
			return zero, err
		}
		if !ownedBy(dig(existing, checkUnwrap), owner, uid) {
			return zero, notOwned(ctx)
		}
		return restCall[T](ctx, p, service, req, unwrap)
	}
	var raw any
	if err := p.REST(service).Do(ctx, req, &raw); err != nil {
		return zero, err
	}
	data := dig(raw, unwrap)
	if !ownedBy(data, owner, uid) {
		return zero, notOwned(ctx)
	}
	return restDecode[T](ctx, service, req.Path, data)
}

func ownedBy(data any, field, uid string) bool {
	m, ok := data.(map[string]any)
	if !ok {
		return false
	}
	s, ok := scalarString(m[field])
	return ok && s == uid
}

func restMultipart[T any](ctx context.Context, p *proxy.Proxy, service string, req rest.Request, fields map[string]any, files []rest.File, unwrap string) (T, error) {
	values := map[string][]string{}
	for k, v := range fields {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				if s, ok := scalarString(rv.Index(i).Interface()); ok {
					values[k] = append(values[k], s)
				}
			}
			continue
		}
		if s, ok := scalarString(v); ok {
			values[k] = []string{s}
		}
	}
	var raw any
	if err := p.REST(service).Multipart(ctx, req, values, files, &raw); err != nil {
		var zero T
		return zero, err
	}
	return restDecode[T](ctx, service, req.Path, dig(raw, unwrap))
}

func restRaw[T any](ctx context.Context, p *proxy.Proxy, service string, req rest.Request, upload graphql.Upload, unwrap string) (T, error) {
	var raw any
	if err := p.REST(service).Raw(ctx, req, upload, &raw); err != nil {
		var zero T
		return zero, err
	}
	return restDecode[T](ctx, service, req.Path, dig(raw, unwrap))
}

func restText[T any](ctx context.Context, p *proxy.Proxy, service string, req rest.Request, text string, unwrap string) (T, error) {
	upload := graphql.Upload{File: strings.NewReader(text), ContentType: "text/plain; charset=utf-8", Size: int64(len(text))}
	return restRaw[T](ctx, p, service, req, upload, unwrap)
}

func restDecode[T any](ctx context.Context, service, path string, raw any) (T, error) {
	var out T
	switch target := any(&out).(type) {
	case *bool:
		if b, ok := raw.(bool); ok {
			*target = b
		} else {
			*target = true
		}
		return out, nil
	case *string:
		switch v := raw.(type) {
		case string:
			*target = v
		case nil:
		default:
			b, _ := json.Marshal(v)
			*target = string(b)
		}
		return out, nil
	}
	if raw == nil {
		return out, nil
	}
	dec, err := mapstructure.NewDecoder(&mapstructure.DecoderConfig{
		TagName:          "json",
		WeaklyTypedInput: true,
		Result:           &out,
		DecodeHook:       mapstructure.ComposeDecodeHookFunc(jsonNumberHook, stringifyHook),
	})
	if err != nil {
		return out, internalError(ctx, service, path, err)
	}
	if err := dec.Decode(raw); err != nil {
		return out, internalError(ctx, service, path, err)
	}
	return out, nil
}

func jsonNumberHook(from, to reflect.Type, data any) (any, error) {
	n, ok := data.(json.Number)
	if !ok {
		return data, nil
	}
	switch to.Kind() {
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if i, err := n.Int64(); err == nil {
			return i, nil
		}
		f, err := n.Float64()
		return int64(f), err
	case reflect.Float32, reflect.Float64:
		return n.Float64()
	case reflect.String:
		return n.String(), nil
	case reflect.Bool:
		return n.String() != "0", nil
	}
	return data, nil
}

func stringifyHook(from, to reflect.Type, data any) (any, error) {
	if to.Kind() != reflect.String {
		return data, nil
	}
	switch data.(type) {
	case map[string]any, []any:
		b, err := json.Marshal(data)
		return string(b), err
	}
	return data, nil
}

func dig(raw any, path string) any {
	if path == "" {
		return raw
	}
	for _, key := range strings.Split(path, ".") {
		m, ok := raw.(map[string]any)
		if !ok {
			return nil
		}
		raw = m[key]
	}
	return raw
}

func restHeader(pairs ...any) http.Header {
	h := http.Header{}
	for i := 0; i+1 < len(pairs); i += 2 {
		if s, ok := scalarString(pairs[i+1]); ok && s != "" {
			h.Set(pairs[i].(string), s)
		}
	}
	return h
}

func restQuery(pairs ...any) url.Values {
	q := url.Values{}
	for i := 0; i+1 < len(pairs); i += 2 {
		addQuery(q, pairs[i].(string), pairs[i+1])
	}
	return q
}

func mergeQuery(q url.Values, fixed string) url.Values {
	if q == nil {
		q = url.Values{}
	}
	extra := map[string]any{}
	_ = json.Unmarshal([]byte(fixed), &extra)
	for k, v := range extra {
		q.Del(k)
		addQuery(q, k, v)
	}
	return q
}

func addQuery(q url.Values, name string, v any) {
	rv := reflect.ValueOf(v)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return
	}
	switch rv.Kind() {
	case reflect.Struct, reflect.Map:
		b, err := json.Marshal(rv.Interface())
		if err != nil {
			return
		}
		var m map[string]any
		if json.Unmarshal(b, &m) != nil {
			return
		}
		for k, val := range m {
			addQuery(q, k, val)
		}
	case reflect.Slice, reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if s, ok := scalarString(rv.Index(i).Interface()); ok {
				q.Add(name, s)
			}
		}
	default:
		if s, ok := scalarString(rv.Interface()); ok {
			q.Set(name, s)
		}
	}
}

func scalarString(v any) (string, bool) {
	rv := reflect.ValueOf(v)
	for rv.IsValid() && rv.Kind() == reflect.Pointer {
		if rv.IsNil() {
			return "", false
		}
		rv = rv.Elem()
	}
	if !rv.IsValid() {
		return "", false
	}
	switch rv.Kind() {
	case reflect.String:
		return rv.String(), true
	case reflect.Bool:
		return strconv.FormatBool(rv.Bool()), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return strconv.FormatInt(rv.Int(), 10), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return strconv.FormatUint(rv.Uint(), 10), true
	case reflect.Float32, reflect.Float64:
		return strconv.FormatFloat(rv.Float(), 'f', -1, 64), true
	case reflect.Map, reflect.Slice, reflect.Struct:
		b, err := json.Marshal(rv.Interface())
		return string(b), err == nil
	}
	return fmt.Sprint(rv.Interface()), true
}

func uploadFiles(pairs ...any) []rest.File {
	var files []rest.File
	for i := 0; i+1 < len(pairs); i += 2 {
		field := pairs[i].(string)
		switch u := pairs[i+1].(type) {
		case graphql.Upload:
			files = append(files, rest.File{Field: field, Upload: u})
		case *graphql.Upload:
			if u != nil {
				files = append(files, rest.File{Field: field, Upload: *u})
			}
		case []*graphql.Upload:
			for _, f := range u {
				if f != nil {
					files = append(files, rest.File{Field: field, Upload: *f})
				}
			}
		}
	}
	return files
}
