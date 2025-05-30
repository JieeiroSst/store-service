package domain

type PartnershipsPartner struct {
	ID            string        `json:"id" db:"id"`
	PartnerId     string        `json:"partner_id" db:"partner_id"`
	PartnershipId string        `json:"partnership_id" db:"partnership_id"`
	Content       string        `json:"content" db:"content"`
	JoinedOn      int           `json:"joined_on" db:"joined_on"`
	LeftOn        int           `json:"left_on" db:"left_on"`
	Partners      []Partner     `json:"partners" gorm:"foreignKey:ID;references:PartnerId"`
	Partnerships  []Partnership `json:"partnerships" gorm:"foreignKey:ID;references:PartnershipId"`
	UserID        string        `json:"user_id" db:"user_id"`
	CreatedAt     int           `json:"created_at" db:"created_at"`
}
type Partnership struct {
	ID          string    `json:"id" db:"id"`
	ProjectId   string    `json:"project_id" db:"project_id"`
	Description string    `json:"description" db:"description"`
	StartedOn   int       `json:"started_on" db:"started_on"`
	ExpiresOn   int       `json:"expires_on" db:"expires_on"`
	Projects    []Project `json:"projects" gorm:"foreignKey:ID;references:ProjectId"`
	UserID      string    `json:"user_id" db:"user_id"`
	CreatedAt   int       `json:"created_at" db:"created_at"`
}

type Partner struct {
	ID        string `json:"id" db:"id"`
	Type      string `json:"type" db:"type"`
	UserID    string `json:"user_id" db:"user_id"`
	CreatedAt int    `json:"created_at" db:"created_at"`
}

type Project struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	UserID    string `json:"user_id" db:"user_id"`
	CreatedAt int    `json:"created_at" db:"created_at"`
}
