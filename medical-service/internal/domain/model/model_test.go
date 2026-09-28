package model

import (
	"testing"
	"time"
)

func TestAllocateFEFOTakesEarliestExpiryFirst(t *testing.T) {
	today := Date("2026-09-28")
	batches := []Batch{
		{ID: 1, ExpiryDate: "2027-06-01", Quantity: 100},
		{ID: 2, ExpiryDate: "2026-12-01", Quantity: 5},
		{ID: 3, ExpiryDate: "2026-09-28", Quantity: 50},
		{ID: 4, ExpiryDate: "2026-10-15", Quantity: 0},
		{ID: 5, ExpiryDate: "2026-12-01", Quantity: 10},
	}

	allocs, shortfall := AllocateFEFO(batches, 12, today)
	if shortfall != 0 {
		t.Fatalf("shortfall = %d, want 0", shortfall)
	}
	want := []struct {
		id  int64
		qty int
	}{{2, 5}, {5, 7}}
	if len(allocs) != len(want) {
		t.Fatalf("allocs = %+v", allocs)
	}
	for i, w := range want {
		if allocs[i].Batch.ID != w.id || allocs[i].Quantity != w.qty {
			t.Errorf("alloc[%d] = batch %d x%d, want batch %d x%d", i, allocs[i].Batch.ID, allocs[i].Quantity, w.id, w.qty)
		}
	}
}

func TestAllocateFEFOReportsShortfall(t *testing.T) {
	_, shortfall := AllocateFEFO([]Batch{{ID: 1, ExpiryDate: "2027-01-01", Quantity: 3}}, 10, "2026-09-28")
	if shortfall != 7 {
		t.Errorf("shortfall = %d, want 7", shortfall)
	}
}

func TestSummarizeStockSplitsExpired(t *testing.T) {
	s := SummarizeStock(1, []Batch{
		{ExpiryDate: "2027-01-01", Quantity: 10},
		{ExpiryDate: "2026-09-01", Quantity: 4},
		{ExpiryDate: "2026-09-28", Quantity: 1},
	}, "2026-09-28")
	if s.Available != 10 || s.Expired != 5 {
		t.Errorf("available=%d expired=%d, want 10/5", s.Available, s.Expired)
	}
}

func TestDate(t *testing.T) {
	if _, err := ParseDate("2026-02-30"); err == nil {
		t.Error("impossible date accepted")
	}
	if d := Date("2026-12-25").AddDays(10); d != "2027-01-04" {
		t.Errorf("AddDays = %s", d)
	}
	var d Date
	if err := d.Scan(time.Date(2027, 3, 1, 0, 0, 0, 0, time.UTC)); err != nil || d != "2027-03-01" {
		t.Errorf("scan time = %q, %v", d, err)
	}
	if err := d.Scan([]byte("2027-04-02")); err != nil || d != "2027-04-02" {
		t.Errorf("scan bytes = %q, %v", d, err)
	}
}

var (
	amoxicillin = Medicine{ID: 1, Code: "AMOX500", Ingredients: []string{"amoxicillin"}, AllergenGroups: []string{"Penicillin", "beta-lactam"}}
	panadol     = Medicine{ID: 2, Code: "PANADOL", Ingredients: []string{"paracetamol"}}
	efferalgan  = Medicine{ID: 3, Code: "EFFER", Ingredients: []string{"Paracetamol"}}
	warfarin    = Medicine{ID: 4, Code: "WARF", Ingredients: []string{"warfarin"}}
	aspirin     = Medicine{ID: 5, Code: "ASPIRIN", Ingredients: []string{"acetylsalicylic acid"}}
)

func TestEvaluateSafetyAllergyByGroupBlocks(t *testing.T) {
	r := EvaluateSafety([]Medicine{amoxicillin}, nil, []string{" PENICILLIN "}, nil, Terminology{})
	if !r.Blocked() || len(r.Warnings) != 1 || r.Warnings[0].Type != WarningAllergy {
		t.Fatalf("report = %+v", r)
	}
}

