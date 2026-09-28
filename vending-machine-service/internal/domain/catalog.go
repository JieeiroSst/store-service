package domain

import (
	"fmt"
	"strings"
	"time"
)

type Category struct {
	ID           string
	Name         string
	Description  string
	DisplayOrder int
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

func (c *Category) Validate() error {
	if strings.TrimSpace(c.Name) == "" {
		return fmt.Errorf("%w: name is required", ErrInvalidInput)
	}
	return nil
}

type Product struct {
	ID          string
	Name        string
	Description string
	PriceCents  int
	CategoryID  string
	ImageURL    string
	Barcode     string
	IsActive    bool
	Attributes  map[string]string
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

func (p *Product) Validate() error {
	if strings.TrimSpace(p.Name) == "" || p.CategoryID == "" {
		return fmt.Errorf("%w: name and category_id are required", ErrInvalidInput)
	}
	if p.PriceCents < 0 {
		return fmt.Errorf("%w: price_cents must not be negative", ErrInvalidInput)
	}
	return nil
}
