package redisstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"maps"
	"math/rand/v2"
	"slices"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/domain"
	goredis "github.com/redis/go-redis/v9"
)

const (
	hNodes      = "nodes"
	hServices   = "services"
	hChecks     = "checks"
	hKV         = "kv"
	hSessions   = "sessions"
	hIntentions = "intentions"
	hLockDelay  = "lockdelay"
	zKVKeys     = "kvkeys"
	kIndex      = "index"
	kVersion    = "version"
	hTableIndex = "tindex"
	chEvents    = "events"

	maxRetries = 100
)

var catalogTables = []string{hNodes, hServices, hChecks}

type Options struct {
	Prefix       string
	PollInterval time.Duration
	OpTimeout    time.Duration
}

type Store struct {
	rdb  goredis.UniversalClient
	opts Options
	now  func() time.Time

	mu       sync.Mutex
	notify   chan struct{}
	lastSeen uint64
	cancel   context.CancelFunc
	done     chan struct{}
}

func New(rdb goredis.UniversalClient, opts Options, now func() time.Time) *Store {
	if opts.PollInterval <= 0 {
		opts.PollInterval = time.Second
	}
	if opts.OpTimeout <= 0 {
		opts.OpTimeout = 5 * time.Second
	}
	return &Store{rdb: rdb, opts: opts, now: now, notify: make(chan struct{})}
}

func (s *Store) k(name string) string { return s.opts.Prefix + name }

func (s *Store) ctx() (context.Context, context.CancelFunc) {
	return context.WithTimeout(context.Background(), s.opts.OpTimeout)
}

func (s *Store) Start(ctx context.Context) error {
	if err := s.rdb.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("redis: %w", err)
	}
	last, err := s.LastIndex()
	if err != nil {
		return err
	}
	s.lastSeen = last

	runCtx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.done = make(chan struct{})
	sub := s.rdb.Subscribe(runCtx, s.k(chEvents))
	if _, err := sub.Receive(ctx); err != nil {
		cancel()
		_ = sub.Close()
		return fmt.Errorf("redis subscribe: %w", err)
	}
	go func() {
		defer close(s.done)
		defer sub.Close()
		msgs := sub.Channel()
		poll := time.NewTicker(s.opts.PollInterval)
		defer poll.Stop()
		for {
			select {
			case <-runCtx.Done():
				return
			case msg, ok := <-msgs:
				if !ok {
					return
				}
				if idx, err := strconv.ParseUint(msg.Payload, 10, 64); err == nil {
					s.observe(idx)
				}
			case <-poll.C:
				if idx, err := s.LastIndex(); err == nil {
					s.observe(idx)
				}
			}
		}
	}()
	return nil
}

func (s *Store) Close() {
	if s.cancel != nil {
		s.cancel()
		<-s.done
	}
}

func (s *Store) observe(idx uint64) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if idx == s.lastSeen {
		return
	}
	s.lastSeen = idx
	close(s.notify)
	s.notify = make(chan struct{})
}

func (s *Store) WatchCh() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.notify
}

func (s *Store) LastIndex() (uint64, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	return getUint(s.rdb.Get(ctx, s.k(kIndex)))
}

func getUint(cmd *goredis.StringCmd) (uint64, error) {
	v, err := cmd.Uint64()
	if errors.Is(err, goredis.Nil) {
		return 0, nil
	}
	return v, err
}

func (s *Store) tableIndex(ctx context.Context, tables ...string) (uint64, error) {
	fields := make([]string, len(tables))
	copy(fields, tables)
	vals, err := s.rdb.HMGet(ctx, s.k(hTableIndex), fields...).Result()
	if err != nil {
		return 0, err
	}
	var out uint64
	for _, v := range vals {
		if str, ok := v.(string); ok {
			n, _ := strconv.ParseUint(str, 10, 64)
			out = max(out, n)
		}
	}
	return out, nil
}

func decodeAll[T any](raw map[string]string) ([]T, error) {
	out := make([]T, 0, len(raw))
	for _, k := range slices.Sorted(maps.Keys(raw)) {
		var v T
		if err := json.Unmarshal([]byte(raw[k]), &v); err != nil {
			return nil, fmt.Errorf("decode %s: %w", k, err)
		}
		out = append(out, v)
	}
	return out, nil
}

