package application

import (
	"context"

	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
)

type employeeService struct {
	repo port.EmployeeRepository
}

func NewEmployeeService(repo port.EmployeeRepository) port.EmployeeUsecase {
	return &employeeService{repo: repo}
}

func (s *employeeService) CreateEmployee(ctx context.Context, employee *model.Employee) (*model.Employee, error) {
	if err := s.repo.Create(ctx, employee); err != nil {
		return nil, err
	}
	return employee, nil
}

func (s *employeeService) GetEmployee(ctx context.Context, id int) (*model.Employee, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *employeeService) ListEmployees(ctx context.Context) ([]model.Employee, error) {
	return s.repo.List(ctx)
}

func (s *employeeService) UpdateEmployee(ctx context.Context, employee *model.Employee) (*model.Employee, error) {
	if _, err := s.repo.GetByID(ctx, employee.EmployeeID); err != nil {
		return nil, err
	}
	if err := s.repo.Update(ctx, employee); err != nil {
		return nil, err
	}
	return employee, nil
}

func (s *employeeService) DeleteEmployee(ctx context.Context, id int) error {
	return s.repo.Delete(ctx, id)
}
