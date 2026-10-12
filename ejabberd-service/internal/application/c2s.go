package application

import (
	"context"
	"log/slog"
	"net"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

type C2SService struct {
	cfg     *config.Config
	matcher *domain.ACLMatcher
	sm      *SessionManager
	router  *Router
	roster  *RosterService
	offline *OfflineService
	uniq    atomic.Uint64
	wg      sync.WaitGroup
}

func NewC2SService(cfg *config.Config, matcher *domain.ACLMatcher, sm *SessionManager, router *Router, roster *RosterService, offline *OfflineService) *C2SService {
	return &C2SService{cfg: cfg, matcher: matcher, sm: sm, router: router, roster: roster, offline: offline}
}

type BindParams struct {
	User       string
	Server     string
	Resource   string
	IP         net.IP
	Conn       string
	AuthModule string
	Access     string
	Lang       string
	Stream     port.StreamConn
}

func (s *C2SService) newUniqID() string {
	return randString() + strconv.FormatUint(s.uniq.Add(1), 10)
}

func (s *C2SService) resourceConflictAction(user, server, resource string) (string, bool) {
	j, err := domain.NewJID(user, server, resource)
	if err != nil {
		return "", false
	}
	option := domain.ConflictAcceptNew
	if s.sm.IsExistingResource(j.LJID()) {
		option = s.cfg.ResourceConflict
	}
	switch option {
	case domain.ConflictCloseNew:
		return "", false
	case domain.ConflictSetResource:
		return s.newUniqID(), true
	}
	return resource, true
}

func (s *C2SService) Bind(ctx context.Context, p BindParams) (*C2SSession, *domain.StanzaError) {
	if p.Resource == "" {
		p.Resource = s.newUniqID()
	}
	resource, ok := s.resourceConflictAction(p.User, p.Server, p.Resource)
	if !ok {
		e := domain.ErrConflict("")
		return nil, &e
	}
	jid, err := domain.NewJID(p.User, p.Server, resource)
	if err != nil {
		e := domain.ErrBadRequest("Malformed resource")
		return nil, &e
	}
	if !s.matcher.MatchRule(p.Access, domain.MatchJIDIP(jid, p.IP)) {
		slog.Warn("forbidden c2s session", "jid", jid.String())
		e := domain.ErrNotAllowedStanza("Access denied by service policy")
		return nil, &e
	}
	sess := &C2SSession{
		svc:    s,
		jid:    jid,
		sid:    s.sm.MakeSID(),
		lang:   p.Lang,
		ip:     p.IP,
		stream: p.Stream,
		box:    newMailbox(),
		presA:  map[domain.LJID]struct{}{},
		done:   make(chan struct{}),
	}
	s.sm.OpenSession(sess.sid, jid, nil, domain.SessionInfo{IP: p.IP, Conn: p.Conn, AuthModule: p.AuthModule}, sess)
	slog.Info("opened c2s session", "jid", jid.String(), "ip", p.IP.String())
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		sess.loop(context.WithoutCancel(ctx))
	}()
	return sess, nil
}

func (s *C2SService) Wait(ctx context.Context) {
	done := make(chan struct{})
	go func() {
		s.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
	case <-ctx.Done():
	}
}

type C2SSession struct {
	svc          *C2SService
	jid          domain.JID
	sid          domain.SID
	lang         string
	ip           net.IP
	stream       port.StreamConn
	box          *mailbox
	presLast     *domain.Element
	presTS       time.Time
	presA        map[domain.LJID]struct{}
	done         chan struct{}
	lastActivity time.Time
	pingID       string
	pingSent     time.Time
}

var _ port.SessionSink = (*C2SSession)(nil)

func (c *C2SSession) JID() domain.JID { return c.jid }

func (c *C2SSession) Done() <-chan struct{} { return c.done }

func (c *C2SSession) Deliver(ev port.SessionEvent) {
	if !c.box.push(mailboxItem{event: &ev}) && ev.Kind == port.EventRoute && ev.Stanza != nil {
		c.svc.router.Route(context.Background(), ev.Stanza)
	}
}