func field(node, id string) string { return node + "\x00" + id }

type catalogView struct {
	nodes    []domain.Node
	services []domain.Service
	checks   []domain.Check
}

func (s *Store) readCatalog(ctx context.Context, withNodes, withServices, withChecks bool) (uint64, catalogView, error) {
	var view catalogView
	idx, err := s.tableIndex(ctx, catalogTables...)
	if err != nil {
		return 0, view, err
	}
	var nodes, services, checks *goredis.MapStringStringCmd
	_, err = s.rdb.TxPipelined(ctx, func(p goredis.Pipeliner) error {
		if withNodes {
			nodes = p.HGetAll(ctx, s.k(hNodes))
		}
		if withServices {
			services = p.HGetAll(ctx, s.k(hServices))
		}
		if withChecks {
			checks = p.HGetAll(ctx, s.k(hChecks))
		}
		return nil
	})
	if err != nil {
		return 0, view, err
	}
	if nodes != nil {
		if view.nodes, err = decodeAll[domain.Node](nodes.Val()); err != nil {
			return 0, view, err
		}
	}
	if services != nil {
		if view.services, err = decodeAll[domain.Service](services.Val()); err != nil {
			return 0, view, err
		}
	}
	if checks != nil {
		if view.checks, err = decodeAll[domain.Check](checks.Val()); err != nil {
			return 0, view, err
		}
	}
	return idx, view, nil
}

func (s *Store) Nodes() (uint64, []domain.Node, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, v, err := s.readCatalog(ctx, true, false, false)
	return idx, v.nodes, err
}

func (s *Store) Node(name string) (uint64, *domain.Node, []domain.Service, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, v, err := s.readCatalog(ctx, true, true, false)
	if err != nil {
		return 0, nil, nil, err
	}
	for _, n := range v.nodes {
		if n.Name == name {
			out := []domain.Service{}
			for _, svc := range v.services {
				if svc.Node == name {
					out = append(out, svc)
				}
			}
			return idx, &n, out, nil
		}
	}
	return idx, nil, nil, nil
}

func (s *Store) Services() (uint64, []domain.Service, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, v, err := s.readCatalog(ctx, false, true, false)
	return idx, v.services, err
}

func (s *Store) ServiceNodes(name string) (uint64, []domain.ServiceEntry, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, v, err := s.readCatalog(ctx, true, true, true)
	if err != nil {
		return 0, nil, err
	}
	nodes := map[string]domain.Node{}
	for _, n := range v.nodes {
		nodes[n.Name] = n
	}
	var out []domain.ServiceEntry
	for _, svc := range v.services {
		n, ok := nodes[svc.Node]
		if svc.Name != name || !ok {
			continue
		}
		entry := domain.ServiceEntry{Node: n, Service: svc}
		for _, c := range v.checks {
			if c.Node == svc.Node && (c.ServiceID == "" || c.ServiceID == svc.ID) {
				entry.Checks = append(entry.Checks, c)
			}
		}
		out = append(out, entry)
	}
	return idx, out, nil
}

func (s *Store) Checks() (uint64, []domain.Check, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, v, err := s.readCatalog(ctx, false, false, true)
	return idx, v.checks, err
}

func (s *Store) KVGet(key string) (uint64, *domain.KVPair, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, err := s.tableIndex(ctx, hKV)
	if err != nil {
		return 0, nil, err
	}
	raw, err := s.rdb.HGet(ctx, s.k(hKV), key).Result()
	if errors.Is(err, goredis.Nil) {
		return idx, nil, nil
	}
	if err != nil {
		return 0, nil, err
	}
	var p domain.KVPair
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		return 0, nil, err
	}
	return idx, &p, nil
}

func lexRange(prefix string) (string, string) {
	if prefix == "" {
		return "-", "+"
	}
	return "[" + prefix, "[" + prefix + "\xff"
}

