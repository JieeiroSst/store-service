package application

import (
	"context"
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

type OfflineService struct {
	cfg     *config.Config
	matcher *domain.ACLMatcher
	repo    port.OfflineRepository
	router  port.Router
	now     func() time.Time
}

func NewOfflineService(cfg *config.Config, matcher *domain.ACLMatcher, repo port.OfflineRepository, router *Router) *OfflineService {
	return &OfflineService{cfg: cfg, matcher: matcher, repo: repo, router: router, now: time.Now}
}

func (s *OfflineService) Enabled() bool { return s.cfg.Modules.Offline.Enabled }

func (s *OfflineService) policy() domain.OfflinePolicy {
	return domain.OfflinePolicy{
		StoreGroupchat: s.cfg.Modules.Offline.StoreGroupchat,
		StoreEmptyBody: s.cfg.Modules.Offline.StoreEmptyBody,
	}
}

func (s *OfflineService) StorePacket(ctx context.Context, msg *domain.Element) OfflineResult {
	if !s.Enabled() || !domain.NeedToStore(s.policy(), msg) {
		return OfflinePass
	}
	ev := domain.CheckEvent(msg)
	if ev.Reply != nil {
		s.router.Route(ctx, ev.Reply)
	}
	if !ev.Store {
		return OfflinePass
	}
	from, _ := domain.StanzaFrom(msg)
	to, _ := domain.StanzaTo(msg)
	now := s.now()
	off := domain.OfflineMessage{
		LUser:     to.LUser,
		LServer:   to.LServer,
		Timestamp: now,
		Expire:    domain.FindExpire(now, msg),
		From:      from,
		To:        to,
		Packet:    msg.Copy(),
	}
	if err := s.store(ctx, off); err != nil {
		s.discardWarnSender(ctx, msg, err)
		return OfflineStop
	}
	return OfflineStored
}

func (s *OfflineService) store(ctx context.Context, msg domain.OfflineMessage) error {
	limit, ok := s.maxUserMessages(msg.To)
	if ok {
		n, err := s.repo.Count(ctx, msg.LUser, msg.LServer)
		if err != nil {
			return err
		}
		if n >= limit {
			return domain.ErrOfflineQueueFull
		}
	}
	return s.repo.Store(ctx, msg)
}

func (s *OfflineService) maxUserMessages(to domain.JID) (int, bool) {
	v := s.matcher.MatchShaper(s.cfg.Modules.Offline.AccessMaxUserMessages, domain.MatchJID(to.Bare()))
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

func (s *OfflineService) discardWarnSender(ctx context.Context, msg *domain.Element, err error) {
	switch domain.StanzaType(msg) {
	case "chat", "normal":
	default:
		return
	}
	if errors.Is(err, domain.ErrOfflineQueueFull) {
		s.router.RouteError(ctx, msg, domain.ErrResourceConstraint("Your contact offline message queue is full. The message has been discarded."))
		return
	}
	slog.Error("offline store failed", "err", err)
	s.router.RouteError(ctx, msg, domain.ErrInternalServerError("Database failure"))
}

func (s *OfflineService) RouteOfflineMessages(ctx context.Context, jid domain.JID) {
	if !s.Enabled() {
		return
	}
	msgs, err := s.repo.Pop(ctx, jid.LUser, jid.LServer)
	if err != nil {
		slog.Error("offline pop failed", "jid", jid.String(), "err", err)
		return
	}
	now := s.now()
	host, _ := domain.NewJID("", jid.LServer, "")
	for _, m := range msgs {
		if m.Expired(now) {
			continue
		}
		pkt := m.Packet.Copy()
		domain.SetFromTo(pkt, m.From, m.To)
		domain.AddDelay(pkt, host, m.Timestamp, "Offline Storage")
		s.router.Route(ctx, pkt)
	}
}

func (s *OfflineService) Count(ctx context.Context, luser, lserver string) int {
	n, err := s.repo.Count(ctx, luser, lserver)
	if err != nil {
		return 0
	}
	return n
}

func (s *OfflineService) RemoveUser(ctx context.Context, luser, lserver string) {
	if err := s.repo.RemoveUser(ctx, luser, lserver); err != nil {
		slog.Error("offline remove_user failed", "user", luser, "server", lserver, "err", err)
	}
}

func (s *OfflineService) RemoveExpired(ctx context.Context) {
	for _, h := range s.cfg.Hosts {
		if n, err := s.repo.RemoveExpired(ctx, h); err != nil {
			slog.Error("remove expired offline messages failed", "host", h, "err", err)
		} else if n > 0 {
			slog.Info("removed expired offline messages", "host", h, "count", n)
		}
	}
}
