package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) BookStoreCreateBook(ctx context.Context, input model.BookStoreBookInput) (*model.BookStoreBook, error) {
	return restCall[*model.BookStoreBook](ctx, r.Proxy, "book-store-service", rest.Request{Method: "POST", Path: "/api/v1/books", Body: input}, "")
}

func (r *mutationResolver) BookStoreUpdateBook(ctx context.Context, id int, input model.BookStoreBookInput) (*model.BookStoreBook, error) {
	return restCall[*model.BookStoreBook](ctx, r.Proxy, "book-store-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/books/%v", id), Body: input}, "")
}

func (r *mutationResolver) BookStoreDeleteBook(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-store-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/books/%v", id)}, "")
}

func (r *mutationResolver) BookStorePurchaseBook(ctx context.Context, id int, quantity int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-store-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/books/%v/purchase", id), Body: map[string]any{"quantity": quantity}}, "")
}

func (r *mutationResolver) BookStoreReadBook(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-store-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/books/%v/read", id), Body: map[string]any{}}, "")
}

func (r *mutationResolver) BookStoreCreateAuthor(ctx context.Context, input model.BookStoreAuthorInput) (*model.BookStoreAuthor, error) {
	return restCall[*model.BookStoreAuthor](ctx, r.Proxy, "book-store-service", rest.Request{Method: "POST", Path: "/api/v1/authors", Body: input}, "")
}

func (r *mutationResolver) BookStoreUpdateAuthor(ctx context.Context, id int, input model.BookStoreAuthorInput) (*model.BookStoreAuthor, error) {
	return restCall[*model.BookStoreAuthor](ctx, r.Proxy, "book-store-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/authors/%v", id), Body: input}, "")
}

func (r *mutationResolver) BookStoreDeleteAuthor(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-store-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/authors/%v", id)}, "")
}

func (r *mutationResolver) BookStoreCreatePublisher(ctx context.Context, input model.BookStorePublisherInput) (*model.BookStorePublisher, error) {
	return restCall[*model.BookStorePublisher](ctx, r.Proxy, "book-store-service", rest.Request{Method: "POST", Path: "/api/v1/publishers", Body: input}, "")
}

func (r *mutationResolver) BookStoreUpdatePublisher(ctx context.Context, id int, input model.BookStorePublisherInput) (*model.BookStorePublisher, error) {
	return restCall[*model.BookStorePublisher](ctx, r.Proxy, "book-store-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/publishers/%v", id), Body: input}, "")
}

func (r *mutationResolver) BookStoreDeletePublisher(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-store-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/publishers/%v", id)}, "")
}

func (r *mutationResolver) BookStoreCreateCategory(ctx context.Context, input model.BookStoreCategoryInput) (*model.BookStoreCategory, error) {
	return restCall[*model.BookStoreCategory](ctx, r.Proxy, "book-store-service", rest.Request{Method: "POST", Path: "/api/v1/categories", Body: input}, "")
}

func (r *mutationResolver) BookStoreUpdateCategory(ctx context.Context, id int, input model.BookStoreCategoryInput) (*model.BookStoreCategory, error) {
	return restCall[*model.BookStoreCategory](ctx, r.Proxy, "book-store-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/categories/%v", id), Body: input}, "")
}

func (r *mutationResolver) BookStoreDeleteCategory(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "book-store-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/categories/%v", id)}, "")
}

func (r *queryResolver) BookStoreBooks(ctx context.Context) ([]*model.BookStoreBook, error) {
	return restCall[[]*model.BookStoreBook](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/books"}, "")
}

func (r *queryResolver) BookStoreBook(ctx context.Context, id int) (*model.BookStoreBook, error) {
	return restCall[*model.BookStoreBook](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/books/%v", id)}, "")
}

func (r *queryResolver) BookStoreHighestPriceBook(ctx context.Context) (*model.BookStoreBook, error) {
	return restCall[*model.BookStoreBook](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/books/highest-price"}, "")
}

func (r *queryResolver) BookStoreMostPurchasedBook(ctx context.Context) (*model.BookStoreBook, error) {
	return restCall[*model.BookStoreBook](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/books/most-purchased"}, "")
}

func (r *queryResolver) BookStoreAuthors(ctx context.Context) ([]*model.BookStoreAuthor, error) {
	return restCall[[]*model.BookStoreAuthor](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/authors"}, "")
}

func (r *queryResolver) BookStoreAuthor(ctx context.Context, id int) (*model.BookStoreAuthor, error) {
	return restCall[*model.BookStoreAuthor](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/authors/%v", id)}, "")
}

func (r *queryResolver) BookStoreMostReadAuthor(ctx context.Context) (*model.BookStoreAuthorStat, error) {
	return restCall[*model.BookStoreAuthorStat](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/authors/most-read"}, "")
}

func (r *queryResolver) BookStoreMostPurchasedAuthor(ctx context.Context) (*model.BookStoreAuthorStat, error) {
	return restCall[*model.BookStoreAuthorStat](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/authors/most-purchased"}, "")
}

func (r *queryResolver) BookStorePublishers(ctx context.Context) ([]*model.BookStorePublisher, error) {
	return restCall[[]*model.BookStorePublisher](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/publishers"}, "")
}

func (r *queryResolver) BookStorePublisher(ctx context.Context, id int) (*model.BookStorePublisher, error) {
	return restCall[*model.BookStorePublisher](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/publishers/%v", id)}, "")
}

func (r *queryResolver) BookStoreCategories(ctx context.Context) ([]*model.BookStoreCategory, error) {
	return restCall[[]*model.BookStoreCategory](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/categories"}, "")
}

func (r *queryResolver) BookStoreCategory(ctx context.Context, id int) (*model.BookStoreCategory, error) {
	return restCall[*model.BookStoreCategory](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/categories/%v", id)}, "")
}

func (r *queryResolver) BookStoreCategoryPriceStats(ctx context.Context) ([]*model.BookStoreCategoryPriceStat, error) {
	return restCall[[]*model.BookStoreCategoryPriceStat](ctx, r.Proxy, "book-store-service", rest.Request{Method: "GET", Path: "/api/v1/categories/price-stats"}, "")
}