func (s *Store) KVList(prefix string) (uint64, []domain.KVPair, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, err := s.tableIndex(ctx, hKV)
	if err != nil {
		return 0, nil, err
	}
	lo, hi := lexRange(prefix)
	keys, err := s.rdb.ZRangeByLex(ctx, s.k(zKVKeys), &goredis.ZRangeBy{Min: lo, Max: hi}).Result()
	if err != nil || len(keys) == 0 {
		return idx, nil, err
	}
	vals, err := s.rdb.HMGet(ctx, s.k(hKV), keys...).Result()
	if err != nil {
		return 0, nil, err
	}
	out := make([]domain.KVPair, 0, len(vals))
	for _, v := range vals {
		str, ok := v.(string)
		if !ok {
			continue
		}
		var p domain.KVPair
		if err := json.Unmarshal([]byte(str), &p); err != nil {
			return 0, nil, err
		}
		out = append(out, p)
	}
	return idx, out, nil
}

func (s *Store) SessionGet(id string) (uint64, *domain.Session, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, err := s.tableIndex(ctx, hSessions)
	if err != nil {
		return 0, nil, err
	}
	raw, err := s.rdb.HGet(ctx, s.k(hSessions), id).Result()
	if errors.Is(err, goredis.Nil) {
		return idx, nil, nil
	}
	if err != nil {
		return 0, nil, err
	}
	var sess domain.Session
	if err := json.Unmarshal([]byte(raw), &sess); err != nil {
		return 0, nil, err
	}
	return idx, &sess, nil
}

func (s *Store) SessionList() (uint64, []domain.Session, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, err := s.tableIndex(ctx, hSessions)
	if err != nil {
		return 0, nil, err
	}
	raw, err := s.rdb.HGetAll(ctx, s.k(hSessions)).Result()
	if err != nil {
		return 0, nil, err
	}
	out, err := decodeAll[domain.Session](raw)
	return idx, out, err
}

func (s *Store) IntentionList() (uint64, []domain.Intention, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	idx, err := s.tableIndex(ctx, hIntentions)
	if err != nil {
		return 0, nil, err
	}
	raw, err := s.rdb.HGetAll(ctx, s.k(hIntentions)).Result()
	if err != nil {
		return 0, nil, err
	}
	out, err := decodeAll[domain.Intention](raw)
	domain.SortIntentions(out)
	return idx, out, err
}

