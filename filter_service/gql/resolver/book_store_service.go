package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) BookStoreQuery() generated.BookStoreQueryResolver {
	return &bookStoreQueryResolver{r}
}

type bookStoreQueryResolver struct{ *Resolver }

func (r *bookStoreQueryResolver) Books(ctx context.Context, obj *model.BookStoreQuery) ([]*model.BookStoreBook, error) {
	return r.Clients.BookStoreService.Books(ctx)
}

func (r *bookStoreQueryResolver) HighestPriceBook(ctx context.Context, obj *model.BookStoreQuery) (*model.BookStoreBook, error) {
	return r.Clients.BookStoreService.HighestPriceBook(ctx)
}

func (r *bookStoreQueryResolver) MostPurchasedBook(ctx context.Context, obj *model.BookStoreQuery) (*model.BookStoreBook, error) {
	return r.Clients.BookStoreService.MostPurchasedBook(ctx)
}

func (r *bookStoreQueryResolver) Book(ctx context.Context, obj *model.BookStoreQuery, id int) (*model.BookStoreBook, error) {
	return r.Clients.BookStoreService.Book(ctx, id)
}

func (r *bookStoreQueryResolver) Authors(ctx context.Context, obj *model.BookStoreQuery) ([]*model.BookStoreAuthor, error) {
	return r.Clients.BookStoreService.Authors(ctx)
}

func (r *bookStoreQueryResolver) MostReadAuthor(ctx context.Context, obj *model.BookStoreQuery) (*model.BookStoreAuthorStat, error) {
	return r.Clients.BookStoreService.MostReadAuthor(ctx)
}

func (r *bookStoreQueryResolver) MostPurchasedAuthor(ctx context.Context, obj *model.BookStoreQuery) (*model.BookStoreAuthorStat, error) {
	return r.Clients.BookStoreService.MostPurchasedAuthor(ctx)
}

func (r *bookStoreQueryResolver) Author(ctx context.Context, obj *model.BookStoreQuery, id int) (*model.BookStoreAuthor, error) {
	return r.Clients.BookStoreService.Author(ctx, id)
}

func (r *bookStoreQueryResolver) Publishers(ctx context.Context, obj *model.BookStoreQuery) ([]*model.BookStorePublisher, error) {
	return r.Clients.BookStoreService.Publishers(ctx)
}

func (r *bookStoreQueryResolver) Publisher(ctx context.Context, obj *model.BookStoreQuery, id int) (*model.BookStorePublisher, error) {
	return r.Clients.BookStoreService.Publisher(ctx, id)
}

func (r *bookStoreQueryResolver) Categories(ctx context.Context, obj *model.BookStoreQuery) ([]*model.BookStoreCategory, error) {
	return r.Clients.BookStoreService.Categories(ctx)
}

func (r *bookStoreQueryResolver) CategoryPriceStats(ctx context.Context, obj *model.BookStoreQuery) ([]*model.BookStoreCategoryPriceStat, error) {
	return r.Clients.BookStoreService.CategoryPriceStats(ctx)
}

func (r *bookStoreQueryResolver) Category(ctx context.Context, obj *model.BookStoreQuery, id int) (*model.BookStoreCategory, error) {
	return r.Clients.BookStoreService.Category(ctx, id)
}
