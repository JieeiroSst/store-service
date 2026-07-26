package repository

import (
	"context"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"gorm.io/gorm"
)

type loanPaymentRepository struct {
	db *gorm.DB
}

func NewLoanPaymentRepository(db *gorm.DB) port.LoanPaymentRepository {
	return &loanPaymentRepository{db: db}
}

func (r *loanPaymentRepository) Create(ctx context.Context, payment *model.LoanPayment) error {
	if err := r.db.WithContext(ctx).Create(payment).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *loanPaymentRepository) GetByID(ctx context.Context, id int) (*model.LoanPayment, error) {
	var payment model.LoanPayment
	err := r.db.WithContext(ctx).First(&payment, "loan_payment_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &payment, nil
}

func (r *loanPaymentRepository) List(ctx context.Context) ([]model.LoanPayment, error) {
	var payments []model.LoanPayment
	if err := r.db.WithContext(ctx).Find(&payments).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return payments, nil
}

func (r *loanPaymentRepository) Update(ctx context.Context, payment *model.LoanPayment) error {
	if err := r.db.WithContext(ctx).Model(&model.LoanPayment{}).Where("loan_payment_id = ?", payment.LoanPaymentID).Updates(payment).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *loanPaymentRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.LoanPayment{}, "loan_payment_id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
