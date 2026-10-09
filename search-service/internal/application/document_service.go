package application

import (
	"context"

	"github.com/JIeeiroSst/search-service/internal/domain"
	"github.com/JIeeiroSst/search-service/internal/port"
)

type DocumentService struct {
	searcher port.DocumentSearcher
	policy   domain.SensitivePolicy
}

func NewDocumentService(searcher port.DocumentSearcher, policy domain.SensitivePolicy) *DocumentService {
	return &DocumentService{searcher: searcher, policy: policy}
}

func (s *DocumentService) Search(ctx context.Context, q domain.DocumentQuery) (domain.DocumentPage, error) {
	return s.query(ctx, q, domain.ModeSearch)
}

func (s *DocumentService) List(ctx context.Context, q domain.DocumentQuery) (domain.DocumentPage, error) {
	return s.query(ctx, q, domain.ModeList)
}

func (s *DocumentService) query(ctx context.Context, q domain.DocumentQuery, mode domain.QueryMode) (domain.DocumentPage, error) {
	if err := q.Normalize(mode); err != nil {
		return domain.DocumentPage{}, err
	}
	if err := s.policy.CheckFields(q.Fields()); err != nil {
		return domain.DocumentPage{}, err
	}
	page, err := s.searcher.Query(ctx, q, mode)
	if err != nil {
		return domain.DocumentPage{}, err
	}
	page.Items = s.policy.RedactAll(page.Items)
	return page, nil
}

func (s *DocumentService) Autocomplete(ctx context.Context, q domain.AutocompleteQuery) ([]domain.Document, error) {
	if err := q.Normalize(); err != nil {
		return nil, err
	}
	docs, err := s.searcher.Autocomplete(ctx, q)
	if err != nil {
		return nil, err
	}
	return s.policy.RedactAll(docs), nil
}

func (s *DocumentService) Get(ctx context.Context, ref domain.DocumentRef) (domain.Document, error) {
	if err := ref.Validate(); err != nil {
		return domain.Document{}, err
	}
	doc, err := s.searcher.Get(ctx, ref)
	if err != nil {
		return domain.Document{}, err
	}
	return s.policy.Redact(doc), nil
}

func (s *DocumentService) Similar(ctx context.Context, ref domain.DocumentRef, size int) ([]domain.Document, error) {
	if err := ref.Validate(); err != nil {
		return nil, err
	}
	size, err := domain.NormalizeSimilarSize(size)
	if err != nil {
		return nil, err
	}
	docs, err := s.searcher.Similar(ctx, ref, size)
	if err != nil {
		return nil, err
	}
	return s.policy.RedactAll(docs), nil
}

func (s *DocumentService) Indices(ctx context.Context) ([]domain.IndexInfo, error) {
	return s.searcher.Indices(ctx)
}

func (s *DocumentService) Fields(ctx context.Context, index string) ([]domain.FieldInfo, error) {
	if err := domain.ValidateIndex(index, true); err != nil {
		return nil, err
	}
	fields, err := s.searcher.Fields(ctx, index)
	if err != nil {
		return nil, err
	}
	out := fields[:0]
	for _, f := range fields {
		if !s.policy.IsSensitive(f.Name) {
			out = append(out, f)
		}
	}
	return out, nil
}
