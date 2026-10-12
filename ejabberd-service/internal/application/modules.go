package application

import (
	"context"
	"errors"
	"log/slog"
	"math"
	"net"
	"runtime"
	"strconv"
	"sync"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
)

type Modules struct {
	cfg     *config.Config
	matcher *domain.ACLMatcher
	iq      *IQRegistry
	router  *Router
	sm      *SessionManager
	auth    *AuthService
	roster  *RosterService
	started time.Time
	regMu   sync.Mutex
	regIPs  map[string]time.Time
}

func NewModules(cfg *config.Config, matcher *domain.ACLMatcher, iq *IQRegistry, router *Router, sm *SessionManager, auth *AuthService, roster *RosterService) *Modules {
	return &Modules{cfg: cfg, matcher: matcher, iq: iq, router: router, sm: sm, auth: auth, roster: roster, started: time.Now(), regIPs: map[string]time.Time{}}
}

func (m *Modules) Register() {
	mods := m.cfg.Modules
	if mods.Roster.Enabled {
		m.iq.Add(ScopeSM, domain.NSRoster, m.roster.ProcessIQ)
	}
	if mods.Ping.Enabled {
		m.iq.Add(ScopeLocal, domain.NSPing, m.ping)
		m.iq.Add(ScopeSM, domain.NSPing, m.ping)
	}
	if mods.Disco {
		m.iq.Add(ScopeLocal, domain.NSDiscoInfo, m.discoInfoLocal)
		m.iq.Add(ScopeLocal, domain.NSDiscoItems, m.discoItemsLocal)
		m.iq.Add(ScopeSM, domain.NSDiscoInfo, m.discoInfoSM)
		m.iq.Add(ScopeSM, domain.NSDiscoItems, m.discoItemsSM)
	}
	if mods.Version.Enabled {
		m.iq.Add(ScopeLocal, domain.NSVersion, m.version)
	}
	if mods.Last {
		m.iq.Add(ScopeLocal, domain.NSLast, m.lastLocal)
	}
	if mods.Register.Enabled {
		m.iq.Add(ScopeLocal, domain.NSRegister, func(ctx context.Context, iq *domain.Element) *domain.Element {
			from, _ := domain.StanzaFrom(iq)
			var ip net.IP
			if from.LUser != "" {
				for _, s := range m.sm.Sessions(from.LUser, from.LServer) {
					if s.USR.Resource == from.LResource {
						ip = s.Info.IP
					}
				}
			}
			return m.ProcessRegister(ctx, iq, ip)
		})
	}
	slog.Info("iq handlers registered", "local", m.iq.Features(ScopeLocal), "sm", m.iq.Features(ScopeSM))
}

func (m *Modules) ping(_ context.Context, iq *domain.Element) *domain.Element {
	if domain.StanzaType(iq) != "get" {
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("Ping query is incorrect"))
	}
	return domain.MakeIQResult(iq, nil)
}

func (m *Modules) version(_ context.Context, iq *domain.Element) *domain.Element {
	if domain.StanzaType(iq) != "get" {
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("Value 'set' of 'type' attribute is not allowed"))
	}
	q := domain.NewElement("query", domain.NSVersion)
	q.Append(domain.NewElement("name", "").AppendText("ejabberd"))
	q.Append(domain.NewElement("version", "").AppendText(config.Version))
	if m.cfg.Modules.Version.ShowOS {
		q.Append(domain.NewElement("os", "").AppendText(runtime.GOOS + "/" + runtime.GOARCH))
	}
	return domain.MakeIQResult(iq, q)
}

func (m *Modules) lastLocal(_ context.Context, iq *domain.Element) *domain.Element {
	if domain.StanzaType(iq) != "get" {
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("Value 'set' of 'type' attribute is not allowed"))
	}
	q := domain.NewElement("query", domain.NSLast)
	q.SetAttr("seconds", strconv.Itoa(int(time.Since(m.started).Seconds())))
	return domain.MakeIQResult(iq, q)
}

func (m *Modules) ServerFeatures() []string {
	feats := []string{"iq", "presence", domain.NSDiscoInfo, domain.NSDiscoItems}
	feats = append(feats, m.iq.Features(ScopeLocal)...)
	if m.cfg.Modules.Offline.Enabled {
		feats = append(feats, domain.NSMsgOffline)
	}
	return dedup(feats)
}

