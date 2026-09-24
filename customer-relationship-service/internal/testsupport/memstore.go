package testsupport

import (
	"context"
	"sync"

	"github.com/JIeeiroSst/customer-relationship-service/common"
)

type MemStore struct {
	mu      sync.Mutex
	Objects map[string][]byte
	Types   map[string]string
	Off     bool
	PutErr  error
	Deleted []string
}

func NewMemStore() *MemStore {
	return &MemStore{Objects: map[string][]byte{}, Types: map[string]string{}}
}

func (m *MemStore) Enabled() bool { return !m.Off }

func (m *MemStore) Put(_ context.Context, key string, data []byte, contentType string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if m.PutErr != nil {
		return m.PutErr
	}
	m.Objects[key] = append([]byte(nil), data...)
	m.Types[key] = contentType
	return nil
}

func (m *MemStore) Get(_ context.Context, key string) ([]byte, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	d, ok := m.Objects[key]
	if !ok {
		return nil, common.ErrNotFound
	}
	return append([]byte(nil), d...), nil
}

func (m *MemStore) Delete(_ context.Context, key string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.Objects, key)
	m.Deleted = append(m.Deleted, key)
	return nil
}

func (m *MemStore) Tamper(key string, data []byte) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.Objects[key] = data
}
