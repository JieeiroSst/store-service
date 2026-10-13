package credential

import (
	"context"
	"net/http"
	"strings"
)

type Credential struct {
	Bearer   string
	Internal string
}

func (c Credential) Empty() bool { return c.Bearer == "" && c.Internal == "" }

func (c Credential) Apply(h http.Header) {
	switch {
	case c.Bearer != "":
		h.Set("Authorization", "Bearer "+c.Bearer)
	case c.Internal != "":
		h.Set("X-Internal-Token", c.Internal)
	}
}

func FromRequest(r *http.Request) Credential {
	if t := strings.TrimSpace(r.Header.Get("X-Internal-Token")); t != "" {
		return Credential{Internal: t}
	}
	if t, ok := strings.CutPrefix(strings.TrimSpace(r.Header.Get("Authorization")), "Bearer "); ok {
		return Credential{Bearer: strings.TrimSpace(t)}
	}
	return Credential{}
}

type key struct{}

func With(ctx context.Context, c Credential) context.Context {
	if c.Empty() {
		return ctx
	}
	return context.WithValue(ctx, key{}, c)
}

func From(ctx context.Context) Credential {
	c, _ := ctx.Value(key{}).(Credential)
	return c
}
