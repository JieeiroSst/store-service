package httpapi

import (
	"slices"
	"strings"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

type caller struct {
	match  domain.Match
	module string
}

func canAccess(cfg *config.Config, matcher *domain.ACLMatcher, cmd string, c caller) bool {
	for _, p := range cfg.APIPermissions {
		if len(p.From) > 0 && !slices.Contains(p.From, c.module) {
			continue
		}
		if !whoMatches(matcher, p.Who, c.match) {
			continue
		}
		if whatMatches(p.What, cmd) {
			return true
		}
	}
	return false
}

func whoMatches(matcher *domain.ACLMatcher, who []config.WhoClause, m domain.Match) bool {
	for _, w := range who {
		switch {
		case w.All:
			return true
		case w.Access != nil:
			if matcher.MatchAccess(w.Access, m) {
				return true
			}
		case len(w.Specs) > 0:
			if matcher.MatchAccess(domain.AccessRule{{Allow: true, Specs: w.Specs}}, m) {
				return true
			}
		}
	}
	return false
}

func whatMatches(what []string, cmd string) bool {
	allowed := false
	for _, w := range what {
		switch {
		case strings.HasPrefix(w, "!"):
			if strings.TrimPrefix(w, "!") == cmd {
				return false
			}
		case w == "*" || w == cmd:
			allowed = true
		}
	}
	return allowed
}
