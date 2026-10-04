package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/domain"
)

type Watcher interface {
	WatchCh() <-chan struct{}
	LastIndex() (uint64, error)
}

type CatalogStore interface {
	Watcher
	EnsureRegistration(r domain.CatalogRegistration) error
	DeregisterNode(node string) error
	DeregisterService(node, serviceID string) error
	DeregisterCheck(node, checkID string) error
	UpdateCheck(node, checkID string, status domain.HealthStatus, output string, ttlExpires time.Time) (domain.Check, error)

	Nodes() (uint64, []domain.Node, error)
	Node(name string) (uint64, *domain.Node, []domain.Service, error)
	Services() (uint64, []domain.Service, error)
	ServiceNodes(name string) (uint64, []domain.ServiceEntry, error)
	Checks() (uint64, []domain.Check, error)
}

type KVStore interface {
	Watcher
	KVGet(key string) (uint64, *domain.KVPair, error)
	KVList(prefix string) (uint64, []domain.KVPair, error)
	KVApply(req domain.KVRequest, now time.Time) (bool, error)
}

type SessionStore interface {
	Watcher
	SessionCreate(s domain.Session) error
	SessionGet(id string) (uint64, *domain.Session, error)
	SessionList() (uint64, []domain.Session, error)
	SessionRenew(id string, expires time.Time) (*domain.Session, error)
	SessionDestroy(id string, now time.Time) error
}

type IntentionStore interface {
	Watcher
	IntentionUpsert(i domain.Intention) error
	IntentionDelete(id string) error
	IntentionList() (uint64, []domain.Intention, error)
}

type StateStore interface {
	CatalogStore
	KVStore
	SessionStore
	IntentionStore
	Snapshot() (domain.Snapshot, error)
	Restore(domain.Snapshot) error
	Stats() (domain.Stats, error)
}

type Leadership interface {
	IsLeader() bool
}

type SnapshotStorage interface {
	Load() (*domain.Snapshot, error)
	Save(s domain.Snapshot) error
}

type Prober interface {
	Probe(ctx context.Context, c domain.Check) (domain.HealthStatus, string)
}

type Metrics interface {
	CheckRun(t domain.CheckType, s domain.HealthStatus)
	SessionInvalidated(reason string)
	BlockingQueryWoken()
}
