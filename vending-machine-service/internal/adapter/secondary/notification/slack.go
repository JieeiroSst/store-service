package notification

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"

	"github.com/JIeeiroSst/vending-machine-service/internal/adapter/secondary/httpx"
	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
)

const priorityHigh = 1

var titles = map[string]string{
	domain.EventMachineStatusChanged: "Vending machine status changed",
	domain.EventInventoryLow:         "Vending slot running low",
	domain.EventInventoryEmpty:       "Vending slot is empty",
	domain.EventOrderFailed:          "Vending machine failed to dispense",
	domain.EventRefundFailed:         "Vending refund failed",
	domain.EventCouponRedeemFailed:   "Vending coupon redemption failed",
}

func title(e domain.Event) string {
	if t, ok := titles[e.EventType]; ok {
		return t
	}
	return "Vending event " + e.EventType
}

func text(e domain.Event) string {
	var b strings.Builder
	fmt.Fprintf(&b, "*machine*: %s\n*%s*: %s\n", e.MachineID, e.RelatedEntity, e.EntityID)
	keys := make([]string, 0, len(e.Data))
	for k := range e.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(&b, "*%s*: %v\n", k, e.Data[k])
	}
	fmt.Fprintf(&b, "_at %s_", e.OccurredAt.Format("2006-01-02 15:04:05 MST"))
	return b.String()
}

type Slack struct {
	client *httpx.Client
}

func NewSlack(client *httpx.Client) *Slack { return &Slack{client: client} }

func (s *Slack) Notify(ctx context.Context, e domain.Event) error {
	return s.client.Do(ctx, http.MethodPost, "/api/v1/notifications/slack", map[string]any{
		"title":    title(e),
		"text":     text(e),
		"priority": priorityHigh,
		"raw_data": map[string]any{
			"event_id":   e.ID,
			"event_type": e.EventType,
			"machine_id": e.MachineID,
			"entity_id":  e.EntityID,
		},
	}, nil)
}

type Log struct{}

func (Log) Notify(_ context.Context, e domain.Event) error {
	log.Printf("alert %s: %s", title(e), strings.ReplaceAll(text(e), "\n", " "))
	return nil
}
