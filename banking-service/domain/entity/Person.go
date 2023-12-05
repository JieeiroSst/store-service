package entity

type Person struct {
	PersonID      int    `json:"person_id" gorm:"column:person_id"`
	LastName      string `json:"last_name" gorm:"column:last_name"`
	FirstName     string `json:"first_name" gorm:"column:first_name"`
	DateOfBirth   int    `json:"date_of_birth" gorm:"column:date_of_birth"`
	Email         string `json:"email" gorm:"column:email"`
	PhoneNumber   string `json:"phone_number" gorm:"column:phone_number"`
	Address       string `json:"address" gorm:"column:address"`
	TaxIdentifier string `json:"tax_identifier" gorm:"column:tax_identifier"`
}
