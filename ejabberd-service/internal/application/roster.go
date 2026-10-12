package application

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"log/slog"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

type RosterService struct {
	cfg     *config.Config
	matcher *domain.ACLMatcher
	repo    port.RosterRepository
	router  port.Router
	sm      *SessionManager
}

func NewRosterService(cfg *config.Config, matcher *domain.ACLMatcher, repo port.RosterRepository, router *Router, sm *SessionManager) *RosterService {
	return &RosterService{cfg: cfg, matcher: matcher, repo: repo, router: router, sm: sm}
}

func (s *RosterService) Enabled() bool { return s.cfg.Modules.Roster.Enabled }

func (s *RosterService) Access() string { return s.cfg.Modules.Roster.Access }

func (s *RosterService) GetRoster(ctx context.Context, luser, lserver string) []domain.RosterItem {
	if !s.Enabled() {
		return nil
	}
	items, err := s.repo.GetRoster(ctx, luser, lserver)
	if err != nil {
		slog.Error("get roster failed", "user", luser, "server", lserver, "err", err)
		return nil
	}
	return items
}

func (s *RosterService) GetFilteredRoster(ctx context.Context, luser, lserver string) []domain.RosterItem {
	return domain.FilterRoster(s.GetRoster(ctx, luser, lserver))
}

func (s *RosterService) getItem(ctx context.Context, repo port.RosterRepository, luser, lserver string, jid domain.LJID) (domain.RosterItem, error) {
	item, ok, err := repo.GetRosterItem(ctx, luser, lserver, jid.Bare())
	if err != nil {
		return domain.RosterItem{}, err
	}
	if !ok {
		return domain.NewRosterItem(luser, lserver, jid), nil
	}
	return item, nil
}

func (s *RosterService) GetSubscription(ctx context.Context, owner domain.JID, contact domain.JID) (domain.Subscription, domain.Ask, []string) {
	if !s.Enabled() {
		return domain.SubNone, domain.AskNone, nil
	}
	item, ok, err := s.repo.GetRosterItem(ctx, owner.LUser, owner.LServer, contact.LJID().Bare())
	if err != nil || !ok {
		return domain.SubNone, domain.AskNone, nil
	}
	return item.Subscription, item.Ask, item.Groups
}

func (s *RosterService) ProcessIQ(ctx context.Context, iq *domain.Element) *domain.Element {
	from, _ := domain.StanzaFrom(iq)
	to, _ := domain.StanzaTo(iq)
	if from.LUser != to.LUser || from.LServer != to.LServer {
		return domain.MakeError(iq, domain.ErrForbidden("Query to another users is forbidden"))
	}
	query := domain.IQPayload(iq)
	items := query.ChildrenNamed("item", "")
	switch domain.StanzaType(iq) {
	case "set":
		if len(items) == 1 {
			if items[0].HasAttr("ask") {
				return domain.MakeError(iq, domain.ErrBadRequest("Possessing 'ask' attribute is not allowed by RFC6121"))
			}
			if domain.HasDuplicatedGroups(itemGroups(items[0])) {
				return domain.MakeError(iq, domain.ErrBadRequest("Duplicated groups are not allowed by RFC6121"))
			}
			if !s.matcher.MatchRule(s.Access(), domain.MatchJID(from)) {
				return domain.MakeError(iq, domain.ErrNotAllowedStanza("Access denied by service policy"))
			}
			return s.processIQSet(ctx, iq, to, items[0])
		}
		if len(items) > 1 {
			return domain.MakeError(iq, domain.ErrBadRequest("Multiple <item/> elements are not allowed by RFC6121"))
		}
	case "get":
		if len(items) == 0 {
			return s.processIQGet(ctx, iq, to, query)
		}
		return domain.MakeError(iq, domain.ErrBadRequest("The query must not contain <item/> elements"))
	}
	return domain.MakeError(iq, domain.ErrServiceUnavailable("No module is handling this query"))
}

func itemGroups(item *domain.Element) []string {
	var out []string
	for _, g := range item.ChildrenNamed("group", "") {
		out = append(out, g.Text())
	}
	return out
}

func (s *RosterService) processIQGet(ctx context.Context, iq *domain.Element, to domain.JID, query *domain.Element) *domain.Element {
	items := s.GetFilteredRoster(ctx, to.LUser, to.LServer)
	ver := ""
	if s.cfg.Modules.Roster.Versioning && query.HasAttr("ver") {
		ver = domain.RosterHash(items)
		if ver == query.Attr("ver") {
			return domain.MakeIQResult(iq, nil)
		}
	}
	res := domain.NewElement("query", domain.NSRoster)
	if ver != "" {
		res.SetAttr("ver", ver)
	}
	for _, it := range items {
		res.Append(it.Encode())
	}
	return domain.MakeIQResult(iq, res)
}

