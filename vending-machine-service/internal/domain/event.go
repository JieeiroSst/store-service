package domain

import "time"

const (
	EventMachineStatusChanged = "machine.status_changed"
	EventMaintenancePerformed = "machine.maintenance_performed"
	EventInventoryRestocked   = "inventory.restocked"
	EventInventoryLow         = "inventory.low"
	EventInventoryEmpty       = "inventory.empty"
	EventSessionStarted       = "session.started"
	EventSessionEnded         = "session.ended"
	EventReservationCreated   = "reservation.created"
	EventReservationReleased  = "reservation.released"
	EventPaymentCompleted     = "payment.completed"
	EventPaymentFailed        = "payment.failed"
	EventPaymentRefunded      = "payment.refunded"
	EventRefundFailed         = "payment.refund_failed"
	EventCouponRedeemFailed   = "coupon.redeem_failed"
	EventOrderCompleted       = "order.completed"
	EventOrderFailed          = "order.failed"
	EventOrderRefunded        = "order.refunded"
)

var alertEvents = map[string]bool{
	EventMachineStatusChanged: true,
	EventInventoryLow:         true,
	EventInventoryEmpty:       true,
	EventOrderFailed:          true,
	EventRefundFailed:         true,
	EventCouponRedeemFailed:   true,
}

func (e *Event) IsAlert() bool { return alertEvents[e.EventType] }

type Event struct {
	ID            string
	EventType     string
	RelatedEntity string
	EntityID      string
	MachineID     string
	Data          map[string]any
	OccurredAt    time.Time
}
