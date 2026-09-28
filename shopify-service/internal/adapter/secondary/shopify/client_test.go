package shopify

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/JIeeiroSst/shopify-service/config"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
)

func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	c := newClient(config.ShopifyConfig{BaseURL: srv.URL, AccessToken: "shpat_test", APIVersion: "2026-07", Timeout: "5s"})
	c.sleep = func(context.Context, time.Duration) error { return nil }
	return c
}

func decodeRequest(t *testing.T, r *http.Request) graphqlRequest {
	t.Helper()
	var req graphqlRequest
	raw, _ := io.ReadAll(r.Body)
	if err := json.Unmarshal(raw, &req); err != nil {
		t.Fatalf("bad request body: %v", err)
	}
	return req
}

func TestListProductsSendsAuthAndMapsNodes(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/admin/api/2026-07/graphql.json" {
			t.Errorf("path = %s", r.URL.Path)
		}
		if r.Header.Get("X-Shopify-Access-Token") != "shpat_test" {
			t.Error("missing access token header")
		}
		req := decodeRequest(t, r)
		if req.Variables["after"] != "cursor-1" {
			t.Errorf("after = %v, want cursor-1", req.Variables["after"])
		}
		io.WriteString(w, `{"data":{"products":{"pageInfo":{"hasNextPage":true,"endCursor":"cursor-2"},"nodes":[
			{"id":"gid://shopify/Product/1","title":"Tee","status":"ACTIVE","tags":["a","b"],
			 "createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:05Z",
			 "variants":{"nodes":[{"id":"gid://shopify/ProductVariant/11","title":"S","sku":null,"price":"19.99","inventoryQuantity":4}]}}]}}}`)
	})

	page, err := c.ListProducts(context.Background(), "cursor-1")
	if err != nil {
		t.Fatal(err)
	}
	if !page.HasNextPage || page.EndCursor != "cursor-2" || len(page.Items) != 1 {
		t.Fatalf("unexpected page: %+v", page)
	}
	p := page.Items[0]
	if p.Tags != "a, b" || p.Variants[0].Price != "19.99" || p.Variants[0].SKU != "" || p.Variants[0].InventoryQuantity != 4 {
		t.Errorf("bad mapping: %+v", p)
	}
}

