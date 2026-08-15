package order

// Status models the order lifecycle:
//
//	PENDING -> AWAITING_PAYMENT -> PAID -> FULFILLING -> COMPLETED
//	                            -> CANCELLED
//	AWAITING_PAYMENT -> FAILED
type Status string

const (
	StatusPending          Status = "pending"
	StatusAwaitingPayment  Status = "awaiting_payment"
	StatusPaid             Status = "paid"
	StatusFulfilling       Status = "fulfilling"
	StatusCompleted        Status = "completed"
	StatusCancelled        Status = "cancelled"
	StatusFailed           Status = "failed"
)

var transitions = map[Status][]Status{
	StatusPending:         {StatusAwaitingPayment, StatusCancelled},
	StatusAwaitingPayment: {StatusPaid, StatusFailed, StatusCancelled},
	StatusPaid:            {StatusFulfilling, StatusCancelled},
	StatusFulfilling:      {StatusCompleted, StatusFailed},
	StatusCompleted:       {},
	StatusCancelled:       {},
	StatusFailed:          {},
}

func (s Status) canTransitionTo(target Status) bool {
	for _, allowed := range transitions[s] {
		if allowed == target {
			return true
		}
	}
	return false
}
