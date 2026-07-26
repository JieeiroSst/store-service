package repository

import (
	"context"

	"github.com/JieeiroSst/banking-service/common"
	"github.com/JieeiroSst/banking-service/internal/domain/model"
	"github.com/JieeiroSst/banking-service/internal/domain/port"
	"gorm.io/gorm"
)

type loanRepository struct {
	db *gorm.DB
}

func NewLoanRepository(db *gorm.DB) port.LoanRepository {
	return &loanRepository{db: db}
}

func (r *loanRepository) Create(ctx context.Context, loan *model.Loan) error {
	if err := r.db.WithContext(ctx).Create(loan).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *loanRepository) GetByID(ctx context.Context, id int) (*model.Loan, error) {
	var loan model.Loan
	err := r.db.WithContext(ctx).First(&loan, "loan_id = ?", id).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &loan, nil
}

func (r *loanRepository) List(ctx context.Context) ([]model.Loan, error) {
	var loans []model.Loan
	if err := r.db.WithContext(ctx).Find(&loans).Error; err != nil {
		return nil, common.ErrDBFailed
	}
	return loans, nil
}

func (r *loanRepository) Update(ctx context.Context, loan *model.Loan) error {
	if err := r.db.WithContext(ctx).Model(&model.Loan{}).Where("loan_id = ?", loan.LoanID).Updates(loan).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *loanRepository) Delete(ctx context.Context, id int) error {
	if err := r.db.WithContext(ctx).Delete(&model.Loan{}, "loan_id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
