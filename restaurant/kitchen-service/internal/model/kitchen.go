package model

type Kitchen struct {
	ID    int    `json:"id"`
	Foods []Food `gorm:"foreignKey:ID" json:"foods"`
}

type Food struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	CategoryID int      `json:"category_id"`
	Category   Category `gorm:"foreignKey:ID;references:CategoryID" json:"category"`
}

type Category struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}
