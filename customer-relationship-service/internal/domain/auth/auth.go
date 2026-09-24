package auth

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
)

type Role string

const (
	Viewer  Role = "viewer"
	Staff   Role = "staff"
	Manager Role = "manager"
	Admin   Role = "admin"
)

var Roles = []Role{Viewer, Staff, Manager, Admin}

func (r Role) rank() int {
	for i, x := range Roles {
		if x == r {
			return i + 1
		}
	}
	return 0
}

func ParseRole(name, prefix string) (Role, bool) {
	name = strings.ToLower(strings.TrimSpace(name))
	if prefix != "" {
		if !strings.HasPrefix(name, strings.ToLower(prefix)) {
			return "", false
		}
		name = strings.TrimPrefix(name, strings.ToLower(prefix))
	}
	for _, r := range Roles {
		if string(r) == name {
			return r, true
		}
	}
	return "", false
}

type Action string

const (
	FileRead      Action = "file.read"
	FileUpload    Action = "file.upload"
	FileDeleteOwn Action = "file.delete.own"
	FileDeleteAny Action = "file.delete.any"
	FileAudit     Action = "file.audit"
)

var restricted = map[string]bool{model.FileKindLegal: true}

type Principal struct {
	Subject string
	Name    string
	Roles   []Role
	Service bool
	IP      string
}

func (p Principal) Highest() Role {
	var best Role
	for _, r := range p.Roles {
		if r.rank() > best.rank() {
			best = r
		}
	}
	return best
}

func (p Principal) atLeast(r Role) bool { return p.Highest().rank() >= r.rank() }

func (p Principal) Can(a Action, kind string) bool {
	if restricted[kind] && !p.atLeast(Manager) {
		return false
	}
	switch a {
	case FileRead:
		return p.atLeast(Viewer)
	case FileUpload, FileDeleteOwn:
		return p.atLeast(Staff)
	case FileDeleteAny, FileAudit:
		return p.atLeast(Manager)
	}
	return false
}

func (p Principal) HiddenKinds() []string {
	var out []string
	for k := range restricted {
		if !p.Can(FileRead, k) {
			out = append(out, k)
		}
	}
	return out
}

func (p Principal) Display() string {
	if p.Name != "" {
		return p.Name
	}
	return p.Subject
}

type ctxKey struct{}

func WithPrincipal(ctx context.Context, p Principal) context.Context {
	return context.WithValue(ctx, ctxKey{}, p)
}

func FromContext(ctx context.Context) (Principal, bool) {
	p, ok := ctx.Value(ctxKey{}).(Principal)
	return p, ok
}
