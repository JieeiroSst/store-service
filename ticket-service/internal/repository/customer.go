package repository

import (
	"context"

	"github.com/JIeeiroSst/ticket-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"gorm.io/gorm"
)

type Customers interface {
	SaveCustomer(ctx context.Context, req model.Customers) error
	Find(ctx context.Context, customerID int) (*model.Customers, error)
}

type CustomerRepository struct {
	db *gorm.DB
}

func NewCustomerRepository(db *gorm.DB) *CustomerRepository {
	return &CustomerRepository{
		db: db,
	}
}

func (r *CustomerRepository) SaveCustomer(ctx context.Context, req model.Customers) error {
	var customer model.Customers

	if err := r.db.Where("customer_id = ?", req.CustomerID).Find(&customer).Error; err != nil {
		logger.Error(err)
		return err
	}

	if customer.CustomerID == 0 {
		if err := r.db.Create(&req).Error; err != nil {
			logger.Error(err)
			return err
		}
		return nil
	}
	if err := r.db.Save(&customer).Error; err != nil {
		logger.Error(err)
		return err
	}
	return nil
}

func (r *CustomerRepository) Find(ctx context.Context, customerID int) (*model.Customers, error) {
	var customer model.Customers

	if err := r.db.Where("customer_id = ?", customerID).Find(&customer).Error; err != nil {
		logger.Error(err)
		return nil, err
	}

	return &customer, nil
}
