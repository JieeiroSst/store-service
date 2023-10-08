package domain

type Partnership struct {
	ID          int       `json:"id" db:"id"`
	ProjectId   int       `json:"project_id" db:"project_id"`
	Description string    `json:"description" db:"description"`
	StartedOn   int       `json:"started_on" db:"started_on"`
	ExpiresOn   int       `json:"expires_on" db:"expires_on"`
	Projects    []Project `json:"projects"`
}

type PartnershipsPartner struct {
	ID             int           `json:"id" db:"id"`
	Partner_id     int           `json:"partner_id" db:"partner_id"`
	Partnership_id int           `json:"partnership_id" db:"partnership_id"`
	Content        string        `json:"content" db:"content"`
	JoinedOn       int           `json:"joined_on" db:"joined_on"`
	LeftOn         int           `json:"left_on" db:"left_on"`
	Partners       []Partner     `json:"partners"`
	Partnerships   []Partnership `json:"partnerships"`
}

type Partner struct {
	ID   int    `json:"id" db:"id"`
	Type string `json:"type" db:"type"`
}

type Project struct {
	ID   int    `json:"id" db:"id"`
	Name string `json:"name" db:"name"`
}
