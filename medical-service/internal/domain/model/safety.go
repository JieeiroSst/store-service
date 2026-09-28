package model

import "fmt"

type WarningType string

const (
	WarningAllergy             WarningType = "ALLERGY"
	WarningInteraction         WarningType = "INTERACTION"
	WarningDuplicateIngredient WarningType = "DUPLICATE_INGREDIENT"
)

type Warning struct {
	Type        WarningType `json:"type"`
	Severity    Severity    `json:"severity"`
	Medicines   []string    `json:"medicines"`
	Ingredients []string    `json:"ingredients"`
	Description string      `json:"description"`
}

type SafetyReport struct {
	Warnings []Warning `json:"warnings"`
}

func (r SafetyReport) Blocked() bool {
	for _, w := range r.Warnings {
		if w.Severity == SeverityContraindicated {
			return true
		}
	}
	return false
}

func (r SafetyReport) NeedsOverride() bool {
	for _, w := range r.Warnings {
		if w.Severity == SeverityMajor {
			return true
		}
	}
	return false
}

func EvaluateSafety(dispensing, current []Medicine, allergies []string, rules []InteractionRule, term Terminology) SafetyReport {
	report := SafetyReport{Warnings: []Warning{}}

	allergySet := toSet(term.CanonicalTerms(allergies))
	for _, m := range dispensing {
		for _, allergen := range term.Allergens(m) {
			if allergySet[allergen] {
				report.Warnings = append(report.Warnings, Warning{
					Type:        WarningAllergy,
					Severity:    SeverityContraindicated,
					Medicines:   []string{m.Code},
					Ingredients: []string{allergen},
					Description: fmt.Sprintf("patient is allergic to %s", allergen),
				})
			}
		}
	}

	ruleIndex := make(map[[2]string]InteractionRule, len(rules))
	for _, r := range rules {
		a, b := IngredientPair(term.Canonical(r.IngredientA), term.Canonical(r.IngredientB))
		ruleIndex[[2]string{a, b}] = r
	}

	checkPair := func(x, y Medicine) {
		for _, ix := range term.CanonicalTerms(x.Ingredients) {
			for _, iy := range term.CanonicalTerms(y.Ingredients) {
				if ix == iy {
					report.Warnings = append(report.Warnings, Warning{
						Type:        WarningDuplicateIngredient,
						Severity:    SeverityMajor,
						Medicines:   []string{x.Code, y.Code},
						Ingredients: []string{ix},
						Description: fmt.Sprintf("both medicines contain %s; risk of overdose", ix),
					})
					continue
				}
				a, b := IngredientPair(ix, iy)
				if r, ok := ruleIndex[[2]string{a, b}]; ok {
					report.Warnings = append(report.Warnings, Warning{
						Type:        WarningInteraction,
						Severity:    r.Severity,
						Medicines:   []string{x.Code, y.Code},
						Ingredients: []string{a, b},
						Description: r.Description,
					})
				}
			}
		}
	}

	for i := range dispensing {
		for j := i + 1; j < len(dispensing); j++ {
			checkPair(dispensing[i], dispensing[j])
		}
		for _, c := range current {
			if c.ID != dispensing[i].ID {
				checkPair(dispensing[i], c)
			}
		}
	}
	return report
}

func toSet(items []string) map[string]bool {
	set := make(map[string]bool, len(items))
	for _, it := range items {
		set[it] = true
	}
	return set
}
