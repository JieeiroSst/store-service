package port

import (
	"context"
	"io"

	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

type UserRepository interface {
	GetPassword(ctx context.Context, luser, lserver string) (domain.Password, error)
	SetPassword(ctx context.Context, luser, lserver string, pass domain.Password) error
	TryRegister(ctx context.Context, luser, lserver string, pass domain.Password) error
	UserExists(ctx context.Context, luser, lserver string) (bool, error)
	RemoveUser(ctx context.Context, luser, lserver string) error
	ListUsers(ctx context.Context, lserver string) ([]string, error)
	CountUsers(ctx context.Context, lserver string) (int, error)
}

type RosterRepository interface {
	GetRoster(ctx context.Context, luser, lserver string) ([]domain.RosterItem, error)
	GetRosterItem(ctx context.Context, luser, lserver string, jid domain.LJID) (domain.RosterItem, bool, error)
	UpdateRoster(ctx context.Context, item domain.RosterItem) error
	DeleteRoster(ctx context.Context, luser, lserver string, jid domain.LJID) error
	RemoveUser(ctx context.Context, luser, lserver string) error
	Transaction(ctx context.Context, luser, lserver string, fn func(tx RosterRepository) error) error
}

type OfflineRepository interface {
	Store(ctx context.Context, msg domain.OfflineMessage) error
	Pop(ctx context.Context, luser, lserver string) ([]domain.OfflineMessage, error)
	Count(ctx context.Context, luser, lserver string) (int, error)
	RemoveUser(ctx context.Context, luser, lserver string) error
	RemoveExpired(ctx context.Context, lserver string) (int, error)
}

type ExternalAuthenticator interface {
	CheckPassword(ctx context.Context, user, luser, lserver, password string) (bool, error)
	UserExists(ctx context.Context, luser, lserver string) (bool, error)
	EnsureUser(ctx context.Context, account domain.Account) (bool, error)
}

type FileStorage interface {
	Exists(ctx context.Context, key string) (bool, error)
	Store(ctx context.Context, key, fileName string, size int64, body io.Reader) error
	Fetch(ctx context.Context, key string) (domain.StoredFile, error)
}

type HealthChecker interface {
	Ping(ctx context.Context) error
}

type SessionEventKind int

const (
	EventRoute SessionEventKind = iota
	EventReplaced
	EventResendPresence
	EventExit
)

type SessionEvent struct {
	Kind   SessionEventKind
	Stanza *domain.Element
	To     *domain.JID
	Reason string
}

type SessionSink interface {
	Deliver(ev SessionEvent)
}

type Router interface {
	Route(ctx context.Context, stanza *domain.Element)
	RouteError(ctx context.Context, stanza *domain.Element, err domain.StanzaError)
}

type StreamConn interface {
	Send(el *domain.Element)
	Close(err *domain.StreamError)
	RemoteIP() string
	Encrypted() bool
}
