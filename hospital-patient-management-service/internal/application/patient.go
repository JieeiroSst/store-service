package application

import (
	"context"
	"strings"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type PatientService struct {
	base[model.Patient]
	repo  port.PatientRepository
	clock port.Clock
	docs  port.DocumentGateway 
}

func NewPatientService(repo port.PatientRepository, clock port.Clock, docs port.DocumentGateway) port.PatientUsecase {
	return &PatientService{base: base[model.Patient]{repo}, repo: repo, clock: clock, docs: docs}
}

func (s *PatientService) Delete(ctx context.Context, id int32) error {
	if s.docs != nil && s.docs.Enabled() && id > 0 {
		_, total, err := s.docs.List(ctx, id, 1, 0)
		if err != nil {
			return err
		}
		if total > 0 {
			return model.Conflict("patient %d still has %d document(s); delete them first", id, total)
		}
	}
	return s.base.Delete(ctx, id)
}

func (s *PatientService) validate(p *model.Patient) error {
	if err := firstErr(
		required("first_name", p.FirstName),
		required("last_name", p.LastName),
	); err != nil {
		return err
	}
	if p.DateOfBirth.IsZero() {
		return model.Invalid("date_of_birth is required")
	}
	if p.DateOfBirth.After(s.clock.Now()) {
		return model.Invalid("date_of_birth cannot be in the future")
	}
	if !p.Gender.Valid() {
		return model.Invalid("gender is required")
	}
	if p.BloodType != nil && !p.BloodType.Valid() {
		return model.Invalid("blood_type is not valid")
	}
	if p.Email = strings.TrimSpace(p.Email); p.Email != "" {
		return validEmail("email", p.Email)
	}
	return nil
}

func (s *PatientService) Create(ctx context.Context, p *model.Patient) (*model.Patient, error) {
	if err := s.validate(p); err != nil {
		return nil, err
	}
	return create[model.Patient](ctx, s.repo, p)
}

func (s *PatientService) Update(ctx context.Context, p *model.Patient) (*model.Patient, error) {
	if err := s.validate(p); err != nil {
		return nil, err
	}
	current, err := s.Get(ctx, p.ID)
	if err != nil {
		return nil, err
	}
	p.UserID = current.UserID
	return update[model.Patient](ctx, s.repo, p.ID, p)
}

func (s *PatientService) List(ctx context.Context, search string, page model.Page) ([]model.Patient, int64, error) {
	return s.repo.List(ctx, strings.TrimSpace(search), page)
}
