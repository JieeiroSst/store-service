package memory

import (
	"maps"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

type table int

const (
	tNodes table = iota
	tServices
	tChecks
	tKV
	tSessions
	tIntentions
	tableCount
)

type Store struct {
	mu     sync.RWMutex
	now    func() time.Time
	index  uint64
	tables [tableCount]uint64
	notify chan struct{}

	nodes    map[string]*domain.Node
	services map[string]map[string]*domain.Service
	checks   map[string]map[string]*domain.Check

	kv        map[string]*domain.KVPair
	lockDelay map[string]time.Time

	sessions   map[string]*domain.Session
	intentions map[string]*domain.Intention
}

func NewStore() *Store {
	return NewStoreWithClock(time.Now)
}

func NewStoreWithClock(now func() time.Time) *Store {
	s := &Store{now: now, notify: make(chan struct{})}
	s.reset()
	return s
}

func (s *Store) reset() {
	s.nodes = map[string]*domain.Node{}
	s.services = map[string]map[string]*domain.Service{}
	s.checks = map[string]map[string]*domain.Check{}
	s.kv = map[string]*domain.KVPair{}
	s.lockDelay = map[string]time.Time{}
	s.sessions = map[string]*domain.Session{}
	s.intentions = map[string]*domain.Intention{}
}

func (s *Store) WatchCh() <-chan struct{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.notify
}

func (s *Store) LastIndex() (uint64, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.index, nil
}

func (s *Store) commit(idx uint64, touched ...table) {
	s.index = idx
	for _, t := range touched {
		s.tables[t] = idx
	}
	close(s.notify)
	s.notify = make(chan struct{})
}

func (s *Store) catalogIndex() uint64 {
	return max(s.tables[tNodes], s.tables[tServices], s.tables[tChecks])
}

func (s *Store) EnsureRegistration(r domain.CatalogRegistration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	name := r.Node.Name
	if name == "" {
		return domain.Invalid("missing node name")
	}
	for _, c := range r.Checks {
		if c.ServiceID == "" || (r.Service != nil && r.Service.ID == c.ServiceID) {
			continue
		}
		if _, ok := s.services[name][c.ServiceID]; !ok {
			return domain.Invalid("check %q refers to unknown service %q on node %q", c.ID, c.ServiceID, name)
		}
	}

	idx := s.index + 1
	now := s.now()
	var touched []table

	if n, changed := domain.MergeNode(s.nodes[name], r.Node, r.SkipNodeUpdate, idx); changed {
		s.nodes[name] = &n
		touched = append(touched, tNodes)
	}
	if r.Service != nil {
		in := *r.Service
		in.Node = name
		if s.services[name] == nil {
			s.services[name] = map[string]*domain.Service{}
		}
		if svc, changed := domain.MergeService(s.services[name][in.ID], in, idx); changed {
			s.services[name][in.ID] = &svc
			touched = append(touched, tServices)
		}
	}
	for _, in := range r.Checks {
		in.Node = name
		if in.ServiceID != "" {
			in.ServiceName = s.services[name][in.ServiceID].Name
		}
		if s.checks[name] == nil {
			s.checks[name] = map[string]*domain.Check{}
		}
		c, changed := domain.MergeCheck(s.checks[name][in.ID], in, now, idx)
		if !changed {
			continue
		}
		s.checks[name][c.ID] = &c
		touched = append(touched, tChecks)
		if c.Status == domain.HealthCritical && s.invalidateSessionsForCheck(name, c.ID, idx, now) {
			touched = append(touched, tSessions, tKV)
		}
	}

	if len(touched) > 0 {
		s.commit(idx, touched...)
	}
	return nil
}

func (s *Store) DeregisterNode(node string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[node]; !ok {
		return domain.ErrNotFound
	}
	idx := s.index + 1
	delete(s.nodes, node)
	delete(s.services, node)
	delete(s.checks, node)
	for id, sess := range s.sessions {
		if sess.Node == node {
			s.destroySession(id, idx, s.now())
		}
	}
	s.commit(idx, tNodes, tServices, tChecks, tSessions, tKV)
	return nil
}

func (s *Store) DeregisterService(node, serviceID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.services[node][serviceID]; !ok {
		return domain.ErrNotFound
	}
	delete(s.services[node], serviceID)
	for id, c := range s.checks[node] {
		if c.ServiceID == serviceID {
			delete(s.checks[node], id)
		}
	}
	s.commit(s.index+1, tServices, tChecks)
	return nil
}

func (s *Store) DeregisterCheck(node, checkID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.checks[node][checkID]; !ok {
		return domain.ErrNotFound
	}
	idx := s.index + 1
	delete(s.checks[node], checkID)
	touched := []table{tChecks}
	if s.invalidateSessionsForCheck(node, checkID, idx, s.now()) {
		touched = append(touched, tSessions, tKV)
	}
	s.commit(idx, touched...)
	return nil
}

func (s *Store) UpdateCheck(node, checkID string, status domain.HealthStatus, output string, ttlExpires time.Time) (domain.Check, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	c, ok := s.checks[node][checkID]
	if !ok {
		return domain.Check{}, domain.ErrNotFound
	}
	c.TTLExpires = ttlExpires
	idx := s.index + 1
	now := s.now()
	if !c.SetStatus(status, output, now, idx) {
		return domain.CloneCheck(*c), nil
	}
	touched := []table{tChecks}
	if status == domain.HealthCritical && s.invalidateSessionsForCheck(node, checkID, idx, now) {
		touched = append(touched, tSessions, tKV)
	}
	s.commit(idx, touched...)
	return domain.CloneCheck(*c), nil
}

func (s *Store) Nodes() (uint64, []domain.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Node, 0, len(s.nodes))
	for _, name := range slices.Sorted(maps.Keys(s.nodes)) {
		out = append(out, domain.CloneNode(*s.nodes[name]))
	}
	return s.catalogIndex(), out, nil
}

