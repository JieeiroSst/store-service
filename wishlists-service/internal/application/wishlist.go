package application

import (
	"context"
	"errors"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
)

const (
	defaultPageSize    = 50
	maxPageSize        = 200
	keyAttempts        = 5
	defaultPopularSize = 10
	maxPopularSize     = 100
	enrichConcurrency  = 8
	enrichTimeout      = 3 * time.Second
)

type wishlistService struct {
	repo  port.WishlistRepository
	users port.UserDirectory
}

func NewWishlistService(repo port.WishlistRepository, users port.UserDirectory) port.WishlistUsecase {
	return &wishlistService{repo: repo, users: users}
}

func (s *wishlistService) Create(ctx context.Context, actor model.Actor, in port.CreateWishlistInput) (*model.Wishlist, error) {
	owner := strings.TrimSpace(in.OwnerID)
	switch {
	case owner == "":
		owner = actor.UserID
	case owner != actor.UserID && !actor.Admin:
		return nil, model.ErrForbidden
	}
	visibility, err := model.ParseVisibility(in.Visibility)
	if err != nil {
		return nil, err
	}
	w, err := model.NewWishlist(owner, in.Name, visibility)
	if err != nil {
		return nil, err
	}
	var profile *model.Owner
	if owner != actor.UserID {
		if profile, err = s.activeOwner(ctx, owner); err != nil {
			return nil, err
		}
	}
	if err := s.create(ctx, w); err != nil {
		return nil, err
	}
	if profile != nil {
		w.Owner = profile
	} else {
		s.enrich(ctx, w)
	}
	return w, nil
}

func (s *wishlistService) Get(ctx context.Context, actor model.Actor, id int64) (*model.Wishlist, error) {
	w, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !w.VisibleTo(actor) {
		return nil, model.ErrNotFound
	}
	redactKeys(actor, w)
	s.enrich(ctx, w)
	return w, nil
}

func (s *wishlistService) GetByKey(ctx context.Context, actor model.Actor, key string) (*model.Wishlist, error) {
	w, err := s.byKey(ctx, actor, key)
	if err != nil {
		return nil, err
	}
	s.enrich(ctx, w)
	return w, nil
}