func TestEvaluateSafetyDuplicateIngredientNeedsOverride(t *testing.T) {
	r := EvaluateSafety([]Medicine{panadol}, []Medicine{efferalgan}, nil, nil, Terminology{})
	if r.Blocked() || !r.NeedsOverride() {
		t.Fatalf("report = %+v", r)
	}
	if r.Warnings[0].Type != WarningDuplicateIngredient {
		t.Errorf("type = %s", r.Warnings[0].Type)
	}
}

func TestEvaluateSafetyInteractionUsesRuleSeverity(t *testing.T) {
	rules := []InteractionRule{{IngredientA: "Warfarin", IngredientB: "acetylsalicylic acid", Severity: SeverityMajor, Description: "bleeding risk"}}
	r := EvaluateSafety([]Medicine{aspirin, warfarin}, nil, nil, rules, Terminology{})
	if len(r.Warnings) != 1 || r.Warnings[0].Severity != SeverityMajor || r.Warnings[0].Description != "bleeding risk" {
		t.Fatalf("report = %+v", r)
	}

	if r := EvaluateSafety([]Medicine{aspirin, panadol}, nil, nil, rules, Terminology{}); len(r.Warnings) != 0 {
		t.Errorf("unrelated rule produced warnings: %+v", r)
	}
}

func TestEvaluateSafetyIgnoresSameMedicineInCurrentList(t *testing.T) {
	r := EvaluateSafety([]Medicine{panadol}, []Medicine{panadol}, nil, nil, Terminology{})
	if len(r.Warnings) != 0 {
		t.Errorf("refill of the same medicine flagged as duplicate: %+v", r)
	}
}

var refTerms = NewTerminology(
	[]IngredientAlias{{Alias: "Aspirin", Canonical: "acetylsalicylic acid"}, {Alias: "acetaminophen", Canonical: "paracetamol"}},
	[]IngredientAllergenGroup{{Ingredient: "amoxicillin", AllergenGroup: "penicillin"}, {Ingredient: "amoxicillin", AllergenGroup: "beta-lactam"}},
)

func TestReferenceGroupCatchesUndeclaredAllergy(t *testing.T) {
	undeclared := Medicine{ID: 9, Code: "AMOX-GEN", Ingredients: []string{"Amoxicillin"}}
	if r := EvaluateSafety([]Medicine{undeclared}, nil, []string{"penicillin"}, nil, Terminology{}); r.Blocked() {
		t.Fatal("without terminology the undeclared group should not match (baseline)")
	}
	if r := EvaluateSafety([]Medicine{undeclared}, nil, []string{"Beta-Lactam"}, nil, refTerms); !r.Blocked() {
		t.Errorf("reference group not applied: %+v", r)
	}
}

func TestSynonymsMatchAllergyDuplicateAndRules(t *testing.T) {
	tylenol := Medicine{ID: 10, Code: "TYLENOL", Ingredients: []string{"Acetaminophen"}}
	if r := EvaluateSafety([]Medicine{tylenol}, []Medicine{panadol}, nil, nil, refTerms); !r.NeedsOverride() {
		t.Errorf("acetaminophen + paracetamol not seen as duplicate: %+v", r)
	}
	if r := EvaluateSafety([]Medicine{aspirin}, nil, []string{"aspirin"}, nil, refTerms); !r.Blocked() {
		t.Errorf("allergy to 'aspirin' missed acetylsalicylic acid: %+v", r)
	}
	rules := []InteractionRule{{IngredientA: "aspirin", IngredientB: "warfarin", Severity: SeverityMajor}}
	if r := EvaluateSafety([]Medicine{aspirin, warfarin}, nil, nil, rules, refTerms); len(r.Warnings) != 1 {
		t.Errorf("rule written with an alias not applied: %+v", r)
	}
}
