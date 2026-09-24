package application

import (
	"context"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type LabResultService struct {
	base[model.LabResult]
	repo   port.LabResultRepository
	notify *PatientNotifier
}

func NewLabResultService(repo port.LabResultRepository, notify *PatientNotifier) port.LabResultUsecase {
	return &LabResultService{base: base[model.LabResult]{repo}, repo: repo, notify: notify}
}

func validateLabResult(r *model.LabResult) error {
	if err := firstErr(
		positive("patient_id", r.PatientID),
		positive("ordered_by", r.OrderedBy),
		required("test_name", r.TestName),
		required("results", r.Results),
	); err != nil {
		return err
	}
	if r.TestDate.IsZero() {
		return model.Invalid("test_date is required")
	}
	return nil
}

func (s *LabResultService) Create(ctx context.Context, r *model.LabResult) (*model.LabResult, error) {
	if err := validateLabResult(r); err != nil {
		return nil, err
	}
	out, err := create[model.LabResult](ctx, s.repo, r)
	if err != nil {
		return nil, err
	}
	s.notify.Notify(ctx, out.PatientID, "Lab results available", "Your "+out.TestName+" results are ready.")
	return out, nil
}

func (s *LabResultService) Update(ctx context.Context, r *model.LabResult) (*model.LabResult, error) {
	if err := validateLabResult(r); err != nil {
		return nil, err
	}
	return update[model.LabResult](ctx, s.repo, r.ID, r)
}

func (s *LabResultService) ListByPatient(ctx context.Context, patientID int32, page model.Page) ([]model.LabResult, int64, error) {
	if err := positive("patient_id", patientID); err != nil {
		return nil, 0, err
	}
	return s.repo.ListByPatient(ctx, patientID, page)
}
