package application

import (
	"context"
	"errors"
	"fmt"
	"slices"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

type CommandError struct {
	Status  int
	Code    int
	Message string
}

func (e *CommandError) Error() string { return e.Message }

func cmdErr(status, code int, format string, args ...any) *CommandError {
	return &CommandError{Status: status, Code: code, Message: fmt.Sprintf(format, args...)}
}

type RosterEntry struct {
	JID          string   `json:"jid"`
	Nick         string   `json:"nick"`
	Subscription string   `json:"subscription"`
	Ask          string   `json:"ask"`
	Groups       []string `json:"groups"`
}

type AdminService struct {
	cfg     *config.Config
	auth    *AuthService
	sm      *SessionManager
	router  *Router
	roster  *RosterService
	offline *OfflineService
}

func NewAdminService(cfg *config.Config, auth *AuthService, sm *SessionManager, router *Router, roster *RosterService, offline *OfflineService) *AdminService {
	return &AdminService{cfg: cfg, auth: auth, sm: sm, router: router, roster: roster, offline: offline}
}

func (a *AdminService) Status() string {
	return fmt.Sprintf("The node ejabberd@%s is started with status: started\nejabberd %s is running in that node", a.cfg.MyName(), config.Version)
}

func (a *AdminService) Register(ctx context.Context, user, host, password string) (string, error) {
	switch err := a.auth.TryRegister(ctx, user, host, password); {
	case err == nil:
		return fmt.Sprintf("User %s@%s successfully registered", user, host), nil
	case errors.Is(err, domain.ErrUserExists):
		return "", cmdErr(409, 10090, "User %s@%s already registered", user, host)
	case errors.Is(err, domain.ErrInvalidJID):
		return "", cmdErr(400, 10, "Malformed username")
	case errors.Is(err, domain.ErrInvalidPassword):
		return "", cmdErr(400, 10, "The password contains unacceptable characters")
	case errors.Is(err, domain.ErrNotMyHost):
		return "", cmdErr(400, 10, "Unknown virtual host: %s", host)
	case errors.Is(err, domain.ErrNotAllowed):
		return "", cmdErr(400, 10, "Accounts are managed by user_service")
	default:
		return "", cmdErr(500, 10001, "Can't register user %s@%s at node ejabberd@%s: %v", user, host, a.cfg.MyName(), err)
	}
}

func (a *AdminService) Unregister(ctx context.Context, user, host string) error {
	if err := a.auth.RemoveUser(ctx, user, host); err != nil && !errors.Is(err, domain.ErrUserNotFound) {
		if errors.Is(err, domain.ErrNotAllowed) {
			return cmdErr(400, 10, "Accounts are managed by user_service")
		}
		return cmdErr(400, 10, "%v", err)
	}
	return nil
}

func (a *AdminService) CheckAccount(ctx context.Context, user, host string) error {
	j, err := domain.NewJID(user, host, "")
	if err != nil || !a.auth.UserExists(ctx, j.LUser, j.LServer) {
		return cmdErr(404, 1, "User does not exist")
	}
	return nil
}

func (a *AdminService) CheckPassword(ctx context.Context, user, host, password string) error {
	if !a.auth.CheckPassword(ctx, user, "", host, password) {
		return cmdErr(400, 1, "Password is incorrect")
	}
	return nil
}

func (a *AdminService) ChangePassword(ctx context.Context, user, host, newpass string) error {
	if err := a.auth.SetPassword(ctx, user, host, newpass); err != nil {
		return cmdErr(400, 1, "%v", err)
	}
	return nil
}

func (a *AdminService) RegisteredUsers(ctx context.Context, host string) ([]string, error) {
	users, err := a.auth.ListUsers(ctx, host)
	if err != nil {
		if errors.Is(err, domain.ErrNotAllowed) {
			return nil, cmdErr(400, 10, "Accounts are managed by user_service")
		}
		return nil, cmdErr(400, 10, "%v", err)
	}
	slices.Sort(users)
	return users, nil
}

func (a *AdminService) ConnectedUsers() []string {
	var out []string
	for _, s := range a.sm.AllSessions() {
		out = append(out, s.USR.String())
	}
	slices.Sort(out)
	return out
}

func (a *AdminService) ConnectedUsersVhost(host string) []string {
	var out []string
	for _, s := range a.sm.AllSessions() {
		if s.USR.Server == host {
			out = append(out, s.USR.String())
		}
	}
	slices.Sort(out)
	return out
}

func (a *AdminService) ConnectedUsersNumber() int {
	return a.sm.Count()
}

func (a *AdminService) UserResources(user, host string) ([]string, error) {
	j, err := domain.NewJID(user, host, "")
	if err != nil {
		return nil, cmdErr(400, 10, "Invalid 'user'/'host'")
	}
	res := a.sm.UserResources(j.LUser, j.LServer)
	slices.Sort(res)
	return res, nil
}

func (a *AdminService) KickUser(user, host string) (int, error) {
	j, err := domain.NewJID(user, host, "")
	if err != nil {
		return 0, cmdErr(400, 10, "Invalid 'user'/'host'")
	}
	n := len(a.sm.UserResources(j.LUser, j.LServer))
	a.sm.DeliverToUser(j.LUser, j.LServer, port.SessionEvent{Kind: port.EventExit, Reason: "Kicked by administrator"})
	return n, nil
}

func (a *AdminService) SendMessage(ctx context.Context, msgType, from, to, subject, body string) error {
	fj, err := domain.ParseJID(from)
	if err != nil {
		return cmdErr(400, 10, "Invalid 'from'")
	}
	tj, err := domain.ParseJID(to)
	if err != nil {
		return cmdErr(400, 10, "Invalid 'to'")
	}
	switch msgType {
	case "chat", "normal", "headline", "groupchat", "":
	default:
		return cmdErr(400, 10, "Invalid 'type'")
	}
	msg := domain.NewMessage(msgType, fj, tj, subject, body)
	msg.SetAttr("id", randString())
	a.router.Route(ctx, msg)
	return nil
}

func (a *AdminService) SendStanza(ctx context.Context, from, to, stanza string) error {
	el, err := domain.ParseElement(stanza)
	if err != nil {
		return cmdErr(400, 10, "invalid xml: %v", err)
	}
	if !domain.IsStanza(el) {
		return cmdErr(400, 10, "incorrect stanza")
	}
	fj, err := domain.ParseJID(from)
	if err != nil {
		return cmdErr(400, 10, "Invalid 'from'")
	}
	tj, err := domain.ParseJID(to)
	if err != nil {
		return cmdErr(400, 10, "Invalid 'to'")
	}
	if el.Namespace() == "" {
		el.SetAttr("xmlns", domain.NSClient)
	}
	domain.SetFromTo(el, fj, tj)
	a.router.Route(ctx, el)
	return nil
}

func (a *AdminService) GetRoster(ctx context.Context, user, host string) ([]RosterEntry, error) {
	j, err := domain.NewJID(user, host, "")
	if err != nil {
		return nil, cmdErr(400, 10, "Invalid 'user'/'server'")
	}
	out := []RosterEntry{}
	for _, it := range a.roster.GetFilteredRoster(ctx, j.LUser, j.LServer) {
		ask := string(domain.AskToPending(it.Ask))
		groups := it.Groups
		if groups == nil {
			groups = []string{}
		}
		out = append(out, RosterEntry{JID: it.JID.String(), Nick: it.Name, Subscription: string(it.Subscription), Ask: ask, Groups: groups})
	}
	return out, nil
}

func (a *AdminService) AddRosterItem(ctx context.Context, localUser, localHost, user, host, nick string, groups []string, subs string) error {
	owner, err := domain.NewJID(localUser, localHost, "")
	if err != nil {
		return cmdErr(400, 10, "Invalid 'localuser'/'localserver'")
	}
	contact, err := domain.NewJID(user, host, "")
	if err != nil {
		return cmdErr(400, 10, "Invalid 'user'/'server'")
	}
	switch domain.Subscription(subs) {
	case domain.SubNone, domain.SubTo, domain.SubFrom, domain.SubBoth:
	default:
		return cmdErr(400, 10, "Invalid 'subs'")
	}
	item := BuildRosterItem(contact, nick, domain.Subscription(subs), groups)
	if err := a.roster.SetItemAndNotifyClients(ctx, owner, item, true); err != nil {
		return cmdErr(500, 1, "%v", err)
	}
	return nil
}

func (a *AdminService) DeleteRosterItem(ctx context.Context, localUser, localHost, user, host string) error {
	owner, err := domain.NewJID(localUser, localHost, "")
	if err != nil {
		return cmdErr(400, 10, "Invalid 'localuser'/'localserver'")
	}
	contact, err := domain.NewJID(user, host, "")
	if err != nil {
		return cmdErr(400, 10, "Invalid 'user'/'server'")
	}
	item := BuildRosterItem(contact, "", domain.SubRemove, nil)
	if err := a.roster.SetItemAndNotifyClients(ctx, owner, item, true); err != nil {
		return cmdErr(500, 1, "%v", err)
	}
	return nil
}

func (a *AdminService) GetOfflineCount(ctx context.Context, user, host string) (int, error) {
	j, err := domain.NewJID(user, host, "")
	if err != nil {
		return 0, cmdErr(400, 10, "Invalid 'user'/'server'")
	}
	return a.offline.Count(ctx, j.LUser, j.LServer), nil
}

func (a *AdminService) DeleteExpiredMessages(ctx context.Context) {
	a.offline.RemoveExpired(ctx)
}

func (a *AdminService) RegisteredVhosts() []string {
	return slices.Clone(a.cfg.Hosts)
}

func (a *AdminService) RegisteredUsersCount(ctx context.Context) (int, error) {
	total := 0
	for _, h := range a.cfg.Hosts {
		n, err := a.auth.CountUsers(ctx, h)
		if err != nil {
			return 0, cmdErr(500, 1, "%v", err)
		}
		total += n
	}
	return total, nil
}