func dedup(in []string) []string {
	seen := map[string]struct{}{}
	var out []string
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

func (m *Modules) discoInfoLocal(_ context.Context, iq *domain.Element) *domain.Element {
	if domain.StanzaType(iq) != "get" {
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("Value 'set' of 'type' attribute is not allowed"))
	}
	q := domain.IQPayload(iq)
	node := q.Attr("node")
	if node != "" {
		return domain.MakeError(iq, domain.ErrItemNotFound(""))
	}
	res := domain.NewElement("query", domain.NSDiscoInfo)
	id := domain.NewElement("identity", "")
	id.SetAttr("category", "server").SetAttr("type", "im").SetAttr("name", "ejabberd")
	res.Append(id)
	for _, f := range m.ServerFeatures() {
		res.Append(domain.NewElement("feature", "").SetAttr("var", f))
	}
	return domain.MakeIQResult(iq, res)
}

func (m *Modules) discoItemsLocal(_ context.Context, iq *domain.Element) *domain.Element {
	if domain.StanzaType(iq) != "get" {
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("Value 'set' of 'type' attribute is not allowed"))
	}
	if domain.IQPayload(iq).Attr("node") != "" {
		return domain.MakeError(iq, domain.ErrItemNotFound(""))
	}
	res := domain.NewElement("query", domain.NSDiscoItems)
	to, _ := domain.StanzaTo(iq)
	for _, s := range m.router.Services(to.LServer) {
		item := domain.NewElement("item", "").SetAttr("jid", s.Host)
		if s.Name != "" {
			item.SetAttr("name", s.Name)
		}
		res.Append(item)
	}
	return domain.MakeIQResult(iq, res)
}

func (m *Modules) isPresenceSubscribed(ctx context.Context, from, to domain.JID) bool {
	if from.LUser == to.LUser && from.LServer == to.LServer {
		return true
	}
	sub, _, _ := m.roster.GetSubscription(ctx, to, from)
	return sub == domain.SubBoth || sub == domain.SubFrom
}

func (m *Modules) discoInfoSM(ctx context.Context, iq *domain.Element) *domain.Element {
	if domain.StanzaType(iq) != "get" {
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("Value 'set' of 'type' attribute is not allowed"))
	}
	from, _ := domain.StanzaFrom(iq)
	to, _ := domain.StanzaTo(iq)
	if !m.isPresenceSubscribed(ctx, from, to) {
		return domain.MakeError(iq, domain.ErrServiceUnavailable(""))
	}
	if domain.IQPayload(iq).Attr("node") != "" {
		return domain.MakeError(iq, domain.ErrItemNotFound(""))
	}
	res := domain.NewElement("query", domain.NSDiscoInfo)
	id := domain.NewElement("identity", "")
	id.SetAttr("category", "account").SetAttr("type", "registered")
	res.Append(id)
	feats := append([]string{domain.NSDiscoInfo, domain.NSDiscoItems}, m.iq.Features(ScopeSM)...)
	if m.cfg.Modules.Offline.Enabled {
		feats = append(feats, domain.NSMsgOffline)
	}
	for _, f := range dedup(feats) {
		res.Append(domain.NewElement("feature", "").SetAttr("var", f))
	}
	return domain.MakeIQResult(iq, res)
}

func (m *Modules) discoItemsSM(ctx context.Context, iq *domain.Element) *domain.Element {
	if domain.StanzaType(iq) != "get" {
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("Value 'set' of 'type' attribute is not allowed"))
	}
	from, _ := domain.StanzaFrom(iq)
	to, _ := domain.StanzaTo(iq)
	if !m.isPresenceSubscribed(ctx, from, to) {
		return domain.MakeError(iq, domain.ErrServiceUnavailable(""))
	}
	res := domain.NewElement("query", domain.NSDiscoItems)
	if from.LUser == to.LUser && from.LServer == to.LServer {
		for _, r := range m.sm.UserResources(to.LUser, to.LServer) {
			full, err := to.WithResource(r)
			if err != nil {
				continue
			}
			res.Append(domain.NewElement("item", "").SetAttr("jid", full.String()))
		}
	}
	return domain.MakeIQResult(iq, res)
}

func (m *Modules) RegisterFeatureEnabled() bool {
	return m.cfg.Modules.Register.Enabled
}

