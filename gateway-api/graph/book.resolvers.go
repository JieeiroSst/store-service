package graph

import (
	"context"

	"github.com/99designs/gqlgen/graphql"
	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) BookCreateBook(ctx context.Context, input model.BookBookInput) (*model.BookBook, error) {
	return restCall[*model.BookBook](ctx, r.Proxy, "book-service", rest.Request{Method: "POST", Path: "/api/v1/books", Body: input}, "")
}

func (r *mutationResolver) BookUpdateBook(ctx context.Context, id int, input model.BookBookInput) (*model.BookBook, error) {
	return restCall[*model.BookBook](ctx, r.Proxy, "book-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/books/%v", id), Body: input}, "")
}

func (r *mutationResolver) BookDeleteBook(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/books/%v", id)}, "")
}

func (r *mutationResolver) BookPutCover(ctx context.Context, id int, file graphql.Upload) (*model.BookBook, error) {
	return restRaw[*model.BookBook](ctx, r.Proxy, "book-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/books/%v/cover", id)}, file, "")
}

func (r *mutationResolver) BookDeleteCover(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/books/%v/cover", id)}, "")
}

func (r *mutationResolver) BookPutChapter(ctx context.Context, id int, number int, input model.BookChapterInput) (*model.BookChapter, error) {
	return restCall[*model.BookChapter](ctx, r.Proxy, "book-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/books/%v/chapters/%v", id, number), Body: input}, "")
}

func (r *mutationResolver) BookDeleteChapter(ctx context.Context, id int, number int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/books/%v/chapters/%v", id, number)}, "")
}

func (r *mutationResolver) BookCreateCategory(ctx context.Context, input model.BookCategoryInput) (*model.BookCategory, error) {
	return restCall[*model.BookCategory](ctx, r.Proxy, "book-service", rest.Request{Method: "POST", Path: "/api/v1/categories", Body: input}, "")
}

func (r *mutationResolver) BookRenameCategory(ctx context.Context, slug string, name string) (*model.BookCategory, error) {
	return restCall[*model.BookCategory](ctx, r.Proxy, "book-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/categories/%v", slug), Body: map[string]any{"name": name}}, "")
}

func (r *mutationResolver) BookDeleteCategory(ctx context.Context, slug string) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/categories/%v", slug)}, "")
}

func (r *queryResolver) BookBooks(ctx context.Context, cursor *string, limit *int, q *string, author *string, year *int, category *string) (*model.BookBookPage, error) {
	return restCall[*model.BookBookPage](ctx, r.Proxy, "book-service", rest.Request{Method: "GET", Path: "/api/v1/books", Query: restQuery("cursor", cursor, "limit", limit, "q", q, "author", author, "year", year, "category", category)}, "")
}

func (r *queryResolver) BookBook(ctx context.Context, id int) (*model.BookBook, error) {
	return restCall[*model.BookBook](ctx, r.Proxy, "book-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/books/%v", id)}, "")
}

func (r *queryResolver) BookChapters(ctx context.Context, id int, cursor *string, limit *int) (*model.BookChapterPage, error) {
	return restCall[*model.BookChapterPage](ctx, r.Proxy, "book-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/books/%v/chapters", id), Query: restQuery("cursor", cursor, "limit", limit)}, "")
}

func (r *queryResolver) BookChapter(ctx context.Context, id int, number int) (*model.BookChapter, error) {
	return restCall[*model.BookChapter](ctx, r.Proxy, "book-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/books/%v/chapters/%v", id, number)}, "")
}

func (r *queryResolver) BookCategories(ctx context.Context) ([]*model.BookCategory, error) {
	return restCall[[]*model.BookCategory](ctx, r.Proxy, "book-service", rest.Request{Method: "GET", Path: "/api/v1/categories"}, "data")
}

func (r *queryResolver) BookCategory(ctx context.Context, slug string) (*model.BookCategory, error) {
	return restCall[*model.BookCategory](ctx, r.Proxy, "book-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/categories/%v", slug)}, "")
}

func (r *queryResolver) BookCategoryBooks(ctx context.Context, slug string, cursor *string, limit *int, q *string) (*model.BookBookPage, error) {
	return restCall[*model.BookBookPage](ctx, r.Proxy, "book-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/categories/%v/books", slug), Query: restQuery("cursor", cursor, "limit", limit, "q", q)}, "")
}
