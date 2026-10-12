package application

import (
	"context"
	"log/slog"
	"slices"
	"strings"
	"sync"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

type SubscriptionHandler interface {
	InSubscription(ctx context.Context, presence *domain.Element) bool
}

type OfflineHandler interface {
	StorePacket(ctx context.Context, msg *domain.Element) OfflineResult
}

type UserChecker interface {
	UserExists(ctx context.Context, luser, lserver string) bool
}

type OfflineResult int

const (
	OfflinePass OfflineResult = iota
	OfflineStored
	OfflineStop
)

type ServiceHandler func(ctx context.Context, stanza *domain.Element)

type ServiceRoute struct {
	Host    string
	Name    string
	Handler ServiceHandler
}

type Router struct {
	servicesMu   sync.RWMutex
	services     map[string]ServiceRoute
	cfg          *config.Config
	sm           *SessionManager
	iq           *IQRegistry
	subscription SubscriptionHandler
	offline      OfflineHandler
	users        UserChecker
}

var _ port.Router = (*Router)(nil)

func NewRouter(cfg *config.Config, sm *SessionManager, iq *IQRegistry) *Router {
	return &Router{cfg: cfg, sm: sm, iq: iq}
}

func (r *Router) RegisterService(route ServiceRoute) {
	r.servicesMu.Lock()
	defer r.servicesMu.Unlock()
	if r.services == nil {
		r.services = map[string]ServiceRoute{}
	}
	r.services[route.Host] = route
}

func (r *Router) Services(serverHost string) []ServiceRoute {
	r.servicesMu.RLock()
	defer r.servicesMu.RUnlock()
	var out []ServiceRoute
	for _, s := range r.services {
		if strings.HasSuffix(s.Host, "."+serverHost) {
			out = append(out, s)
		}
	}
	slices.SortFunc(out, func(a, b ServiceRoute) int { return strings.Compare(a.Host, b.Host) })
	return out
}

func (r *Router) service(host string) (ServiceRoute, bool) {
	r.servicesMu.RLock()
	defer r.servicesMu.RUnlock()
	s, ok := r.services[host]
	return s, ok
}

func (r *Router) Use(sub SubscriptionHandler, off OfflineHandler, users UserChecker) {
	r.subscription, r.offline, r.users = sub, off, users
}

func (r *Router) Route(ctx context.Context, stanza *domain.Element) {
	to, ok := domain.StanzaTo(stanza)
	if !ok {
		slog.Debug("dropping stanza without valid to", "stanza", stanza.String())
		return
	}
	if s, ok := r.service(to.LServer); ok {
		s.Handler(ctx, stanza)
		return
	}
	if !r.cfg.IsMyHost(to.LServer) {
		r.RouteError(ctx, stanza, domain.ErrRemoteServerNotFound("Server-to-server connections are not supported"))
		return
	}
	r.localRoute(ctx, stanza, to)
}

func (r *Router) RouteError(ctx context.Context, stanza *domain.Element, err domain.StanzaError) {
	switch domain.StanzaType(stanza) {
	case "error", "result":
		return
	}
	if stanza.Attr("from") == "" {
		return
	}
	r.Route(ctx, domain.MakeError(stanza, err))
}

func (r *Router) localRoute(ctx context.Context, stanza *domain.Element, to domain.JID) {
	stype := domain.StanzaType(stanza)
	switch {
	case to.LUser != "":
		r.smRoute(ctx, stanza, to)
	case stanza.Name == domain.KindIQ && to.LResource == "":
		if stype == "get" || stype == "set" {
			if res := r.iq.Handle(ctx, ScopeLocal, stanza); res != nil {
				r.Route(ctx, res)
			}
		}
	case stype == "result" || stype == "error":
	default:
		r.bounceResourcePacket(ctx, stanza, to)
	}
}

func (r *Router) bounceResourcePacket(ctx context.Context, stanza *domain.Element, to domain.JID) {
	if to.LResource == "" && (stanza.Name == domain.KindPresence ||
		(stanza.Name == domain.KindMessage && domain.StanzaType(stanza) == "headline")) {
		return
	}
	r.RouteError(ctx, stanza, domain.ErrItemNotFound("No available resource found"))
}

func (r *Router) deliver(sink port.SessionSink, stanza *domain.Element) {
	sink.Deliver(port.SessionEvent{Kind: port.EventRoute, Stanza: stanza})
}

func (r *Router) smRoute(ctx context.Context, stanza *domain.Element, to domain.JID) {
	stype := domain.StanzaType(stanza)
	switch {
	case stanza.Name == domain.KindPresence && domain.IsSubscriptionPresence(stype):
		if r.subscription == nil || !r.subscription.InSubscription(ctx, stanza) {
			return
		}
		for _, e := range r.sm.rawSessions(to.LUser, to.LServer) {
			if e.session.Priority == nil {
				continue
			}
			full, err := to.WithResource(e.session.USR.Resource)
			if err != nil {
				continue
			}
			p := stanza.Copy()
			p.SetAttr("to", full.String())
			r.deliver(e.sink, p)
		}
	case stanza.Name == domain.KindPresence && to.LResource == "":
		for _, pr := range r.sm.PresentResources(to.LUser, to.LServer) {
			full, err := to.WithResource(pr.Resource)
			if err != nil {
				continue
			}
			p := stanza.Copy()
			p.SetAttr("to", full.String())
			r.smRoute(ctx, p, full)
		}
	case stanza.Name == domain.KindMessage && to.LResource == "":
		switch stype {
		case "chat", "headline", "normal":
			r.routeMessage(ctx, stanza, to)
		default:
			r.bounceSMPacket(ctx, stanza, true)
		}
	case stanza.Name == domain.KindIQ && to.LResource == "":
		if stype == "get" || stype == "set" {
			if res := r.iq.Handle(ctx, ScopeSM, stanza); res != nil {
				r.Route(ctx, res)
			}
		}
	default:
		e := r.sm.resourceEntry(to.LJID())
		if e != nil {
			r.deliver(e.sink, stanza)
			return
		}
		switch {
		case stanza.Name == domain.KindMessage && (stype == "chat" || stype == "normal"):
			r.routeMessage(ctx, stanza, to)
		case stanza.Name == domain.KindMessage && stype == "headline":
			r.bounceSMPacket(ctx, stanza, false)
		case stanza.Name == domain.KindPresence:
			r.bounceSMPacket(ctx, stanza, false)
		default:
			r.bounceSMPacket(ctx, stanza, true)
		}
	}
}

func (r *Router) bounceSMPacket(ctx context.Context, stanza *domain.Element, bounce bool) {
	if !bounce {
		slog.Debug("dropping packet to unavailable resource", "stanza", stanza.String())
		return
	}
	r.RouteError(ctx, stanza, domain.ErrServiceUnavailable("User session not found"))
}

func (r *Router) routeMessage(ctx context.Context, msg *domain.Element, to domain.JID) {
	stype := domain.StanzaType(msg)
	present := r.sm.PresentResources(to.LUser, to.LServer)
	maxPrio, found := 0, false
	for _, p := range present {
		if !found || p.Priority > maxPrio {
			maxPrio, found = p.Priority, true
		}
	}
	if found && maxPrio >= 0 {
		for _, p := range present {
			if p.Priority != maxPrio && !(p.Priority >= 0 && stype == "headline") {
				continue
			}
			full, err := to.WithResource(p.Resource)
			if err != nil {
				continue
			}
			if e := r.sm.resourceEntry(full.LJID()); e != nil {
				r.deliver(e.sink, msg.Copy())
			}
		}
		return
	}
	if r.users != nil && r.users.UserExists(ctx, to.LUser, to.LServer) {
		result := OfflinePass
		if r.offline != nil {
			result = r.offline.StorePacket(ctx, msg)
		}
		if result == OfflinePass && slices.Contains([]string{"chat", "groupchat", "normal"}, stype) {
			r.bounceSMPacket(ctx, msg, true)
		}
		return
	}
	r.RouteError(ctx, msg, domain.ErrServiceUnavailable(""))
}