func (s *Store) Node(name string) (uint64, *domain.Node, []domain.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	n, ok := s.nodes[name]
	if !ok {
		return s.catalogIndex(), nil, nil, nil
	}
	node := domain.CloneNode(*n)
	return s.catalogIndex(), &node, s.nodeServices(name), nil
}

func (s *Store) nodeServices(node string) []domain.Service {
	byID := s.services[node]
	out := make([]domain.Service, 0, len(byID))
	for _, id := range slices.Sorted(maps.Keys(byID)) {
		out = append(out, domain.CloneService(*byID[id]))
	}
	return out
}

func (s *Store) Services() (uint64, []domain.Service, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Service
	for _, node := range slices.Sorted(maps.Keys(s.services)) {
		out = append(out, s.nodeServices(node)...)
	}
	return s.catalogIndex(), out, nil
}

func (s *Store) ServiceNodes(name string) (uint64, []domain.ServiceEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.ServiceEntry
	for _, node := range slices.Sorted(maps.Keys(s.services)) {
		for _, svc := range s.nodeServices(node) {
			if svc.Name != name {
				continue
			}
			entry := domain.ServiceEntry{Node: domain.CloneNode(*s.nodes[node]), Service: svc}
			for _, id := range slices.Sorted(maps.Keys(s.checks[node])) {
				c := s.checks[node][id]
				if c.ServiceID == "" || c.ServiceID == svc.ID {
					entry.Checks = append(entry.Checks, domain.CloneCheck(*c))
				}
			}
			out = append(out, entry)
		}
	}
	return s.catalogIndex(), out, nil
}

func (s *Store) Checks() (uint64, []domain.Check, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.Check
	for _, node := range slices.Sorted(maps.Keys(s.checks)) {
		for _, id := range slices.Sorted(maps.Keys(s.checks[node])) {
			out = append(out, domain.CloneCheck(*s.checks[node][id]))
		}
	}
	return s.catalogIndex(), out, nil
}

