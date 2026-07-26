package repository

import (
	"context"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"gorm.io/gorm"
)

type customerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) port.CustomerRepository {
	return &customerRepository{db: db}
}

func (r *customerRepository) Create(ctx context.Context, customer *model.Customer) error {
	if err := r.db.WithContext(ctx).Create(customer).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *customerRepository) GetByID(ctx context.Context, id int) (*model.Customer, error) {
	var customer model.Customer
	err := r.db.WithContext(ctx).First(&customer, "customer_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &customer, nil
}

func (r *customerRepository) List(ctx context.Context) ([]model.Customer, error) {
	var customers []model.Customer
	if err := r.db.WithContext(ctx).Find(&customers).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return customers, nil
}

func (r *customerRepository) Update(ctx context.Context, customer *model.Customer) error {
	if err := r.db.WithContext(ctx).Model(&model.Customer{}).Where("customer_id = ?", customer.CustomerID).Updates(customer).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *customerRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Customer{}, "customer_id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
