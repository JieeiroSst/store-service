package application

import (
	"context"

	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

type hooks[T any] struct {
	prepare     func(*T)
	keepManaged func(incoming, stored *T)
	created     func(ctx context.Context, entity *T)
	updated     func(ctx context.Context, entity *T)
	deleted     func(ctx context.Context, id uint)
}

type crudService[T any] struct {
	repo port.Repository[T]
	h    hooks[T]
}

func NewCRUDService[T any](repo port.Repository[T]) port.Usecase[T] {
	return &crudService[T]{repo: repo}
}

func newCRUD[T any](repo port.Repository[T], h hooks[T]) port.Usecase[T] {
	return &crudService[T]{repo: repo, h: h}
}

func (s *crudService[T]) Create(ctx context.Context, entity *T) (*T, error) {
	if s.h.prepare != nil {
		s.h.prepare(entity)
	}
	if err := s.repo.Create(ctx, entity); err != nil {
		return nil, err
	}
	if s.h.created != nil {
		s.h.created(ctx, entity)
	}
	return entity, nil
}

func (s *crudService[T]) Get(ctx context.Context, id uint) (*T, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *crudService[T]) List(ctx context.Context, q port.ListQuery) ([]T, int64, error) {
	return s.repo.List(ctx, q)
}

func (s *crudService[T]) Update(ctx context.Context, id uint, entity *T) (*T, error) {
	stored, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	if s.h.keepManaged != nil {
		s.h.keepManaged(entity, stored)
	}
	if err := s.repo.Update(ctx, id, entity); err != nil {
		return nil, err
	}
	updated, err := s.repo.GetByID(ctx, id)
	if err == nil && s.h.updated != nil {
		s.h.updated(ctx, updated)
	}
	return updated, err
}

func (s *crudService[T]) Delete(ctx context.Context, id uint) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if s.h.deleted != nil {
		s.h.deleted(ctx, id)
	}
	return nil
}
