package shopify

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/JIeeiroSst/shopify-service/config"
	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
)

const maxAttempts = 3

type Client struct {
	endpoint    string
	accessToken string
	httpClient  *http.Client
	sleep       func(context.Context, time.Duration) error
}

func NewClient(cfg *config.Config) port.ShopifyClient {
	return newClient(cfg.Shopify)
}

func newClient(cfg config.ShopifyConfig) *Client {
	return &Client{
		endpoint:    cfg.GraphQLEndpoint(),
		accessToken: cfg.AccessToken,
		httpClient:  &http.Client{Timeout: cfg.TimeoutDuration()},
		sleep:       sleepCtx,
	}
}

type graphqlRequest struct {
	Query     string         `json:"query"`
	Variables map[string]any `json:"variables,omitempty"`
}

type graphqlError struct {
	Message    string `json:"message"`
	Extensions struct {
		Code string `json:"code"`
	} `json:"extensions"`
}

type graphqlResponse struct {
	Data   json.RawMessage `json:"data"`
	Errors []graphqlError  `json:"errors"`
}

func (c *Client) do(ctx context.Context, query string, variables map[string]any, out any) error {
	body, err := json.Marshal(graphqlRequest{Query: query, Variables: variables})
	if err != nil {
		return err
	}

	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		throttled, err := c.attempt(ctx, body, out)
		if err == nil {
			return nil
		}
		lastErr = err
		if !throttled || attempt == maxAttempts {
			break
		}
		if err := c.sleep(ctx, time.Duration(attempt)*time.Second); err != nil {
			return err
		}
	}
	return lastErr
}

func (c *Client) attempt(ctx context.Context, body []byte, out any) (throttled bool, err error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.endpoint, bytes.NewReader(body))
	if err != nil {
		return false, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Shopify-Access-Token", c.accessToken)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return false, fmt.Errorf("%w: %v", port.ErrShopifyRequestFailed, err)
	}
	defer resp.Body.Close()

	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return false, fmt.Errorf("%w: read body: %v", port.ErrShopifyRequestFailed, err)
	}
	if resp.StatusCode == http.StatusTooManyRequests {
		return true, fmt.Errorf("%w: rate limited", port.ErrShopifyRequestFailed)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return false, fmt.Errorf("%w: status %d: %s", port.ErrShopifyRequestFailed, resp.StatusCode, truncate(raw))
	}

	var gql graphqlResponse
	if err := json.Unmarshal(raw, &gql); err != nil {
		return false, fmt.Errorf("%w: decode response: %v", port.ErrShopifyRequestFailed, err)
	}
	if len(gql.Errors) > 0 {
		msgs := make([]string, 0, len(gql.Errors))
		for _, e := range gql.Errors {
			if e.Extensions.Code == "THROTTLED" {
				throttled = true
			}
			msgs = append(msgs, e.Message)
		}
		return throttled, fmt.Errorf("%w: %s", port.ErrShopifyRequestFailed, strings.Join(msgs, "; "))
	}
	if err := json.Unmarshal(gql.Data, out); err != nil {
		return false, fmt.Errorf("%w: decode data: %v", port.ErrShopifyRequestFailed, err)
	}
	return false, nil
}

func (c *Client) ListProducts(ctx context.Context, after string) (*port.Page[model.Product], error) {
	var data struct {
		Products struct {
			PageInfo pageInfo      `json:"pageInfo"`
			Nodes    []productNode `json:"nodes"`
		} `json:"products"`
	}
	if err := c.do(ctx, listProductsQuery, pageVariables(after), &data); err != nil {
		return nil, err
	}

	page := &port.Page[model.Product]{
		HasNextPage: data.Products.PageInfo.HasNextPage,
		EndCursor:   data.Products.PageInfo.EndCursor,
	}
	for i := range data.Products.Nodes {
		n := &data.Products.Nodes[i]
		if err := c.completeVariants(ctx, n); err != nil {
			if errors.Is(err, port.ErrNotFound) {
				continue // deleted in Shopify mid-sync; skip rather than store it truncated
			}
			return nil, err
		}
		page.Items = append(page.Items, *n.toModel())
	}
	return page, nil
}

func (c *Client) GetProduct(ctx context.Context, shopifyID string) (*model.Product, error) {
	var data struct {
		Product *productNode `json:"product"`
	}
	if err := c.do(ctx, getProductQuery, map[string]any{"id": shopifyID}, &data); err != nil {
		return nil, err
	}
	if data.Product == nil {
		return nil, port.ErrNotFound
	}
	if err := c.completeVariants(ctx, data.Product); err != nil {
		return nil, err
	}
	return data.Product.toModel(), nil
}

