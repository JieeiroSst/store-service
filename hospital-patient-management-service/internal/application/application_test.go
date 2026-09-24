package application

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
)

type fixedClock time.Time

func (c fixedClock) Now() time.Time { return time.Time(c) }

var now = time.Date(2026, 6, 15, 10, 0, 0, 0, time.UTC)

// memAppointments is an in-memory port.AppointmentRepository.
type memAppointments struct {
	rows map[int32]model.Appointment
	next int32
}

func newMemAppointments() *memAppointments {
	return &memAppointments{rows: map[int32]model.Appointment{}}
}

func (m *memAppointments) Create(_ context.Context, a *model.Appointment) error {
	m.next++
	a.ID = m.next
	m.rows[a.ID] = *a
	return nil
}

func (m *memAppointments) Get(_ context.Context, id int32) (*model.Appointment, error) {
	a, ok := m.rows[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return &a, nil
}

func (m *memAppointments) Update(_ context.Context, a *model.Appointment) error {
	if _, ok := m.rows[a.ID]; !ok {
		return model.ErrNotFound
	}
	m.rows[a.ID] = *a
	return nil
}

func (m *memAppointments) Delete(_ context.Context, id int32) error { delete(m.rows, id); return nil }

func (m *memAppointments) List(context.Context, model.AppointmentFilter, model.Page) ([]model.Appointment, int64, error) {
	return nil, 0, nil
}

func (m *memAppointments) HasScheduled(_ context.Context, staffID int32, at time.Time, excludeID int32) (bool, error) {
	for _, a := range m.rows {
		if a.ID != excludeID && a.StaffID == staffID && a.AppointmentDate.Equal(at) && a.Status == model.AppointmentScheduled {
			return true, nil
		}
	}
	return false, nil
}

func (m *memAppointments) Stats(context.Context, *time.Time, *time.Time, int32) (*model.AppointmentStats, error) {
	return &model.AppointmentStats{}, nil
}

func newAppointment(at time.Time) *model.Appointment {
	return &model.Appointment{PatientID: 1, StaffID: 7, DepartmentID: 1, AppointmentDate: at}
}

func TestAppointmentCreate(t *testing.T) {
	ctx := context.Background()
	svc := NewAppointmentService(newMemAppointments(), fixedClock(now), nil)

	tomorrow := now.Add(24 * time.Hour)
	a, err := svc.Create(ctx, newAppointment(tomorrow))
	if err != nil || a.Status != model.AppointmentScheduled {
		t.Fatalf("create = %+v, %v", a, err)
	}

	if _, err := svc.Create(ctx, newAppointment(tomorrow)); !errors.Is(err, model.ErrConflict) {
		t.Errorf("double booking: got %v, want ErrConflict", err)
	}
	if _, err := svc.Create(ctx, newAppointment(now.Add(-time.Hour))); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("past date: got %v, want ErrInvalid", err)
	}
	bad := newAppointment(tomorrow.Add(time.Hour))
	bad.StaffID = 0
	if _, err := svc.Create(ctx, bad); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("missing staff: got %v, want ErrInvalid", err)
	}
}

func TestAppointmentStatusTransitions(t *testing.T) {
	ctx := context.Background()
	svc := NewAppointmentService(newMemAppointments(), fixedClock(now), nil)
	a, _ := svc.Create(ctx, newAppointment(now.Add(time.Hour)))

	if _, err := svc.UpdateStatus(ctx, a.ID, model.AppointmentScheduled); !errors.Is(err, model.ErrConflict) {
		t.Errorf("scheduled -> scheduled: got %v, want ErrConflict", err)
	}
	if _, err := svc.UpdateStatus(ctx, a.ID, "Bogus"); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("bogus status: got %v, want ErrInvalid", err)
	}
	got, err := svc.UpdateStatus(ctx, a.ID, model.AppointmentCompleted)
	if err != nil || got.Status != model.AppointmentCompleted {
		t.Fatalf("complete = %+v, %v", got, err)
	}
	if _, err := svc.UpdateStatus(ctx, a.ID, model.AppointmentCancelled); !errors.Is(err, model.ErrConflict) {
		t.Errorf("completed -> cancelled: got %v, want ErrConflict", err)
	}
	if _, err := svc.Update(ctx, &model.Appointment{ID: a.ID, PatientID: 1, StaffID: 7, DepartmentID: 1, AppointmentDate: now.Add(2 * time.Hour)}); !errors.Is(err, model.ErrConflict) {
		t.Errorf("edit completed appointment: got %v, want ErrConflict", err)
	}
}

func TestAppointmentRescheduleFreesOldSlot(t *testing.T) {
	ctx := context.Background()
	svc := NewAppointmentService(newMemAppointments(), fixedClock(now), nil)
	slot := now.Add(time.Hour)
	a, _ := svc.Create(ctx, newAppointment(slot))

	// Re-saving an appointment onto its own slot is not a clash with itself.
	same := *a
	if _, err := svc.Update(ctx, &same); err != nil {
		t.Fatalf("update in place: %v", err)
	}
	// Cancelling releases the slot for someone else.
	if _, err := svc.UpdateStatus(ctx, a.ID, model.AppointmentCancelled); err != nil {
		t.Fatal(err)
	}
	if _, err := svc.Create(ctx, newAppointment(slot)); err != nil {
		t.Errorf("slot should be free after cancel: %v", err)
	}
}

// memBilling is an in-memory port.BillingRepository.
type memBilling struct {
	rows map[int32]model.Billing
	next int32
	tick int
}

