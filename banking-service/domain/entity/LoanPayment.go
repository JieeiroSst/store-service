package entity

type LoanPayment struct {
	LoanPaymentID        int  `json:"loan_payment_id" gorm:"column:loan_payment_id"`
	ScheduledPaymentDate int  `json:"scheduled_payment_date" gorm:"column:scheduled_payment_date"`
	PaymentAccount       uint `json:"payment_account" gorm:"column:payment_account"`
	PrincipalAmount      uint `json:"principal_amount" gorm:"column:principal_amount"`
	InterestAmount       uint `json:"interest_amount" gorm:"column:interest_amount"`
	PaidAmount           uint `json:"paid_amount" gorm:"column:paid_amount"`
	PaidDate             int  `json:"paid_date" gorm:"column:paid_date"`
}

func (LoanPayment) TableName() string {
	return "loan_payments"
}