func (s *Store) KVGet(key string) (uint64, *domain.KVPair, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	p, ok := s.kv[key]
	if !ok {
		return s.tables[tKV], nil, nil
	}
	c := domain.CloneKV(*p)
	return s.tables[tKV], &c, nil
}

func (s *Store) KVList(prefix string) (uint64, []domain.KVPair, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var out []domain.KVPair
	for _, k := range slices.Sorted(maps.Keys(s.kv)) {
		if strings.HasPrefix(k, prefix) {
			out = append(out, domain.CloneKV(*s.kv[k]))
		}
	}
	return s.tables[tKV], out, nil
}

func (s *Store) KVApply(req domain.KVRequest, now time.Time) (bool, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	idx := s.index + 1

	if req.Op == domain.KVDelTree {
		deleted := false
		for k := range s.kv {
			if strings.HasPrefix(k, req.Pair.Key) {
				delete(s.kv, k)
				deleted = true
			}
		}
		if deleted {
			s.commit(idx, tKV)
		}
		return true, nil
	}

	key := req.Pair.Key
	_, sessionExists := s.sessions[req.Session]
	until := s.lockDelay[key]
	if !until.IsZero() && !now.Before(until) {
		delete(s.lockDelay, key)
	}
	res, err := domain.ApplyKV(req, s.kv[key], sessionExists, until, now, idx)
	if err != nil || !res.Changed {
		return res.OK, err
	}
	if res.Delete {
		delete(s.kv, key)
	} else {
		s.kv[key] = res.Put
	}
	s.commit(idx, tKV)
	return res.OK, nil
}

func (s *Store) SessionCreate(sess domain.Session) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.nodes[sess.Node]; !ok {
		return domain.Invalid("missing node registration %q", sess.Node)
	}
	for _, id := range sess.NodeChecks {
		c, ok := s.checks[sess.Node][id]
		if err := domain.CheckBindable(id, c, ok); err != nil {
			return err
		}
	}
	if _, ok := s.sessions[sess.ID]; ok {
		return domain.Invalid("session %q already exists", sess.ID)
	}
	idx := s.index + 1
	sess = domain.CloneSession(sess)
	sess.CreateIndex, sess.ModifyIndex = idx, idx
	s.sessions[sess.ID] = &sess
	s.commit(idx, tSessions)
	return nil
}

func (s *Store) SessionGet(id string) (uint64, *domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sess, ok := s.sessions[id]
	if !ok {
		return s.tables[tSessions], nil, nil
	}
	c := domain.CloneSession(*sess)
	return s.tables[tSessions], &c, nil
}

func (s *Store) SessionList() (uint64, []domain.Session, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Session, 0, len(s.sessions))
	for _, id := range slices.Sorted(maps.Keys(s.sessions)) {
		out = append(out, domain.CloneSession(*s.sessions[id]))
	}
	return s.tables[tSessions], out, nil
}

func (s *Store) SessionRenew(id string, expires time.Time) (*domain.Session, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	sess, ok := s.sessions[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	sess.Expires = expires
	c := domain.CloneSession(*sess)
	return &c, nil
}

func (s *Store) SessionDestroy(id string, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.sessions[id]; !ok {
		return domain.ErrNotFound
	}
	idx := s.index + 1
	s.destroySession(id, idx, now)
	s.commit(idx, tSessions, tKV)
	return nil
}

func (s *Store) destroySession(id string, idx uint64, now time.Time) {
	sess := s.sessions[id]
	delete(s.sessions, id)
	for k, p := range s.kv {
		if p.Session != id {
			continue
		}
		released, del := domain.ReleaseHeldKey(*p, *sess, idx)
		if del {
			delete(s.kv, k)
			continue
		}
		s.kv[k] = &released
		if sess.LockDelay > 0 {
			s.lockDelay[k] = now.Add(sess.LockDelay)
		}
	}
}

func (s *Store) invalidateSessionsForCheck(node, checkID string, idx uint64, now time.Time) bool {
	hit := false
	for id, sess := range s.sessions {
		if sess.BindsCheck(node, checkID) {
			s.destroySession(id, idx, now)
			hit = true
		}
	}
	return hit
}

func (s *Store) IntentionUpsert(in domain.Intention) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, other := range s.intentions {
		if id != in.ID && other.SourceName == in.SourceName && other.DestinationName == in.DestinationName {
			return domain.Invalid("duplicate intention found: %s => %s", in.SourceName, in.DestinationName)
		}
	}
	idx := s.index + 1
	in = domain.CloneIntention(in)
	in.CreateIndex, in.ModifyIndex = idx, idx
	if old, ok := s.intentions[in.ID]; ok {
		in.CreateIndex = old.CreateIndex
	}
	s.intentions[in.ID] = &in
	s.commit(idx, tIntentions)
	return nil
}