func (c *C2SSession) HandleClientPacket(el *domain.Element) {
	c.box.push(mailboxItem{client: el})
}

func (c *C2SSession) Terminate(reason string) {
	c.box.push(mailboxItem{terminate: &reason})
}

func (c *C2SSession) loop(ctx context.Context) {
	defer close(c.done)
	c.lastActivity = time.Now()
	var tick <-chan time.Time
	if ping := c.svc.cfg.Modules.Ping; ping.Enabled && ping.SendPings && ping.PingInterval > 0 {
		t := time.NewTicker(ping.PingInterval / 2)
		defer t.Stop()
		tick = t.C
	}
	for {
		select {
		case <-c.box.signal:
		case <-tick:
			c.box.push(mailboxItem{tick: true})
			continue
		}
		for _, it := range c.box.take() {
			if reason, stop := c.handle(ctx, it); stop {
				c.terminate(ctx, reason)
				return
			}
		}
	}
}

func (c *C2SSession) handle(ctx context.Context, it mailboxItem) (string, bool) {
	switch {
	case it.terminate != nil:
		return *it.terminate, true
	case it.tick:
		return c.checkPing(ctx)
	case it.client != nil:
		c.lastActivity = time.Now()
		return c.handleAuthenticatedPacket(ctx, it.client)
	case it.event != nil:
		return c.processInfo(ctx, *it.event)
	}
	return "", false
}

func (c *C2SSession) checkPing(ctx context.Context) (string, bool) {
	ping := c.svc.cfg.Modules.Ping
	now := time.Now()
	if c.pingID != "" {
		if now.Sub(c.pingSent) >= ping.PingAckTimeout {
			c.pingID = ""
			if ping.TimeoutAction == "kill" {
				slog.Info("ping timeout, closing session", "jid", c.jid.String())
				c.stream.Close(nil)
				return "ping timeout", true
			}
		}
		return "", false
	}
	if now.Sub(c.lastActivity) >= ping.PingInterval {
		host, _ := domain.NewJID("", c.jid.LServer, "")
		c.pingID = "ping-" + randString()
		c.pingSent = now
		c.stream.Send(domain.NewIQ("get", c.pingID, host, c.jid, domain.NewElement("ping", domain.NSPing)))
	}
	return "", false
}

func (c *C2SSession) handleAuthenticatedPacket(ctx context.Context, pkt *domain.Element) (string, bool) {
	if !domain.IsStanza(pkt) {
		return "", false
	}
	if pkt.Name == domain.KindIQ && c.pingID != "" && pkt.Attr("id") == c.pingID {
		c.pingID = ""
		return "", false
	}
	if v := pkt.Attr("from"); v != "" {
		from, err := domain.ParseJID(v)
		if err != nil || from.LUser != c.jid.LUser || from.LServer != c.jid.LServer ||
			(from.LResource != "" && from.LResource != c.jid.LResource) {
			serr := domain.StreamErr("invalid-from", "")
			c.stream.Close(&serr)
			return "invalid-from", true
		}
	} else {
		pkt.SetAttr("from", c.jid.String())
	}
	var to domain.JID
	if v := pkt.Attr("to"); v != "" {
		j, err := domain.ParseJID(v)
		if err != nil {
			c.sendError(pkt, domain.ErrJIDMalformed("Malformed 'to' attribute"))
			return "", false
		}
		to = j
	} else {
		to = c.jid.Bare()
		pkt.SetAttr("to", to.String())
	}
	switch {
	case pkt.Name == domain.KindIQ && domain.StanzaType(pkt) == "set" && len(pkt.Elements()) == 1 &&
		pkt.Elements()[0].Name == "session" && pkt.Elements()[0].Namespace() == domain.NSSession:
		res := pkt.Copy()
		host, _ := domain.NewJID("", c.jid.LServer, "")
		res.SetAttr("to", host.String())
		c.stream.Send(domain.MakeIQResult(res, nil))
	case pkt.Name == domain.KindPresence && to.LUser == c.jid.LUser && to.LServer == c.jid.LServer && to.LResource == "":
		c.processSelfPresence(ctx, pkt)
	case pkt.Name == domain.KindPresence:
		c.processPresenceOut(ctx, pkt)
	default:
		c.svc.router.Route(ctx, pkt)
	}
	return "", false
}

