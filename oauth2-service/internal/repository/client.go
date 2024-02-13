package repository

import (
	"context"
	"errors"
	"sync"

	"github.com/JIeeiroSst/oauth2-service/pkg/token"
)

type IClientStore interface {
	GetByID(ctx context.Context, id string) (token.ClientInfo, error)
}

type ClientStore struct {
	sync.RWMutex
	data map[string]token.ClientInfo
}

func NewClientStore() *ClientStore {
	return &ClientStore{
		data: make(map[string]token.ClientInfo),
	}
}

func (cs *ClientStore) GetByID(ctx context.Context, id string) (token.ClientInfo, error) {
	cs.RLock()
	defer cs.RUnlock()

	if c, ok := cs.data[id]; ok {
		return c, nil
	}
	return nil, errors.New("not found")
}

func (cs *ClientStore) Set(id string, cli token.ClientInfo) (err error) {
	cs.Lock()
	defer cs.Unlock()

	cs.data[id] = cli
	return
}
