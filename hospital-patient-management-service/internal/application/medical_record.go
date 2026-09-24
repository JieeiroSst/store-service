package application

import (
	"context"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type MedicalRecordService struct {
	base[model.MedicalRecord]
	repo port.MedicalRecordRepository
}

func NewMedicalRecordService(repo port.MedicalRecordRepository) port.MedicalRecordUsecase {
	return &MedicalRecordService{base: base[model.MedicalRecord]{repo}, repo: repo}
}

func validateMedicalRecord(r *model.MedicalRecord) error {
	if r.CreatedBy != nil && *r.CreatedBy <= 0 {
		r.CreatedBy = nil
	}
	return firstErr(positive("patient_id", r.PatientID), required("diagnosis", r.Diagnosis))
}

func (s *MedicalRecordService) Create(ctx context.Context, r *model.MedicalRecord) (*model.MedicalRecord, error) {
	if err := validateMedicalRecord(r); err != nil {
		return nil, err
	}
	return create[model.MedicalRecord](ctx, s.repo, r)
}

func (s *MedicalRecordService) Update(ctx context.Context, r *model.MedicalRecord) (*model.MedicalRecord, error) {
	if err := validateMedicalRecord(r); err != nil {
		return nil, err
	}
	return update[model.MedicalRecord](ctx, s.repo, r.ID, r)
}

func (s *MedicalRecordService) ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.MedicalRecord, int64, error) {
	if err := positive("patient_id", patientID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListByPatient(ctx, patientID, page)
}
