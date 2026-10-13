package port

import (
	"context"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
)

type ListFilter struct {
	OwnerID    string
	PublicOnly bool
	Limit      int
	Offset     int
}

type WishlistRepository interface {
	Create(ctx context.Context, w *model.Wishlist) error
	FirstOrCreate(ctx context.Context, w *model.Wishlist) (*model.Wishlist, bool, error)
	Get(ctx context.Context, id int64) (*model.Wishlist, error)
	GetByKey(ctx context.Context, key string) (*model.Wishlist, error)
	List(ctx context.Context, f ListFilter) ([]model.Wishlist, int64, error)
	Update(ctx context.Context, w *model.Wishlist) error
	Delete(ctx context.Context, id int64) error
	AddLine(ctx context.Context, l *model.Line) (*model.Line, error)
	UpdateLineQuantity(ctx context.Context, wishlistID, lineID int64, quantity int) (*model.Line, error)
	DeleteLine(ctx context.Context, wishlistID, lineID int64) error
	ClearLines(ctx context.Context, wishlistID int64) (int64, error)
	MoveLine(ctx context.Context, fromWishlistID, lineID, toWishlistID int64) (*model.Line, error)
	FindProduct(ctx context.Context, ownerID string, productID int64) ([]model.ProductMembership, error)
	PopularProducts(ctx context.Context, limit int) ([]model.ProductStat, error)
}

type UserDirectory interface {
	GetUser(ctx context.Context, id string) (*model.Owner, error)
}

type RateLimiter interface {
	Allow(ctx context.Context, key string, limit int, window time.Duration) (bool, error)
}

type HealthChecker interface {
	Name() string
	Critical() bool
	Ping(ctx context.Context) error
}
