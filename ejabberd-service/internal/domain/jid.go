package domain

import (
	"errors"
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var ErrInvalidJID = errors.New("invalid jid")

const maxJIDPartSize = 1023

type LJID struct {
	User     string
	Server   string
	Resource string
}

func (l LJID) Bare() LJID {
	return LJID{User: l.User, Server: l.Server}
}

func (l LJID) String() string {
	return encodeJID(l.User, l.Server, l.Resource)
}

func (l LJID) JID() JID {
	return JID{User: l.User, Server: l.Server, Resource: l.Resource, LUser: l.User, LServer: l.Server, LResource: l.Resource}
}

type JID struct {
	User      string
	Server    string
	Resource  string
	LUser     string
	LServer   string
	LResource string
}

func NewJID(user, server, resource string) (JID, error) {
	lu, ok := Nodeprep(user)
	if !ok {
		return JID{}, ErrInvalidJID
	}
	ls, ok := Nameprep(server)
	if !ok {
		return JID{}, ErrInvalidJID
	}
	lr, ok := Resourceprep(resource)
	if !ok {
		return JID{}, ErrInvalidJID
	}
	return JID{User: user, Server: server, Resource: resource, LUser: lu, LServer: ls, LResource: lr}, nil
}

func MustJID(user, server, resource string) JID {
	j, err := NewJID(user, server, resource)
	if err != nil {
		panic(err)
	}
	return j
}

func ParseJID(s string) (JID, error) {
	if s == "" {
		return JID{}, ErrInvalidJID
	}
	rest, resource, hasRes := strings.Cut(s, "/")
	if hasRes && resource == "" {
		return JID{}, ErrInvalidJID
	}
	user, server, hasUser := strings.Cut(rest, "@")
	if !hasUser {
		user, server = "", rest
	} else if user == "" {
		return JID{}, ErrInvalidJID
	}
	return NewJID(user, server, resource)
}

func (j JID) IsZero() bool {
	return j.LServer == ""
}

func (j JID) IsBare() bool {
	return j.LResource == ""
}

func (j JID) Bare() JID {
	return JID{User: j.User, Server: j.Server, LUser: j.LUser, LServer: j.LServer}
}

func (j JID) WithResource(resource string) (JID, error) {
	return NewJID(j.User, j.Server, resource)
}

func (j JID) LJID() LJID {
	return LJID{User: j.LUser, Server: j.LServer, Resource: j.LResource}
}

func (j JID) String() string {
	return encodeJID(j.User, j.Server, j.Resource)
}

func (j JID) Equal(o JID) bool {
	return j.LJID() == o.LJID()
}

func encodeJID(user, server, resource string) string {
	var b strings.Builder
	if user != "" {
		b.WriteString(user)
		b.WriteByte('@')
	}
	b.WriteString(server)
	if resource != "" {
		b.WriteByte('/')
		b.WriteString(resource)
	}
	return b.String()
}

func Nodeprep(s string) (string, bool) {
	if s == "" {
		return "", true
	}
	out, ok := prep(s, true)
	if !ok {
		return "", false
	}
	for _, r := range out {
		if strings.ContainsRune("\"&'/:<>@", r) || unicode.IsSpace(r) {
			return "", false
		}
	}
	return out, true
}

func Nameprep(s string) (string, bool) {
	if s == "" {
		return "", false
	}
	out, ok := prep(strings.TrimSuffix(s, "."), true)
	if !ok || out == "" {
		return "", false
	}
	for _, r := range out {
		if strings.ContainsRune("/@", r) || unicode.IsSpace(r) {
			return "", false
		}
	}
	return out, true
}

func Resourceprep(s string) (string, bool) {
	if s == "" {
		return "", true
	}
	return prep(s, false)
}

func prep(s string, fold bool) (string, bool) {
	if !utf8.ValidString(s) {
		return "", false
	}
	out := norm.NFKC.String(s)
	if fold {
		out = strings.ToLower(out)
	}
	if len(out) > maxJIDPartSize {
		return "", false
	}
	for _, r := range out {
		if unicode.IsControl(r) || r == utf8.RuneError {
			return "", false
		}
	}
	return out, true
}
