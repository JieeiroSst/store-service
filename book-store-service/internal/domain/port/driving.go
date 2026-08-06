package port

import (
	"context"

	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
)

type BookUsecase interface {
	CreateBook(ctx context.Context, book *model.Book) (*model.Book, error)
	GetBook(ctx context.Context, id int) (*model.Book, error)
	ListBooks(ctx context.Context) ([]model.Book, error)
	UpdateBook(ctx context.Context, book *model.Book) (*model.Book, error)
	DeleteBook(ctx context.Context, id int) error

	HighestPriceBook(ctx context.Context) (*model.Book, error)
	MostPurchasedBook(ctx context.Context) (*model.Book, error)
	PurchaseBook(ctx context.Context, id int, qty int) error
	ReadBook(ctx context.Context, id int) error
	CategoryPriceStats(ctx context.Context) ([]model.CategoryPriceStat, error)
}

type AuthorUsecase interface {
	CreateAuthor(ctx context.Context, author *model.Author) (*model.Author, error)
	GetAuthor(ctx context.Context, id int) (*model.Author, error)
	ListAuthors(ctx context.Context) ([]model.Author, error)
	UpdateAuthor(ctx context.Context, author *model.Author) (*model.Author, error)
	DeleteAuthor(ctx context.Context, id int) error

	MostReadAuthor(ctx context.Context) (*model.AuthorStat, error)
	MostPurchasedAuthor(ctx context.Context) (*model.AuthorStat, error)
}

type PublisherUsecase interface {
	CreatePublisher(ctx context.Context, publisher *model.Publisher) (*model.Publisher, error)
	GetPublisher(ctx context.Context, id int) (*model.Publisher, error)
	ListPublishers(ctx context.Context) ([]model.Publisher, error)
	UpdatePublisher(ctx context.Context, publisher *model.Publisher) (*model.Publisher, error)
	DeletePublisher(ctx context.Context, id int) error
}

type CategoryUsecase interface {
	CreateCategory(ctx context.Context, category *model.Category) (*model.Category, error)
	GetCategory(ctx context.Context, id int) (*model.Category, error)
	ListCategories(ctx context.Context) ([]model.Category, error)
	UpdateCategory(ctx context.Context, category *model.Category) (*model.Category, error)
	DeleteCategory(ctx context.Context, id int) error
}
