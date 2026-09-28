package model

import "time"

type TermSource string

const (
	SourceReference TermSource = "reference"

	SourceLocal TermSource = "local"
)

type IngredientAlias struct {
	ID        int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Alias     string     `json:"alias" gorm:"uniqueIndex"`
	Canonical string     `json:"canonical"`
	Source    TermSource `json:"source"`
	CreatedAt time.Time  `json:"created_at"`
}

func (IngredientAlias) TableName() string { return "ingredient_aliases" }

type IngredientAllergenGroup struct {
	ID            int64      `json:"id" gorm:"primaryKey;autoIncrement"`
	Ingredient    string     `json:"ingredient"`
	AllergenGroup string     `json:"allergen_group"`
	Source        TermSource `json:"source"`
	CreatedAt     time.Time  `json:"created_at"`
}

func (IngredientAllergenGroup) TableName() string { return "ingredient_allergen_groups" }

type Terminology struct {
	aliases map[string]string
	groups  map[string][]string
}

func NewTerminology(aliases []IngredientAlias, groups []IngredientAllergenGroup) Terminology {
	t := Terminology{
		aliases: make(map[string]string, len(aliases)),
		groups:  make(map[string][]string, len(groups)),
	}
	for _, a := range aliases {
		t.aliases[NormalizeTerm(a.Alias)] = NormalizeTerm(a.Canonical)
	}
	for _, g := range groups {
		ing := t.Canonical(g.Ingredient)
		t.groups[ing] = append(t.groups[ing], t.Canonical(g.AllergenGroup))
	}
	return t
}

func (t Terminology) Canonical(term string) string {
	term = NormalizeTerm(term)
	if c, ok := t.aliases[term]; ok {
		return c
	}
	return term
}

func (t Terminology) CanonicalTerms(terms []string) []string {
	out := make([]string, len(terms))
	for i, term := range terms {
		out[i] = t.Canonical(term)
	}
	return NormalizeTerms(out)
}

func (t Terminology) IsAlias(term string) bool {
	_, ok := t.aliases[NormalizeTerm(term)]
	return ok
}

func (t Terminology) IsCanonicalTarget(term string) bool {
	term = NormalizeTerm(term)
	for _, c := range t.aliases {
		if c == term {
			return true
		}
	}
	return false
}

func (t Terminology) Allergens(m Medicine) []string {
	ingredients := t.CanonicalTerms(m.Ingredients)
	terms := append([]string{}, ingredients...)
	terms = append(terms, t.CanonicalTerms(m.AllergenGroups)...)
	for _, ing := range ingredients {
		terms = append(terms, t.groups[ing]...)
	}
	return NormalizeTerms(terms)
}
