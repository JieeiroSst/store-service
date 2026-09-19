package model

// PrepStatus tracks a Kitchen ticket (an order translated into work for the
// kitchen, created from the "kitchen.create" event) through preparation.
const (
	PrepStatusReceived  = "received"
	PrepStatusPreparing = "preparing"
	PrepStatusReady     = "ready"
	PrepStatusServed    = "served"
)

type Kitchen struct {
	ID     int    `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Foods  []Food `gorm:"foreignKey:ID" json:"foods"`
}

type Food struct {
	ID         int      `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name       string   `json:"name"`
	CategoryID int      `json:"category_id"`
	Price      float64  `json:"price"`
	Status     int      `json:"status"`
	Category   Category `gorm:"foreignKey:CategoryID" json:"category,omitempty"`
}

type Category struct {
	ID   int    `gorm:"primaryKey;autoIncrement:false" json:"id"`
	Name string `json:"name"`
}
