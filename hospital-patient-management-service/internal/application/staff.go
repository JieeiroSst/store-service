package application

import (
	"context"
	"strings"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/port"
)

type StaffService struct {
	base[model.Staff]
	repo port.StaffRepository
}

func NewStaffService(repo port.StaffRepository) port.StaffUsecase {
	return &StaffService{base: base[model.Staff]{repo}, repo: repo}
}

func validateStaff(s *model.Staff) error {
	if s.LicenseNumber != nil && strings.TrimSpace(*s.LicenseNumber) == "" {
		s.LicenseNumber = nil // blank must not collide on the unique index
	}
	s.Email = strings.TrimSpace(s.Email)
	return firstErr(
		positive("department_id", s.DepartmentID),
		required("first_name", s.FirstName),
		required("last_name", s.LastName),
		required("role", s.Role),
		required("email", s.Email),
		validEmail("email", s.Email),
		func() error {
			if s.HireDate.IsZero() {
				return model.Invalid("hire_date is required")
			}
			return nil
		}(),
	)
}

func (s *StaffService) Create(ctx context.Context, st *model.Staff) (*model.Staff, error) {
	if err := validateStaff(st); err != nil {
		return nil, err
	}
	return create[model.Staff](ctx, s.repo, st)
}

func (s *StaffService) Update(ctx context.Context, st *model.Staff) (*model.Staff, error) {
	if err := validateStaff(st); err != nil {
		return nil, err
	}
	return update[model.Staff](ctx, s.repo, st.ID, st)
}

func (s *StaffService) List(ctx context.Context, departmentID int32, page model.Page) ([]model.Staff, int64, error) {
	return s.repo.List(ctx, departmentID, page)
}
