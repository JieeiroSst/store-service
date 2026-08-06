package application

import (
	"context"

	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
	"github.com/JIeeiroSst/bookStore-service/internal/domain/port"
)

type bookService struct {
	repo port.BookRepository
}

func NewBookService(repo port.BookRepository) port.BookUsecase {
	return &bookService{repo: repo}
}

func (s *bookService) CreateBook(ctx context.Context, book *model.Book) (*model.Book, error) {
	if err := s.repo.Create(ctx, book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *bookService) GetBook(ctx context.Context, id int) (*model.Book, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *bookService) ListBooks(ctx context.Context) ([]model.Book, error) {
	return s.repo.List(ctx)
}

func (s *bookService) UpdateBook(ctx context.Context, book *model.Book) (*model.Book, error) {
	if _, err := s.repo.GetByID(ctx, book.ID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, book); err != nil {
		return nil, err
	}
	return book, nil
}

func (s *bookService) DeleteBook(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}

func (s *bookService) HighestPriceBook(ctx context.Context) (*model.Book, error) {
	return s.repo.GetHighestPrice(ctx)
}

func (s *bookService) MostPurchasedBook(ctx context.Context) (*model.Book, error) {
	return s.repo.GetMostPurchased(ctx)
}

func (s *bookService) PurchaseBook(ctx context.Context, id int, qty int) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.repo.IncrementPurchaseCount(ctx, id, qty)
}

func (s *bookService) ReadBook(ctx context.Context, id int) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}
	return s.repo.IncrementReadCount(ctx, id)
}

func (s *bookService) CategoryPriceStats(ctx context.Context) ([]model.CategoryPriceStat, error) {
	return s.repo.PriceStatsByCategory(ctx)
}