func (c *Client) CreateProduct(ctx context.Context, in port.CreateProductInput) (*model.Product, error) {
	input := map[string]any{"title": in.Title}
	if in.DescriptionHTML != "" {
		input["descriptionHtml"] = in.DescriptionHTML
	}
	if in.Vendor != "" {
		input["vendor"] = in.Vendor
	}
	if in.ProductType != "" {
		input["productType"] = in.ProductType
	}
	if len(in.Tags) > 0 {
		input["tags"] = in.Tags
	}
	if in.Status != "" {
		input["status"] = string(in.Status)
	}

	var data struct {
		ProductCreate struct {
			Product    *productNode `json:"product"`
			UserErrors []struct {
				Field   []string `json:"field"`
				Message string   `json:"message"`
			} `json:"userErrors"`
		} `json:"productCreate"`
	}
	if err := c.do(ctx, createProductMutation, map[string]any{"product": input}, &data); err != nil {
		return nil, err
	}
	if errs := data.ProductCreate.UserErrors; len(errs) > 0 {
		msgs := make([]string, 0, len(errs))
		for _, e := range errs {
			msgs = append(msgs, fmt.Sprintf("%s: %s", strings.Join(e.Field, "."), e.Message))
		}
		return nil, fmt.Errorf("%w: %s", port.ErrShopifyUserError, strings.Join(msgs, "; "))
	}
	if data.ProductCreate.Product == nil {
		return nil, fmt.Errorf("%w: productCreate returned no product", port.ErrShopifyRequestFailed)
	}
	if err := c.completeVariants(ctx, data.ProductCreate.Product); err != nil {
		return nil, err
	}
	return data.ProductCreate.Product.toModel(), nil
}

func (c *Client) ListOrders(ctx context.Context, after string) (*port.Page[model.Order], error) {
	var data struct {
		Orders struct {
			PageInfo pageInfo    `json:"pageInfo"`
			Nodes    []orderNode `json:"nodes"`
		} `json:"orders"`
	}
	if err := c.do(ctx, listOrdersQuery, pageVariables(after), &data); err != nil {
		return nil, err
	}

	page := &port.Page[model.Order]{
		HasNextPage: data.Orders.PageInfo.HasNextPage,
		EndCursor:   data.Orders.PageInfo.EndCursor,
	}
	for i := range data.Orders.Nodes {
		n := &data.Orders.Nodes[i]
		if err := c.completeLineItems(ctx, n); err != nil {
			if errors.Is(err, port.ErrNotFound) {
				continue // deleted in Shopify mid-sync; skip rather than store it truncated
			}
			return nil, err
		}
		page.Items = append(page.Items, *n.toModel())
	}
	return page, nil
}

func (c *Client) GetOrder(ctx context.Context, shopifyID string) (*model.Order, error) {
	var data struct {
		Order *orderNode `json:"order"`
	}
	if err := c.do(ctx, getOrderQuery, map[string]any{"id": shopifyID}, &data); err != nil {
		return nil, err
	}
	if data.Order == nil {
		return nil, port.ErrNotFound
	}
	if err := c.completeLineItems(ctx, data.Order); err != nil {
		return nil, err
	}
	return data.Order.toModel(), nil
}

func (c *Client) completeVariants(ctx context.Context, n *productNode) error {
	for info := n.Variants.PageInfo; info.HasNextPage; {
		var data struct {
			Product *struct {
				Variants variantConnection `json:"variants"`
			} `json:"product"`
		}
		vars := map[string]any{"id": n.ID, "first": nestedPageSize, "after": info.EndCursor}
		if err := c.do(ctx, productVariantsQuery, vars, &data); err != nil {
			return fmt.Errorf("fetch variants of %s: %w", n.ID, err)
		}
		if data.Product == nil {
			return fmt.Errorf("fetch variants of %s: %w", n.ID, port.ErrNotFound)
		}
		n.Variants.Nodes = append(n.Variants.Nodes, data.Product.Variants.Nodes...)
		info = data.Product.Variants.PageInfo
	}
	return nil
}

func (c *Client) completeLineItems(ctx context.Context, n *orderNode) error {
	for info := n.LineItems.PageInfo; info.HasNextPage; {
		var data struct {
			Order *struct {
				LineItems lineItemConnection `json:"lineItems"`
			} `json:"order"`
		}
		vars := map[string]any{"id": n.ID, "first": nestedPageSize, "after": info.EndCursor}
		if err := c.do(ctx, orderLineItemsQuery, vars, &data); err != nil {
			return fmt.Errorf("fetch line items of %s: %w", n.ID, err)
		}
		if data.Order == nil {
			return fmt.Errorf("fetch line items of %s: %w", n.ID, port.ErrNotFound)
		}
		n.LineItems.Nodes = append(n.LineItems.Nodes, data.Order.LineItems.Nodes...)
		info = data.Order.LineItems.PageInfo
	}
	return nil
}

func pageVariables(after string) map[string]any {
	vars := map[string]any{"first": pageSize}
	if after != "" {
		vars["after"] = after
	}
	return vars
}

func truncate(b []byte) string {
	const max = 512
	if len(b) > max {
		return string(b[:max]) + "..."
	}
	return string(b)
}

func sleepCtx(ctx context.Context, d time.Duration) error {
	t := time.NewTimer(d)
	defer t.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-t.C:
		return nil
	}
}
