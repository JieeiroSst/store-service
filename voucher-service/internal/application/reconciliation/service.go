package reconciliation

import (
	"context"
	"time"
)

type Service struct {
	runs    RunRepository
	payments PaymentRecordSource
}

func NewService(runs RunRepository, payments PaymentRecordSource) ReconciliationService {
	return &Service{runs: runs, payments: payments}
}

func (s *Service) RunPaymentReconciliation(ctx context.Context) (*Run, error) {
	now := time.Now().UTC()
	run := &Run{RunType: "payment", Status: RunStatusRunning, StartedAt: &now}
	if err := s.runs.Create(ctx, run); err != nil {
		return nil, err
	}

	since := now.Add(-24 * time.Hour)
	records, err := s.payments.ListSettledSince(ctx, since)
	if err != nil {
		run.Status = RunStatusFailed
		_ = s.runs.Save(ctx, run)
		return nil, err
	}

	var discrepancies []Discrepancy
	for _, rec := range records {
		if rec.ProviderTxnRef == "" {
			discrepancies = append(discrepancies, Discrepancy{
				Reference: rec.PaymentID,
				Reason:    "settled payment missing provider_txn_ref",
			})
		}
	}

	finished := time.Now().UTC()
	run.Status = RunStatusCompleted
	run.DiscrepancyCount = len(discrepancies)
	run.Discrepancies = discrepancies
	run.FinishedAt = &finished
	if err := s.runs.Save(ctx, run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *Service) GetRun(ctx context.Context, id string) (*Run, error) {
	return s.runs.FindByID(ctx, id)
}
