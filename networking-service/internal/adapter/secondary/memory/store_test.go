package memory

import (
	"testing"
	"time"

	"github.com/JIeeiroSst/networking-service/internal/adapter/secondary/storetest"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

func TestContract(t *testing.T) {
	storetest.Run(t, func(t *testing.T, now func() time.Time) port.StateStore {
		return NewStoreWithClock(now)
	})
}

func TestWatchFiresOnWriteOnly(t *testing.T) {
	s := NewStore()
	ch := s.WatchCh()
	if err := s.IntentionDelete("missing"); err == nil {
		t.Fatal("expected not found")
	}
	select {
	case <-ch:
		t.Fatal("watch fired without a write")
	default:
	}
	_ = s.SessionDestroy("missing", time.Now())
	if err := s.EnsureRegistration(registration()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	default:
		t.Fatal("watch did not fire on a write")
	}
}

func registration() domain.CatalogRegistration {
	return domain.CatalogRegistration{Node: domain.Node{Name: "n1", Address: "10.0.0.1"}}
}
