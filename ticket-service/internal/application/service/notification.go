package service

import (
	"context"
	"log/slog"
	"time"

	"github.com/JIeeiroSst/ticket-service/internal/application/port/inbound"
	"github.com/JIeeiroSst/ticket-service/internal/application/port/outbound"
	"github.com/JIeeiroSst/ticket-service/internal/domain"
)

type Notifier struct {
	repo  outbound.NotificationRepository
	log   *slog.Logger
	email bool
	loc   *time.Location
}

func NewNotifier(repo outbound.NotificationRepository, log *slog.Logger) *Notifier {
	if repo == nil {
		return nil
	}
	return &Notifier{repo: repo, log: defaultLog(log), loc: time.Local}
}

func (n *Notifier) WithEmail(on bool) *Notifier {
	if n != nil {
		n.email = on
	}
	return n
}

func (n *Notifier) WithLocation(loc *time.Location) *Notifier {
	if n != nil && loc != nil {
		n.loc = loc
	}
	return n
}

func (n *Notifier) EmailOn() bool { return n != nil && n.email }

func (n *Notifier) Format(t time.Time) string {
	loc := time.Local
	if n != nil {
		loc = n.loc
	}
	return t.In(loc).Format("Mon 2 Jan 2006, 15:04")
}

func (n *Notifier) Send(ctx context.Context, userID int64, kind, title, body string, orderID, eventID int64) {
	n.SendMail(ctx, userID, kind, title, body, orderID, eventID, nil)
}

func (n *Notifier) SendMail(ctx context.Context, userID int64, kind, title, body string, orderID, eventID int64, mail *domain.Email) {
	if n == nil || (userID == 0 && mail == nil) {
		return
	}
	if !n.email || mail == nil || mail.To == "" {
		mail = nil
	}
	if userID == 0 && mail == nil {
		return
	}
	if err := n.repo.Add(ctx, domain.Notification{UserID: userID, Kind: kind, Title: title, Body: body, OrderID: orderID, EventID: eventID, Email: mail}); err != nil {
		n.log.Warn("notification not stored", "user", userID, "kind", kind, "err", err)
	}
}

func (n *Notifier) MailOnly(ctx context.Context, to, template, title string, eventID int64, data map[string]string) {
	n.SendMail(ctx, 0, "email", title, title, 0, eventID, &domain.Email{To: to, Template: template, Data: data})
}

const pushLease = 30 * time.Second

type NotificationService struct {
	repo  outbound.NotificationRepository
	push  outbound.PushGateway
	email outbound.EmailGateway
	log   *slog.Logger
}

var _ inbound.NotificationUseCase = (*NotificationService)(nil)

func NewNotificationService(repo outbound.NotificationRepository, push outbound.PushGateway, email outbound.EmailGateway, log *slog.Logger) *NotificationService {
	return &NotificationService{repo: repo, push: push, email: email, log: defaultLog(log)}
}

func (s *NotificationService) List(ctx context.Context, a inbound.Principal, unreadOnly bool, afterID int64, limit int) (domain.Page[domain.Notification], error) {
	limit = page(limit)
	rows, err := s.repo.List(ctx, a.UserID, unreadOnly, afterID, limit+1)
	if err != nil {
		return domain.Page[domain.Notification]{}, err
	}
	return domain.PageOf(rows, limit, func(n domain.Notification) int64 { return n.ID }), nil
}

func (s *NotificationService) UnreadCount(ctx context.Context, a inbound.Principal) (int, error) {
	return s.repo.UnreadCount(ctx, a.UserID)
}

func (s *NotificationService) MarkRead(ctx context.Context, a inbound.Principal, id int64) error {
	return s.repo.MarkRead(ctx, a.UserID, id)
}

func (s *NotificationService) MarkAllRead(ctx context.Context, a inbound.Principal) (int, error) {
	return s.repo.MarkAllRead(ctx, a.UserID)
}

func (s *NotificationService) PushPending(ctx context.Context) (int, error) {
	if s.push == nil {
		return 0, nil
	}
	pending, err := s.repo.ClaimUnpushed(ctx, 100, pushLease)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, n := range pending {
		if err := s.push.Push(ctx, n); err != nil {
			s.log.Warn("push to notification-service failed", "notification", n.ID, "err", err)
			continue
		}
		if err := s.repo.MarkPushed(ctx, n.ID); err != nil {
			s.log.Error("push recorded as failed though it was sent", "notification", n.ID, "err", err)
			continue
		}
		sent++
	}
	return sent, nil
}

func (s *NotificationService) SendEmails(ctx context.Context) (int, error) {
	if s.email == nil {
		return 0, nil
	}
	pending, err := s.repo.ClaimUnemailed(ctx, 100, pushLease)
	if err != nil {
		return 0, err
	}
	sent := 0
	for _, n := range pending {
		if err := s.email.Email(ctx, n); err != nil {
			s.log.Warn("e-mail to notification-service failed", "notification", n.ID, "err", err)
			continue
		}
		if err := s.repo.MarkEmailed(ctx, n.ID); err != nil {
			s.log.Error("e-mail recorded as unsent though it was sent", "notification", n.ID, "err", err)
			continue
		}
		sent++
	}
	return sent, nil
}
