package snapshot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/JIeeiroSst/networking-service/config"
	"github.com/JIeeiroSst/networking-service/internal/domain"
	"github.com/JIeeiroSst/networking-service/internal/port"
)

const (
	fileName      = "state.json"
	formatVersion = 1
)

type envelope struct {
	Version  int
	Snapshot domain.Snapshot
}

type File struct {
	dir string
}

func New(cfg *config.Config) (port.SnapshotStorage, error) {
	if cfg.Snapshot.DataDir == "" {
		return Noop{}, nil
	}
	if err := os.MkdirAll(cfg.Snapshot.DataDir, 0o750); err != nil {
		return nil, fmt.Errorf("snapshot dir: %w", err)
	}
	return &File{dir: cfg.Snapshot.DataDir}, nil
}

func (f *File) Load() (*domain.Snapshot, error) {
	b, err := os.ReadFile(filepath.Join(f.dir, fileName))
	if errors.Is(err, fs.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	var env envelope
	if err := json.Unmarshal(b, &env); err != nil {
		return nil, fmt.Errorf("decode snapshot: %w", err)
	}
	if env.Version != formatVersion {
		return nil, fmt.Errorf("snapshot format version %d not supported", env.Version)
	}
	return &env.Snapshot, nil
}

func (f *File) Save(s domain.Snapshot) error {
	tmp, err := os.CreateTemp(f.dir, fileName+".*.tmp")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	if err := json.NewEncoder(tmp).Encode(envelope{Version: formatVersion, Snapshot: s}); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Sync(); err != nil {
		tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), filepath.Join(f.dir, fileName))
}

type Noop struct{}

func (Noop) Load() (*domain.Snapshot, error) { return nil, nil }
func (Noop) Save(domain.Snapshot) error      { return nil }
