package entity

type Account struct {
	AccountID      int    `json:"account_id" gorm:"column:account_id"`
	AccountNumber  string `json:"account_number" gorm:"column:account_number"`
	AccountType    string `json:"account_type" gorm:"column:account_type"`
	CurrentBalance uint   `json:"current_balance" gorm:"column:current_balance"`
	DateOpened     int    `json:"date_opened" gorm:"column:date_opened"`
	DateClosed     int    `json:"date_closed" gorm:"column:date_closed"`
	AccountStatus  string `json:"account_status" gorm:"column:account_status"`
}