func (m *Modules) ProcessRegister(ctx context.Context, iq *domain.Element, ip net.IP) *domain.Element {
	from, _ := domain.StanzaFrom(iq)
	to, _ := domain.StanzaTo(iq)
	server := to.LServer
	reg := m.cfg.Modules.Register
	allowRemove := m.matcher.MatchRule(reg.AccessRemove, domain.MatchJID(from)) && from.LServer == server
	q := domain.IQPayload(iq)
	username := q.Child("username", "")
	password := q.Child("password", "")
	switch domain.StanzaType(iq) {
	case "set":
		if q.Child("remove", "") != nil {
			if !allowRemove {
				return domain.MakeError(iq, domain.ErrForbidden("Access denied by service policy"))
			}
			return m.processRemove(ctx, iq, from, server, username, password)
		}
		if username != nil && password != nil {
			return m.tryRegisterOrSetPassword(ctx, iq, from, server, username.Text(), password.Text(), ip)
		}
		return domain.MakeError(iq, domain.ErrBadRequest(""))
	case "get":
		isRegistered, user := false, ""
		if from.LServer == server && from.User != "" {
			user = from.User
			isRegistered = m.auth.UserExists(ctx, from.LUser, from.LServer)
		}
		res := domain.NewElement("query", domain.NSRegister)
		if reg.RedirectURL != "" && !isRegistered {
			res.Append(domain.NewElement("instructions", "").AppendText("To register, visit " + reg.RedirectURL))
			x := domain.NewElement("x", "jabber:x:oob")
			x.Append(domain.NewElement("url", "").AppendText(reg.RedirectURL))
			res.Append(x)
			return domain.MakeIQResult(iq, res)
		}
		res.Append(domain.NewElement("instructions", "").AppendText("Choose a username and password to register with this server"))
		if isRegistered {
			res.Append(domain.NewElement("registered", ""))
		}
		res.Append(domain.NewElement("username", "").AppendText(user))
		res.Append(domain.NewElement("password", ""))
		return domain.MakeIQResult(iq, res)
	}
	return domain.MakeError(iq, domain.ErrBadRequest(""))
}

func (m *Modules) processRemove(ctx context.Context, iq *domain.Element, from domain.JID, server string, username, password *domain.Element) *domain.Element {
	if username != nil {
		user := username.Text()
		lu, _ := domain.Nodeprep(user)
		if from.LUser != "" && from.LUser == lu && from.LServer == server {
			m.router.Route(ctx, domain.MakeIQResult(iq, nil))
			_ = m.auth.RemoveUser(ctx, user, server)
			return nil
		}
		if password == nil {
			return domain.MakeError(iq, domain.ErrBadRequest("No 'password' found in this query"))
		}
		if !m.auth.CheckPassword(ctx, user, "", server, password.Text()) {
			return domain.MakeError(iq, domain.ErrForbidden("Incorrect password"))
		}
		m.router.Route(ctx, domain.MakeIQResult(iq, nil))
		_ = m.auth.RemoveUser(ctx, user, server)
		return nil
	}
	if from.LUser != "" && from.LServer == server {
		res := domain.MakeIQResult(iq, nil)
		res.SetAttr("from", from.String())
		res.SetAttr("to", from.String())
		m.router.Route(ctx, res)
		_ = m.auth.RemoveUser(ctx, from.LUser, server)
		return nil
	}
	return domain.MakeError(iq, domain.ErrNotAllowedStanza("The query is only allowed from local users"))
}

func (m *Modules) tryRegisterOrSetPassword(ctx context.Context, iq *domain.Element, from domain.JID, server, user, password string, ip net.IP) *domain.Element {
	lu, _ := domain.Nodeprep(user)
	if from.LUser != "" && from.LUser == lu && from.LServer == server {
		return m.trySetPassword(ctx, iq, user, server, password)
	}
	if from.LServer != "" && !m.matcher.MatchRule(m.cfg.Modules.Register.AccessFrom, domain.MatchJID(from)) {
		return domain.MakeError(iq, domain.ErrForbidden("Access denied by service policy"))
	}
	if err := m.tryRegister(ctx, user, server, password, ip); err != nil {
		var se domain.StanzaError
		if errors.As(err, &se) {
			return domain.MakeError(iq, se)
		}
		return domain.MakeError(iq, domain.ErrInternalServerError("Database failure"))
	}
	slog.Info("account registered", "jid", user+"@"+server, "ip", ipString(ip))
	return domain.MakeIQResult(iq, nil)
}

func ipString(ip net.IP) string {
	if ip == nil {
		return ""
	}
	return ip.String()
}

