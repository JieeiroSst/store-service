package snapshot

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/domain"
)

func TestFileRoundTrip(t *testing.T) {
	dir := t.TempDir()
	cfg := &config.Config{Snapshot: config.SnapshotConfig{DataDir: dir}}
	st, err := New(cfg)
	if err != nil {
		t.Fatal(err)
	}
	if snap, err := st.Load(); err != nil || snap != nil {
		t.Fatalf("empty dir: %v, %v", snap, err)
	}

	in := domain.Snapshot{Index: 7, KV: []domain.KVPair{{Key: "k", Value: []byte{0, 1, 2}}}}
	if err := st.Save(in); err != nil {
		t.Fatal(err)
	}
	out, err := st.Load()
	if err != nil || out.Index != 7 || string(out.KV[0].Value) != "\x00\x01\x02" {
		t.Fatalf("loaded %+v, %v", out, err)
	}

	entries, _ := os.ReadDir(dir)
	if len(entries) != 1 {
		t.Fatalf("dir has %d entries", len(entries))
	}
}

func TestCorruptSnapshotFailsLoudly(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, fileName), []byte("{nope"), 0o600)
	st, _ := New(&config.Config{Snapshot: config.SnapshotConfig{DataDir: dir}})
	if _, err := st.Load(); err == nil {
		t.Fatal("corrupt snapshot loaded without error")
	}
}

func TestNoDataDirIsNoop(t *testing.T) {
	st, err := New(&config.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := st.(Noop); !ok {
		t.Fatalf("storage = %T, want Noop", st)
	}
}