func (s *Store) EnsureRegistration(r domain.CatalogRegistration) error {
	name := r.Node.Name
	if name == "" {
		return domain.Invalid("missing node name")
	}
	return s.update(func(t *txn) error {
		serviceNames := map[string]string{}
		if r.Service != nil {
			serviceNames[r.Service.ID] = r.Service.Name
		}
		for _, c := range r.Checks {
			if c.ServiceID == "" || serviceNames[c.ServiceID] != "" {
				continue
			}
			var svc domain.Service
			ok, err := t.get(hServices, field(name, c.ServiceID), &svc)
			if err != nil {
				return err
			}
			if !ok {
				return domain.Invalid("check %q refers to unknown service %q on node %q", c.ID, c.ServiceID, name)
			}
			serviceNames[c.ServiceID] = svc.Name
		}

		var oldNode domain.Node
		ok, err := t.get(hNodes, name, &oldNode)
		if err != nil {
			return err
		}
		if n, changed := domain.MergeNode(ptrIf(ok, &oldNode), r.Node, r.SkipNodeUpdate, t.idx); changed {
			t.put(hNodes, name, n)
			t.touch(hNodes)
		}

		if r.Service != nil {
			in := *r.Service
			in.Node = name
			var old domain.Service
			ok, err := t.get(hServices, field(name, in.ID), &old)
			if err != nil {
				return err
			}
			if svc, changed := domain.MergeService(ptrIf(ok, &old), in, t.idx); changed {
				t.put(hServices, field(name, in.ID), svc)
				t.touch(hServices)
			}
		}

		for _, in := range r.Checks {
			in.Node = name
			if in.ServiceID != "" {
				in.ServiceName = serviceNames[in.ServiceID]
			}
			var old domain.Check
			ok, err := t.get(hChecks, field(name, in.ID), &old)
			if err != nil {
				return err
			}
			c, changed := domain.MergeCheck(ptrIf(ok, &old), in, t.now, t.idx)
			if !changed {
				continue
			}
			t.put(hChecks, field(name, c.ID), c)
			t.touch(hChecks)
			if c.Status == domain.HealthCritical {
				if err := t.invalidateSessionsForCheck(name, c.ID); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func ptrIf[T any](ok bool, v *T) *T {
	if ok {
		return v
	}
	return nil
}

func (s *Store) DeregisterNode(node string) error {
	return s.update(func(t *txn) error {
		var n domain.Node
		ok, err := t.get(hNodes, node, &n)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrNotFound
		}
		t.del(hNodes, node)
		for _, h := range []string{hServices, hChecks} {
			all, err := t.all(h)
			if err != nil {
				return err
			}
			for f := range all {
				if strings.HasPrefix(f, node+"\x00") {
					t.del(h, f)
				}
			}
		}
		sessions, err := t.sessions()
		if err != nil {
			return err
		}
		for _, sess := range sessions {
			if sess.Node == node {
				if err := t.destroySession(sess); err != nil {
					return err
				}
			}
		}
		t.touch(hNodes, hServices, hChecks, hSessions, hKV)
		return nil
	})
}

func (s *Store) DeregisterService(node, serviceID string) error {
	return s.update(func(t *txn) error {
		var svc domain.Service
		ok, err := t.get(hServices, field(node, serviceID), &svc)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrNotFound
		}
		t.del(hServices, field(node, serviceID))
		checks, err := t.all(hChecks)
		if err != nil {
			return err
		}
		for f, raw := range checks {
			var c domain.Check
			if err := json.Unmarshal([]byte(raw), &c); err != nil {
				return err
			}
			if c.Node == node && c.ServiceID == serviceID {
				t.del(hChecks, f)
			}
		}
		t.touch(hServices, hChecks)
		return nil
	})
}

func (s *Store) DeregisterCheck(node, checkID string) error {
	return s.update(func(t *txn) error {
		var c domain.Check
		ok, err := t.get(hChecks, field(node, checkID), &c)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrNotFound
		}
		t.del(hChecks, field(node, checkID))
		t.touch(hChecks)
		return t.invalidateSessionsForCheck(node, checkID)
	})
}

func (s *Store) UpdateCheck(node, checkID string, status domain.HealthStatus, output string, ttlExpires time.Time) (domain.Check, error) {
	var out domain.Check
	err := s.update(func(t *txn) error {
		var c domain.Check
		ok, err := t.get(hChecks, field(node, checkID), &c)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrNotFound
		}
		deadlineMoved := !c.TTLExpires.Equal(ttlExpires)
		c.TTLExpires = ttlExpires
		changed := c.SetStatus(status, output, t.now, t.idx)
		out = c
		if !changed {
			if deadlineMoved {
				t.put(hChecks, field(node, checkID), c)
			}
			return nil
		}
		t.put(hChecks, field(node, checkID), c)
		t.touch(hChecks)
		if status == domain.HealthCritical {
			return t.invalidateSessionsForCheck(node, checkID)
		}
		return nil
	})
	return out, err
}

func (s *Store) KVApply(req domain.KVRequest, now time.Time) (bool, error) {
	var ok bool
	err := s.update(func(t *txn) error {
		ok = false
		if req.Op == domain.KVDelTree {
			keys, err := t.kvKeys(req.Pair.Key)
			if err != nil {
				return err
			}
			for _, k := range keys {
				t.del(hKV, k)
			}
			if len(keys) > 0 {
				t.touch(hKV)
			}
			ok = true
			return nil
		}

		key := req.Pair.Key
		var existing domain.KVPair
		found, err := t.get(hKV, key, &existing)
		if err != nil {
			return err
		}
		sessionExists := false
		if req.Session != "" {
			var sess domain.Session
			if sessionExists, err = t.get(hSessions, req.Session, &sess); err != nil {
				return err
			}
		}
		var until time.Time
		var untilNanos int64
		hasDelay, err := t.get(hLockDelay, key, &untilNanos)
		if err != nil {
			return err
		}
		if hasDelay {
			until = time.Unix(0, untilNanos)
			if !now.Before(until) {
				t.del(hLockDelay, key)
			}
		}
		res, err := domain.ApplyKV(req, ptrIf(found, &existing), sessionExists, until, now, t.idx)
		if err != nil {
			return err
		}
		ok = res.OK
		if !res.Changed {
			return nil
		}
		if res.Delete {
			t.del(hKV, key)
		} else {
			t.put(hKV, key, *res.Put)
		}
		t.touch(hKV)
		return nil
	})
	return ok, err
}

func (s *Store) SessionCreate(sess domain.Session) error {
	return s.update(func(t *txn) error {
		var n domain.Node
		ok, err := t.get(hNodes, sess.Node, &n)
		if err != nil {
			return err
		}
		if !ok {
			return domain.Invalid("missing node registration %q", sess.Node)
		}
		for _, id := range sess.NodeChecks {
			var c domain.Check
			ok, err := t.get(hChecks, field(sess.Node, id), &c)
			if err != nil {
				return err
			}
			if err := domain.CheckBindable(id, &c, ok); err != nil {
				return err
			}
		}
		var existing domain.Session
		if ok, err := t.get(hSessions, sess.ID, &existing); err != nil || ok {
			if err != nil {
				return err
			}
			return domain.Invalid("session %q already exists", sess.ID)
		}
		out := domain.CloneSession(sess)
		out.CreateIndex, out.ModifyIndex = t.idx, t.idx
		t.put(hSessions, sess.ID, out)
		t.touch(hSessions)
		return nil
	})
}

func (s *Store) SessionRenew(id string, expires time.Time) (*domain.Session, error) {
	var out *domain.Session
	err := s.update(func(t *txn) error {
		var sess domain.Session
		ok, err := t.get(hSessions, id, &sess)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrNotFound
		}
		sess.Expires = expires
		t.put(hSessions, id, sess)
		out = &sess
		return nil
	})
	return out, err
}

