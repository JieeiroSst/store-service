package domain

import (
	"net"
	"regexp"
)

type ACLKind string

const (
	ACLRef            ACLKind = "acl"
	ACLUser           ACLKind = "user"
	ACLServer         ACLKind = "server"
	ACLResource       ACLKind = "resource"
	ACLUserRegexp     ACLKind = "user_regexp"
	ACLServerRegexp   ACLKind = "server_regexp"
	ACLResourceRegexp ACLKind = "resource_regexp"
	ACLNodeRegexp     ACLKind = "node_regexp"
	ACLUserGlob       ACLKind = "user_glob"
	ACLServerGlob     ACLKind = "server_glob"
	ACLResourceGlob   ACLKind = "resource_glob"
	ACLNodeGlob       ACLKind = "node_glob"
	ACLIP             ACLKind = "ip"
)

type ACLSpec struct {
	Kind     ACLKind
	Name     string
	User     string
	Server   string
	Resource string
	UserRE   *regexp.Regexp
	ServerRE *regexp.Regexp
	Net      *net.IPNet
}

type AccessClause struct {
	Allow bool
	Specs []ACLSpec
}

type AccessRule []AccessClause

type ShaperClause struct {
	Value string
	Specs []ACLSpec
}

type ShaperRule []ShaperClause

type Match struct {
	USR *LJID
	IP  net.IP
}

func MatchJID(j JID) Match {
	l := j.LJID()
	return Match{USR: &l}
}

func MatchJIDIP(j JID, ip net.IP) Match {
	l := j.LJID()
	return Match{USR: &l, IP: ip}
}

func MatchIP(ip net.IP) Match {
	return Match{IP: ip}
}

type ACLMatcher struct {
	ACLs        map[string][]ACLSpec
	AccessRules map[string]AccessRule
	ShaperRules map[string]ShaperRule
	IsMyHost    func(string) bool
}

func (m *ACLMatcher) MatchRule(name string, match Match) bool {
	switch name {
	case "all", "":
		return true
	case "none":
		return false
	}
	rule, ok := m.AccessRules[name]
	if !ok {
		return false
	}
	return m.MatchAccess(rule, match)
}

func (m *ACLMatcher) MatchAccess(rule AccessRule, match Match) bool {
	for _, c := range rule {
		if m.matchAll(c.Specs, match) {
			return c.Allow
		}
	}
	return false
}

func (m *ACLMatcher) MatchShaper(name string, match Match) string {
	switch name {
	case "none", "infinity":
		return name
	}
	rule, ok := m.ShaperRules[name]
	if !ok {
		return name
	}
	for _, c := range rule {
		if m.matchAll(c.Specs, match) {
			return c.Value
		}
	}
	return "none"
}

func (m *ACLMatcher) matchAll(specs []ACLSpec, match Match) bool {
	if len(specs) == 0 {
		return false
	}
	for _, s := range specs {
		if !m.matchSpec(s, match, 0) {
			return false
		}
	}
	return true
}

func (m *ACLMatcher) matchSpec(s ACLSpec, match Match, depth int) bool {
	if depth > 16 {
		return false
	}
	if s.Kind == ACLRef {
		switch s.Name {
		case "all":
			return true
		case "none":
			return false
		}
		for _, sub := range m.ACLs[s.Name] {
			if m.matchSpec(sub, match, depth+1) {
				return true
			}
		}
		return false
	}
	if s.Kind == ACLIP {
		return match.IP != nil && s.Net != nil && s.Net.Contains(match.IP)
	}
	if match.USR == nil {
		return false
	}
	u, srv, r := match.USR.User, match.USR.Server, match.USR.Resource
	isMy := func(h string) bool { return m.IsMyHost != nil && m.IsMyHost(h) }
	switch s.Kind {
	case ACLUser:
		if s.Server != "" {
			return u == s.User && srv == s.Server
		}
		return u == s.User && isMy(srv)
	case ACLServer:
		return srv == s.Server
	case ACLResource:
		return r == s.Resource
	case ACLUserRegexp, ACLUserGlob:
		if s.Server != "" {
			return srv == s.Server && s.UserRE.MatchString(u)
		}
		return isMy(srv) && s.UserRE.MatchString(u)
	case ACLServerRegexp, ACLServerGlob:
		return s.ServerRE.MatchString(srv)
	case ACLResourceRegexp, ACLResourceGlob:
		return s.UserRE.MatchString(r)
	case ACLNodeRegexp, ACLNodeGlob:
		return s.UserRE.MatchString(u) && s.ServerRE.MatchString(srv)
	}
	return false
}
