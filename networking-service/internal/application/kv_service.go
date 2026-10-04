package application

import (
	"context"
	"strconv"
	"strings"
	"time"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

type ErrValueTooLarge struct{ Limit int }

func (e ErrValueTooLarge) Error() string {
	return "Value exceeds " + strconv.Itoa(e.Limit) + " byte limit"
}

type KVService struct {
	store   port.KVStore
	blocker *Blocker
	maxSize int
	now     func() time.Time
}

func NewKVService(store port.KVStore, blocker *Blocker, cfg *config.Config) *KVService {
	return &KVService{store: store, blocker: blocker, maxSize: cfg.KV.MaxValueBytes, now: time.Now}
}

func (s *KVService) MaxValueBytes() int { return s.maxSize }

func (s *KVService) Get(ctx context.Context, q QueryOptions, key string) (*domain.KVPair, QueryMeta, error) {
	var out *domain.KVPair
	m, err := s.blocker.Query(ctx, s.store, q, func() (uint64, error) {
		var (
			idx uint64
			err error
		)
		idx, out, err = s.store.KVGet(key)
		return idx, err
	})
	return out, m, err
}

func (s *KVService) List(ctx context.Context, q QueryOptions, prefix string) ([]domain.KVPair, QueryMeta, error) {
	var out []domain.KVPair
	m, err := s.blocker.Query(ctx, s.store, q, func() (uint64, error) {
		var (
			idx uint64
			err error
		)
		idx, out, err = s.store.KVList(prefix)
		return idx, err
	})
	return out, m, err
}

func (s *KVService) Keys(ctx context.Context, q QueryOptions, prefix, sep string) ([]string, QueryMeta, error) {
	var out []string
	m, err := s.blocker.Query(ctx, s.store, q, func() (uint64, error) {
		idx, pairs, err := s.store.KVList(prefix)
		out = make([]string, 0, len(pairs))
		for _, p := range pairs {
			k := p.Key
			if sep != "" {
				if i := strings.Index(k[len(prefix):], sep); i >= 0 {
					k = k[:len(prefix)+i+len(sep)]
				}
			}
			if n := len(out); n == 0 || out[n-1] != k {
				out = append(out, k)
			}
		}
		return idx, err
	})
	return out, m, err
}

func (s *KVService) Put(pair domain.KVPair, cas *uint64, acquire, release string) (bool, error) {
	if err := domain.ValidateKey(pair.Key, false); err != nil {
		return false, err
	}
	if len(pair.Value) > s.maxSize {
		return false, ErrValueTooLarge{Limit: s.maxSize}
	}
	req := domain.KVRequest{Op: domain.KVSet, Pair: pair}
	switch {
	case acquire != "" && release != "":
		return false, domain.Invalid("conflicting flags: acquire and release")
	case acquire != "":
		req.Op, req.Session = domain.KVLock, acquire
	case release != "":
		req.Op, req.Session = domain.KVUnlock, release
	case cas != nil:
		req.Op, req.CAS = domain.KVCAS, *cas
	}
	return s.store.KVApply(req, s.now())
}

func (s *KVService) Delete(key string, recurse bool, cas *uint64) (bool, error) {
	if err := domain.ValidateKey(key, recurse); err != nil {
		return false, err
	}
	req := domain.KVRequest{Op: domain.KVDelete, Pair: domain.KVPair{Key: key}}
	switch {
	case recurse && cas != nil:
		return false, domain.Invalid("conflicting flags: recurse and cas")
	case recurse:
		req.Op = domain.KVDelTree
	case cas != nil:
		req.Op, req.CAS = domain.KVDelCAS, *cas
	}
	return s.store.KVApply(req, s.now())
}
