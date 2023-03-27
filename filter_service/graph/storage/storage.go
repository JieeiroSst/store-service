package storage

import (
	"context"

	"github.com/JIeeiroSst/filter-service/graph/model"
)

type Storage interface {
	PutUser(ctx context.Context, usr *model.User) error
	PutTodo(ctx context.Context, todo *model.Todo) error
	GetUsers(ctx context.Context, ids []string) ([]*model.User, error)
	GetTodos(ctx context.Context, ids []string) ([]*model.Todo, error)
	GetAllTodos(ctx context.Context) ([]*model.Todo, error)
}

type MemoryStorage struct {
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{}
}

func (m *MemoryStorage) PutUser(ctx context.Context, usr *model.User) error {
	return nil
}

func (m *MemoryStorage) PutTodo(ctx context.Context, todo *model.Todo) error {
	return nil
}

func (m *MemoryStorage) GetUsers(ctx context.Context, ids []string) ([]*model.User, error) {
	output := make([]*model.User, 0, len(ids))

	return output, nil
}

func (m *MemoryStorage) GetTodos(ctx context.Context, ids []string) ([]*model.Todo, error) {
	output := make([]*model.Todo, 0, len(ids))

	return output, nil
}

func (m *MemoryStorage) GetAllTodos(ctx context.Context) ([]*model.Todo, error) {
	output := make([]*model.Todo, 0)

	return output, nil
}
