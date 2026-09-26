package temporalx

import (
	"fmt"
	"time"
)

const (
	WorkflowOrder      = "OrderLifecycle"
	SignalOrderChanged = "order-changed"

	ActivityAdvance   = "AdvanceOrder"
	ActivityDocuments = "GenerateOrderDocuments"
	ActivityRemind    = "RemindOrder"
)

func WorkflowID(queue string, orderID int64) string {
	return fmt.Sprintf("%s/order-%d", queue, orderID)
}

type OrderInput struct {
	OrderID int64
}

type Progress struct {
	Status         string
	ExpiresAt      time.Time
	RetryAfter     time.Duration
	EventStartsAt  time.Time
	EventCancelled bool
}
