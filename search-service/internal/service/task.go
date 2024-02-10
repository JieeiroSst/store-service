package service

import (
	"context"

	"github.com/JIeeiroSst/search-service/internal"
	"github.com/JIeeiroSst/search-service/internal/repository"
)

type TaskService struct {
	search repository.TaskRepository
}

type Tasks interface {
	Search(ctx context.Context, args internal.SearchParams) (_ internal.SearchResults, err error)
}

func NewTask(search repository.TaskRepository) *TaskService {
	return &TaskService{
		search: search,
	}
}

func (t *TaskService) Search(ctx context.Context, args internal.SearchParams) (_ internal.SearchResults, err error) {
	res, err := t.search.Search(ctx, args)
	if err != nil {
		return internal.SearchResults{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "search")
	}

	return res, nil
}
