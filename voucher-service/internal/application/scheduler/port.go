package scheduler

import "context"

type SchedulerService interface {
	RegisterJobs(ctx context.Context) error
}

type VoucherExpirer interface {
	ExpireDueVouchers(ctx context.Context) (expired int, err error)
}

type ReconciliationTrigger interface {
	RunPaymentReconciliation(ctx context.Context) error
}

type LowStockAlerter interface {
	CheckAndAlertLowStock(ctx context.Context, threshold int) error
}
