package shopify

import (
	"strings"
	"time"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
)

type pageInfo struct {
	HasNextPage bool   `json:"hasNextPage"`
	EndCursor   string `json:"endCursor"`
}

type productNode struct {
	ID          string            `json:"id"`
	Title       string            `json:"title"`
	Handle      string            `json:"handle"`
	Status      string            `json:"status"`
	Vendor      string            `json:"vendor"`
	ProductType string            `json:"productType"`
	Tags        []string          `json:"tags"`
	CreatedAt   time.Time         `json:"createdAt"`
	UpdatedAt   time.Time         `json:"updatedAt"`
	Variants    variantConnection `json:"variants"`
}

type variantNode struct {
	ID                string  `json:"id"`
	Title             string  `json:"title"`
	SKU               *string `json:"sku"`
	Price             string  `json:"price"`
	InventoryQuantity *int    `json:"inventoryQuantity"`
}

type variantConnection struct {
	PageInfo pageInfo      `json:"pageInfo"`
	Nodes    []variantNode `json:"nodes"`
}

type money struct {
	ShopMoney struct {
		Amount string `json:"amount"`
	} `json:"shopMoney"`
}

type orderNode struct {
	ID                       string             `json:"id"`
	Name                     string             `json:"name"`
	Email                    *string            `json:"email"`
	DisplayFinancialStatus   *string            `json:"displayFinancialStatus"`
	DisplayFulfillmentStatus string             `json:"displayFulfillmentStatus"`
	CurrencyCode             string             `json:"currencyCode"`
	TotalPriceSet            money              `json:"totalPriceSet"`
	CancelledAt              *time.Time         `json:"cancelledAt"`
	ProcessedAt              time.Time          `json:"processedAt"`
	CreatedAt                time.Time          `json:"createdAt"`
	UpdatedAt                time.Time          `json:"updatedAt"`
	LineItems                lineItemConnection `json:"lineItems"`
}

type lineItemNode struct {
	ID       string  `json:"id"`
	Title    string  `json:"title"`
	SKU      *string `json:"sku"`
	Quantity int     `json:"quantity"`
	Variant  *struct {
		ID string `json:"id"`
	} `json:"variant"`
	OriginalUnitPriceSet money `json:"originalUnitPriceSet"`
}

type lineItemConnection struct {
	PageInfo pageInfo       `json:"pageInfo"`
	Nodes    []lineItemNode `json:"nodes"`
}

func (n productNode) toModel() *model.Product {
	product := &model.Product{
		ShopifyID:        n.ID,
		Title:            n.Title,
		Handle:           n.Handle,
		Status:           model.ProductStatus(n.Status),
		Vendor:           n.Vendor,
		ProductType:      n.ProductType,
		Tags:             strings.Join(n.Tags, ", "),
		ShopifyCreatedAt: n.CreatedAt,
		ShopifyUpdatedAt: n.UpdatedAt,
	}
	for _, v := range n.Variants.Nodes {
		product.Variants = append(product.Variants, model.ProductVariant{
			ShopifyID:         v.ID,
			Title:             v.Title,
			SKU:               deref(v.SKU),
			Price:             decimalOrZero(v.Price),
			InventoryQuantity: derefInt(v.InventoryQuantity),
		})
	}
	return product
}

func (n orderNode) toModel() *model.Order {
	order := &model.Order{
		ShopifyID:         n.ID,
		Name:              n.Name,
		Email:             deref(n.Email),
		FinancialStatus:   deref(n.DisplayFinancialStatus),
		FulfillmentStatus: n.DisplayFulfillmentStatus,
		Currency:          n.CurrencyCode,
		TotalPrice:        decimalOrZero(n.TotalPriceSet.ShopMoney.Amount),
		CancelledAt:       n.CancelledAt,
		ProcessedAt:       n.ProcessedAt,
		ShopifyCreatedAt:  n.CreatedAt,
		ShopifyUpdatedAt:  n.UpdatedAt,
	}
	for _, li := range n.LineItems.Nodes {
		item := model.LineItem{
			ShopifyID: li.ID,
			Title:     li.Title,
			SKU:       deref(li.SKU),
			Quantity:  li.Quantity,
			UnitPrice: decimalOrZero(li.OriginalUnitPriceSet.ShopMoney.Amount),
		}
		if li.Variant != nil {
			item.VariantShopifyID = li.Variant.ID
		}
		order.LineItems = append(order.LineItems, item)
	}
	return order
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func derefInt(i *int) int {
	if i == nil {
		return 0
	}
	return *i
}

func decimalOrZero(s string) string {
	if s == "" {
		return "0"
	}
	return s
}
