package service

import (
	"context"

	"github.com/JIeeiroSst/search-service/internal"
)

type TaskSearchRepository interface {
	Search(ctx context.Context, args internal.SearchParams) (internal.SearchResults, error)
}

type Task struct {
	search TaskSearchRepository
}

func NewTask(search TaskSearchRepository) *Task {
	return &Task{
		search: search,
	}
}

func (t *Task) By(ctx context.Context, args internal.SearchParams) (_ internal.SearchResults, err error) {
	res, err := t.search.Search(ctx, args)
	if err != nil {
		return internal.SearchResults{}, internal.WrapErrorf(err, internal.ErrorCodeUnknown, "search")
	}

	return res, nil
}
