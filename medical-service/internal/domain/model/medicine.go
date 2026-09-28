package model

import (
	"strings"
	"time"
)

type Medicine struct {
	ID   int64  `json:"id" gorm:"primaryKey;autoIncrement"`
	Code string `json:"code" gorm:"uniqueIndex"`
	Name string `json:"name"`

	Ingredients []string `json:"ingredients" gorm:"serializer:json;type:text"`

	AllergenGroups       []string  `json:"allergen_groups" gorm:"serializer:json;type:text"`
	DosageForm           string    `json:"dosage_form"`
	Strength             string    `json:"strength"`
	Unit                 string    `json:"unit"`
	RequiresPrescription bool      `json:"requires_prescription"`
	ReorderLevel         int       `json:"reorder_level"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

func (Medicine) TableName() string { return "medicines" }

func NormalizeTerm(s string) string {
	return strings.Join(strings.Fields(strings.ToLower(s)), " ")
}

func NormalizeTerms(terms []string) []string {
	seen := make(map[string]bool, len(terms))
	out := make([]string, 0, len(terms))
	for _, t := range terms {
		t = NormalizeTerm(t)
		if t == "" || seen[t] {
			continue
		}
		seen[t] = true
		out = append(out, t)
	}
	return out
}

type LowStockItem struct {
	Medicine  Medicine `json:"medicine"`
	Available int      `json:"available"`
}
