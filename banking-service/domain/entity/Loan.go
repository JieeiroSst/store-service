package entity

type Loan struct {
	LoanID      int    `json:"loan_id" gorm:"column:loan_id"`
	LoanType    string `json:"loan_type" gorm:"column:loan_type"`
	LoanAmount  uint   `json:"loan_amount" gorm:"column:loan_amount"`
	InteresRate uint   `json:"interes_rate" gorm:"column:interes_rate"`
	Term        int    `json:"term" gorm:"column:term"`
	StartDate   int    `json:"start_date" gorm:"column:start_date"`
	EndDate     int    `json:"end_date" gorm:"column:end_date"`
	LoanStatus  string `json:"loan_status" gorm:"column:loan_status"`
}

func (Loan) TableName() string {
	return "loans"
}