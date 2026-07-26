package repository

import (
	"context"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"gorm.io/gorm"
)

type employeeRepository struct {
	db *gorm.DB
}

func NewEmployeeRepository(db *gorm.DB) port.EmployeeRepository {
	return &employeeRepository{db: db}
}

func (r *employeeRepository) Create(ctx context.Context, employee *model.Employee) error {
	if err := r.db.WithContext(ctx).Create(employee).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *employeeRepository) GetByID(ctx context.Context, id int) (*model.Employee, error) {
	var employee model.Employee
	err := r.db.WithContext(ctx).First(&employee, "employee_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &employee, nil
}

func (r *employeeRepository) List(ctx context.Context) ([]model.Employee, error) {
	var employees []model.Employee
	if err := r.db.WithContext(ctx).Find(&employees).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return employees, nil
}

func (r *employeeRepository) Update(ctx context.Context, employee *model.Employee) error {
	if err := r.db.WithContext(ctx).Model(&model.Employee{}).Where("employee_id = ?", employee.EmployeeID).Updates(employee).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *employeeRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Employee{}, "employee_id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