func (s *RosterService) processIQSet(ctx context.Context, iq *domain.Element, to domain.JID, item *domain.Element) *domain.Element {
	if err := s.SetItemAndNotifyClients(ctx, to, item, false); err != nil {
		return domain.MakeError(iq, domain.ErrInternalServerError("Database failure"))
	}
	return domain.MakeIQResult(iq, nil)
}

func decodeItem(el *domain.Element, r domain.RosterItem, managed bool) (domain.RosterItem, error) {
	j, err := domain.ParseJID(el.Attr("jid"))
	if err != nil {
		return r, err
	}
	r.JID = j.LJID()
	if domain.Subscription(el.Attr("subscription")) == domain.SubRemove {
		r.Name, r.Subscription, r.Ask, r.Groups, r.AskMessage, r.Xs = "", domain.SubRemove, domain.AskNone, nil, "", nil
		return r, nil
	}
	r.Name = el.Attr("name")
	if managed {
		if sub := domain.Subscription(el.Attr("subscription")); sub != "" {
			r.Subscription = sub
		}
	}
	r.Groups = itemGroups(el)
	return r, nil
}

func (s *RosterService) SetItemAndNotifyClients(ctx context.Context, owner domain.JID, item *domain.Element, override bool) error {
	peer, err := domain.ParseJID(item.Attr("jid"))
	if err != nil {
		return err
	}
	peerL := peer.LJID()
	var oldItem, newItem domain.RosterItem
	err = s.repo.Transaction(ctx, owner.LUser, owner.LServer, func(tx port.RosterRepository) error {
		item1, err := s.getItem(ctx, tx, owner.LUser, owner.LServer, peerL)
		if err != nil {
			return err
		}
		item2, err := decodeItem(item, item1, override)
		if err != nil {
			return err
		}
		item2.JID = peerL.Bare()
		if item2.Subscription == domain.SubRemove {
			err = tx.DeleteRoster(ctx, owner.LUser, owner.LServer, peerL.Bare())
		} else {
			err = tx.UpdateRoster(ctx, item2)
		}
		oldItem, newItem = item1, item2
		return err
	})
	if err != nil {
		return err
	}
	s.pushItem(ctx, owner.Bare(), oldItem, newItem)
	if newItem.Subscription == domain.SubRemove {
		s.sendUnsubscribingPresence(ctx, owner, oldItem)
	}
	return nil
}

func (s *RosterService) pushItem(ctx context.Context, owner domain.JID, oldItem, newItem domain.RosterItem) {
	ver := ""
	if s.cfg.Modules.Roster.Versioning {
		ver = domain.RosterHash(s.GetFilteredRoster(ctx, owner.LUser, owner.LServer))
	}
	for _, res := range s.sm.UserResources(owner.LUser, owner.LServer) {
		full, err := owner.WithResource(res)
		if err != nil {
			continue
		}
		s.routePresenceChange(ctx, full, oldItem, newItem)
		q := domain.NewElement("query", domain.NSRoster)
		if ver != "" {
			q.SetAttr("ver", ver)
		}
		q.Append(newItem.Encode())
		s.router.Route(ctx, domain.NewIQ("set", "push"+randString(), owner.Bare(), full, q))
	}
}

func (s *RosterService) routePresenceChange(ctx context.Context, from domain.JID, oldItem, newItem domain.RosterItem) {
	newIsFrom, oldIsFrom := newItem.IsFrom(), oldItem.IsFrom()
	to := newItem.JID.JID()
	switch {
	case newIsFrom && !oldIsFrom:
		if sink, ok := s.sm.Sink(from.LJID()); ok {
			sink.Deliver(port.SessionEvent{Kind: port.EventResendPresence, To: &to})
		}
	case oldIsFrom && !newIsFrom:
		s.router.Route(ctx, domain.NewPresence("unavailable", from, to))
	}
}

func (s *RosterService) sendUnsubscribingPresence(ctx context.Context, from domain.JID, item domain.RosterItem) {
	to := item.JID.JID()
	if item.IsTo() {
		s.router.Route(ctx, domain.NewPresence(domain.PresUnsubscribe, from.Bare(), to))
	}
	if item.IsFrom() {
		s.router.Route(ctx, domain.NewPresence(domain.PresUnsubscribed, from.Bare(), to))
	}
}

