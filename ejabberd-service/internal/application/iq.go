package application

import (
	"context"
	"slices"
	"sync"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

type IQHandler func(ctx context.Context, iq *domain.Element) *domain.Element

type IQScope int

const (
	ScopeLocal IQScope = iota
	ScopeSM
)

type IQRegistry struct {
	mu       sync.RWMutex
	handlers map[IQScope]map[string]IQHandler
}

func NewIQRegistry() *IQRegistry {
	return &IQRegistry{handlers: map[IQScope]map[string]IQHandler{ScopeLocal: {}, ScopeSM: {}}}
}

func (r *IQRegistry) Add(scope IQScope, ns string, h IQHandler) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.handlers[scope][ns] = h
}

func (r *IQRegistry) Features(scope IQScope) []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]string, 0, len(r.handlers[scope]))
	for ns := range r.handlers[scope] {
		out = append(out, ns)
	}
	slices.Sort(out)
	return out
}

func (r *IQRegistry) Handle(ctx context.Context, scope IQScope, iq *domain.Element) *domain.Element {
	els := iq.Elements()
	if len(els) != 1 {
		return domain.MakeError(iq, domain.ErrBadRequest("Invalid number of sub-elements"))
	}
	r.mu.RLock()
	h, ok := r.handlers[scope][els[0].Namespace()]
	r.mu.RUnlock()
	if !ok {
		return domain.MakeError(iq, domain.ErrServiceUnavailable("No module is handling this query"))
	}
	return h(ctx, iq)
}
