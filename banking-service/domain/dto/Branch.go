package entity

type Branch struct {
	BranchID    int    `json:"branch_id" gorm:"column:branch_id"`
	BranchName  string `json:"branch_name" gorm:"column:branch_name"`
	BranchCode  string `json:"branch_code" gorm:"column:branch_code"`
	Address     string `json:"address" gorm:"column:address"`
	PhoneNumber string `json:"phone_number" gorm:"column:phone_number"`
}
