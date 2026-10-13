package postgres

import (
	"context"
	_ "embed"
	"errors"
	"time"

	"github.com/JIeeiroSst/wishlists-service/internal/domain/model"
	"github.com/JIeeiroSst/wishlists-service/internal/domain/port"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

//go:embed schema.sql
var schema string

const (
	migrationLockID = 7_432_019_011
	ownerLockSpace  = 74_320
)

type wishlistRecord struct {
	ID          int64        `gorm:"column:id;primaryKey"`
	Name        string       `gorm:"column:name"`
	Key         string       `gorm:"column:key"`
	Visibility  string       `gorm:"column:visibility"`
	DateCreated time.Time    `gorm:"column:date_created"`
	OwnerID     string       `gorm:"column:owner_id"`
	Lines       []lineRecord `gorm:"foreignKey:WishlistID"`
}

func (wishlistRecord) TableName() string { return "wishlists_wishlist" }

type lineRecord struct {
	ID         int64  `gorm:"column:id;primaryKey"`
	WishlistID int64  `gorm:"column:wishlist_id"`
	ProductID  int64  `gorm:"column:product_id"`
	Title      string `gorm:"column:title"`
	Quantity   int    `gorm:"column:quantity"`
}

func (lineRecord) TableName() string { return "wishlists_line" }

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Migrate(ctx context.Context) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?)", migrationLockID).Error; err != nil {
			return err
		}
		return tx.Exec(schema).Error
	})
}

func (r *Repository) Ping(ctx context.Context) error {
	sqlDB, err := r.db.DB()
	if err != nil {
		return err
	}
	return sqlDB.PingContext(ctx)
}

func (r *Repository) Name() string { return "postgres" }

func (r *Repository) Critical() bool { return true }

func (r *Repository) Create(ctx context.Context, w *model.Wishlist) error {
	return translate(r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return insertWishlist(tx, w)
	}))
}

func (r *Repository) FirstOrCreate(ctx context.Context, w *model.Wishlist) (*model.Wishlist, bool, error) {
	var out *model.Wishlist
	created := false
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Exec("SELECT pg_advisory_xact_lock(?, hashtext(?))", ownerLockSpace, w.OwnerID).Error; err != nil {
			return err
		}
		var rec wishlistRecord
		err := preloadLines(tx).Where("owner_id = ?", w.OwnerID).Order("id").First(&rec).Error
		switch {
		case err == nil:
			out = toModel(&rec)
			return nil
		case !errors.Is(err, gorm.ErrRecordNotFound):
			return err
		}
		if err := insertWishlist(tx, w); err != nil {
			return err
		}
		out, created = w, true
		return nil
	})
	if err != nil {
		return nil, false, translate(err)
	}
	return out, created, nil
}

func (r *Repository) Get(ctx context.Context, id int64) (*model.Wishlist, error) {
	return r.first(ctx, "id = ?", id)
}

func (r *Repository) GetByKey(ctx context.Context, key string) (*model.Wishlist, error) {
	return r.first(ctx, `"key" = ?`, key)
}

func (r *Repository) List(ctx context.Context, f port.ListFilter) ([]model.Wishlist, int64, error) {
	scope := func(q *gorm.DB) *gorm.DB {
		if f.OwnerID != "" {
			q = q.Where("owner_id = ?", f.OwnerID)
		}
		if f.PublicOnly {
			q = q.Where("visibility = ?", string(model.VisibilityPublic))
		}
		return q
	}
	var total int64
	if err := scope(r.db.WithContext(ctx).Model(&wishlistRecord{})).Count(&total).Error; err != nil {
		return nil, 0, translate(err)
	}
	var recs []wishlistRecord
	if err := scope(r.withLines(ctx)).Order("id").Limit(f.Limit).Offset(f.Offset).Find(&recs).Error; err != nil {
		return nil, 0, translate(err)
	}
	out := make([]model.Wishlist, 0, len(recs))
	for i := range recs {
		out = append(out, *toModel(&recs[i]))
	}
	return out, total, nil
}

func (r *Repository) Update(ctx context.Context, w *model.Wishlist) error {
	res := r.db.WithContext(ctx).Model(&wishlistRecord{}).Where("id = ?", w.ID).Updates(map[string]any{
		"name":       w.Name,
		"key":        w.Key,
		"visibility": string(w.Visibility),
	})
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *Repository) Delete(ctx context.Context, id int64) error {
	res := r.db.WithContext(ctx).Delete(&wishlistRecord{}, id)
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *Repository) AddLine(ctx context.Context, l *model.Line) (*model.Line, error) {
	out, err := upsertLine(r.db.WithContext(ctx), l)
	if err != nil {
		return nil, translate(err)
	}
	return out, nil
}

func (r *Repository) UpdateLineQuantity(ctx context.Context, wishlistID, lineID int64, quantity int) (*model.Line, error) {
	var recs []lineRecord
	err := r.db.WithContext(ctx).Raw(`
		UPDATE wishlists_line SET quantity = ?
		WHERE id = ? AND wishlist_id = ?
		RETURNING id, wishlist_id, product_id, title, quantity`,
		quantity, lineID, wishlistID).Scan(&recs).Error
	if err != nil {
		return nil, translate(err)
	}
	if len(recs) == 0 {
		return nil, model.ErrNotFound
	}
	return lineToModel(&recs[0]), nil
}

