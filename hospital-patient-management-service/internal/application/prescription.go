package application

import (
	"context"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type PrescriptionService struct {
	base[model.Prescription]
	repo port.PrescriptionRepository
}

func NewPrescriptionService(repo port.PrescriptionRepository) port.PrescriptionUsecase {
	return &PrescriptionService{base: base[model.Prescription]{repo}, repo: repo}
}

func validatePrescription(p *model.Prescription) error {
	if err := firstErr(
		positive("patient_id", p.PatientID),
		positive("prescribed_by", p.PrescribedBy),
		required("medication_name", p.MedicationName),
		required("dosage", p.Dosage),
		required("frequency", p.Frequency),
	); err != nil {
		return err
	}
	if p.StartDate.IsZero() {
		return model.Invalid("start_date is required")
	}
	if p.EndDate != nil && p.EndDate.Before(p.StartDate) {
		return model.Invalid("end_date is before start_date")
	}
	return nil
}

func (s *PrescriptionService) Create(ctx context.Context, p *model.Prescription) (*model.Prescription, error) {
	if err := validatePrescription(p); err != nil {
		return nil, err
	}
	return create[model.Prescription](ctx, s.repo, p)
}

func (s *PrescriptionService) Update(ctx context.Context, p *model.Prescription) (*model.Prescription, error) {
	if err := validatePrescription(p); err != nil {
		return nil, err
	}
	return update[model.Prescription](ctx, s.repo, p.ID, p)
}

func (s *PrescriptionService) ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.Prescription, int64, error) {
	if err := positive("patient_id", patientID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListByPatient(ctx, patientID, page)
}
