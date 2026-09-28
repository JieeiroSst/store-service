package postgres

import (
	"context"
	"time"

	"github.com/JIeeiroSst/vending-machine-service/internal/domain"
	"github.com/JIeeiroSst/vending-machine-service/internal/port"
	"github.com/jackc/pgx/v5"
)

const machineColumns = `machine_id, location, model, status, last_maintenance_at, created_at, updated_at`

type machineRepository struct{ db *DB }

func NewMachineRepository(db *DB) port.MachineRepository { return &machineRepository{db: db} }

func scanMachine(row pgx.Row) (domain.Machine, error) {
	var m domain.Machine
	err := row.Scan(&m.ID, &m.Location, &m.Model, &m.Status, &m.LastMaintenance, &m.CreatedAt, &m.UpdatedAt)
	return m, err
}

func (r *machineRepository) Create(ctx context.Context, m *domain.Machine) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO machines (`+machineColumns+`) VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		m.ID, m.Location, m.Model, m.Status, m.LastMaintenance, m.CreatedAt, m.UpdatedAt)
	return mapErr(err)
}

func (r *machineRepository) Get(ctx context.Context, id string) (*domain.Machine, error) {
	m, err := scanMachine(r.db.q(ctx).QueryRow(ctx, `SELECT `+machineColumns+` FROM machines WHERE machine_id = $1`, id))
	if err != nil {
		return nil, mapErr(err)
	}
	return &m, nil
}

func (r *machineRepository) List(ctx context.Context, status string, page domain.PageRequest) (domain.Page[domain.Machine], error) {
	after, err := timeCursorArgs(page.Cursor)
	if err != nil {
		return domain.Page[domain.Machine]{}, err
	}
	size := page.Size()
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+machineColumns+` FROM machines
		WHERE ($1 = '' OR status = $1)
		  AND ($2::timestamptz IS NULL OR (created_at, machine_id) > ($2::timestamptz, $3::text))
		ORDER BY created_at, machine_id LIMIT $4`, status, after[0], after[1], size+1)
	if err != nil {
		return domain.Page[domain.Machine]{}, mapErr(err)
	}
	items, err := collect(rows, scanMachine)
	if err != nil {
		return domain.Page[domain.Machine]{}, err
	}
	return paginate(items, size, func(m *domain.Machine) []string { return []string{timeKey(m.CreatedAt), m.ID} }), nil
}

func (r *machineRepository) Update(ctx context.Context, m *domain.Machine) error {
	tag, err := r.db.q(ctx).Exec(ctx, `UPDATE machines
		SET location = $2, model = $3, status = $4, last_maintenance_at = $5, updated_at = $6
		WHERE machine_id = $1`,
		m.ID, m.Location, m.Model, m.Status, m.LastMaintenance, m.UpdatedAt)
	if err == nil && tag.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return mapErr(err)
}

type maintenanceRepository struct{ db *DB }

func NewMaintenanceRepository(db *DB) port.MaintenanceRepository {
	return &maintenanceRepository{db: db}
}

func (r *maintenanceRepository) Create(ctx context.Context, l *domain.MaintenanceLog) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO maintenance_logs
		(log_id, machine_id, technician_id, maintenance_type, notes, performed_at)
		VALUES ($1, $2, $3, $4, $5, $6)`,
		l.ID, l.MachineID, l.TechnicianID, l.MaintenanceType, l.Notes, l.PerformedAt)
	return mapErr(err)
}

func (r *maintenanceRepository) ListByMachine(ctx context.Context, machineID string) ([]domain.MaintenanceLog, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT log_id, machine_id, technician_id, maintenance_type, notes, performed_at
		FROM maintenance_logs WHERE machine_id = $1 ORDER BY performed_at DESC`, machineID)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, func(row pgx.Row) (domain.MaintenanceLog, error) {
		var l domain.MaintenanceLog
		err := row.Scan(&l.ID, &l.MachineID, &l.TechnicianID, &l.MaintenanceType, &l.Notes, &l.PerformedAt)
		return l, err
	})
}

type eventRepository struct{ db *DB }

func NewEventRepository(db *DB) port.EventRepository { return &eventRepository{db: db} }

func (r *eventRepository) Append(ctx context.Context, e *domain.Event) error {
	_, err := r.db.q(ctx).Exec(ctx, `INSERT INTO events
		(event_id, event_type, related_entity, entity_id, machine_id, data, occurred_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)`,
		e.ID, e.EventType, e.RelatedEntity, e.EntityID, e.MachineID, e.Data, e.OccurredAt)
	return mapErr(err)
}

const eventColumns = `event_id, event_type, related_entity, entity_id, machine_id, data, occurred_at`

func scanEvent(row pgx.Row) (domain.Event, error) {
	var e domain.Event
	err := row.Scan(&e.ID, &e.EventType, &e.RelatedEntity, &e.EntityID, &e.MachineID, &e.Data, &e.OccurredAt)
	return e, err
}

func (r *eventRepository) ListByMachine(ctx context.Context, machineID string, limit int) ([]domain.Event, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+eventColumns+`
		FROM events WHERE machine_id = $1 ORDER BY occurred_at DESC LIMIT $2`, machineID, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, scanEvent)
}

func (r *eventRepository) LockUndispatched(ctx context.Context, limit int) ([]domain.Event, error) {
	rows, err := r.db.q(ctx).Query(ctx, `SELECT `+eventColumns+`
		FROM events WHERE dispatched_at IS NULL ORDER BY occurred_at LIMIT $1 FOR UPDATE SKIP LOCKED`, limit)
	if err != nil {
		return nil, mapErr(err)
	}
	return collect(rows, scanEvent)
}

func (r *eventRepository) MarkDispatched(ctx context.Context, ids []string, at time.Time) error {
	_, err := r.db.q(ctx).Exec(ctx, `UPDATE events SET dispatched_at = $2 WHERE event_id = ANY($1)`, ids, at)
	return mapErr(err)
}
