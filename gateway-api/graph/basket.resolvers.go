package graph

import (
	"context"

	"github.com/JIeeiroSst/gateway-api/graph/model"
	"github.com/JIeeiroSst/gateway-api/proxy/rest"
)

func (r *mutationResolver) BasketCreateBasket(ctx context.Context, input model.BasketBasketInput) (*model.BasketBasket, error) {
	return restCall[*model.BasketBasket](ctx, r.Proxy, "basket-service", rest.Request{Method: "POST", Path: "/api/v1/baskets", Body: input}, "")
}

func (r *mutationResolver) BasketUpdateBasket(ctx context.Context, id int, input model.BasketBasketInput) (*model.BasketBasket, error) {
	return restCall[*model.BasketBasket](ctx, r.Proxy, "basket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/baskets/%v", id), Body: input}, "")
}

func (r *mutationResolver) BasketDeleteBasket(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "basket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/baskets/%v", id)}, "")
}

func (r *mutationResolver) BasketCreateLine(ctx context.Context, input model.BasketLineInput) (*model.BasketLine, error) {
	return restCall[*model.BasketLine](ctx, r.Proxy, "basket-service", rest.Request{Method: "POST", Path: "/api/v1/basket-lines", Body: input}, "")
}

func (r *mutationResolver) BasketUpdateLine(ctx context.Context, id int, input model.BasketLineInput) (*model.BasketLine, error) {
	return restCall[*model.BasketLine](ctx, r.Proxy, "basket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/basket-lines/%v", id), Body: input}, "")
}

func (r *mutationResolver) BasketDeleteLine(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "basket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/basket-lines/%v", id)}, "")
}

func (r *mutationResolver) BasketCreateLineAttribute(ctx context.Context, input model.BasketLineAttributeInput) (*model.BasketLineAttribute, error) {
	return restCall[*model.BasketLineAttribute](ctx, r.Proxy, "basket-service", rest.Request{Method: "POST", Path: "/api/v1/basket-line-attributes", Body: input}, "")
}

func (r *mutationResolver) BasketUpdateLineAttribute(ctx context.Context, id int, input model.BasketLineAttributeInput) (*model.BasketLineAttribute, error) {
	return restCall[*model.BasketLineAttribute](ctx, r.Proxy, "basket-service", rest.Request{Method: "PUT", Path: rest.P("/api/v1/basket-line-attributes/%v", id), Body: input}, "")
}

func (r *mutationResolver) BasketDeleteLineAttribute(ctx context.Context, id int) (bool, error) {
	return restCall[bool](ctx, r.Proxy, "basket-service", rest.Request{Method: "DELETE", Path: rest.P("/api/v1/basket-line-attributes/%v", id)}, "")
}

func (r *mutationResolver) BasketCheckout(ctx context.Context, id int, walletID string, pin string, shippingAddressID *int, shippingFee *float64) (*model.BasketBasket, error) {
	return restCall[*model.BasketBasket](ctx, r.Proxy, "basket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/baskets/%v/checkout", id), Body: map[string]any{"wallet_id": walletID, "pin": pin, "shipping_address_id": shippingAddressID, "shipping_fee": shippingFee}}, "")
}

func (r *queryResolver) BasketBaskets(ctx context.Context) ([]*model.BasketBasket, error) {
	return restCall[[]*model.BasketBasket](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: "/api/v1/baskets"}, "")
}

func (r *queryResolver) BasketBasket(ctx context.Context, id int) (*model.BasketBasket, error) {
	return restCall[*model.BasketBasket](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/baskets/%v", id)}, "")
}

func (r *queryResolver) BasketLines(ctx context.Context) ([]*model.BasketLine, error) {
	return restCall[[]*model.BasketLine](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: "/api/v1/basket-lines"}, "")
}

func (r *queryResolver) BasketLine(ctx context.Context, id int) (*model.BasketLine, error) {
	return restCall[*model.BasketLine](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/basket-lines/%v", id)}, "")
}

func (r *queryResolver) BasketLineAttributes(ctx context.Context) ([]*model.BasketLineAttribute, error) {
	return restCall[[]*model.BasketLineAttribute](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: "/api/v1/basket-line-attributes"}, "")
}

func (r *queryResolver) BasketLineAttribute(ctx context.Context, id int) (*model.BasketLineAttribute, error) {
	return restCall[*model.BasketLineAttribute](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/basket-line-attributes/%v", id)}, "")
}

func (r *queryResolver) BasketOrders(ctx context.Context) ([]*model.BasketOrder, error) {
	return restCall[[]*model.BasketOrder](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: "/api/v1/orders"}, "")
}

func (r *queryResolver) BasketOrder(ctx context.Context, id int) (*model.BasketOrder, error) {
	return restCall[*model.BasketOrder](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/orders/%v", id)}, "")
}

func (r *queryResolver) BasketUser(ctx context.Context, id int) (*model.BasketUser, error) {
	return restCall[*model.BasketUser](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/users/%v", id)}, "")
}

func (r *queryResolver) BasketBasketLines(ctx context.Context, id int) ([]*model.BasketLine, error) {
	return restCall[[]*model.BasketLine](ctx, r.Proxy, "basket-service", rest.Request{Method: "GET", Path: rest.P("/api/v1/baskets/%v/lines", id)}, "")
}

func (r *queryResolver) BasketShippingQuote(ctx context.Context, id int, shippingAddressID int) (*model.BasketShippingQuote, error) {
	return restCall[*model.BasketShippingQuote](ctx, r.Proxy, "basket-service", rest.Request{Method: "POST", Path: rest.P("/api/v1/baskets/%v/shipping-quote", id), Body: map[string]any{"shipping_address_id": shippingAddressID}}, "")
}
