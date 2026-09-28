package application

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/google/uuid"
)

const (
	maxEventLimit  = 500
	maxReportRange = 366 * 24 * time.Hour
)

type machineService struct {
	tx          port.Transactor
	machines    port.MachineRepository
	maintenance port.MaintenanceRepository
	events      port.EventRepository
	reports     port.ReportRepository
	now         func() time.Time
}

func NewMachineService(
	tx port.Transactor,
	machines port.MachineRepository,
	maintenance port.MaintenanceRepository,
	events port.EventRepository,
	reports port.ReportRepository,
) port.MachineService {
	return &machineService{tx: tx, machines: machines, maintenance: maintenance, events: events, reports: reports, now: time.Now}
}

func (s *machineService) Register(ctx context.Context, m *domain.Machine) (*domain.Machine, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	now := s.now().UTC()
	m.ID = uuid.NewString()
	if m.Status == "" {
		m.Status = domain.MachineActive
	}
	m.CreatedAt, m.UpdatedAt = now, now
	if err := s.machines.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *machineService) Get(ctx context.Context, id string) (*domain.Machine, error) {
	return s.machines.Get(ctx, id)
}

func (s *machineService) List(ctx context.Context, status string, page domain.PageRequest) (domain.Page[domain.Machine], error) {
	if status != "" {
		if _, err := domain.ParseMachineStatus(status); err != nil {
			return domain.Page[domain.Machine]{}, err
		}
	}
	return s.machines.List(ctx, status, page)
}

func (s *machineService) Update(ctx context.Context, id string, patch domain.MachinePatch) (*domain.Machine, error) {
	var m *domain.Machine
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if m, err = s.machines.Get(ctx, id); err != nil {
			return err
		}
		if err := m.Apply(patch); err != nil {
			return err
		}
		m.UpdatedAt = s.now().UTC()
		return s.machines.Update(ctx, m)
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *machineService) ChangeStatus(ctx context.Context, id string, status domain.MachineStatus) (*domain.Machine, error) {
	var m *domain.Machine
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var err error
		if m, err = s.machines.Get(ctx, id); err != nil {
			return err
		}
		if m.Status == status {
			return nil
		}
		from := m.Status
		m.Status = status
		m.UpdatedAt = s.now().UTC()
		if err := s.machines.Update(ctx, m); err != nil {
			return err
		}
		return s.events.Append(ctx, newEvent(m.UpdatedAt, domain.EventMachineStatusChanged, "machine", m.ID, m.ID,
			map[string]any{"from": from, "to": status, "location": m.Location}))
	})
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (s *machineService) RecordMaintenance(ctx context.Context, l *domain.MaintenanceLog) (*domain.MaintenanceLog, error) {
	if err := l.Validate(); err != nil {
		return nil, err
	}
	err := s.tx.WithinTx(ctx, func(ctx context.Context) error {
		m, err := s.machines.Get(ctx, l.MachineID)
		if err != nil {
			return err
		}
		now := s.now().UTC()
		l.ID = uuid.NewString()
		l.PerformedAt = now
		if err := s.maintenance.Create(ctx, l); err != nil {
			return err
		}
		m.LastMaintenance = &now
		m.UpdatedAt = now
		if err := s.machines.Update(ctx, m); err != nil {
			return err
		}
		return s.events.Append(ctx, newEvent(now, domain.EventMaintenancePerformed, "machine", m.ID, m.ID,
			map[string]any{"log_id": l.ID, "maintenance_type": l.MaintenanceType}))
	})
	if err != nil {
		return nil, err
	}
	return l, nil
}

func (s *machineService) ListMaintenance(ctx context.Context, machineID string) ([]domain.MaintenanceLog, error) {
	if _, err := s.machines.Get(ctx, machineID); err != nil {
		return nil, err
	}
	return s.maintenance.ListByMachine(ctx, machineID)
}

func (s *machineService) ListEvents(ctx context.Context, machineID string, limit int) ([]domain.Event, error) {
	if limit <= 0 || limit > maxEventLimit {
		limit = maxEventLimit
	}
	return s.events.ListByMachine(ctx, machineID, limit)
}

func (s *machineService) SalesReport(ctx context.Context, machineID string, from, to time.Time) (*domain.SalesReport, error) {
	if !from.Before(to) || to.Sub(from) > maxReportRange {
		return nil, fmt.Errorf("%w: from must be before to and the range at most 366 days", domain.ErrInvalidInput)
	}
	if _, err := s.machines.Get(ctx, machineID); err != nil {
		return nil, err
	}
	products, err := s.reports.SalesByProduct(ctx, machineID, from, to)
	if err != nil {
		return nil, err
	}
	report := &domain.SalesReport{MachineID: machineID, From: from, To: to, Products: products}
	for _, p := range products {
		report.Orders += p.Units
		report.RevenueCents += p.RevenueCents
		report.DiscountCents += p.DiscountCents
	}
	return report, nil
}
