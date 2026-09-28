package application

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/JIeeiroSst/shopify-service/internal/domain/model"
	"github.com/JIeeiroSst/shopify-service/internal/domain/port"
)

const (
	headerHmac      = "X-Shopify-Hmac-Sha256"
	headerTopic     = "X-Shopify-Topic"
	headerWebhookID = "X-Shopify-Webhook-Id"
	headerShop      = "X-Shopify-Shop-Domain"
)

type webhookService struct {
	verifier port.WebhookVerifier
	events   port.WebhookEventRepository
	products port.ProductRepository
	orders   port.OrderRepository
	shopify  port.ShopifyClient
	now      func() time.Time
}

func NewWebhookService(
	verifier port.WebhookVerifier,
	events port.WebhookEventRepository,
	products port.ProductRepository,
	orders port.OrderRepository,
	shopify port.ShopifyClient,
) port.WebhookUsecase {
	return &webhookService{
		verifier: verifier,
		events:   events,
		products: products,
		orders:   orders,
		shopify:  shopify,
		now:      time.Now,
	}
}

type webhookPayload struct {
	ID                json.Number `json:"id"`
	AdminGraphqlAPIID string      `json:"admin_graphql_api_id"`
}

func (s *webhookService) HandleWebhook(ctx context.Context, payload []byte, headers http.Header) error {
	if !s.verifier.Verify(payload, headers.Get(headerHmac)) {
		return port.ErrInvalidWebhook
	}

	topic := headers.Get(headerTopic)
	webhookID := headers.Get(headerWebhookID)
	if topic == "" || webhookID == "" {
		return fmt.Errorf("%w: missing topic or webhook id header", port.ErrInvalidWebhook)
	}

	seen, err := s.events.Exists(ctx, webhookID)
	if err != nil {
		return err
	}
	if seen {
		return nil
	}

	var body webhookPayload
	dec := json.NewDecoder(bytes.NewReader(payload))
	dec.UseNumber()
	if err := dec.Decode(&body); err != nil {
		return fmt.Errorf("%w: %v", port.ErrInvalidWebhook, err)
	}

	if err := s.dispatch(ctx, topic, body); err != nil {
		return err
	}

	return s.events.Create(ctx, &model.WebhookEvent{
		WebhookID:   webhookID,
		Topic:       topic,
		ShopDomain:  headers.Get(headerShop),
		ProcessedAt: s.now(),
	})
}

func (s *webhookService) dispatch(ctx context.Context, topic string, body webhookPayload) error {
	switch topic {
	case model.TopicProductsCreate, model.TopicProductsUpdate:
		return s.refreshProduct(ctx, resourceGID("Product", body))
	case model.TopicProductsDelete:
		return s.products.DeleteByShopifyID(ctx, resourceGID("Product", body))
	case model.TopicOrdersCreate, model.TopicOrdersUpdated, model.TopicOrdersPaid,
		model.TopicOrdersCancelled, model.TopicOrdersFulfilled:
		return s.refreshOrder(ctx, resourceGID("Order", body))
	default:
		// Subscribed but unhandled topics are acknowledged so Shopify stops retrying.
		return nil
	}
}

func (s *webhookService) refreshProduct(ctx context.Context, gid string) error {
	product, err := s.shopify.GetProduct(ctx, gid)
	if errors.Is(err, port.ErrNotFound) {
		return s.products.DeleteByShopifyID(ctx, gid)
	}
	if err != nil {
		return err
	}
	_, err = s.products.Upsert(ctx, product)
	return err
}

func (s *webhookService) refreshOrder(ctx context.Context, gid string) error {
	order, err := s.shopify.GetOrder(ctx, gid)
	if errors.Is(err, port.ErrNotFound) {
		return nil
	}
	if err != nil {
		return err
	}
	_, err = s.orders.Upsert(ctx, order)
	return err
}

func resourceGID(resource string, body webhookPayload) string {
	if body.AdminGraphqlAPIID != "" {
		return body.AdminGraphqlAPIID
	}
	return fmt.Sprintf("gid://shopify/%s/%s", resource, body.ID.String())
}