func (s *Store) SessionDestroy(id string, now time.Time) error {
	return s.update(func(t *txn) error {
		t.now = now
		var sess domain.Session
		ok, err := t.get(hSessions, id, &sess)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrNotFound
		}
		return t.destroySession(sess)
	})
}

func (s *Store) IntentionUpsert(in domain.Intention) error {
	return s.update(func(t *txn) error {
		all, err := t.all(hIntentions)
		if err != nil {
			return err
		}
		list, err := decodeAll[domain.Intention](all)
		if err != nil {
			return err
		}
		out := domain.CloneIntention(in)
		out.CreateIndex, out.ModifyIndex = t.idx, t.idx
		for _, other := range list {
			if other.ID == in.ID {
				out.CreateIndex = other.CreateIndex
			} else if other.SourceName == in.SourceName && other.DestinationName == in.DestinationName {
				return domain.Invalid("duplicate intention found: %s => %s", in.SourceName, in.DestinationName)
			}
		}
		t.put(hIntentions, in.ID, out)
		t.touch(hIntentions)
		return nil
	})
}

func (s *Store) IntentionDelete(id string) error {
	return s.update(func(t *txn) error {
		var in domain.Intention
		ok, err := t.get(hIntentions, id, &in)
		if err != nil {
			return err
		}
		if !ok {
			return domain.ErrNotFound
		}
		t.del(hIntentions, id)
		t.touch(hIntentions)
		return nil
	})
}

func (s *Store) Snapshot() (domain.Snapshot, error) {
	ctx, cancel := s.ctx()
	defer cancel()
	cmds := map[string]*goredis.MapStringStringCmd{}
	var index *goredis.StringCmd
	_, err := s.rdb.TxPipelined(ctx, func(p goredis.Pipeliner) error {
		index = p.Get(ctx, s.k(kIndex))
		for _, h := range []string{hNodes, hServices, hChecks, hKV, hSessions, hIntentions} {
			cmds[h] = p.HGetAll(ctx, s.k(h))
		}
		return nil
	})
	if err != nil && !errors.Is(err, goredis.Nil) {
		return domain.Snapshot{}, err
	}
	idx, err := getUint(index)
	if err != nil {
		return domain.Snapshot{}, err
	}
	snap := domain.Snapshot{Index: idx, TakenAt: s.now()}
	if snap.Nodes, err = decodeAll[domain.Node](cmds[hNodes].Val()); err != nil {
		return snap, err
	}
	if snap.Services, err = decodeAll[domain.Service](cmds[hServices].Val()); err != nil {
		return snap, err
	}
	if snap.Checks, err = decodeAll[domain.Check](cmds[hChecks].Val()); err != nil {
		return snap, err
	}
	if snap.KV, err = decodeAll[domain.KVPair](cmds[hKV].Val()); err != nil {
		return snap, err
	}
	if snap.Sessions, err = decodeAll[domain.Session](cmds[hSessions].Val()); err != nil {
		return snap, err
	}
	snap.Intentions, err = decodeAll[domain.Intention](cmds[hIntentions].Val())
	return snap, err
}