func (c *C2SSession) processInfo(ctx context.Context, ev port.SessionEvent) (string, bool) {
	switch ev.Kind {
	case port.EventReplaced:
		serr := domain.StreamErr("conflict", "Replaced by new connection")
		c.stream.Close(&serr)
		return "Replaced by new connection", true
	case port.EventExit:
		serr := domain.StreamErr("policy-violation", ev.Reason)
		c.stream.Close(&serr)
		return ev.Reason, true
	case port.EventResendPresence:
		if c.presLast == nil {
			return "", false
		}
		if ev.To == nil || (ev.To.LUser == c.jid.LUser && ev.To.LServer == c.jid.LServer && ev.To.LResource == "") {
			c.processSelfPresence(ctx, c.presLast.Copy())
			return "", false
		}
		p := c.presLast.Copy()
		p.SetAttr("to", ev.To.String())
		c.processPresenceOut(ctx, p)
	case port.EventRoute:
		pkt := ev.Stanza
		pass := true
		if pkt.Name == domain.KindPresence {
			pass = c.processPresenceIn(ctx, pkt)
		}
		if pass {
			c.stream.Send(pkt)
		}
	}
	return "", false
}

func (c *C2SSession) processPresenceIn(ctx context.Context, pres *domain.Element) bool {
	from, _ := domain.StanzaFrom(pres)
	switch domain.StanzaType(pres) {
	case "probe":
		c.routeProbeReply(ctx, from)
		return false
	case "error":
		delete(c.presA, from.LJID())
	}
	return true
}

func (c *C2SSession) routeProbeReply(ctx context.Context, from domain.JID) {
	if c.presLast == nil || from.IsZero() {
		return
	}
	isAnotherResource := from.LUser == c.jid.LUser && from.LServer == c.jid.LServer && from.LResource != c.jid.LResource
	sub, _, _ := c.svc.roster.GetSubscription(ctx, c.jid, from)
	if isAnotherResource || sub == domain.SubBoth || sub == domain.SubFrom {
		p := c.presLast.Copy()
		domain.SetFromTo(p, c.jid, from)
		p.RemoveChildren("delay", domain.NSDelay)
		domain.AddDelay(p, c.jid, c.presTS, "")
		c.svc.router.Route(ctx, p)
	}
}

func (c *C2SSession) processPresenceOut(ctx context.Context, pres *domain.Element) {
	from, _ := domain.StanzaFrom(pres)
	to, _ := domain.StanzaTo(pres)
	ptype := domain.StanzaType(pres)
	if domain.IsSubscriptionPresence(ptype) {
		if !c.svc.matcher.MatchRule(c.svc.roster.Access(), domain.MatchJID(c.jid.Bare())) {
			c.sendError(pres, domain.ErrForbidden("Access denied by service policy"))
		} else {
			c.svc.roster.OutSubscription(ctx, pres)
		}
	}
	switch ptype {
	case domain.PresSubscribe, domain.PresSubscribed, domain.PresUnsubscribe, domain.PresUnsubscribed:
		p := pres.Copy()
		domain.SetFromTo(p, from.Bare(), to)
		c.svc.router.Route(ctx, p)
	case "error", "probe":
		c.svc.router.Route(ctx, pres)
	default:
		c.svc.router.Route(ctx, pres)
		lto := to.LJID()
		if lto.Bare() == from.LJID().Bare() {
			return
		}
		sub, _, _ := c.svc.roster.GetSubscription(ctx, from, to)
		if sub == domain.SubBoth || sub == domain.SubFrom {
			return
		}
		switch ptype {
		case "available":
			c.presA[lto] = struct{}{}
		case "unavailable":
			delete(c.presA, lto)
		}
	}
}

