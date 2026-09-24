package repository

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"github.com/JIeeiroSst/catalogues-service/internal/domain/port"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type productRepo struct{ db *gorm.DB }

func NewProductRepo(db *gorm.DB) port.ProductRepository { return &productRepo{db: db} }

func productToEntity(p *model.Product) productEntity {
	return productEntity{
		ID: p.ID, Structure: string(p.Structure), UPC: p.UPC, Title: p.Title,
		Slug: p.Slug, Description: p.Description, Rating: p.Rating,
		IsDiscountable: p.IsDiscountable, IsPublic: p.IsPublic,
		ParentID: p.ParentID, ProductClassID: p.ProductClassID,
		MetaTitle: p.MetaTitle, MetaDescription: p.MetaDescription,
	}
}

func (r *productRepo) Create(ctx context.Context, p *model.Product) error {
	e := productToEntity(p)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Omit("Parent", "ProductClass").Create(&e).Error; err != nil {
			return err
		}
		return setLinks(tx, e.ID, p.CategoryIDs, p.OptionIDs)
	})
	if err != nil {
		return writeErr(err)
	}
	p.ID = e.ID
	return nil
}

func (r *productRepo) Update(ctx context.Context, p *model.Product) error {
	e := productToEntity(p)
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.Model(&productEntity{ID: p.ID}).
			Select("*").Omit("id", "date_created", "Parent", "ProductClass").Updates(&e)
		if res.Error != nil {
			return res.Error
		}
		if res.RowsAffected == 0 {
			return model.ErrNotFound
		}
		return setLinks(tx, p.ID, p.CategoryIDs, p.OptionIDs)
	})
	return writeErr(err)
}

func setLinks(tx *gorm.DB, productID int64, categoryIDs, optionIDs []int64) error {
	if err := tx.Where("product_id = ?", productID).Delete(&productCategoryEntity{}).Error; err != nil {
		return err
	}
	if err := tx.Where("product_id = ?", productID).Delete(&productOptionEntity{}).Error; err != nil {
		return err
	}
	if len(categoryIDs) > 0 {
		rows := make([]productCategoryEntity, len(categoryIDs))
		for i, id := range categoryIDs {
			rows[i] = productCategoryEntity{ProductID: productID, CategoryID: id}
		}
		if err := tx.Omit("Product", "Category").Create(&rows).Error; err != nil {
			return err
		}
	}
	if len(optionIDs) > 0 {
		rows := make([]productOptionEntity, len(optionIDs))
		for i, id := range optionIDs {
			rows[i] = productOptionEntity{ProductID: productID, OptionID: id}
		}
		if err := tx.Omit("Product", "Option").Create(&rows).Error; err != nil {
			return err
		}
	}
	return nil
}

func (r *productRepo) Get(ctx context.Context, id int64) (*model.Product, error) {
	var e productEntity
	if err := r.db.WithContext(ctx).First(&e, id).Error; err != nil {
		return nil, translate(err, nil)
	}
	ps, err := r.hydrate(ctx, []productEntity{e})
	if err != nil {
		return nil, err
	}
	return &ps[0], nil
}

func (r *productRepo) List(ctx context.Context, f model.ProductFilter) ([]model.Product, error) {
	q := r.db.WithContext(ctx).Model(&productEntity{})
	if f.OnlyPublic {
		q = q.Where("is_public")
	}
	if f.CategoryID != nil {
		q = q.Where("id IN (?)", r.db.Model(&productCategoryEntity{}).
			Select("product_id").Where("category_id = ?", *f.CategoryID))
	}
	if f.Query != "" {
		like := "%" + escapeLike(strings.ToLower(f.Query)) + "%"
		q = q.Where("lower(title) LIKE ? OR lower(upc) LIKE ?", like, like)
	}
	var es []productEntity
	if err := q.Order("id").Limit(f.Limit).Offset(f.Offset).Find(&es).Error; err != nil {
		return nil, err
	}
	return r.hydrate(ctx, es)
}

