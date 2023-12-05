package entity

type Employee struct {
	EmployeeID int    `json:"employee_id" gorm:"column:employee_id"`
	Position   string `json:"position" gorm:"column:position"`
}

func (Employee) TableName() string {
	return "employees"
}