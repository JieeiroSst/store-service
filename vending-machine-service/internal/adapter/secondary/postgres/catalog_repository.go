package postgres

import (
	"context"
	"strconv"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type categoryRepository struct{ db *DB }

func NewCategoryRepository(db *DB) port.CategoryRepository { return &categoryRepository{db: db} }

func (r *categoryRepository) Create(ctx context.Context, c *domain.Category) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO categories
		(category_id, name, description, display_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		c.ID, c.Name, c.Description, c.DisplayOrder, c.CreatedAt, c.UpdatedAt)
	return mapErr(err)
}

func (r *categoryRepository) List(ctx context.Context, page domain.PageRequest) (domain.Page[domain.Category], error) {
	after, err := intCursorArgs(page.Cursor)
	if err != nil {
		return domain.Page[domain.Category]{}, err
	}
	size := page.Size()
	rows, err := r.db.q(ctx).Query(ctx, `SELECT category_id, name, description, display_order, created_at, updated_at
		FROM categories
		WHERE $1::int IS NULL OR (display_order, name) > ($1::int, $2::text)
		ORDER BY display_order, name LIMIT $3`, after[0], after[1], size+1)
	if err != nil {
		return domain.Page[domain.Category]{}, mapErr(err)
	}
	items, err := collect(rows, func(row pgx.Row) (domain.Category, error) {
		var c domain.Category
		err := row.Scan(&c.ID, &c.Name, &c.Description, &c.DisplayOrder, &c.CreatedAt, &c.UpdatedAt)
		return c, err
	})
	if err != nil {
		return domain.Page[domain.Category]{}, err
	}
	return paginate(items, size, func(c *domain.Category) []string {
		return []string{strconv.Itoa(c.DisplayOrder), c.Name}
	}), nil
}

const productColumns = `p.product_id, p.name, p.description, p.price_cents, p.category_id,
	p.image_url, p.barcode, p.is_active, p.created_at, p.updated_at`

func productDest(p *domain.Product) []any {
	return []any{&p.ID, &p.Name, &p.Description, &p.PriceCents, &p.CategoryID,
		&p.ImageURL, &p.Barcode, &p.IsActive, &p.CreatedAt, &p.UpdatedAt}
}

type productRepository struct{ db *DB }

func NewProductRepository(db *DB) port.ProductRepository { return &productRepository{db: db} }

func (r *productRepository) Create(ctx context.Context, p *domain.Product) error {
	return r.db.WithinTx(ctx, func(ctx context.Context) error {
		_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO products
			(product_id, name, description, price_cents, category_id, image_url, barcode, is_active, created_at, updated_at)
			VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)`,
			p.ID, p.Name, p.Description, p.PriceCents, p.CategoryID, p.ImageURL, p.Barcode, p.IsActive, p.CreatedAt, p.UpdatedAt)
		if err != nil {
			return mapErr(err)
		}
		for k, v := range p.Attributes {
			if _, err := r.db.q(ctx).Exec(ctx, `INSERT INTO product_attributes (attribute_id, product_id, key, value)
				VALUES ($1, $2, $3, $4)`, uuid.NewString(), p.ID, k, v); err != nil {
				return mapErr(err)
			}
		}
		return nil
	})
}

func (r *productRepository) Update(ctx context.Context, p *domain.Product) error {
	return updated(r.db.q(ctx).Exec(ctx, `UPDATE products
		SET name = $2, description = $3, price_cents = $4, image_url = $5, barcode = $6, is_active = $7, updated_at = $8
		WHERE product_id = $1`,
		p.ID, p.Name, p.Description, p.PriceCents, p.ImageURL, p.Barcode, p.IsActive, p.UpdatedAt))
}

func (r *productRepository) Get(ctx context.Context, id string) (*domain.Product, error) {
	var p domain.Product
	err := r.db.q(ctx).QueryRow(ctx, `SELECT `+productColumns+` FROM products p WHERE p.product_id = $1`, id).
		Scan(productDest(&p)...)
	if err != nil {
		return nil, mapErr(err)
	}
	products := []domain.Product{p}
	if err := r.loadAttributes(ctx, products); err != nil {
		return nil, err
	}
	return &products[0], nil
}

func (r *productRepository) List(ctx context.Context, categoryID string, page domain.PageRequest) (domain.Page[domain.Product], error) {
	after, err := cursorArgs(page.Cursor, 2)
	if err != nil {
		return domain.Page[domain.Product]{}, err
	}
	size := page.Size()
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+productColumns+` FROM products p
		WHERE ($1 = '' OR p.category_id = $1)
		  AND ($2::text IS NULL OR (p.name, p.product_id) > ($2::text, $3::text))
		ORDER BY p.name, p.product_id LIMIT $4`, categoryID, after[0], after[1], size+1)
	if err != nil {
		return domain.Page[domain.Product]{}, mapErr(err)
	}
	products, err := collect(rows, func(row pgx.Row) (domain.Product, error) {
		var p domain.Product
		err := row.Scan(productDest(&p)...)
		return p, err
	})
	if err != nil {
		return domain.Page[domain.Product]{}, err
	}
	result := paginate(products, size, func(p *domain.Product) []string { return []string{p.Name, p.ID} })
	return result, r.loadAttributes(ctx, result.Items)
}

func (r *productRepository) loadAttributes(ctx context.Context, products []domain.Product) error {
	if len(products) == 0 {
		return nil
	}
	index := make(map[string]*domain.Product, len(products))
	ids := make([]string, len(products))
	for i := range products {
		products[i].Attributes = map[string]string{}
		index[products[i].ID] = &products[i]
		ids[i] = products[i].ID
	}
	rows, err := r.db.q(ctx).Query(ctx, `SELECT product_id, key, value FROM product_attributes
		WHERE product_id = ANY($1)`, ids)
	if err != nil {
		return mapErr(err)
	}
	defer rows.Close()
	for rows.Next() {
		var productID, k, v string
		if err := rows.Scan(&productID, &k, &v); err != nil {
			return err
		}
		index[productID].Attributes[k] = v
	}
	return mapErr(rows.Err())
}
