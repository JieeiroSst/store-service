package port

import (
	"context"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
)

type CreateWishlistInput struct {
	OwnerID    string `json:"owner_id"`
	Name       string `json:"name"`
	Visibility string `json:"visibility"`
}

type UpdateWishlistInput struct {
	Name          *string `json:"name"`
	Visibility    *string `json:"visibility"`
	RegenerateKey bool    `json:"regenerate_key"`
}

type ListWishlistsInput struct {
	OwnerID string
	Limit   int
	Offset  int
}

type WishlistPage struct {
	Items  []model.Wishlist `json:"items"`
	Total  int64            `json:"total"`
	Limit  int              `json:"limit"`
	Offset int              `json:"offset"`
}

type AddLineInput struct {
	ProductID int64  `json:"product_id"`
	Title     string `json:"title"`
	Quantity  int    `json:"quantity"`
}

type WishlistUsecase interface {
	Create(ctx context.Context, actor model.Actor, in CreateWishlistInput) (*model.Wishlist, error)
	Get(ctx context.Context, actor model.Actor, id int64) (*model.Wishlist, error)
	GetByKey(ctx context.Context, actor model.Actor, key string) (*model.Wishlist, error)
	List(ctx context.Context, actor model.Actor, in ListWishlistsInput) (*WishlistPage, error)
	Update(ctx context.Context, actor model.Actor, id int64, in UpdateWishlistInput) (*model.Wishlist, error)
	Delete(ctx context.Context, actor model.Actor, id int64) error
	Default(ctx context.Context, actor model.Actor) (*model.Wishlist, error)
	CopyShared(ctx context.Context, actor model.Actor, key string) (*model.Wishlist, error)
	AddLine(ctx context.Context, actor model.Actor, wishlistID int64, in AddLineInput) (*model.Line, error)
	AddToDefault(ctx context.Context, actor model.Actor, in AddLineInput) (*model.Line, error)
	UpdateLine(ctx context.Context, actor model.Actor, wishlistID, lineID int64, quantity int) (*model.Line, error)
	RemoveLine(ctx context.Context, actor model.Actor, wishlistID, lineID int64) error
	ClearLines(ctx context.Context, actor model.Actor, wishlistID int64) (int64, error)
	MoveLine(ctx context.Context, actor model.Actor, wishlistID, lineID, targetWishlistID int64) (*model.Line, error)
	FindProduct(ctx context.Context, actor model.Actor, productID int64) ([]model.ProductMembership, error)
	PopularProducts(ctx context.Context, limit int) ([]model.ProductStat, error)
}
