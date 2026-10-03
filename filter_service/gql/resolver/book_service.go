package resolver

import (
	"context"

	"github.com/JIeeiroSst/filter-service/gql/generated"
	"github.com/JIeeiroSst/filter-service/gql/model"
)

func (r *Resolver) BookSvcQuery() generated.BookSvcQueryResolver { return &bookSvcQueryResolver{r} }

type bookSvcQueryResolver struct{ *Resolver }

func (r *bookSvcQueryResolver) Books(ctx context.Context, obj *model.BookSvcQuery, cursor *string, limit *int, qArg *string, author *string, year *int, category *string) (*model.BookSvcBookPage, error) {
	return r.Clients.BookService.Books(ctx, cursor, limit, qArg, author, year, category)
}

func (r *bookSvcQueryResolver) Book(ctx context.Context, obj *model.BookSvcQuery, id int) (*model.BookSvcBookDetail, error) {
	return r.Clients.BookService.Book(ctx, id)
}

func (r *bookSvcQueryResolver) Chapters(ctx context.Context, obj *model.BookSvcQuery, id int, cursor *string, limit *int) (*model.BookSvcChapterPage, error) {
	return r.Clients.BookService.Chapters(ctx, id, cursor, limit)
}

func (r *bookSvcQueryResolver) Chapter(ctx context.Context, obj *model.BookSvcQuery, id int, number int) (*model.BookSvcChapter, error) {
	return r.Clients.BookService.Chapter(ctx, id, number)
}

func (r *bookSvcQueryResolver) Categories(ctx context.Context, obj *model.BookSvcQuery) ([]*model.BookSvcCategory, error) {
	return r.Clients.BookService.Categories(ctx)
}

func (r *bookSvcQueryResolver) Category(ctx context.Context, obj *model.BookSvcQuery, slug string) (*model.BookSvcCategory, error) {
	return r.Clients.BookService.Category(ctx, slug)
}

func (r *bookSvcQueryResolver) CategoryBooks(ctx context.Context, obj *model.BookSvcQuery, slug string, cursor *string, limit *int, qArg *string) (*model.BookSvcBookPage, error) {
	return r.Clients.BookService.CategoryBooks(ctx, slug, cursor, limit, qArg)
}