func (r *Repository) DeleteLine(ctx context.Context, wishlistID, lineID int64) error {
	res := r.db.WithContext(ctx).Where("id = ? AND wishlist_id = ?", lineID, wishlistID).Delete(&lineRecord{})
	if res.Error != nil {
		return translate(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *Repository) ClearLines(ctx context.Context, wishlistID int64) (int64, error) {
	res := r.db.WithContext(ctx).Where("wishlist_id = ?", wishlistID).Delete(&lineRecord{})
	if res.Error != nil {
		return 0, translate(res.Error)
	}
	return res.RowsAffected, nil
}

func (r *Repository) MoveLine(ctx context.Context, fromWishlistID, lineID, toWishlistID int64) (*model.Line, error) {
	var out *model.Line
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var src []lineRecord
		if err := tx.Raw(`
			DELETE FROM wishlists_line WHERE id = ? AND wishlist_id = ?
			RETURNING id, wishlist_id, product_id, title, quantity`,
			lineID, fromWishlistID).Scan(&src).Error; err != nil {
			return err
		}
		if len(src) == 0 {
			return model.ErrNotFound
		}
		moved, err := upsertLine(tx, &model.Line{
			WishlistID: toWishlistID,
			ProductID:  src[0].ProductID,
			Title:      src[0].Title,
			Quantity:   src[0].Quantity,
		})
		out = moved
		return err
	})
	if err != nil {
		return nil, translate(err)
	}
	return out, nil
}

func (r *Repository) FindProduct(ctx context.Context, ownerID string, productID int64) ([]model.ProductMembership, error) {
	out := []model.ProductMembership{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT w.id AS wishlist_id, w.name AS wishlist_name, l.id AS line_id, l.quantity
		FROM wishlists_line l
		JOIN wishlists_wishlist w ON w.id = l.wishlist_id
		WHERE w.owner_id = ? AND l.product_id = ?
		ORDER BY w.id`, ownerID, productID).Scan(&out).Error
	if err != nil {
		return nil, translate(err)
	}
	return out, nil
}

func (r *Repository) PopularProducts(ctx context.Context, limit int) ([]model.ProductStat, error) {
	out := []model.ProductStat{}
	err := r.db.WithContext(ctx).Raw(`
		SELECT product_id, MAX(title) AS title,
		       COUNT(DISTINCT wishlist_id) AS wishlist_count,
		       SUM(quantity) AS total_quantity
		FROM wishlists_line
		GROUP BY product_id
		ORDER BY wishlist_count DESC, total_quantity DESC, product_id
		LIMIT ?`, limit).Scan(&out).Error
	if err != nil {
		return nil, translate(err)
	}
	return out, nil
}

func insertWishlist(tx *gorm.DB, w *model.Wishlist) error {
	var created wishlistRecord
	err := tx.Raw(`
		INSERT INTO wishlists_wishlist (name, "key", visibility, owner_id)
		VALUES (?, ?, ?, ?)
		RETURNING id, date_created`,
		w.Name, w.Key, string(w.Visibility), w.OwnerID).Scan(&created).Error
	if err != nil {
		return err
	}
	w.ID, w.DateCreated = created.ID, created.DateCreated
	for i := range w.Lines {
		w.Lines[i].WishlistID = w.ID
		l, err := upsertLine(tx, &w.Lines[i])
		if err != nil {
			return err
		}
		w.Lines[i] = *l
	}
	return nil
}

func upsertLine(tx *gorm.DB, l *model.Line) (*model.Line, error) {
	var rec lineRecord
	err := tx.Raw(`
		INSERT INTO wishlists_line (wishlist_id, product_id, title, quantity)
		VALUES (?, ?, ?, ?)
		ON CONFLICT (wishlist_id, product_id)
		DO UPDATE SET quantity = LEAST(wishlists_line.quantity + EXCLUDED.quantity, ?), title = EXCLUDED.title
		RETURNING id, wishlist_id, product_id, title, quantity`,
		l.WishlistID, l.ProductID, l.Title, l.Quantity, model.MaxQuantity).Scan(&rec).Error
	if err != nil {
		return nil, err
	}
	return lineToModel(&rec), nil
}

func preloadLines(db *gorm.DB) *gorm.DB {
	return db.Preload("Lines", func(db *gorm.DB) *gorm.DB { return db.Order("id") })
}

func (r *Repository) first(ctx context.Context, query string, arg any) (*model.Wishlist, error) {
	var rec wishlistRecord
	if err := r.withLines(ctx).Where(query, arg).First(&rec).Error; err != nil {
		return nil, translate(err)
	}
	return toModel(&rec), nil
}

func (r *Repository) withLines(ctx context.Context) *gorm.DB {
	return preloadLines(r.db.WithContext(ctx))
}

func translate(err error) error {
	var pgErr *pgconn.PgError
	switch {
	case err == nil:
		return nil
	case errors.Is(err, model.ErrNotFound), errors.Is(err, model.ErrConflict):
		return err
	case errors.Is(err, gorm.ErrRecordNotFound), errors.Is(err, gorm.ErrForeignKeyViolated):
		return model.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return model.ErrConflict
	case errors.As(err, &pgErr) && pgErr.Code == "23505":
		return model.ErrConflict
	case errors.As(err, &pgErr) && pgErr.Code == "23503":
		return model.ErrNotFound
	}
	return err
}

func toModel(rec *wishlistRecord) *model.Wishlist {
	w := &model.Wishlist{
		ID:          rec.ID,
		Name:        rec.Name,
		Key:         rec.Key,
		Visibility:  model.Visibility(rec.Visibility),
		DateCreated: rec.DateCreated,
		OwnerID:     rec.OwnerID,
		Lines:       make([]model.Line, 0, len(rec.Lines)),
	}
	for i := range rec.Lines {
		w.Lines = append(w.Lines, *lineToModel(&rec.Lines[i]))
	}
	return w
}

func lineToModel(rec *lineRecord) *model.Line {
	return &model.Line{
		ID:         rec.ID,
		WishlistID: rec.WishlistID,
		ProductID:  rec.ProductID,
		Title:      rec.Title,
		Quantity:   rec.Quantity,
	}
}
