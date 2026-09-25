package domain

import (
	"fmt"
	"time"
)

// RentalModel is how a homestay is rented. Each model has its own rate set by the manager.
type RentalModel string

const (
	ModelDay   RentalModel = "day"
	ModelWeek  RentalModel = "week"
	ModelMonth RentalModel = "month"
	ModelYear  RentalModel = "year"
)

func (m RentalModel) Valid() bool {
	switch m {
	case ModelDay, ModelWeek, ModelMonth, ModelYear:
		return true
	}
	return false
}

// IsLease reports whether the model is a long-term lease billed period by period.
func (m RentalModel) IsLease() bool { return m == ModelWeek || m == ModelMonth || m == ModelYear }

// MaxPeriods bounds how long one lease can run.
func (m RentalModel) MaxPeriods() int {
	switch m {
	case ModelWeek:
		return 104 // two years
	case ModelMonth:
		return 120 // ten years
	case ModelYear:
		return 10
	}
	return 0
}

// Rate is the manager-set price of one rental model of a homestay. Price is per
// night, week, month or year, as a decimal string.
type Rate struct {
	HomestayID int64
	Model      RentalModel
	Price      string
	Currency   string
	MinPeriods int
	Active     bool
}

type InvoiceStatus int16

const (
	InvoiceUnpaid InvoiceStatus = 1
	InvoicePaid   InvoiceStatus = 2
	InvoiceVoid   InvoiceStatus = 3
	// InvoiceRefunded: paid, then the lease ended before this period began, so it is given back.
	InvoiceRefunded InvoiceStatus = 4
)

type LeaseStatus int16

const (
	LeasePending   LeaseStatus = 1 // nights held, waiting for the first invoice
	LeaseActive    LeaseStatus = 2
	LeaseEnded     LeaseStatus = 3
	LeaseCancelled LeaseStatus = 4
)

type Invoice struct {
	ID            int64
	LeaseID       int64
	PeriodNo      int
	PeriodStart   time.Time
	PeriodEnd     time.Time // exclusive
	DueDate       time.Time
	Amount        string
	Status        InvoiceStatus
	PaymentMethod PaymentMethod
	PaymentRef    string
	PaidAt        *time.Time
	Version       int64

	// Filled when invoices are listed across leases.
	UserID     int64
	HomestayID int64
}

// Overdue reports whether the invoice is unpaid past its due date.
func (i Invoice) Overdue(today time.Time) bool {
	return i.Status == InvoiceUnpaid && i.DueDate.Before(today)
}

type Lease struct {
	ID         int64
	UserID     int64
	HomestayID int64
	Model      RentalModel
	StartDate  time.Time
	EndDate    time.Time // exclusive
	Periods    int
	BillingDay int
	Currency   string
	Rent       string
	Status     LeaseStatus
	Note       string
	RequestID  string
	ExpiresAt  *time.Time
	CreatedAt  time.Time
	Invoices   []Invoice
}

// InvoicePlan is one row of a lease's billing schedule before it is stored.
type InvoicePlan struct {
	PeriodNo    int
	PeriodStart time.Time
	PeriodEnd   time.Time
	DueDate     time.Time
}

// addMonths adds n months to start, clamping to the last day of a shorter month
// (Jan 31 + 1 month = Feb 28) without drifting: always call it on the original start.
func addMonths(start time.Time, n int) time.Time {
	y, m, d := start.Date()
	first := time.Date(y, m+time.Month(n), 1, 0, 0, 0, 0, time.UTC)
	last := time.Date(first.Year(), first.Month()+1, 0, 0, 0, 0, 0, time.UTC).Day()
	if d > last {
		d = last
	}
	return time.Date(first.Year(), first.Month(), d, 0, 0, 0, 0, time.UTC)
}

func isoWeekday(t time.Time) int {
	if w := int(t.Weekday()); w != 0 {
		return w
	}
	return 7
}

// DefaultBillingDay is used when the manager does not choose a collection day.
func DefaultBillingDay(model RentalModel, start time.Time) int {
	switch model {
	case ModelMonth:
		return min(start.Day(), 28)
	case ModelWeek:
		return isoWeekday(start)
	}
	return 0
}

// BuildSchedule lays out the invoices of a lease.
//
// Periods run back to back from start (weeks are 7 days, months and years are calendar
// months from the start date). The first invoice is due on the start date. After that:
//   - month: due on billingDay (1-28) of the month the period starts in
//   - week:  due on the first billingDay weekday (1=Mon..7=Sun) on or after the period start
//   - year:  due on the period start
//
// It returns the plans and the exclusive end date of the lease.
func BuildSchedule(model RentalModel, start time.Time, periods, billingDay int) ([]InvoicePlan, time.Time, error) {
	if !model.IsLease() {
		return nil, time.Time{}, fmt.Errorf("%w: %q is not a lease model", ErrInvalid, model)
	}
	if periods < 1 || periods > model.MaxPeriods() {
		return nil, time.Time{}, fmt.Errorf("%w: periods must be between 1 and %d for %s rentals", ErrInvalid, model.MaxPeriods(), model)
	}
	switch model {
	case ModelMonth:
		if billingDay < 1 || billingDay > 28 {
			return nil, time.Time{}, fmt.Errorf("%w: billing day must be between 1 and 28", ErrInvalid)
		}
	case ModelWeek:
		if billingDay < 1 || billingDay > 7 {
			return nil, time.Time{}, fmt.Errorf("%w: billing day must be a weekday from 1 (Mon) to 7 (Sun)", ErrInvalid)
		}
	}

	at := func(k int) time.Time {
		switch model {
		case ModelWeek:
			return start.AddDate(0, 0, 7*k)
		case ModelMonth:
			return addMonths(start, k)
		default:
			return addMonths(start, 12*k)
		}
	}
	plans := make([]InvoicePlan, periods)
	for k := range plans {
		ps, pe := at(k), at(k+1)
		due := ps
		if k > 0 {
			switch model {
			case ModelMonth:
				due = time.Date(ps.Year(), ps.Month(), billingDay, 0, 0, 0, 0, time.UTC)
			case ModelWeek:
				due = ps.AddDate(0, 0, (billingDay-isoWeekday(ps)+7)%7)
			}
		}
		plans[k] = InvoicePlan{PeriodNo: k, PeriodStart: ps, PeriodEnd: pe, DueDate: due}
	}
	return plans, at(periods), nil
}
