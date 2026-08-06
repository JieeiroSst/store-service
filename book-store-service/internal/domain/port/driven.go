package port

import (
	"context"

	"github.com/JIeeiroSst/bookStore-service/internal/domain/model"
)

// BookRepository persists books owned by this service.
type BookRepository interface {
	Create(ctx context.Context, book *model.Book) error
	GetByID(ctx context.Context, id int) (*model.Book, error)
	Update(ctx context.Context, book *model.Book) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Book, error)

	// GetHighestPrice returns the book with the highest price.
	GetHighestPrice(ctx context.Context) (*model.Book, error)
	// GetMostPurchased returns the book with the highest purchase_count.
	GetMostPurchased(ctx context.Context) (*model.Book, error)
	// IncrementPurchaseCount records a purchase of qty units, bumping
	// purchase_count and decrementing quantity.
	IncrementPurchaseCount(ctx context.Context, id int, qty int) error
	// IncrementReadCount records a read of the book.
	IncrementReadCount(ctx context.Context, id int) error
	// PriceStatsByCategory aggregates min/max/avg price and book count per category.
	PriceStatsByCategory(ctx context.Context) ([]model.CategoryPriceStat, error)
}

// AuthorRepository persists authors owned by this service.
type AuthorRepository interface {
	Create(ctx context.Context, author *model.Author) error
	GetByID(ctx context.Context, id int) (*model.Author, error)
	Update(ctx context.Context, author *model.Author) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Author, error)

	// MostRead returns the author whose books have the highest total read_count.
	MostRead(ctx context.Context) (*model.AuthorStat, error)
	// MostPurchased returns the author whose books have the highest total purchase_count.
	MostPurchased(ctx context.Context) (*model.AuthorStat, error)
}

// PublisherRepository persists publishers owned by this service.
type PublisherRepository interface {
	Create(ctx context.Context, publisher *model.Publisher) error
	GetByID(ctx context.Context, id int) (*model.Publisher, error)
	Update(ctx context.Context, publisher *model.Publisher) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Publisher, error)
}

// CategoryRepository persists categories owned by this service.
type CategoryRepository interface {
	Create(ctx context.Context, category *model.Category) error
	GetByID(ctx context.Context, id int) (*model.Category, error)
	Update(ctx context.Context, category *model.Category) error
	Delete(ctx context.Context, id int) error
	List(ctx context.Context) ([]model.Category, error)
}