func escapeLike(s string) string {
	return strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`).Replace(s)
}

// hydrate converts entities and loads their links, images and
// recommendations with one query per relation, however many products.
func (r *productRepo) hydrate(ctx context.Context, es []productEntity) ([]model.Product, error) {
	out := make([]model.Product, len(es))
	idx := make(map[int64]int, len(es))
	ids := make([]int64, len(es))
	for i, e := range es {
		out[i] = model.Product{
			ID: e.ID, Structure: model.Structure(e.Structure), UPC: e.UPC,
			Title: e.Title, Slug: e.Slug, Description: e.Description,
			Rating: e.Rating, IsDiscountable: e.IsDiscountable, IsPublic: e.IsPublic,
			ParentID: e.ParentID, ProductClassID: e.ProductClassID,
			MetaTitle: e.MetaTitle, MetaDescription: e.MetaDescription,
			DateCreated: e.DateCreated, DateUpdated: e.DateUpdated,
			CategoryIDs: []int64{}, OptionIDs: []int64{},
			Images: []model.ProductImage{}, Recommendations: []model.Recommendation{},
		}
		idx[e.ID], ids[i] = i, e.ID
	}
	if len(ids) == 0 {
		return out, nil
	}
	db := r.db.WithContext(ctx)

	var cats []productCategoryEntity
	if err := db.Where("product_id IN ?", ids).Order("category_id").Find(&cats).Error; err != nil {
		return nil, err
	}
	for _, c := range cats {
		i := idx[c.ProductID]
		out[i].CategoryIDs = append(out[i].CategoryIDs, c.CategoryID)
	}

	var opts []productOptionEntity
	if err := db.Where("product_id IN ?", ids).Order("option_id").Find(&opts).Error; err != nil {
		return nil, err
	}
	for _, o := range opts {
		i := idx[o.ProductID]
		out[i].OptionIDs = append(out[i].OptionIDs, o.OptionID)
	}

	var imgs []productImageEntity
	if err := db.Where("product_id IN ?", ids).Order("display_order, id").Find(&imgs).Error; err != nil {
		return nil, err
	}
	for _, m := range imgs {
		i := idx[m.ProductID]
		out[i].Images = append(out[i].Images, m.toModel())
	}

	var recs []recommendationEntity
	if err := db.Where("primary_id IN ?", ids).Order("ranking, id").Find(&recs).Error; err != nil {
		return nil, err
	}
	for _, rc := range recs {
		i := idx[rc.PrimaryID]
		out[i].Recommendations = append(out[i].Recommendations,
			model.Recommendation{RecommendedID: rc.RecommendedID, Ranking: rc.Ranking})
	}
	return out, nil
}

func (m productImageEntity) toModel() model.ProductImage {
	return model.ProductImage{
		ID: m.ID, ProductID: m.ProductID, Original: m.Original, Caption: m.Caption,
		DisplayOrder: m.DisplayOrder, DateCreated: m.DateCreated,
	}
}

func (r *productRepo) Delete(ctx context.Context, id int64) error {
	return deleteByID[productEntity](r.db.WithContext(ctx), id)
}

func (r *productRepo) AddImage(ctx context.Context, img *model.ProductImage) error {
	e := productImageEntity{
		ProductID: img.ProductID, Original: img.Original,
		Caption: img.Caption, DisplayOrder: img.DisplayOrder,
	}
	// A missing product surfaces as a foreign-key violation.
	if err := r.db.WithContext(ctx).Omit("Product").Create(&e).Error; err != nil {
		return translate(err, model.ErrNotFound)
	}
	*img = e.toModel()
	return nil
}

func (r *productRepo) DeleteImage(ctx context.Context, productID, imageID int64) error {
	res := r.db.WithContext(ctx).Where("id = ? AND product_id = ?", imageID, productID).Delete(&productImageEntity{})
	if res.Error != nil {
		return res.Error
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r *productRepo) SetRecommendations(ctx context.Context, productID int64, recs []model.Recommendation) error {
	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var n int64
		if err := tx.Model(&productEntity{}).Where("id = ?", productID).Count(&n).Error; err != nil {
			return err
		}
		if n == 0 {
			return model.ErrNotFound
		}
		if err := tx.Where("primary_id = ?", productID).Delete(&recommendationEntity{}).Error; err != nil {
			return err
		}
		if len(recs) == 0 {
			return nil
		}
		rows := make([]recommendationEntity, len(recs))
		for i, rc := range recs {
			rows[i] = recommendationEntity{PrimaryID: productID, RecommendedID: rc.RecommendedID, Ranking: rc.Ranking}
		}
		return tx.Clauses(clause.OnConflict{DoNothing: true}).Omit("Primary", "Recommended").Create(&rows).Error
	})
	return writeErr(err)
}