func (s *RosterService) InSubscription(ctx context.Context, presence *domain.Element) bool {
	if !s.Enabled() {
		return false
	}
	from, ok := domain.StanzaFrom(presence)
	if !ok {
		return false
	}
	to, _ := domain.StanzaTo(presence)
	ptype := domain.StanzaType(presence)
	reason := ""
	if ptype == domain.PresSubscribe {
		reason = domain.StatusText(presence)
	}
	return s.processSubscription(ctx, true, to, from, ptype, reason, presence.Elements())
}

func (s *RosterService) OutSubscription(ctx context.Context, presence *domain.Element) bool {
	if !s.Enabled() {
		return false
	}
	from, _ := domain.StanzaFrom(presence)
	to, ok := domain.StanzaTo(presence)
	if !ok {
		return false
	}
	return s.processSubscription(ctx, false, from, to, domain.StanzaType(presence), "", nil)
}

func (s *RosterService) processSubscription(ctx context.Context, in bool, owner, contact domain.JID, ptype, reason string, subEls []*domain.Element) bool {
	lj := contact.LJID().Bare()
	var (
		autoReply        string
		pushed           bool
		oldItem, newItem domain.RosterItem
	)
	err := s.repo.Transaction(ctx, owner.LUser, owner.LServer, func(tx port.RosterRepository) error {
		item, err := s.getItem(ctx, tx, owner.LUser, owner.LServer, lj)
		if err != nil {
			return err
		}
		var (
			sub     domain.Subscription
			ask     domain.Ask
			changed bool
		)
		if in {
			sub, ask, changed = domain.InStateChange(item.Subscription, item.Ask, ptype)
			autoReply = domain.InAutoReply(item.Subscription, item.Ask, ptype)
		} else {
			sub, ask, changed = domain.OutStateChange(item.Subscription, item.Ask, ptype)
		}
		if !changed {
			return nil
		}
		if sub == domain.SubNone && ask == domain.AskNone && item.Subscription == domain.SubNone && domain.AskToPending(item.Ask) == domain.AskIn {
			return tx.DeleteRoster(ctx, owner.LUser, owner.LServer, lj)
		}
		askMessage := ""
		if ask == domain.AskBoth || ask == domain.AskIn {
			askMessage = reason
		}
		next := item
		next.Subscription, next.Ask = sub, ask
		next.Name = domain.NickFromSubEls(subEls, item.Name)
		next.Xs = subEls
		next.AskMessage = askMessage
		if err := tx.UpdateRoster(ctx, next); err != nil {
			return err
		}
		oldItem, newItem, pushed = item, next, true
		return nil
	})
	if err != nil {
		slog.Error("roster subscription failed", "owner", owner.String(), "contact", contact.String(), "err", err)
		return false
	}
	if autoReply != "" {
		s.router.Route(ctx, domain.NewPresence(autoReply, owner.Bare(), contact))
	}
	if !pushed {
		return false
	}
	if !newItem.IsPendingIn() {
		s.pushItem(ctx, owner.Bare(), oldItem, newItem)
	}
	return true
}

func (s *RosterService) PendingSubscriptions(ctx context.Context, jid domain.JID) []*domain.Element {
	var out []*domain.Element
	for _, r := range s.GetRoster(ctx, jid.LUser, jid.LServer) {
		ask := domain.AskToPending(r.Ask)
		if ask != domain.AskIn && ask != domain.AskBoth {
			continue
		}
		p := domain.NewPresence(domain.PresSubscribe, r.JID.JID(), jid.Bare())
		for _, x := range r.Xs {
			p.Append(x.Copy())
		}
		if r.AskMessage != "" {
			p.Append(domain.NewElement("status", "").AppendText(r.AskMessage))
		}
		out = append(out, p)
	}
	return out
}

func (s *RosterService) RemoveUser(ctx context.Context, luser, lserver string) {
	from, err := domain.NewJID(luser, lserver, "")
	if err != nil {
		return
	}
	for _, it := range s.GetFilteredRoster(ctx, luser, lserver) {
		s.sendUnsubscribingPresence(ctx, from, it)
	}
	if err := s.repo.RemoveUser(ctx, luser, lserver); err != nil {
		slog.Error("roster remove_user failed", "user", luser, "server", lserver, "err", err)
	}
}

func BuildRosterItem(jid domain.JID, nick string, sub domain.Subscription, groups []string) *domain.Element {
	el := domain.NewElement("item", "")
	el.SetAttr("jid", jid.Bare().String())
	if nick != "" {
		el.SetAttr("name", nick)
	}
	el.SetAttr("subscription", string(sub))
	for _, g := range groups {
		if g != "" {
			el.Append(domain.NewElement("group", "").AppendText(g))
		}
	}
	return el
}

func randString() string {
	b := make([]byte, 8)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
