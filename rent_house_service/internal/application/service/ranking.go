package service

import (
	"context"
	"sort"
	"sync"
	"time"

	"github.com/JIeerioSst/rent-house-service/internal/application/port/outbound"
	"github.com/JIeerioSst/rent-house-service/internal/domain"
)

const (
	rankCandidates = 500 // homestays re-ranked per recommended search
	tierLookupPool = 8   // parallel calls to recompense-service
	defaultTier    = 1   // bronze: no boost
)

type HostTiers struct {
	loyalty outbound.LoyaltyGateway
	ttl     time.Duration
	now     func() time.Time

	mu    sync.Mutex
	cache map[int64]tierEntry
}

type tierEntry struct {
	tier    int
	expires time.Time
}

func NewHostTiers(l outbound.LoyaltyGateway, ttl time.Duration) *HostTiers {
	if l == nil {
		return nil
	}
	return &HostTiers{loyalty: l, ttl: ttl, now: time.Now, cache: map[int64]tierEntry{}}
}

func (t *HostTiers) Tiers(ctx context.Context, hostIDs []int64) map[int64]int {
	out := make(map[int64]int, len(hostIDs))
	var missing []int64
	if t == nil {
		for _, id := range hostIDs {
			out[id] = defaultTier
		}
		return out
	}
	t.mu.Lock()
	for _, id := range hostIDs {
		if e, ok := t.cache[id]; ok && t.now().Before(e.expires) {
			out[id] = e.tier
		} else if _, dup := out[id]; !dup {
			out[id] = defaultTier
			missing = append(missing, id)
		}
	}
	t.mu.Unlock()

	var wg sync.WaitGroup
	sem := make(chan struct{}, tierLookupPool)
	for _, id := range missing {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			m, err := t.loyalty.Status(ctx, id)
			if err != nil {
				return 
			}
			t.mu.Lock()
			t.cache[id] = tierEntry{tier: m.Tier, expires: t.now().Add(t.ttl)}
			out[id] = m.Tier
			t.mu.Unlock()
		}()
	}
	wg.Wait()
	return out
}

type scored struct {
	h     domain.Homestay
	score float64
}

func rank(list []domain.Homestay, tiers map[int64]int) []scored {
	out := make([]scored, len(list))
	for i, h := range list {
		out[i] = scored{h, domain.RankScore(h.Rating, h.ReviewCount, tiers[h.HostID])}
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if a.score != b.score {
			return a.score > b.score
		}
		if a.h.ReviewCount != b.h.ReviewCount {
			return a.h.ReviewCount > b.h.ReviewCount
		}
		return a.h.ID > b.h.ID
	})
	return out
}

func after(list []scored, c domain.HomestayCursor) []scored {
	for i, s := range list {
		if s.score < c.Score ||
			(s.score == c.Score && s.h.ReviewCount < c.Count) ||
			(s.score == c.Score && s.h.ReviewCount == c.Count && s.h.ID < c.ID) {
			return list[i:]
		}
	}
	return nil
}
