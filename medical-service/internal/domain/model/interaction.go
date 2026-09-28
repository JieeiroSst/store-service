package model

import "time"

type Severity string

const (
	SeverityMinor           Severity = "MINOR"
	SeverityModerate        Severity = "MODERATE"
	SeverityMajor           Severity = "MAJOR"
	SeverityContraindicated Severity = "CONTRAINDICATED"
)

func (s Severity) Valid() bool {
	switch s {
	case SeverityMinor, SeverityModerate, SeverityMajor, SeverityContraindicated:
		return true
	}
	return false
}

type InteractionRule struct {
	ID          int64    `json:"id" gorm:"primaryKey;autoIncrement"`
	IngredientA string   `json:"ingredient_a"`
	IngredientB string   `json:"ingredient_b"`
	Severity    Severity `json:"severity"`
	Description string   `json:"description"`

	Source    TermSource `json:"source"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
}

func (InteractionRule) TableName() string { return "interaction_rules" }

func IngredientPair(a, b string) (string, string) {
	a, b = NormalizeTerm(a), NormalizeTerm(b)
	if a > b {
		a, b = b, a
	}
	return a, b
}