func TestGetProductNullIsNotFound(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"data":{"product":null}}`)
	})
	_, err := c.GetProduct(context.Background(), "gid://shopify/Product/404")
	if !errors.Is(err, port.ErrNotFound) {
		t.Fatalf("got %v, want ErrNotFound", err)
	}
}

func TestThrottledRequestIsRetried(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		if calls == 1 {
			io.WriteString(w, `{"errors":[{"message":"Throttled","extensions":{"code":"THROTTLED"}}]}`)
			return
		}
		io.WriteString(w, `{"data":{"order":{"id":"gid://shopify/Order/1","name":"#1001","totalPriceSet":{"shopMoney":{"amount":"10.00"}},
			"processedAt":"2026-01-02T03:04:05Z","createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:05Z",
			"lineItems":{"nodes":[{"id":"gid://shopify/LineItem/5","title":"Tee","quantity":2,"variant":null,"originalUnitPriceSet":{"shopMoney":{"amount":"5.00"}}}]}}}}`)
	})

	order, err := c.GetOrder(context.Background(), "gid://shopify/Order/1")
	if err != nil {
		t.Fatal(err)
	}
	if calls != 2 {
		t.Errorf("calls = %d, want 2", calls)
	}
	if order.TotalPrice != "10.00" || order.LineItems[0].UnitPrice != "5.00" || order.LineItems[0].VariantShopifyID != "" {
		t.Errorf("bad mapping: %+v", order)
	}
}

func TestNonThrottleErrorIsNotRetried(t *testing.T) {
	calls := 0
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		calls++
		w.WriteHeader(http.StatusUnauthorized)
		io.WriteString(w, `{"errors":"[API] Invalid API key or access token"}`)
	})
	_, err := c.GetOrder(context.Background(), "gid://shopify/Order/1")
	if !errors.Is(err, port.ErrShopifyRequestFailed) || calls != 1 {
		t.Fatalf("err=%v calls=%d, want ErrShopifyRequestFailed after 1 call", err, calls)
	}
}

func TestCreateProductUserErrors(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		req := decodeRequest(t, r)
		product := req.Variables["product"].(map[string]any)
		if product["title"] != "Tee" || product["status"] != "DRAFT" {
			t.Errorf("unexpected input: %v", product)
		}
		if _, ok := product["vendor"]; ok {
			t.Error("empty vendor should be omitted")
		}
		io.WriteString(w, `{"data":{"productCreate":{"product":null,"userErrors":[{"field":["title"],"message":"Title is too long"}]}}}`)
	})
	_, err := c.CreateProduct(context.Background(), port.CreateProductInput{Title: "Tee", Status: "DRAFT"})
	if !errors.Is(err, port.ErrShopifyUserError) || !strings.Contains(err.Error(), "title: Title is too long") {
		t.Fatalf("got %v", err)
	}
}

func TestWebhookVerifier(t *testing.T) {
	v := NewWebhookVerifier(&config.Config{Shopify: config.ShopifyConfig{APISecret: "hush"}})
	body := []byte(`{"id":1}`)
	mac := hmac.New(sha256.New, []byte("hush"))
	mac.Write(body)
	sig := base64.StdEncoding.EncodeToString(mac.Sum(nil))

	if !v.Verify(body, sig) {
		t.Error("valid signature rejected")
	}
	if v.Verify([]byte(`{"id":2}`), sig) {
		t.Error("tampered body accepted")
	}
	if v.Verify(body, "") || v.Verify(body, "not-base64!") {
		t.Error("missing/garbage signature accepted")
	}
	empty := NewWebhookVerifier(&config.Config{})
	if empty.Verify(body, sig) {
		t.Error("verifier without secret must reject everything")
	}
}

func TestListProductsFetchesRemainingVariants(t *testing.T) {
	var followUps []map[string]any
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		req := decodeRequest(t, r)
		switch {
		case strings.Contains(req.Query, "query ListProducts"):
			io.WriteString(w, `{"data":{"products":{"pageInfo":{"hasNextPage":false,"endCursor":"p"},"nodes":[
				{"id":"gid://shopify/Product/1","createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:05Z",
				 "variants":{"pageInfo":{"hasNextPage":true,"endCursor":"v1"},"nodes":[{"id":"gid://shopify/ProductVariant/1","price":"1.00"}]}},
				{"id":"gid://shopify/Product/2","createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:05Z",
				 "variants":{"pageInfo":{"hasNextPage":true,"endCursor":"gone"},"nodes":[]}}]}}}`)
		case strings.Contains(req.Query, "query ProductVariants"):
			followUps = append(followUps, req.Variables)
			switch req.Variables["after"] {
			case "v1":
				io.WriteString(w, `{"data":{"product":{"variants":{"pageInfo":{"hasNextPage":true,"endCursor":"v2"},"nodes":[{"id":"gid://shopify/ProductVariant/2","price":"2.00"}]}}}}`)
			case "v2":
				io.WriteString(w, `{"data":{"product":{"variants":{"pageInfo":{"hasNextPage":false,"endCursor":"v3"},"nodes":[{"id":"gid://shopify/ProductVariant/3","price":"3.00"}]}}}}`)
			default: // product 2 was deleted after the list call
				io.WriteString(w, `{"data":{"product":null}}`)
			}
		default:
			t.Errorf("unexpected query: %s", req.Query)
		}
	})

	page, err := c.ListProducts(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if len(page.Items) != 1 {
		t.Fatalf("got %d products, want 1 (deleted product skipped)", len(page.Items))
	}
	if n := len(page.Items[0].Variants); n != 3 {
		t.Errorf("got %d variants, want 3", n)
	}
	if followUps[0]["id"] != "gid://shopify/Product/1" || followUps[0]["first"] != float64(nestedPageSize) {
		t.Errorf("bad follow-up variables: %v", followUps[0])
	}
}

func TestGetOrderFetchesRemainingLineItems(t *testing.T) {
	c := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		req := decodeRequest(t, r)
		if strings.Contains(req.Query, "query OrderLineItems") {
			if req.Variables["after"] != "l1" {
				t.Errorf("after = %v, want l1", req.Variables["after"])
			}
			io.WriteString(w, `{"data":{"order":{"lineItems":{"pageInfo":{"hasNextPage":false},"nodes":[{"id":"gid://shopify/LineItem/2","quantity":1}]}}}}`)
			return
		}
		io.WriteString(w, `{"data":{"order":{"id":"gid://shopify/Order/1","processedAt":"2026-01-02T03:04:05Z","createdAt":"2026-01-02T03:04:05Z","updatedAt":"2026-01-02T03:04:05Z",
			"lineItems":{"pageInfo":{"hasNextPage":true,"endCursor":"l1"},"nodes":[{"id":"gid://shopify/LineItem/1","quantity":1}]}}}}`)
	})

	order, err := c.GetOrder(context.Background(), "gid://shopify/Order/1")
	if err != nil {
		t.Fatal(err)
	}
	if len(order.LineItems) != 2 || order.LineItems[1].ShopifyID != "gid://shopify/LineItem/2" {
		t.Errorf("line items = %+v, want both pages", order.LineItems)
	}
}
