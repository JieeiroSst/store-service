package domain

import (
	"strings"
	"time"
)

type KVPair struct {
	Key       string
	Value     []byte
	Flags     uint64
	Session   string
	LockIndex uint64
	RaftIndex
}

type KVOp string

const (
	KVSet     KVOp = "set"
	KVCAS     KVOp = "cas"
	KVLock    KVOp = "lock"
	KVUnlock  KVOp = "unlock"
	KVDelete  KVOp = "delete"
	KVDelCAS  KVOp = "delete-cas"
	KVDelTree KVOp = "delete-tree"
)

type KVRequest struct {
	Op      KVOp
	Pair    KVPair
	Session string
	CAS     uint64
}

func ValidateKey(key string, allowEmpty bool) error {
	if key == "" && !allowEmpty {
		return Invalid("missing key name")
	}
	if strings.HasPrefix(key, "/") {
		return Invalid("key must not start with '/'")
	}
	return nil
}

type SessionBehavior string

const (
	SessionRelease SessionBehavior = "release"
	SessionDelete  SessionBehavior = "delete"
)

const (
	SessionTTLMin    = 10 * time.Second
	SessionTTLMax    = 24 * time.Hour
	DefaultLockDelay = 15 * time.Second
)

type Session struct {
	ID         string
	Name       string
	Node       string
	Behavior   SessionBehavior
	TTL        time.Duration
	LockDelay  time.Duration
	NodeChecks []string
	Expires    time.Time
	RaftIndex
}

func (s *Session) Validate() error {
	if s.Behavior == "" {
		s.Behavior = SessionRelease
	}
	if s.Behavior != SessionRelease && s.Behavior != SessionDelete {
		return Invalid("invalid Behavior %q", s.Behavior)
	}
	if s.TTL != 0 && (s.TTL < SessionTTLMin || s.TTL > SessionTTLMax) {
		return Invalid("invalid session TTL %s, must be between %s and %s", s.TTL, SessionTTLMin, SessionTTLMax)
	}
	if s.LockDelay < 0 || s.LockDelay > 60*time.Second {
		return Invalid("invalid LockDelay %s, must be between 0s and 60s", s.LockDelay)
	}
	return nil
}