func (s *Store) Restore(snap domain.Snapshot) error {
	ctx, cancel := s.ctx()
	defer cancel()
	put := func(p goredis.Pipeliner, h, f string, v any) error {
		b, err := json.Marshal(v)
		if err != nil {
			return err
		}
		p.HSet(ctx, s.k(h), f, b)
		return nil
	}
	_, err := s.rdb.TxPipelined(ctx, func(p goredis.Pipeliner) error {
		for _, h := range []string{hNodes, hServices, hChecks, hKV, hSessions, hIntentions, hLockDelay, zKVKeys, hTableIndex} {
			p.Del(ctx, s.k(h))
		}
		for _, n := range snap.Nodes {
			if err := put(p, hNodes, n.Name, n); err != nil {
				return err
			}
		}
		for _, svc := range snap.Services {
			if err := put(p, hServices, field(svc.Node, svc.ID), svc); err != nil {
				return err
			}
		}
		for _, c := range snap.Checks {
			if err := put(p, hChecks, field(c.Node, c.ID), c); err != nil {
				return err
			}
		}
		for _, kv := range snap.KV {
			if err := put(p, hKV, kv.Key, kv); err != nil {
				return err
			}
			p.ZAdd(ctx, s.k(zKVKeys), goredis.Z{Member: kv.Key})
		}
		for _, sess := range snap.Sessions {
			if err := put(p, hSessions, sess.ID, sess); err != nil {
				return err
			}
		}
		for _, in := range snap.Intentions {
			if err := put(p, hIntentions, in.ID, in); err != nil {
				return err
			}
		}
		p.Set(ctx, s.k(kIndex), snap.Index, 0)
		for _, h := range []string{hNodes, hServices, hChecks, hKV, hSessions, hIntentions} {
			p.HSet(ctx, s.k(hTableIndex), h, snap.Index)
		}
		p.Incr(ctx, s.k(kVersion))
		return nil
	})
	if err != nil {
		return err
	}
	s.publish(ctx, snap.Index)
	return nil
}

func (s *Store) Stats() (domain.Stats, error) {
	snap, err := s.Snapshot()
	if err != nil {
		return domain.Stats{}, err
	}
	return domain.StatsOf(snap), nil
}

func (s *Store) publish(ctx context.Context, idx uint64) {
	s.observe(idx)
	if err := s.rdb.Publish(ctx, s.k(chEvents), idx).Err(); err != nil {
		log.Printf("redis publish: %v", err)
	}
}

func (s *Store) update(fn func(t *txn) error) error {
	ctx, cancel := s.ctx()
	defer cancel()
	for attempt := 0; attempt < maxRetries; attempt++ {
		var committed *txn
		err := s.rdb.Watch(ctx, func(tx *goredis.Tx) error {
			cur, err := getUint(tx.Get(ctx, s.k(kIndex)))
			if err != nil {
				return err
			}
			t := &txn{s: s, ctx: ctx, tx: tx, idx: cur + 1, now: s.now(), staged: map[string]map[string]*string{}, touched: map[string]bool{}}
			if err := fn(t); err != nil {
				return err
			}
			if err := t.commit(); err != nil {
				return err
			}
			committed = t
			return nil
		}, s.k(kVersion))
		if errors.Is(err, goredis.TxFailedErr) {
			time.Sleep(time.Duration(rand.IntN(5)+1) * time.Millisecond)
			continue
		}
		if err == nil && committed != nil && len(committed.touched) > 0 {
			s.publish(ctx, committed.idx)
		}
		return err
	}
	return errors.New("redis: too much write contention, giving up")
}

type txn struct {
	s       *Store
	ctx     context.Context
	tx      *goredis.Tx
	idx     uint64
	now     time.Time
	staged  map[string]map[string]*string
	touched map[string]bool
}

func (t *txn) touch(tables ...string) {
	for _, tb := range tables {
		t.touched[tb] = true
	}
}