func (c *C2SSession) processSelfPresence(ctx context.Context, pres *domain.Element) {
	switch domain.StanzaType(pres) {
	case "unavailable":
		c.svc.sm.UnsetPresence(c.sid, c.jid)
		c.broadcastPresenceUnavailable(ctx, pres, true)
		c.presLast = nil
		c.presTS = time.Time{}
	case "available":
		prev := c.presLast
		prio := domain.Priority(pres)
		c.svc.sm.SetPresence(c.sid, c.jid, prio)
		c.selfPresenceHooks(ctx, pres, prev)
		c.presLast = pres
		c.presTS = time.Now()
		c.broadcastPresenceAvailable(ctx, pres, prev == nil)
	}
}

func (c *C2SSession) selfPresenceHooks(ctx context.Context, pres, prev *domain.Element) {
	newPrio := domain.Priority(pres)
	if prev == nil && newPrio >= 0 {
		for _, sub := range c.svc.roster.PendingSubscriptions(ctx, c.jid) {
			c.stream.Send(sub)
		}
	}
	lastPrio := -1
	if prev != nil {
		lastPrio = domain.Priority(prev)
	}
	if lastPrio < 0 && newPrio >= 0 {
		c.svc.offline.RouteOfflineMessages(ctx, c.jid)
	}
}

func (c *C2SSession) routeMultiple(ctx context.Context, jids []domain.JID, pkt *domain.Element) {
	for _, j := range jids {
		p := pkt.Copy()
		p.SetAttr("to", j.String())
		c.svc.router.Route(ctx, p)
	}
}

func (c *C2SSession) broadcastPresenceUnavailable(ctx context.Context, pres *domain.Element, toRoster bool) {
	bare := c.jid.Bare()
	seen := map[domain.LJID]struct{}{bare.LJID(): {}}
	targets := []domain.JID{bare}
	add := func(l domain.LJID) {
		if _, ok := seen[l]; ok {
			return
		}
		seen[l] = struct{}{}
		targets = append(targets, l.JID())
	}
	if toRoster {
		for _, it := range c.svc.roster.GetFilteredRoster(ctx, c.jid.LUser, c.jid.LServer) {
			if it.IsFrom() {
				add(it.JID)
			}
		}
	}
	for l := range c.presA {
		add(l)
	}
	p := pres.Copy()
	p.SetAttr("from", c.jid.String())
	c.routeMultiple(ctx, targets, p)
	c.presA = map[domain.LJID]struct{}{}
}

func (c *C2SSession) broadcastPresenceAvailable(ctx context.Context, pres *domain.Element, fromUnavailable bool) {
	bare := c.jid.Bare()
	items := c.svc.roster.GetFilteredRoster(ctx, c.jid.LUser, c.jid.LServer)
	p := pres.Copy()
	p.SetAttr("from", c.jid.String())
	fjids := []domain.JID{bare}
	tjids := []domain.JID{bare}
	for _, it := range items {
		if it.IsFrom() {
			fjids = append(fjids, it.JID.JID())
		}
		if fromUnavailable && it.IsTo() {
			tjids = append(tjids, it.JID.JID())
		}
	}
	if fromUnavailable {
		c.routeMultiple(ctx, tjids, domain.NewPresence("probe", c.jid, domain.JID{}))
	}
	c.routeMultiple(ctx, fjids, p)
}

func (c *C2SSession) sendError(pkt *domain.Element, se domain.StanzaError) {
	switch domain.StanzaType(pkt) {
	case "error", "result":
		return
	}
	if se.Lang == "" {
		se.Lang = c.lang
	}
	c.stream.Send(domain.MakeError(pkt, se))
}

func (c *C2SSession) terminate(ctx context.Context, reason string) {
	slog.Info("closing c2s session", "jid", c.jid.String(), "reason", reason)
	pres := domain.NewPresence("unavailable", c.jid, c.jid.Bare())
	if c.presLast != nil {
		c.svc.sm.CloseSession(c.sid, c.jid)
		c.broadcastPresenceUnavailable(ctx, pres, true)
	} else {
		c.svc.sm.CloseSession(c.sid, c.jid)
		c.broadcastPresenceUnavailable(ctx, pres, false)
	}
	for _, it := range c.box.close() {
		if it.event != nil && it.event.Kind == port.EventRoute && it.event.Stanza != nil {
			c.svc.router.Route(ctx, it.event.Stanza)
		}
	}
	c.stream.Close(nil)
}