func (s *Store) IntentionDelete(id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.intentions[id]; !ok {
		return domain.ErrNotFound
	}
	delete(s.intentions, id)
	s.commit(s.index+1, tIntentions)
	return nil
}

func (s *Store) IntentionList() (uint64, []domain.Intention, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]domain.Intention, 0, len(s.intentions))
	for _, in := range s.intentions {
		out = append(out, domain.CloneIntention(*in))
	}
	domain.SortIntentions(out)
	return s.tables[tIntentions], out, nil
}

func (s *Store) Snapshot() (domain.Snapshot, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap := domain.Snapshot{Index: s.index, TakenAt: s.now()}
	for _, n := range s.nodes {
		snap.Nodes = append(snap.Nodes, domain.CloneNode(*n))
	}
	for _, byID := range s.services {
		for _, svc := range byID {
			snap.Services = append(snap.Services, domain.CloneService(*svc))
		}
	}
	for _, byID := range s.checks {
		for _, c := range byID {
			snap.Checks = append(snap.Checks, domain.CloneCheck(*c))
		}
	}
	for _, p := range s.kv {
		snap.KV = append(snap.KV, domain.CloneKV(*p))
	}
	for _, sess := range s.sessions {
		snap.Sessions = append(snap.Sessions, domain.CloneSession(*sess))
	}
	for _, in := range s.intentions {
		snap.Intentions = append(snap.Intentions, domain.CloneIntention(*in))
	}
	return snap, nil
}

func (s *Store) Restore(snap domain.Snapshot) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.reset()
	for _, n := range snap.Nodes {
		n := domain.CloneNode(n)
		s.nodes[n.Name] = &n
	}
	for _, svc := range snap.Services {
		svc := domain.CloneService(svc)
		if s.services[svc.Node] == nil {
			s.services[svc.Node] = map[string]*domain.Service{}
		}
		s.services[svc.Node][svc.ID] = &svc
	}
	for _, c := range snap.Checks {
		c := domain.CloneCheck(c)
		if s.checks[c.Node] == nil {
			s.checks[c.Node] = map[string]*domain.Check{}
		}
		s.checks[c.Node][c.ID] = &c
	}
	for _, p := range snap.KV {
		p := domain.CloneKV(p)
		s.kv[p.Key] = &p
	}
	for _, sess := range snap.Sessions {
		sess := domain.CloneSession(sess)
		s.sessions[sess.ID] = &sess
	}
	for _, in := range snap.Intentions {
		in := domain.CloneIntention(in)
		s.intentions[in.ID] = &in
	}
	all := make([]table, tableCount)
	for i := range all {
		all[i] = table(i)
	}
	s.commit(snap.Index, all...)
	return nil
}

func (s *Store) Stats() (domain.Stats, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var snap domain.Snapshot
	for _, byID := range s.services {
		for _, svc := range byID {
			snap.Services = append(snap.Services, *svc)
		}
	}
	for _, byID := range s.checks {
		for _, c := range byID {
			snap.Checks = append(snap.Checks, *c)
		}
	}
	st := domain.StatsOf(snap)
	st.Nodes, st.Keys, st.Sessions, st.Intentions = len(s.nodes), len(s.kv), len(s.sessions), len(s.intentions)
	return st, nil
}