func (t *txn) get(h, f string, v any) (bool, error) {
	if fields, ok := t.staged[h]; ok {
		if raw, ok := fields[f]; ok {
			if raw == nil {
				return false, nil
			}
			return true, json.Unmarshal([]byte(*raw), v)
		}
	}
	raw, err := t.tx.HGet(t.ctx, t.s.k(h), f).Result()
	if errors.Is(err, goredis.Nil) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	return true, json.Unmarshal([]byte(raw), v)
}

func (t *txn) all(h string) (map[string]string, error) {
	out, err := t.tx.HGetAll(t.ctx, t.s.k(h)).Result()
	if err != nil {
		return nil, err
	}
	for f, raw := range t.staged[h] {
		if raw == nil {
			delete(out, f)
		} else {
			out[f] = *raw
		}
	}
	return out, nil
}

func (t *txn) kvKeys(prefix string) ([]string, error) {
	lo, hi := lexRange(prefix)
	keys, err := t.tx.ZRangeByLex(t.ctx, t.s.k(zKVKeys), &goredis.ZRangeBy{Min: lo, Max: hi}).Result()
	if err != nil {
		return nil, err
	}
	set := map[string]bool{}
	for _, k := range keys {
		set[k] = true
	}
	for k, raw := range t.staged[hKV] {
		if strings.HasPrefix(k, prefix) {
			set[k] = raw != nil
		}
	}
	out := make([]string, 0, len(set))
	for k, live := range set {
		if live {
			out = append(out, k)
		}
	}
	slices.Sort(out)
	return out, nil
}

func (t *txn) put(h, f string, v any) {
	b, err := json.Marshal(v)
	if err != nil {
		panic(fmt.Sprintf("redisstore: encode %T: %v", v, err))
	}
	str := string(b)
	t.stage(h, f, &str)
}

func (t *txn) del(h, f string) { t.stage(h, f, nil) }

func (t *txn) stage(h, f string, v *string) {
	if t.staged[h] == nil {
		t.staged[h] = map[string]*string{}
	}
	t.staged[h][f] = v
}

func (t *txn) sessions() ([]domain.Session, error) {
	all, err := t.all(hSessions)
	if err != nil {
		return nil, err
	}
	return decodeAll[domain.Session](all)
}

func (t *txn) invalidateSessionsForCheck(node, checkID string) error {
	sessions, err := t.sessions()
	if err != nil {
		return err
	}
	for _, sess := range sessions {
		if sess.BindsCheck(node, checkID) {
			if err := t.destroySession(sess); err != nil {
				return err
			}
		}
	}
	return nil
}

func (t *txn) destroySession(sess domain.Session) error {
	t.del(hSessions, sess.ID)
	t.touch(hSessions, hKV)
	all, err := t.all(hKV)
	if err != nil {
		return err
	}
	for key, raw := range all {
		var p domain.KVPair
		if err := json.Unmarshal([]byte(raw), &p); err != nil {
			return err
		}
		if p.Session != sess.ID {
			continue
		}
		released, del := domain.ReleaseHeldKey(p, sess, t.idx)
		if del {
			t.del(hKV, key)
			continue
		}
		t.put(hKV, key, released)
		if sess.LockDelay > 0 {
			t.put(hLockDelay, key, t.now.Add(sess.LockDelay).UnixNano())
		}
	}
	return nil
}

func (t *txn) commit() error {
	if len(t.staged) == 0 && len(t.touched) == 0 {
		return nil
	}
	ctx, s := t.ctx, t.s
	_, err := t.tx.TxPipelined(ctx, func(p goredis.Pipeliner) error {
		for h, fields := range t.staged {
			for f, raw := range fields {
				if raw == nil {
					p.HDel(ctx, s.k(h), f)
					if h == hKV {
						p.ZRem(ctx, s.k(zKVKeys), f)
					}
					continue
				}
				p.HSet(ctx, s.k(h), f, *raw)
				if h == hKV {
					p.ZAdd(ctx, s.k(zKVKeys), goredis.Z{Member: f})
				}
			}
		}
		p.Incr(ctx, s.k(kVersion))
		if len(t.touched) > 0 {
			p.Set(ctx, s.k(kIndex), t.idx, 0)
			for tb := range t.touched {
				p.HSet(ctx, s.k(hTableIndex), tb, t.idx)
			}
		}
		return nil
	})
	return err
}
