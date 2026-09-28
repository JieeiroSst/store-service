package application

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
)

func TestCreateProductValidatesInput(t *testing.T) {
	svc := NewProductService(newFakeProductRepo(), newFakeShopify())

	_, err := svc.CreateProduct(context.Background(), port.CreateProductInput{Title: "  "})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("empty title: got %v, want ErrInvalidInput", err)
	}

	_, err = svc.CreateProduct(context.Background(), port.CreateProductInput{Title: "Tee", Status: "PUBLISHED"})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("bad status: got %v, want ErrInvalidInput", err)
	}
}

func TestCreateProductStoresShopifyResult(t *testing.T) {
	repo := newFakeProductRepo()
	shop := newFakeShopify()
	svc := NewProductService(repo, shop)

	product, err := svc.CreateProduct(context.Background(), port.CreateProductInput{Title: " Tee ", Status: model.ProductStatusDraft})
	if err != nil {
		t.Fatal(err)
	}
	if shop.created.Title != "Tee" {
		t.Errorf("title sent to Shopify = %q, want trimmed %q", shop.created.Title, "Tee")
	}
	if product.ID == 0 || repo.byShopifyID["gid://shopify/Product/99"] == nil {
		t.Errorf("created product was not persisted: %+v", product)
	}
}

func TestSyncProductsFollowsCursorAndUpserts(t *testing.T) {
	repo := newFakeProductRepo()
	shop := newFakeShopify()
	shop.productPages = []port.Page[model.Product]{
		{Items: []model.Product{{ShopifyID: "gid://shopify/Product/1"}, {ShopifyID: "gid://shopify/Product/2"}}, HasNextPage: true, EndCursor: "1"},
		{Items: []model.Product{{ShopifyID: "gid://shopify/Product/2", Title: "renamed"}, {ShopifyID: "gid://shopify/Product/3"}}},
	}

	result, err := NewProductService(repo, shop).SyncProducts(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Synced != 4 {
		t.Errorf("synced = %d, want 4", result.Synced)
	}
	if len(repo.byShopifyID) != 3 {
		t.Errorf("stored %d products, want 3 (upsert by shopify id)", len(repo.byShopifyID))
	}
	if repo.byShopifyID["gid://shopify/Product/2"].Title != "renamed" {
		t.Error("later page did not update existing product")
	}
}

func TestSyncOrdersFollowsCursor(t *testing.T) {
	repo := newFakeOrderRepo()
	shop := newFakeShopify()
	shop.orderPages = []port.Page[model.Order]{
		{Items: []model.Order{{ShopifyID: "gid://shopify/Order/1"}}, HasNextPage: true, EndCursor: "1"},
		{Items: []model.Order{{ShopifyID: "gid://shopify/Order/2"}}},
	}

	result, err := NewOrderService(repo, shop).SyncOrders(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if result.Synced != 2 || len(repo.byShopifyID) != 2 {
		t.Errorf("synced=%d stored=%d, want 2/2", result.Synced, len(repo.byShopifyID))
	}
}

type webhookFixture struct {
	svc      port.WebhookUsecase
	products *fakeProductRepo
	orders   *fakeOrderRepo
	events   *fakeEventRepo
	shop     *fakeShopify
}

func newWebhookFixture(verified bool) *webhookFixture {
	f := &webhookFixture{
		products: newFakeProductRepo(),
		orders:   newFakeOrderRepo(),
		events:   newFakeEventRepo(),
		shop:     newFakeShopify(),
	}
	f.svc = NewWebhookService(fakeVerifier{ok: verified}, f.events, f.products, f.orders, f.shop)
	return f
}

func webhookHeaders(topic, id string) http.Header {
	h := http.Header{}
	h.Set("X-Shopify-Topic", topic)
	h.Set("X-Shopify-Webhook-Id", id)
	h.Set("X-Shopify-Hmac-Sha256", "sig")
	return h
}

func TestWebhookRejectsBadSignature(t *testing.T) {
	f := newWebhookFixture(false)
	err := f.svc.HandleWebhook(context.Background(), []byte(`{}`), webhookHeaders(model.TopicProductsCreate, "w1"))
	if !errors.Is(err, port.ErrInvalidWebhook) {
		t.Fatalf("got %v, want ErrInvalidWebhook", err)
	}
}

func TestWebhookProductUpdateRefetchesAndIsIdempotent(t *testing.T) {
	f := newWebhookFixture(true)
	gid := "gid://shopify/Product/632910392"
	f.shop.products[gid] = &model.Product{ShopifyID: gid, Title: "IPod Nano"}
	payload := []byte(`{"id":632910392,"admin_graphql_api_id":"` + gid + `","title":"stale"}`)

	for i := 0; i < 2; i++ {
		if err := f.svc.HandleWebhook(context.Background(), payload, webhookHeaders(model.TopicProductsUpdate, "w1")); err != nil {
			t.Fatal(err)
		}
	}
	if got := f.products.byShopifyID[gid]; got == nil || got.Title != "IPod Nano" {
		t.Fatalf("product not stored from Shopify refetch: %+v", got)
	}
	if f.shop.getCalls != 1 {
		t.Errorf("duplicate delivery refetched: getCalls = %d, want 1", f.shop.getCalls)
	}
}

func TestWebhookProductDeleteUsesNumericID(t *testing.T) {
	f := newWebhookFixture(true)
	gid := "gid://shopify/Product/9007199254740993" // beyond float64 precision
	f.products.byShopifyID[gid] = &model.Product{ID: 1, ShopifyID: gid}

	err := f.svc.HandleWebhook(context.Background(), []byte(`{"id":9007199254740993}`), webhookHeaders(model.TopicProductsDelete, "w2"))
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := f.products.byShopifyID[gid]; ok {
		t.Error("product was not deleted")
	}
}

func TestWebhookOrderCreateStoresOrder(t *testing.T) {
	f := newWebhookFixture(true)
	gid := "gid://shopify/Order/820982911946154508"
	f.shop.orders[gid] = &model.Order{ShopifyID: gid, Name: "#1001"}

	err := f.svc.HandleWebhook(context.Background(), []byte(`{"id":820982911946154508,"admin_graphql_api_id":"`+gid+`"}`), webhookHeaders(model.TopicOrdersCreate, "w3"))
	if err != nil {
		t.Fatal(err)
	}
	if f.orders.byShopifyID[gid] == nil {
		t.Error("order not stored")
	}
}

func TestWebhookUnknownTopicIsAcknowledged(t *testing.T) {
	f := newWebhookFixture(true)
	if err := f.svc.HandleWebhook(context.Background(), []byte(`{}`), webhookHeaders("app/uninstalled", "w4")); err != nil {
		t.Fatal(err)
	}
	if !f.events.seen["w4"] {
		t.Error("unknown topic delivery not recorded")
	}
}
