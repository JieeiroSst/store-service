package domain

import (
	"maps"
	"reflect"
	"slices"
	"strings"
	"time"
)

func CloneNode(n Node) Node {
	n.Meta = maps.Clone(n.Meta)
	return n
}

func CloneService(s Service) Service {
	s.Tags = slices.Clone(s.Tags)
	s.Meta = maps.Clone(s.Meta)
	return s
}

func CloneCheck(c Check) Check {
	if c.Header != nil {
		h := make(map[string][]string, len(c.Header))
		for k, v := range c.Header {
			h[k] = slices.Clone(v)
		}
		c.Header = h
	}
	return c
}

func CloneKV(p KVPair) KVPair {
	p.Value = slices.Clone(p.Value)
	return p
}

func CloneSession(s Session) Session {
	s.NodeChecks = slices.Clone(s.NodeChecks)
	return s
}

func CloneIntention(i Intention) Intention {
	i.Meta = maps.Clone(i.Meta)
	return i
}

func MergeNode(old *Node, in Node, skipUpdate bool, idx uint64) (Node, bool) {
	if old == nil {
		n := CloneNode(in)
		n.CreateIndex, n.ModifyIndex = idx, idx
		return n, true
	}
	if skipUpdate || (old.Address == in.Address && maps.Equal(old.Meta, in.Meta)) {
		return *old, false
	}
	n := *old
	n.Address = in.Address
	n.Meta = maps.Clone(in.Meta)
	n.ModifyIndex = idx
	return n, true
}

func MergeService(old *Service, in Service, idx uint64) (Service, bool) {
	svc := CloneService(in)
	if old != nil && sameService(*old, svc) {
		return *old, false
	}
	svc.CreateIndex, svc.ModifyIndex = idx, idx
	if old != nil {
		svc.CreateIndex = old.CreateIndex
	}
	return svc, true
}

func MergeCheck(old *Check, in Check, now time.Time, idx uint64) (Check, bool) {
	c := CloneCheck(in)
	if c.Status == "" {
		c.Status = HealthCritical
		if old != nil {
			c.Status, c.Output = old.Status, old.Output
		}
	}
	if old != nil {
		c.CriticalSince, c.TTLExpires = old.CriticalSince, old.TTLExpires
		if sameCheck(*old, c) {
			return *old, false
		}
		c.CreateIndex = old.CreateIndex
	} else {
		c.CreateIndex = idx
	}
	switch {
	case c.Status == HealthCritical && c.CriticalSince.IsZero():
		c.CriticalSince = now
	case c.Status != HealthCritical:
		c.CriticalSince = time.Time{}
	}
	c.ModifyIndex = idx
	return c, true
}

func (c *Check) SetStatus(status HealthStatus, output string, now time.Time, idx uint64) bool {
	if c.Status == status && c.Output == output {
		return false
	}
	switch {
	case status == HealthCritical && c.Status != HealthCritical:
		c.CriticalSince = now
	case status != HealthCritical:
		c.CriticalSince = time.Time{}
	}
	c.Status, c.Output, c.ModifyIndex = status, output, idx
	return true
}

func (s Session) BindsCheck(node, checkID string) bool {
	return s.Node == node && slices.Contains(s.NodeChecks, checkID)
}

type KVResult struct {
	OK      bool
	Changed bool
	Put     *KVPair
	Delete  bool
}

func ApplyKV(req KVRequest, existing *KVPair, sessionExists bool, lockedUntil, now time.Time, idx uint64) (KVResult, error) {
	put := func() *KVPair {
		p := &KVPair{Key: req.Pair.Key, Value: slices.Clone(req.Pair.Value), Flags: req.Pair.Flags}
		p.CreateIndex, p.ModifyIndex = idx, idx
		if existing != nil {
			p.CreateIndex = existing.CreateIndex
			p.Session = existing.Session
			p.LockIndex = existing.LockIndex
		}
		return p
	}
	written := func(p *KVPair) (KVResult, error) { return KVResult{OK: true, Changed: true, Put: p}, nil }
	deleted := KVResult{OK: true, Changed: true, Delete: true}

	switch req.Op {
	case KVSet:
		return written(put())

	case KVCAS:
		if req.CAS == 0 && existing != nil {
			return KVResult{}, nil
		}
		if req.CAS != 0 && (existing == nil || existing.ModifyIndex != req.CAS) {
			return KVResult{}, nil
		}
		return written(put())

	case KVLock:
		if !sessionExists {
			return KVResult{}, ErrInvalidSession
		}
		if now.Before(lockedUntil) {
			return KVResult{}, nil
		}
		if existing != nil && existing.Session != "" && existing.Session != req.Session {
			return KVResult{}, nil
		}
		p := put()
		if existing == nil || existing.Session != req.Session {
			p.LockIndex++
		}
		p.Session = req.Session
		return written(p)

	case KVUnlock:
		if existing == nil || existing.Session != req.Session {
			return KVResult{}, nil
		}
		p := put()
		p.Session = ""
		return written(p)

	case KVDelete:
		if existing == nil {
			return KVResult{OK: true}, nil
		}
		return deleted, nil

	case KVDelCAS:
		if existing == nil {
			return KVResult{OK: true}, nil
		}
		if existing.ModifyIndex != req.CAS {
			return KVResult{}, nil
		}
		return deleted, nil
	}
	return KVResult{}, Invalid("unknown KV operation %q", req.Op)
}

func ReleaseHeldKey(p KVPair, sess Session, idx uint64) (KVPair, bool) {
	if sess.Behavior == SessionDelete {
		return KVPair{}, true
	}
	p.Session = ""
	p.ModifyIndex = idx
	return p, false
}

func sameService(a, b Service) bool {
	a.RaftIndex, b.RaftIndex = RaftIndex{}, RaftIndex{}
	return reflect.DeepEqual(a, b)
}

func sameCheck(a, b Check) bool {
	a.RaftIndex, b.RaftIndex = RaftIndex{}, RaftIndex{}
	return reflect.DeepEqual(a, b)
}

func CheckBindable(id string, c *Check, ok bool) error {
	if !ok || c.ServiceID != "" {
		return Invalid("missing check %q registration", id)
	}
	if c.Status == HealthCritical {
		return Invalid("check %q is in critical state", id)
	}
	return nil
}

func SortIntentions(list []Intention) {
	slices.SortFunc(list, func(a, b Intention) int {
		if a.Precedence != b.Precedence {
			return b.Precedence - a.Precedence
		}
		if c := strings.Compare(a.DestinationName, b.DestinationName); c != 0 {
			return c
		}
		return strings.Compare(a.SourceName, b.SourceName)
	})
}

func StatsOf(snap Snapshot) Stats {
	st := Stats{
		Nodes:      len(snap.Nodes),
		Checks:     map[HealthStatus]int{},
		Keys:       len(snap.KV),
		Sessions:   len(snap.Sessions),
		Intentions: len(snap.Intentions),
		Instances:  len(snap.Services),
	}
	names := map[string]struct{}{}
	for _, svc := range snap.Services {
		names[svc.Name] = struct{}{}
	}
	st.Services = len(names)
	for _, c := range snap.Checks {
		st.Checks[c.Status]++
	}
	return st
}
