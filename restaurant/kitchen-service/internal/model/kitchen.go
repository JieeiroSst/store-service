package model

type Kitchen struct {
	ID    int    `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name  string `json:"name"`
	Foods []Food `gorm:"foreignKey:ID" json:"foods"`
}

type Food struct {
	ID         int      `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name       string   `json:"name"`
	CategoryID int      `json:"category_id"`
	Status     int      `json:"status"`
	Category   Category `gorm:"foreignKey:CategoryID"`
}

type Category struct {
	ID   int    `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name string `json:"name"`
}
