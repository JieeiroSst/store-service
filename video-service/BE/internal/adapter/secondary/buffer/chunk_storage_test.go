package buffer

import (
	"bytes"
	"context"
	"io"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JIeeiroSst/video-service/config"
)

type countingStorage struct {
	data  []byte
	calls atomic.Int64
	delay time.Duration
}

func (s *countingStorage) Put(context.Context, string, io.Reader, int64, string) (int64, error) {
	return 0, nil
}
func (s *countingStorage) Delete(context.Context, string) error       { return nil }
func (s *countingStorage) DeletePrefix(context.Context, string) error { return nil }
func (s *countingStorage) Open(context.Context, string) (io.ReadCloser, error) {
	return nil, nil
}
func (s *countingStorage) Get(context.Context, string) ([]byte, error) {
	s.calls.Add(1)
	time.Sleep(s.delay)
	return s.data, nil
}
func (s *countingStorage) ReadAt(_ context.Context, _ string, size int64, p []byte, off int64) (int, error) {
	s.calls.Add(1)
	time.Sleep(s.delay)
	return copy(p, s.data[off:size]), nil
}

func newTestStorage(next *countingStorage, chunk int64, prefetch int) *ChunkStorage {
	cfg := &config.Config{}
	cfg.Stream.ChunkSize = chunk
	cfg.Stream.CacheBytes = 1 << 20
	cfg.Stream.PrefetchChunks = prefetch
	return NewChunkStorage(next, cfg)
}

func testData(n int) []byte {
	b := make([]byte, n)
	for i := range b {
		b[i] = byte(i * 7)
	}
	return b
}

func TestReadAtAcrossChunkBoundaries(t *testing.T) {
	data := testData(1000)
	cs := newTestStorage(&countingStorage{data: data}, 64, 0)

	for _, tc := range []struct{ off, n int }{{0, 10}, {60, 10}, {63, 130}, {990, 10}, {0, 1000}} {
		p := make([]byte, tc.n)
		n, err := cs.ReadAt(context.Background(), "k", 1000, p, int64(tc.off))
		if err != nil || n != tc.n || !bytes.Equal(p, data[tc.off:tc.off+tc.n]) {
			t.Fatalf("off=%d n=%d: got n=%d err=%v", tc.off, tc.n, n, err)
		}
	}

	if _, err := cs.ReadAt(context.Background(), "k", 1000, make([]byte, 4), 1000); err != io.EOF {
		t.Fatalf("read at EOF: got %v", err)
	}
}

func TestConcurrentReadersShareOneUpstreamRead(t *testing.T) {
	next := &countingStorage{data: testData(1000), delay: 50 * time.Millisecond}
	cs := newTestStorage(next, 1000, 0)

	var wg sync.WaitGroup
	for i := 0; i < 200; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			p := make([]byte, 100)
			if _, err := cs.ReadAt(context.Background(), "k", 1000, p, 0); err != nil {
				t.Error(err)
			}
		}()
	}
	wg.Wait()

	if got := next.calls.Load(); got != 1 {
		t.Fatalf("upstream calls = %d, want 1", got)
	}
}

func TestPrefetchWarmsNextChunk(t *testing.T) {
	next := &countingStorage{data: testData(1000)}
	cs := newTestStorage(next, 100, 2)

	if _, err := cs.ReadAt(context.Background(), "k", 1000, make([]byte, 10), 0); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(2 * time.Second)
	for !(cs.cache.contains(chunkKey{"k", 1}) && cs.cache.contains(chunkKey{"k", 2})) {
		if time.Now().After(deadline) {
			t.Fatal("chunks 1 and 2 were not prefetched")
		}
		time.Sleep(5 * time.Millisecond)
	}
}

func TestLRUEvictsWithinByteBudget(t *testing.T) {
	c := newLRU(shardCount * 100) // 100 bytes per shard
	for i := int64(0); i < 100; i++ {
		c.add(chunkKey{"k", i}, make([]byte, 60))
	}
	var total int64
	for _, s := range c.shards {
		if s.bytes > s.maxBytes+60 {
			t.Fatalf("shard holds %d bytes, budget %d", s.bytes, s.maxBytes)
		}
		total += s.bytes
	}
	if total >= 100*60 {
		t.Fatal("nothing was evicted")
	}
}

func TestGetSharesOneUpstreamRead(t *testing.T) {
	next := &countingStorage{data: testData(100), delay: 30 * time.Millisecond}
	cs := newTestStorage(next, 64, 0)

	var wg sync.WaitGroup
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if b, err := cs.Get(context.Background(), "seg"); err != nil || len(b) != 100 {
				t.Error(len(b), err)
			}
		}()
	}
	wg.Wait()
	if _, err := cs.Get(context.Background(), "seg"); err != nil {
		t.Fatal(err)
	}
	if got := next.calls.Load(); got != 1 {
		t.Fatalf("upstream calls = %d, want 1", got)
	}
}
