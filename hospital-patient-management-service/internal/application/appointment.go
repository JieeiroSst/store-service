package application

import (
	"context"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type AppointmentService struct {
	base[model.Appointment]
	repo   port.AppointmentRepository
	clock  port.Clock
	notify *PatientNotifier
}

func NewAppointmentService(repo port.AppointmentRepository, clock port.Clock, notify *PatientNotifier) port.AppointmentUsecase {
	return &AppointmentService{base: base[model.Appointment]{repo}, repo: repo, clock: clock, notify: notify}
}

func when(t time.Time) string { return t.UTC().Format("Mon 02 Jan 2006 15:04 MST") }

func validateAppointment(a *model.Appointment) error {
	if a.Status == "" {
		a.Status = model.AppointmentScheduled
	}
	if !a.Status.Valid() {
		return model.Invalid("status is not valid")
	}
	if a.AppointmentDate.IsZero() {
		return model.Invalid("appointment_date is required")
	}
	return firstErr(
		positive("patient_id", a.PatientID),
		positive("staff_id", a.StaffID),
		positive("department_id", a.DepartmentID),
	)
}

// ensureFree rejects double booking of a doctor at the same time.
func (s *AppointmentService) ensureFree(ctx context.Context, a *model.Appointment, excludeID int32) error {
	if a.Status != model.AppointmentScheduled {
		return nil
	}
	busy, err := s.repo.HasScheduled(ctx, a.StaffID, a.AppointmentDate, excludeID)
	if err != nil {
		return err
	}
	if busy {
		return model.Conflict("staff %d already has an appointment at %s", a.StaffID, a.AppointmentDate.Format(time.RFC3339))
	}
	return nil
}

func (s *AppointmentService) Create(ctx context.Context, a *model.Appointment) (*model.Appointment, error) {
	a.Status = model.AppointmentScheduled
	if err := validateAppointment(a); err != nil {
		return nil, err
	}
	if !a.AppointmentDate.After(s.clock.Now()) {
		return nil, model.Invalid("appointment_date must be in the future")
	}
	if err := s.ensureFree(ctx, a, 0); err != nil {
		return nil, err
	}
	out, err := create[model.Appointment](ctx, s.repo, a)
	if err != nil {
		return nil, err
	}
	s.notify.Notify(ctx, out.PatientID, "Appointment scheduled", "Your appointment is scheduled for "+when(out.AppointmentDate)+".")
	return out, nil
}

func (s *AppointmentService) Update(ctx context.Context, a *model.Appointment) (*model.Appointment, error) {
	current, err := s.Get(ctx, a.ID)
	if err != nil {
		return nil, err
	}
	if current.Status != model.AppointmentScheduled {
		return nil, model.Conflict("a %s appointment can no longer be changed", current.Status)
	}
	a.Status = current.Status
	if err := validateAppointment(a); err != nil {
		return nil, err
	}
	if !a.AppointmentDate.Equal(current.AppointmentDate) || a.StaffID != current.StaffID {
		if !a.AppointmentDate.After(s.clock.Now()) {
			return nil, model.Invalid("appointment_date must be in the future")
		}
		if err := s.ensureFree(ctx, a, a.ID); err != nil {
			return nil, err
		}
	}
	out, err := update[model.Appointment](ctx, s.repo, a.ID, a)
	if err != nil {
		return nil, err
	}
	if !out.AppointmentDate.Equal(current.AppointmentDate) {
		s.notify.Notify(ctx, out.PatientID, "Appointment rescheduled", "Your appointment moved to "+when(out.AppointmentDate)+".")
	}
	return out, nil
}

func (s *AppointmentService) UpdateStatus(ctx context.Context, id int32, status model.AppointmentStatus) (*model.Appointment, error) {
	current, err := s.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if !status.Valid() {
		return nil, model.Invalid("status is not valid")
	}
	if !current.Status.CanTransitionTo(status) {
		return nil, model.Conflict("cannot move appointment from %s to %s", current.Status, status)
	}
	current.Status = status
	out, err := update[model.Appointment](ctx, s.repo, id, current)
	if err != nil {
		return nil, err
	}
	switch status {
	case model.AppointmentCancelled:
		s.notify.Notify(ctx, out.PatientID, "Appointment cancelled", "Your appointment on "+when(out.AppointmentDate)+" was cancelled.")
	case model.AppointmentNoShow:
		s.notify.Notify(ctx, out.PatientID, "Missed appointment", "You missed your appointment on "+when(out.AppointmentDate)+". Please book a new one.")
	}
	return out, nil
}

func (s *AppointmentService) List(ctx context.Context, f model.AppointmentFilter, page model.Page) ([]model.Appointment, int64, error) {
	if f.Status != "" && !f.Status.Valid() {
		return nil, 0, model.Invalid("status is not valid")
	}
	if f.From != nil && f.To != nil && f.To.Before(*f.From) {
		return nil, 0, model.Invalid("date_to is before date_from")
	}
	return s.repo.List(ctx, f, page)
}

func (s *AppointmentService) Stats(ctx context.Context, from, to *time.Time, departmentID int32) (*model.AppointmentStats, error) {
	if from != nil && to != nil && to.Before(*from) {
		return nil, model.Invalid("date_to is before date_from")
	}
	return s.repo.Stats(ctx, from, to, departmentID)
}
