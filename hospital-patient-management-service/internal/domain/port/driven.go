package port

import (
	"context"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
)

type Store[T any] interface {
	Create(ctx context.Context, v *T) error
	Get(ctx context.Context, id int32) (*T, error)
	Update(ctx context.Context, v *T) error
	Delete(ctx context.Context, id int32) error
}

type DepartmentRepository interface {
	Store[model.Department]
	List(ctx context.Context, page model.Page) ([]model.Department, int64, error)
	Stats(ctx context.Context) ([]model.DepartmentStat, error)
}

type StaffRepository interface {
	Store[model.Staff]
	// List filters by department when departmentID > 0.
	List(ctx context.Context, departmentID int32, page model.Page) ([]model.Staff, int64, error)
}

type PatientRepository interface {
	Store[model.Patient]
	// List matches search case-insensitively against name, phone and email.
	List(ctx context.Context, search string, page model.Page) ([]model.Patient, int64, error)
}

type MedicalRecordRepository interface {
	Store[model.MedicalRecord]
	ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.MedicalRecord, int64, error)
}

type AppointmentRepository interface {
	Store[model.Appointment]
	List(ctx context.Context, f model.AppointmentFilter, page model.Page) ([]model.Appointment, int64, error)
	// HasScheduled reports whether staff already has a scheduled appointment at
	// exactly that time, ignoring excludeID (0 to ignore nothing).
	HasScheduled(ctx context.Context, staffID int32, at time.Time, excludeID int32) (bool, error)
	Stats(ctx context.Context, from, to *time.Time, departmentID int32) (*model.AppointmentStats, error)
}

type PrescriptionRepository interface {
	Store[model.Prescription]
	ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.Prescription, int64, error)
}

type LabResultRepository interface {
	Store[model.LabResult]
	ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.LabResult, int64, error)
}

type BillingRepository interface {
	Store[model.Billing]
	List(ctx context.Context, patientID int32, status model.BillingStatus, page model.Page) ([]model.Billing, int64, error)
	Stats(ctx context.Context, from, to *time.Time) (*model.BillingStats, error)
}

type Clock interface {
	Now() time.Time
}
