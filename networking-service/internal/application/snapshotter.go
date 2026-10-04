package application

import (
	"context"
	"log"
	"sync"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

type Snapshotter struct {
	store    port.StateStore
	storage  port.SnapshotStorage
	interval time.Duration

	mu     sync.Mutex
	saved  uint64
	cancel context.CancelFunc
	done   chan struct{}
}

func NewSnapshotter(store port.StateStore, storage port.SnapshotStorage, cfg *config.Config) *Snapshotter {
	return &Snapshotter{store: store, storage: storage, interval: cfg.Snapshot.Interval}
}

func (s *Snapshotter) Restore() error {
	snap, err := s.storage.Load()
	if err != nil || snap == nil {
		return err
	}
	last, err := s.store.LastIndex()
	if err != nil {
		return err
	}
	if last > 0 {
		log.Printf("snapshot: store already holds index %d, not restoring snapshot index %d", last, snap.Index)
		s.saved = last
		return nil
	}
	if err := s.store.Restore(*snap); err != nil {
		return err
	}
	s.saved = snap.Index
	log.Printf("snapshot: restored index %d (%d keys, %d services) taken %s",
		snap.Index, len(snap.KV), len(snap.Services), snap.TakenAt.Format(time.RFC3339))
	return nil
}

func (s *Snapshotter) Start() {
	if s.interval <= 0 {
		return
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.cancel = cancel
	s.done = make(chan struct{})
	go func() {
		defer close(s.done)
		t := time.NewTicker(s.interval)
		defer t.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-t.C:
				if err := s.Save(); err != nil {
					log.Printf("snapshot: %v", err)
				}
			}
		}
	}()
}

func (s *Snapshotter) Stop() error {
	if s.cancel != nil {
		s.cancel()
		<-s.done
	}
	return s.Save()
}

func (s *Snapshotter) Save() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	last, err := s.store.LastIndex()
	if err != nil || last == s.saved {
		return err
	}
	snap, err := s.store.Snapshot()
	if err != nil {
		return err
	}
	if err := s.storage.Save(snap); err != nil {
		return err
	}
	s.saved = snap.Index
	return nil
}

func (s *Snapshotter) Export() (domain.Snapshot, error) {
	return s.store.Snapshot()
}

func (s *Snapshotter) Import(snap domain.Snapshot) error {
	last, err := s.store.LastIndex()
	if err != nil {
		return err
	}
	snap.Index = max(snap.Index, last+1)
	if err := s.store.Restore(snap); err != nil {
		return err
	}
	return s.Save()
}
