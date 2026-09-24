package application

import (
	"context"
	"errors"
	"strconv"
	"time"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
)

type contractLifecycle struct {
	repo         port.Repository[model.Contract]
	expirer      port.ContractExpirer
	orchestrator port.ContractOrchestrator
	notifier     port.Notifier
	now          func() time.Time
}

func NewContractLifecycle(
	repo port.Repository[model.Contract], e port.ContractExpirer, o port.ContractOrchestrator, n port.Notifier,
) port.ContractLifecycleUsecase {
	return &contractLifecycle{repo: repo, expirer: e, orchestrator: o, notifier: n, now: time.Now}
}

func (s *contractLifecycle) State(ctx context.Context, id uint) (port.ContractLifecycleState, error) {
	c, err := s.repo.GetByID(ctx, id)
	if errors.Is(err, common.ErrNotFound) {
		return port.ContractLifecycleState{}, nil
	}
	if err != nil {
		return port.ContractLifecycleState{}, err
	}
	return port.ContractLifecycleState{Exists: true, Status: c.Status, EndDate: c.EndDate}, nil
}

func (s *contractLifecycle) Expire(ctx context.Context, id uint, at time.Time) (bool, error) {
	expired, err := s.expirer.ExpireOne(ctx, id, at)
	if err != nil || !expired {
		return false, err
	}
	notify(ctx, s.notifier, "Contract expired", "Contract #%d passed its end date and was closed; it needs a new signature to be reopened", id)
	return true, nil
}

func (s *contractLifecycle) Remind(ctx context.Context, id uint, remaining time.Duration) error {
	c, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if c.Status != model.ContractApproved {
		return nil
	}
	notify(ctx, s.notifier, "Contract ending soon", "Contract #%d ends in about %s", id, humanize(remaining))
	return nil
}

const resyncPage = 200

// Resync starts (or wakes) the lifecycle of every contract that can still
// expire. It repairs missed syncs and populates a fresh Temporal namespace.
func (s *contractLifecycle) Resync(ctx context.Context) (int, error) {
	synced := 0
	for _, status := range []string{model.ContractDraft, model.ContractPending, model.ContractApproved} {
		for offset := 0; ; offset += resyncPage {
			page, total, err := s.repo.List(ctx, port.ListQuery{
				Offset: offset, Limit: resyncPage, Equals: map[string]any{"contract_status": status},
			})
			if err != nil {
				return synced, err
			}
			for _, c := range page {
				if c.EndDate == nil {
					continue
				}
				if err := s.orchestrator.Sync(ctx, c.ID); err != nil {
					return synced, err
				}
				synced++
			}
			if int64(offset+resyncPage) >= total {
				break
			}
		}
	}
	return synced, nil
}

func humanize(d time.Duration) string {
	switch days := int(d.Round(time.Hour).Hours() / 24); {
	case days >= 1:
		return plural(days, "day")
	case d >= time.Hour:
		return plural(int(d.Round(time.Hour).Hours()), "hour")
	default:
		return d.Round(time.Second).String()
	}
}

func plural(n int, unit string) string {
	if n == 1 {
		return "1 " + unit
	}
	return strconv.Itoa(n) + " " + unit + "s"
}
