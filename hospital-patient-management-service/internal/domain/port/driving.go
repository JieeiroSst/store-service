package port

import (
	"context"
	"time"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
)

type DepartmentUsecase interface {
	Create(ctx context.Context, d *model.Department) (*model.Department, error)
	Get(ctx context.Context, id int32) (*model.Department, error)
	List(ctx context.Context, page model.Page) ([]model.Department, int64, error)
	Update(ctx context.Context, d *model.Department) (*model.Department, error)
	Delete(ctx context.Context, id int32) error
	Stats(ctx context.Context) ([]model.DepartmentStat, error)
}

type StaffUsecase interface {
	Create(ctx context.Context, s *model.Staff) (*model.Staff, error)
	Get(ctx context.Context, id int32) (*model.Staff, error)
	List(ctx context.Context, departmentID int32, page model.Page) ([]model.Staff, int64, error)
	Update(ctx context.Context, s *model.Staff) (*model.Staff, error)
	Delete(ctx context.Context, id int32) error
}

type PatientUsecase interface {
	Create(ctx context.Context, p *model.Patient) (*model.Patient, error)
	Get(ctx context.Context, id int32) (*model.Patient, error)
	List(ctx context.Context, search string, page model.Page) ([]model.Patient, int64, error)
	Update(ctx context.Context, p *model.Patient) (*model.Patient, error)
	Delete(ctx context.Context, id int32) error
}

type MedicalRecordUsecase interface {
	Create(ctx context.Context, r *model.MedicalRecord) (*model.MedicalRecord, error)
	Get(ctx context.Context, id int32) (*model.MedicalRecord, error)
	ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.MedicalRecord, int64, error)
	Update(ctx context.Context, r *model.MedicalRecord) (*model.MedicalRecord, error)
	Delete(ctx context.Context, id int32) error
}

type AppointmentUsecase interface {
	Create(ctx context.Context, a *model.Appointment) (*model.Appointment, error)
	Get(ctx context.Context, id int32) (*model.Appointment, error)
	List(ctx context.Context, f model.AppointmentFilter, page model.Page) ([]model.Appointment, int64, error)
	Update(ctx context.Context, a *model.Appointment) (*model.Appointment, error)
	UpdateStatus(ctx context.Context, id int32, status model.AppointmentStatus) (*model.Appointment, error)
	Delete(ctx context.Context, id int32) error
	Stats(ctx context.Context, from, to *time.Time, departmentID int32) (*model.AppointmentStats, error)
}

type PrescriptionUsecase interface {
	Create(ctx context.Context, p *model.Prescription) (*model.Prescription, error)
	Get(ctx context.Context, id int32) (*model.Prescription, error)
	ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.Prescription, int64, error)
	Update(ctx context.Context, p *model.Prescription) (*model.Prescription, error)
	Delete(ctx context.Context, id int32) error
}

type LabResultUsecase interface {
	Create(ctx context.Context, r *model.LabResult) (*model.LabResult, error)
	Get(ctx context.Context, id int32) (*model.LabResult, error)
	ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.LabResult, int64, error)
	Update(ctx context.Context, r *model.LabResult) (*model.LabResult, error)
	Delete(ctx context.Context, id int32) error
}

type BillingUsecase interface {
	Create(ctx context.Context, b *model.Billing) (*model.Billing, error)
	Get(ctx context.Context, id int32) (*model.Billing, error)
	List(ctx context.Context, patientID int32, status model.BillingStatus, page model.Page) ([]model.Billing, int64, error)
	Update(ctx context.Context, b *model.Billing) (*model.Billing, error)
	// UpdateStatus sets paidDate only when marking Paid; nil defaults to today.
	UpdateStatus(ctx context.Context, id int32, status model.BillingStatus, paidDate *time.Time) (*model.Billing, error)
	Delete(ctx context.Context, id int32) error
	Stats(ctx context.Context, from, to *time.Time) (*model.BillingStats, error)
}

type IdentityUsecase interface {
	LinkUser(ctx context.Context, patientID int32, userID string) (*model.Patient, error)
	Status(ctx context.Context, patientID int32) (*model.IdentityStatus, error)
	SubmitCitizenCard(ctx context.Context, patientID int32, front, back []byte) (*model.IdentityStatus, error)
	SubmitFaceScan(ctx context.Context, patientID int32, frames [][]byte) (*model.IdentityStatus, error)
	Verify(ctx context.Context, patientID int32) (*model.IdentityStatus, error)
}

type DocumentUsecase interface {
	Upload(ctx context.Context, patientID int32, in DocumentUpload) (*model.Document, error)
	List(ctx context.Context, patientID int32, limit, offset int) ([]model.Document, int64, error)
	Open(ctx context.Context, patientID int32, docID string) (*DocumentDownload, error)
	Delete(ctx context.Context, patientID int32, docID string) error
}
