package ports

import (
	"context"

	"github.com/JIeeiroSst/shortlink-service/internal/domain"
)

type WebhookSender interface {
	Send(ctx context.Context, webhook *domain.Webhook, payload domain.WebhookPayload) domain.WebhookDeliveryResult
}

type GeoIPLookup interface {
	Lookup(ip string) domain.GeoLocation
}

type QRCodeGenerator interface {
	PNG(data string, size int, darkColor, lightColor string) ([]byte, error)
	SVG(data string, size int, darkColor, lightColor string) (string, error)
}
