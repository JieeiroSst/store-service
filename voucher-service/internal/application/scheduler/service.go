package scheduler

import (
	"context"

	platformscheduler "github.com/JIeeiroSst/voucher-service/internal/platform/scheduler"
	"go.uber.org/zap"
)

type Service struct {
	cron          *platformscheduler.Scheduler
	voucherExpirer     VoucherExpirer
	reconciliation     ReconciliationTrigger
	lowStock           LowStockAlerter
	lowStockThreshold  int
	log                *zap.Logger
}

func NewService(cron *platformscheduler.Scheduler, voucherExpirer VoucherExpirer, reconciliation ReconciliationTrigger, lowStock LowStockAlerter, log *zap.Logger) SchedulerService {
	return &Service{
		cron:              cron,
		voucherExpirer:    voucherExpirer,
		reconciliation:    reconciliation,
		lowStock:          lowStock,
		lowStockThreshold: 20,
		log:               log,
	}
}

func (s *Service) RegisterJobs(ctx context.Context) error {
	if err := s.cron.AddFunc("*/5 * * * *", func() {
		expired, err := s.voucherExpirer.ExpireDueVouchers(context.Background())
		if err != nil {
			s.log.Error("scheduled voucher expiry sweep failed", zap.Error(err))
			return
		}
		s.log.Info("voucher expiry sweep completed", zap.Int("expired", expired))
	}); err != nil {
		return err
	}

	if err := s.cron.AddFunc("0 * * * *", func() {
		if err := s.reconciliation.RunPaymentReconciliation(context.Background()); err != nil {
			s.log.Error("scheduled reconciliation run failed", zap.Error(err))
		}
	}); err != nil {
		return err
	}

	if err := s.cron.AddFunc("0 */6 * * *", func() {
		if err := s.lowStock.CheckAndAlertLowStock(context.Background(), s.lowStockThreshold); err != nil {
			s.log.Error("scheduled low-stock check failed", zap.Error(err))
		}
	}); err != nil {
		return err
	}

	return nil
}
