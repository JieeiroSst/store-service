package model

type Status string

const (
	StatusPending        Status = "PENDING"
	StatusRejected       Status = "REJECTED"
	StatusCreated        Status = "CREATED"
	StatusPicking        Status = "PICKING"
	StatusPicked         Status = "PICKED"
	StatusInTransit      Status = "IN_TRANSIT"
	StatusDelivering     Status = "DELIVERING"
	StatusDelivered      Status = "DELIVERED"
	StatusDeliveryFailed Status = "DELIVERY_FAILED"
	StatusReturning      Status = "RETURNING"
	StatusReturned       Status = "RETURNED"
	StatusCancelled      Status = "CANCELLED"
	StatusException      Status = "EXCEPTION"
)

var transitions = map[Status][]Status{
	StatusPending:        {StatusCreated, StatusRejected, StatusCancelled},
	StatusRejected:       {StatusPending, StatusCreated, StatusCancelled},
	StatusCreated:        {StatusPicking, StatusPicked, StatusInTransit, StatusDelivering, StatusCancelled, StatusException},
	StatusPicking:        {StatusPicked, StatusInTransit, StatusCancelled, StatusException},
	StatusPicked:         {StatusInTransit, StatusDelivering, StatusReturning, StatusException},
	StatusInTransit:      {StatusDelivering, StatusDelivered, StatusDeliveryFailed, StatusReturning, StatusException},
	StatusDelivering:     {StatusDelivered, StatusDeliveryFailed, StatusInTransit, StatusException},
	StatusDeliveryFailed: {StatusDelivering, StatusInTransit, StatusReturning, StatusException},
	StatusReturning:      {StatusReturned, StatusException},
	StatusException:      {StatusInTransit, StatusDelivering, StatusDelivered, StatusReturning, StatusReturned, StatusCancelled},
}

func (s Status) Terminal() bool {
	return s == StatusDelivered || s == StatusReturned || s == StatusCancelled
}

func CanTransition(from, to Status) bool {
	for _, next := range transitions[from] {
		if next == to {
			return true
		}
	}
	return false
}

func (s Status) CancellableAtCarrier() bool {
	return s == StatusCreated || s == StatusPicking
}

func (s Status) CancellableLocally() bool {
	return s == StatusPending || s == StatusRejected
}

var ghnStatuses = map[string]Status{
	"ready_to_pick":            StatusCreated,
	"picking":                  StatusPicking,
	"money_collect_picking":    StatusPicking,
	"picked":                   StatusPicked,
	"storing":                  StatusInTransit,
	"transporting":             StatusInTransit,
	"sorting":                  StatusInTransit,
	"delivering":               StatusDelivering,
	"money_collect_delivering": StatusDelivering,
	"delivered":                StatusDelivered,
	"delivery_fail":            StatusDeliveryFailed,
	"waiting_to_return":        StatusDeliveryFailed,
	"return":                   StatusReturning,
	"return_transporting":      StatusReturning,
	"return_sorting":           StatusReturning,
	"returning":                StatusReturning,
	"return_fail":              StatusReturning,
	"returned":                 StatusReturned,
	"cancel":                   StatusCancelled,
	"exception":                StatusException,
	"damage":                   StatusException,
	"lost":                     StatusException,
}

func MapGHNStatus(carrierStatus string) (Status, bool) {
	s, ok := ghnStatuses[carrierStatus]
	return s, ok
}