func (m *memBilling) Create(_ context.Context, b *model.Billing) error {
	if m.rows == nil {
		m.rows = map[int32]model.Billing{}
	}
	m.next++
	b.ID = m.next
	m.rows[b.ID] = *b
	return nil
}

func (m *memBilling) Get(_ context.Context, id int32) (*model.Billing, error) {
	b, ok := m.rows[id]
	if !ok {
		return nil, model.ErrNotFound
	}
	return &b, nil
}

// Update stamps updated_at the way the database does.
func (m *memBilling) Update(_ context.Context, b *model.Billing) error {
	m.tick++
	b.UpdatedAt = now.Add(time.Duration(m.tick) * time.Second)
	m.rows[b.ID] = *b
	return nil
}
func (m *memBilling) Delete(_ context.Context, id int32) error { delete(m.rows, id); return nil }
func (m *memBilling) List(context.Context, int32, model.BillingStatus, model.Page) ([]model.Billing, int64, error) {
	return nil, 0, nil
}
func (m *memBilling) Stats(context.Context, *time.Time, *time.Time) (*model.BillingStats, error) {
	return &model.BillingStats{}, nil
}

func TestBillingAmounts(t *testing.T) {
	ctx := context.Background()
	svc := newRig().svc

	b, err := svc.Create(ctx, &model.Billing{PatientID: 1, Amount: 100.5, InsuranceCovered: 40})
	if err != nil {
		t.Fatal(err)
	}
	if b.PatientResponsibility != 60.5 || b.Status != model.BillingPending {
		t.Errorf("defaults = %+v", b)
	}

	cases := map[string]model.Billing{
		"zero amount":        {PatientID: 1, Amount: 0},
		"insurance too high": {PatientID: 1, Amount: 10, InsuranceCovered: 11},
		"shares do not add":  {PatientID: 1, Amount: 100, InsuranceCovered: 40, PatientResponsibility: 30},
		"unknown status":     {PatientID: 1, Amount: 10, Status: "Weird"},
		"missing patient":    {Amount: 10},
		"negative insurance": {PatientID: 1, Amount: 10, InsuranceCovered: -1},
	}
	for name, in := range cases {
		in := in
		if _, err := svc.Create(ctx, &in); !errors.Is(err, model.ErrInvalid) {
			t.Errorf("%s: got %v, want ErrInvalid", name, err)
		}
	}
}

func TestBillingPaidDate(t *testing.T) {
	ctx := context.Background()
	svc := newRig().svc
	b, _ := svc.Create(ctx, &model.Billing{PatientID: 1, Amount: 50})

	paid, err := svc.UpdateStatus(ctx, b.ID, model.BillingPaid, nil)
	if err != nil || paid.PaidDate == nil || !paid.PaidDate.Equal(dateOnly(now)) {
		t.Fatalf("paid = %+v, %v", paid, err)
	}
	reopened, err := svc.UpdateStatus(ctx, b.ID, model.BillingOverdue, nil)
	if err != nil || reopened.PaidDate != nil {
		t.Errorf("leaving Paid must clear paid_date: %+v, %v", reopened, err)
	}
	// Edits cannot smuggle in a status change.
	edit := *reopened
	edit.Status = model.BillingPaid
	got, err := svc.Update(ctx, &edit)
	if err != nil || got.Status != model.BillingOverdue {
		t.Errorf("update must keep status: %+v, %v", got, err)
	}
}

func TestPatientValidation(t *testing.T) {
	svc := &PatientService{clock: fixedClock(now)}
	ok := model.Patient{FirstName: "A", LastName: "B", DateOfBirth: now.AddDate(-30, 0, 0), Gender: model.GenderFemale}
	if err := svc.validate(&ok); err != nil {
		t.Fatalf("valid patient rejected: %v", err)
	}

	mut := map[string]func(*model.Patient){
		"future birth": func(p *model.Patient) { p.DateOfBirth = now.Add(time.Hour) },
		"no gender":    func(p *model.Patient) { p.Gender = "" },
		"bad email":    func(p *model.Patient) { p.Email = "not-an-email" },
		"bad blood":    func(p *model.Patient) { b := model.BloodType("Z"); p.BloodType = &b },
		"no name":      func(p *model.Patient) { p.FirstName = " " },
	}
	for name, f := range mut {
		p := ok
		f(&p)
		if err := svc.validate(&p); !errors.Is(err, model.ErrInvalid) {
			t.Errorf("%s: got %v, want ErrInvalid", name, err)
		}
	}
}

func TestPrescriptionDates(t *testing.T) {
	p := model.Prescription{PatientID: 1, PrescribedBy: 1, MedicationName: "x", Dosage: "1", Frequency: "1", StartDate: now}
	if err := validatePrescription(&p); err != nil {
		t.Fatal(err)
	}
	before := now.AddDate(0, 0, -1)
	p.EndDate = &before
	if err := validatePrescription(&p); !errors.Is(err, model.ErrInvalid) {
		t.Errorf("end before start: got %v, want ErrInvalid", err)
	}
}

func TestPageNormalization(t *testing.T) {
	cases := []struct {
		in         model.Page
		limit, off int
	}{
		{model.Page{}, 20, 0},
		{model.Page{Number: 3, Size: 10}, 10, 20},
		{model.Page{Number: 1, Size: 5000}, 100, 0},
		{model.Page{Number: -4, Size: -1}, 20, 0},
	}
	for _, c := range cases {
		if c.in.Limit() != c.limit || c.in.Offset() != c.off {
			t.Errorf("%+v: limit=%d offset=%d, want %d/%d", c.in, c.in.Limit(), c.in.Offset(), c.limit, c.off)
		}
	}
}
