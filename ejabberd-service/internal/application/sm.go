package application

import (
	"slices"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JIeeiroSst/ejabberd-service/internal/config"
	"github.com/JIeeiroSst/ejabberd-service/internal/domain"
	"github.com/JIeeiroSst/ejabberd-service/internal/port"
)

type usKey struct {
	user   string
	server string
}

type sessionEntry struct {
	session domain.Session
	sink    port.SessionSink
}

type PresentResource struct {
	Priority int
	Resource string
}

type SessionManager struct {
	cfg     *config.Config
	matcher *domain.ACLMatcher
	seq     atomic.Uint64
	mu      sync.RWMutex
	byUS    map[usKey][]*sessionEntry
}

func NewSessionManager(cfg *config.Config, matcher *domain.ACLMatcher) *SessionManager {
	return &SessionManager{cfg: cfg, matcher: matcher, byUS: map[usKey][]*sessionEntry{}}
}

func (m *SessionManager) MakeSID() domain.SID {
	return domain.SID{Time: time.Now(), Seq: m.seq.Add(1)}
}

func (m *SessionManager) OpenSession(sid domain.SID, jid domain.JID, prio *int, info domain.SessionInfo, sink port.SessionSink) {
	k := usKey{jid.LUser, jid.LServer}
	m.mu.Lock()
	m.byUS[k] = append(m.byUS[k], &sessionEntry{
		session: domain.Session{SID: sid, USR: jid.LJID(), Priority: prio, Info: info},
		sink:    sink,
	})
	m.mu.Unlock()
	m.checkForSessionsToReplace(jid)
}

func (m *SessionManager) CloseSession(sid domain.SID, jid domain.JID) {
	k := usKey{jid.LUser, jid.LServer}
	m.mu.Lock()
	defer m.mu.Unlock()
	list := m.byUS[k]
	out := list[:0]
	for _, e := range list {
		if e.session.SID != sid {
			out = append(out, e)
		}
	}
	if len(out) == 0 {
		delete(m.byUS, k)
		return
	}
	m.byUS[k] = out
}

func (m *SessionManager) SetPresence(sid domain.SID, jid domain.JID, prio int) {
	m.update(sid, jid, &prio)
}

func (m *SessionManager) UnsetPresence(sid domain.SID, jid domain.JID) {
	m.update(sid, jid, nil)
}

func (m *SessionManager) update(sid domain.SID, jid domain.JID, prio *int) {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, e := range m.byUS[usKey{jid.LUser, jid.LServer}] {
		if e.session.SID == sid {
			e.session.Priority = prio
		}
	}
}

func (m *SessionManager) checkForSessionsToReplace(jid domain.JID) {
	var same []*sessionEntry
	all := m.rawSessions(jid.LUser, jid.LServer)
	for _, e := range all {
		if e.session.USR.Resource == jid.LResource {
			same = append(same, e)
		}
	}
	if len(same) > 0 {
		max := maxEntry(same)
		for _, e := range same {
			if e != max {
				e.sink.Deliver(port.SessionEvent{Kind: port.EventReplaced})
			}
		}
	}
	if limit, ok := m.maxUserSessions(jid); ok && len(all) > limit {
		minE := all[0]
		for _, e := range all[1:] {
			if e.session.SID.Less(minE.session.SID) {
				minE = e
			}
		}
		minE.sink.Deliver(port.SessionEvent{Kind: port.EventReplaced})
	}
}

func (m *SessionManager) maxUserSessions(jid domain.JID) (int, bool) {
	v := m.matcher.MatchShaper("max_user_sessions", domain.MatchJID(jid.Bare()))
	n, err := strconv.Atoi(v)
	if err != nil {
		return 0, false
	}
	return n, true
}

func maxEntry(list []*sessionEntry) *sessionEntry {
	best := list[0]
	for _, e := range list[1:] {
		if best.session.SID.Less(e.session.SID) {
			best = e
		}
	}
	return best
}

func (m *SessionManager) cleanSessions(luser, lserver string) []*sessionEntry {
	list := m.rawSessions(luser, lserver)
	byRes := map[string]*sessionEntry{}
	var order []string
	for _, e := range list {
		r := e.session.USR.Resource
		cur, ok := byRes[r]
		if !ok {
			order = append(order, r)
			byRes[r] = e
			continue
		}
		if cur.session.SID.Less(e.session.SID) {
			byRes[r] = e
		}
	}
	out := make([]*sessionEntry, 0, len(order))
	for _, r := range order {
		out = append(out, byRes[r])
	}
	return out
}

func (m *SessionManager) Sessions(luser, lserver string) []domain.Session {
	var out []domain.Session
	for _, e := range m.cleanSessions(luser, lserver) {
		out = append(out, e.session)
	}
	return out
}

func (m *SessionManager) rawSessions(luser, lserver string) []*sessionEntry {
	m.mu.RLock()
	defer m.mu.RUnlock()
	list := m.byUS[usKey{luser, lserver}]
	out := make([]*sessionEntry, len(list))
	for i, e := range list {
		snapshot := *e
		out[i] = &snapshot
	}
	return out
}

func (m *SessionManager) resourceEntry(l domain.LJID) *sessionEntry {
	var same []*sessionEntry
	for _, e := range m.rawSessions(l.User, l.Server) {
		if e.session.USR.Resource == l.Resource {
			same = append(same, e)
		}
	}
	if len(same) == 0 {
		return nil
	}
	return maxEntry(same)
}

func (m *SessionManager) Sink(l domain.LJID) (port.SessionSink, bool) {
	e := m.resourceEntry(l)
	if e == nil {
		return nil, false
	}
	return e.sink, true
}

func (m *SessionManager) UserResources(luser, lserver string) []string {
	var out []string
	for _, e := range m.cleanSessions(luser, lserver) {
		out = append(out, e.session.USR.Resource)
	}
	return out
}

func (m *SessionManager) PresentResources(luser, lserver string) []PresentResource {
	var out []PresentResource
	for _, e := range m.cleanSessions(luser, lserver) {
		if e.session.Priority != nil {
			out = append(out, PresentResource{Priority: *e.session.Priority, Resource: e.session.USR.Resource})
		}
	}
	return out
}

func (m *SessionManager) IsExistingResource(l domain.LJID) bool {
	return m.resourceEntry(l) != nil
}

func (m *SessionManager) DeliverToUser(luser, lserver string, ev port.SessionEvent) {
	for _, e := range m.cleanSessions(luser, lserver) {
		e.sink.Deliver(ev)
	}
}

func (m *SessionManager) AllSessions() []domain.Session {
	m.mu.RLock()
	defer m.mu.RUnlock()
	var out []domain.Session
	for _, list := range m.byUS {
		for _, e := range list {
			out = append(out, e.session)
		}
	}
	slices.SortFunc(out, func(a, b domain.Session) int {
		if a.SID.Less(b.SID) {
			return -1
		}
		if b.SID.Less(a.SID) {
			return 1
		}
		return 0
	})
	return out
}

func (m *SessionManager) Count() int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for _, list := range m.byUS {
		n += len(list)
	}
	return n
}

func (m *SessionManager) VhostCount(lserver string) int {
	m.mu.RLock()
	defer m.mu.RUnlock()
	n := 0
	for k, list := range m.byUS {
		if k.server == lserver {
			n += len(list)
		}
	}
	return n
}
