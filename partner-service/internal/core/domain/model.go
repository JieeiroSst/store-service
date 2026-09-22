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

const (
	PartnerStatusActive   = "active"
	PartnerStatusInactive = "inactive"
	PartnerStatusPending  = "pending"
)

const (
	PartnerScoreMax                 = 100
	PartnerScoreThreshold           = 50
	PartnerScorePartnershipComplete = 10
)

type Partner struct {
	ID        string `json:"id" db:"id"`
	Type      string `json:"type" db:"type"`
	Name      string `json:"name" db:"name"`
	Email     string `json:"email" db:"email"`
	Phone     string `json:"phone" db:"phone"`
	Address   string `json:"address" db:"address"`
	Status    string `json:"status" db:"status"`
	Score     int    `json:"score" db:"score"`
	UserID    string `json:"user_id" db:"user_id"`
	CreatedAt int    `json:"created_at" db:"created_at"`
	UpdatedAt int    `json:"updated_at" db:"updated_at"`
	DeletedAt int    `json:"deleted_at" db:"deleted_at"`
}

type PartnerFilter struct {
	Type   string `json:"type" form:"type"`
	Status string `json:"status" form:"status"`
	Name   string `json:"name" form:"name"`
	UserID string `json:"user_id" form:"user_id"`
}

type PartnerStatistics struct {
	Total    int64            `json:"total"`
	ByStatus map[string]int64 `json:"by_status"`
	ByType   map[string]int64 `json:"by_type"`
}

type PartnerActivity struct {
	PartnerID          string `json:"partner_id"`
	Status             string `json:"status"`
	Score              int    `json:"score"`
	HasOpenPartnership bool   `json:"has_open_partnership"`
	LastActivityAt     int    `json:"last_activity_at"`
	EligibleForClosure bool   `json:"eligible_for_closure"`
	IsActive           bool   `json:"is_active"`
}

type Project struct {
	ID        string `json:"id" db:"id"`
	Name      string `json:"name" db:"name"`
	UserID    string `json:"user_id" db:"user_id"`
	CreatedAt int    `json:"created_at" db:"created_at"`
}
