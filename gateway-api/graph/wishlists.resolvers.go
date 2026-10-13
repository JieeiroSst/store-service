package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) WishlistCreate(ctx context.Context, input model.WishlistCreateInput) (*model.Wishlist, error) {
	return restCall[*model.Wishlist](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "POST", Path: "/api/v1/wishlists", Body: input}, "data")
}

func (r *mutationResolver) WishlistUpdate(ctx context.Context, id int, input model.WishlistUpdateInput) (*model.Wishlist, error) {
	return restCall[*model.Wishlist](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/wishlists/%v", id), Body: input}, "data")
}

func (r *mutationResolver) WishlistDelete(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/wishlists/%v", id)}, "")
}

func (r *mutationResolver) WishlistAddLine(ctx context.Context, id int, input model.WishlistLineInput) (*model.WishlistLine, error) {
	return restCall[*model.WishlistLine](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wishlists/%v/lines", id), Body: input}, "data")
}

func (r *mutationResolver) WishlistUpdateLine(ctx context.Context, id int, lineID int, quantity int) (*model.WishlistLine, error) {
	return restCall[*model.WishlistLine](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "PATCH", Path: rest.P("/api/v1/wishlists/%v/lines/%v", id, lineID), Body: map[string]any{"quantity": quantity}}, "data")
}

func (r *mutationResolver) WishlistRemoveLine(ctx context.Context, id int, lineID int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/wishlists/%v/lines/%v", id, lineID)}, "")
}

func (r *mutationResolver) WishlistClearLines(ctx context.Context, id int) (*model.WishlistClearResult, error) {
	return restCall[*model.WishlistClearResult](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/wishlists/%v/lines", id)}, "data")
}

func (r *mutationResolver) WishlistMoveLine(ctx context.Context, id int, lineID int, targetWishlistID int) (*model.WishlistLine, error) {
	return restCall[*model.WishlistLine](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/wishlists/%v/lines/%v/move", id, lineID), Body: map[string]any{"target_wishlist_id": targetWishlistID}}, "data")
}

func (r *mutationResolver) WishlistCopyShared(ctx context.Context, key string) (*model.Wishlist, error) {
	return restCall[*model.Wishlist](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/shared-wishlists/%v/copy", key), Body: map[string]any{}}, "data")
}

func (r *mutationResolver) WishlistQuickAdd(ctx context.Context, input model.WishlistLineInput) (*model.WishlistLine, error) {
	return restCall[*model.WishlistLine](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "POST", Path: "/api/v1/me/wishlist-items", Body: input}, "data")
}

func (r *queryResolver) Wishlists(ctx context.Context, ownerID *string, limit *int, offset *int) (*model.WishlistPage, error) {
	return restCall[*model.WishlistPage](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: "/api/v1/wishlists", Query: restQuery("owner_id", ownerID, "limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) Wishlist(ctx context.Context, id int) (*model.Wishlist, error) {
	return restCall[*model.Wishlist](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/wishlists/%v", id)}, "data")
}

func (r *queryResolver) WishlistDefault(ctx context.Context) (*model.Wishlist, error) {
	return restCall[*model.Wishlist](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: "/api/v1/me/wishlists/default"}, "data")
}

func (r *queryResolver) WishlistsOfUser(ctx context.Context, userID string, limit *int, offset *int) (*model.WishlistPage, error) {
	return restCall[*model.WishlistPage](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v/wishlists", userID), Query: restQuery("limit", limit, "offset", offset)}, "data")
}

func (r *queryResolver) WishlistShared(ctx context.Context, key string) (*model.Wishlist, error) {
	return restCall[*model.Wishlist](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/shared-wishlists/%v", key)}, "data")
}

func (r *queryResolver) WishlistProductLookup(ctx context.Context, productID int) (*model.WishlistProductLookup, error) {
	return restCall[*model.WishlistProductLookup](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/me/products/%v/wishlists", productID)}, "data")
}

func (r *queryResolver) WishlistPopularProducts(ctx context.Context, limit *int) ([]*model.WishlistProductStat, error) {
	return restCall[[]*model.WishlistProductStat](ctx, r.Proxy, "wishlists-service", rest.Request{Method: "GET", Path: "/api/v1/wishlist-stats/popular-products", Query: restQuery("limit", limit)}, "data")
}
