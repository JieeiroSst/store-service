package application

import (
	"context"

	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
)

type authorService struct {
	repo port.AuthorRepository
}

func NewAuthorService(repo port.AuthorRepository) port.AuthorUsecase {
	return &authorService{repo: repo}
}

func (s *authorService) CreateAuthor(ctx context.Context, author *model.Author) (*model.Author, error) {
	if err := s.repo.Create(ctx, author); err != nil {
		return nil, err
	}
	return author, nil
}

func (s *authorService) GetAuthor(ctx context.Context, id int) (*model.Author, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *authorService) ListAuthors(ctx context.Context) ([]model.Author, error) {
	return s.repo.List(ctx)
}

func (s *authorService) UpdateAuthor(ctx context.Context, author *model.Author) (*model.Author, error) {
	if _, err := s.repo.GetByID(ctx, author.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, author); err != nil {
		return nil, err
	}
	return author, nil
}

func (s *authorService) DeleteAuthor(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *authorService) MostReadAuthor(ctx context.Context) (*model.AuthorStat, error) {
	return s.repo.MostRead(ctx)
}

func (s *authorService) MostPurchasedAuthor(ctx context.Context) (*model.AuthorStat, error) {
	return s.repo.MostPurchased(ctx)
}