func (s *wishlistService) List(ctx context.Context, actor model.Actor, in port.ListWishlistsInput) (*port.WishlistPage, error) {
	f := port.ListFilter{OwnerID: strings.TrimSpace(in.OwnerID), Limit: in.Limit, Offset: in.Offset}
	if f.Limit <= 0 {
		f.Limit = defaultPageSize
	}
	if f.Limit > maxPageSize {
		f.Limit = maxPageSize
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	if !actor.Admin {
		switch {
		case f.OwnerID == "" && actor.UserID == "":
			return nil, model.ErrForbidden
		case f.OwnerID == "":
			f.OwnerID = actor.UserID
		case f.OwnerID != actor.UserID:
			f.PublicOnly = true
		}
	}
	items, total, err := s.repo.List(ctx, f)
	if err != nil {
		return nil, err
	}
	ptrs := make([]*model.Wishlist, len(items))
	for i := range items {
		ptrs[i] = &items[i]
	}
	redactKeys(actor, ptrs...)
	s.enrich(ctx, ptrs...)
	return &port.WishlistPage{Items: items, Total: total, Limit: f.Limit, Offset: f.Offset}, nil
}

func (s *wishlistService) Update(ctx context.Context, actor model.Actor, id int64, in port.UpdateWishlistInput) (*model.Wishlist, error) {
	w, err := s.owned(ctx, actor, id)
	if err != nil {
		return nil, err
	}
	if in.Name != nil {
		if err := w.Rename(*in.Name); err != nil {
			return nil, err
		}
	}
	if in.Visibility != nil {
		if w.Visibility, err = model.ParseVisibility(*in.Visibility); err != nil {
			return nil, err
		}
	}
	attempts := 1
	if in.RegenerateKey {
		attempts = keyAttempts
	}
	for i := 0; i < attempts; i++ {
		if in.RegenerateKey {
			if err := w.RegenerateKey(); err != nil {
				return nil, err
			}
		}
		if err = s.repo.Update(ctx, w); !errors.Is(err, model.ErrConflict) {
			break
		}
	}
	if err != nil {
		return nil, err
	}
	s.enrich(ctx, w)
	return w, nil
}

func (s *wishlistService) Delete(ctx context.Context, actor model.Actor, id int64) error {
	if _, err := s.owned(ctx, actor, id); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *wishlistService) Default(ctx context.Context, actor model.Actor) (*model.Wishlist, error) {
	w, err := s.defaultFor(ctx, actor)
	if err != nil {
		return nil, err
	}
	s.enrich(ctx, w)
	return w, nil
}

func (s *wishlistService) CopyShared(ctx context.Context, actor model.Actor, key string) (*model.Wishlist, error) {
	if actor.UserID == "" {
		return nil, &model.ValidationError{Field: "owner_id", Reason: "copying requires a user session"}
	}
	src, err := s.byKey(ctx, actor, key)
	if err != nil {
		return nil, err
	}
	w, err := src.CopyFor(actor.UserID)
	if err != nil {
		return nil, err
	}
	if err := s.create(ctx, w); err != nil {
		return nil, err
	}
	s.enrich(ctx, w)
	return w, nil
}

func (s *wishlistService) AddLine(ctx context.Context, actor model.Actor, wishlistID int64, in port.AddLineInput) (*model.Line, error) {
	w, err := s.owned(ctx, actor, wishlistID)
	if err != nil {
		return nil, err
	}
	return s.addLine(ctx, w, in)
}

func (s *wishlistService) AddToDefault(ctx context.Context, actor model.Actor, in port.AddLineInput) (*model.Line, error) {
	w, err := s.defaultFor(ctx, actor)
	if err != nil {
		return nil, err
	}
	return s.addLine(ctx, w, in)
}

func (s *wishlistService) UpdateLine(ctx context.Context, actor model.Actor, wishlistID, lineID int64, quantity int) (*model.Line, error) {
	if err := model.ValidateQuantity(quantity); err != nil {
		return nil, err
	}
	if _, err := s.owned(ctx, actor, wishlistID); err != nil {
		return nil, err
	}
	return s.repo.UpdateLineQuantity(ctx, wishlistID, lineID, quantity)
}

func (s *wishlistService) RemoveLine(ctx context.Context, actor model.Actor, wishlistID, lineID int64) error {
	if _, err := s.owned(ctx, actor, wishlistID); err != nil {
		return err
	}
	return s.repo.DeleteLine(ctx, wishlistID, lineID)
}

func (s *wishlistService) ClearLines(ctx context.Context, actor model.Actor, wishlistID int64) (int64, error) {
	if _, err := s.owned(ctx, actor, wishlistID); err != nil {
		return 0, err
	}
	return s.repo.ClearLines(ctx, wishlistID)
}

func (s *wishlistService) MoveLine(ctx context.Context, actor model.Actor, wishlistID, lineID, targetWishlistID int64) (*model.Line, error) {
	if targetWishlistID == wishlistID {
		return nil, &model.ValidationError{Field: "target_wishlist_id", Reason: "must differ from the source wishlist"}
	}
	src, err := s.owned(ctx, actor, wishlistID)
	if err != nil {
		return nil, err
	}
	var line *model.Line
	for i := range src.Lines {
		if src.Lines[i].ID == lineID {
			line = &src.Lines[i]
		}
	}
	if line == nil {
		return nil, model.ErrNotFound
	}
	dst, err := s.owned(ctx, actor, targetWishlistID)
	if err != nil {
		return nil, err
	}
	if dst.OwnerID != src.OwnerID {
		return nil, &model.ValidationError{Field: "target_wishlist_id", Reason: "must belong to the same owner"}
	}
	if err := dst.CanAdd(line.ProductID); err != nil {
		return nil, err
	}
	return s.repo.MoveLine(ctx, wishlistID, lineID, targetWishlistID)
}

func (s *wishlistService) FindProduct(ctx context.Context, actor model.Actor, productID int64) ([]model.ProductMembership, error) {
	if actor.UserID == "" {
		return nil, &model.ValidationError{Field: "owner_id", Reason: "requires a user session"}
	}
	if productID <= 0 {
		return nil, &model.ValidationError{Field: "product_id", Reason: "must be a positive integer"}
	}
	return s.repo.FindProduct(ctx, actor.UserID, productID)
}

func (s *wishlistService) PopularProducts(ctx context.Context, limit int) ([]model.ProductStat, error) {
	if limit <= 0 {
		limit = defaultPopularSize
	}
	if limit > maxPopularSize {
		limit = maxPopularSize
	}
	return s.repo.PopularProducts(ctx, limit)
}

func (s *wishlistService) create(ctx context.Context, w *model.Wishlist) error {
	var err error
	for i := 0; i < keyAttempts; i++ {
		if i > 0 {
			if err := w.RegenerateKey(); err != nil {
				return err
			}
		}
		if err = s.repo.Create(ctx, w); !errors.Is(err, model.ErrConflict) {
			return err
		}
	}
	return err
}

func (s *wishlistService) defaultFor(ctx context.Context, actor model.Actor) (*model.Wishlist, error) {
	if actor.UserID == "" {
		return nil, &model.ValidationError{Field: "owner_id", Reason: "a default wishlist requires a user session"}
	}
	var lastErr error
	for i := 0; i < keyAttempts; i++ {
		candidate, err := model.NewWishlist(actor.UserID, model.DefaultName, model.VisibilityPrivate)
		if err != nil {
			return nil, err
		}
		w, _, err := s.repo.FirstOrCreate(ctx, candidate)
		if !errors.Is(err, model.ErrConflict) {
			return w, err
		}
		lastErr = err
	}
	return nil, lastErr
}

func (s *wishlistService) addLine(ctx context.Context, w *model.Wishlist, in port.AddLineInput) (*model.Line, error) {
	l, err := model.NewLine(w.ID, in.ProductID, in.Title, in.Quantity)
	if err != nil {
		return nil, err
	}
	if err := w.CanAdd(l.ProductID); err != nil {
		return nil, err
	}
	return s.repo.AddLine(ctx, l)
}

func (s *wishlistService) byKey(ctx context.Context, actor model.Actor, key string) (*model.Wishlist, error) {
	key = strings.TrimSpace(key)
	if key == "" {
		return nil, model.ErrNotFound
	}
	w, err := s.repo.GetByKey(ctx, key)
	if err != nil {
		return nil, err
	}
	if !w.VisibleByKey(actor) {
		return nil, model.ErrNotFound
	}
	return w, nil
}

func (s *wishlistService) owned(ctx context.Context, actor model.Actor, id int64) (*model.Wishlist, error) {
	w, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if actor.Owns(w) {
		return w, nil
	}
	if w.Visibility == model.VisibilityPublic {
		return nil, model.ErrForbidden
	}
	return nil, model.ErrNotFound
}

func (s *wishlistService) activeOwner(ctx context.Context, id string) (*model.Owner, error) {
	u, err := s.users.GetUser(ctx, id)
	switch {
	case errors.Is(err, model.ErrNotFound):
		return nil, &model.ValidationError{Field: "owner_id", Reason: "user does not exist"}
	case err != nil:
		return nil, err
	case !u.Active:
		return nil, &model.ValidationError{Field: "owner_id", Reason: "user is not active"}
	}
	return u, nil
}

func (s *wishlistService) enrich(ctx context.Context, ws ...*model.Wishlist) {
	var ids []string
	seen := map[string]bool{}
	for _, w := range ws {
		if w.OwnerID != "" && !seen[w.OwnerID] {
			seen[w.OwnerID] = true
			ids = append(ids, w.OwnerID)
		}
	}
	if len(ids) == 0 {
		return
	}
	ctx, cancel := context.WithTimeout(ctx, enrichTimeout)
	defer cancel()
	owners := make([]*model.Owner, len(ids))
	sem := make(chan struct{}, enrichConcurrency)
	var wg sync.WaitGroup
	for i, id := range ids {
		wg.Add(1)
		sem <- struct{}{}
		go func() {
			defer wg.Done()
			defer func() { <-sem }()
			u, err := s.users.GetUser(ctx, id)
			switch {
			case err == nil:
				owners[i] = u
			case errors.Is(err, model.ErrNotFound), errors.Is(err, model.ErrUnauthenticated):
			default:
				slog.Warn("owner lookup failed", "owner_id", id, "err", err)
			}
		}()
	}
	wg.Wait()
	byID := make(map[string]*model.Owner, len(ids))
	for i, id := range ids {
		byID[id] = owners[i]
	}
	for _, w := range ws {
		w.Owner = byID[w.OwnerID]
	}
}

func redactKeys(actor model.Actor, ws ...*model.Wishlist) {
	for _, w := range ws {
		if !actor.Owns(w) {
			w.Key = ""
		}
	}
}