func (m *Modules) trySetPassword(ctx context.Context, iq *domain.Element, user, server, password string) *domain.Element {
	if _, ok := domain.Resourceprep(password); !ok {
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("The password contains unacceptable characters"))
	}
	if !m.isStrongPassword(password) {
		return domain.MakeError(iq, domain.ErrNotAcceptable("The password is too weak"))
	}
	switch err := m.auth.SetPassword(ctx, user, server, password); {
	case err == nil:
		slog.Info("password changed", "jid", user+"@"+server)
		return domain.MakeIQResult(iq, nil)
	case errors.Is(err, domain.ErrInvalidJID):
		return domain.MakeError(iq, domain.ErrJIDMalformed("Malformed username"))
	case errors.Is(err, domain.ErrInvalidPassword):
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("The password contains unacceptable characters"))
	case errors.Is(err, domain.ErrUserNotFound), errors.Is(err, domain.ErrNotAllowed):
		return domain.MakeError(iq, domain.ErrNotAllowedStanza("Changing password is not allowed"))
	default:
		return domain.MakeError(iq, domain.ErrInternalServerError("Database failure"))
	}
}

func (m *Modules) tryRegister(ctx context.Context, user, server, password string, ip net.IP) error {
	jid, err := domain.NewJID(user, server, "")
	if err != nil || jid.LUser == "" {
		return domain.ErrJIDMalformed("Malformed username")
	}
	reg := m.cfg.Modules.Register
	if !m.matcher.MatchRule(reg.Access, domain.MatchJID(jid)) || !m.matcher.MatchRule(reg.IPAccess, domain.MatchIP(ip)) {
		return domain.ErrForbidden("Access denied by service policy")
	}
	if !m.checkTimeout(ip) {
		return domain.ErrResourceConstraint("Users are not allowed to register accounts so quickly")
	}
	if _, ok := domain.Resourceprep(password); !ok {
		m.removeTimeout(ip)
		return domain.ErrNotAcceptable("The password contains unacceptable characters")
	}
	if !m.isStrongPassword(password) {
		m.removeTimeout(ip)
		return domain.ErrNotAcceptable("The password is too weak")
	}
	switch err := m.auth.TryRegister(ctx, user, server, password); {
	case err == nil:
		return nil
	case errors.Is(err, domain.ErrUserExists):
		m.removeTimeout(ip)
		return domain.ErrConflict("User already exists")
	case errors.Is(err, domain.ErrInvalidJID):
		m.removeTimeout(ip)
		return domain.ErrJIDMalformed("Malformed username")
	case errors.Is(err, domain.ErrInvalidPassword):
		m.removeTimeout(ip)
		return domain.ErrNotAcceptable("The password contains unacceptable characters")
	case errors.Is(err, domain.ErrNotMyHost), errors.Is(err, domain.ErrNotAllowed):
		m.removeTimeout(ip)
		return domain.ErrNotAllowedStanza("Not allowed")
	default:
		m.removeTimeout(ip)
		return domain.ErrInternalServerError("Database failure")
	}
}

func (m *Modules) checkTimeout(ip net.IP) bool {
	timeout := m.cfg.RegistrationTimeout
	if ip == nil || timeout <= 0 {
		return true
	}
	now := time.Now()
	key := ip.String()
	m.regMu.Lock()
	defer m.regMu.Unlock()
	for k, t := range m.regIPs {
		if now.Sub(t) > timeout {
			delete(m.regIPs, k)
		}
	}
	if _, ok := m.regIPs[key]; ok {
		return false
	}
	m.regIPs[key] = now
	return true
}

func (m *Modules) removeTimeout(ip net.IP) {
	if ip == nil {
		return
	}
	m.regMu.Lock()
	delete(m.regIPs, ip.String())
	m.regMu.Unlock()
}

func (m *Modules) isStrongPassword(password string) bool {
	strength := m.cfg.Modules.Register.PasswordStrength
	return strength == 0 || Entropy(password) >= float64(strength)
}

func Entropy(s string) float64 {
	if s == "" {
		return 0
	}
	var digit, printable, lower, upper, other int
	for i := 0; i < len(s); i++ {
		c := s[i]
		switch {
		case c >= 'a' && c <= 'z':
			lower = 26
		case c >= '0' && c <= '9':
			digit = 9
		case c >= 'A' && c <= 'Z':
			upper = 26
		case c >= 33 && c <= 126:
			printable = 33
		default:
			other = 128
		}
	}
	return float64(len(s)) * math.Log2(float64(digit+printable+lower+upper+other))
}
