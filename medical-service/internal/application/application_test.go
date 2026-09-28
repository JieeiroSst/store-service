package application

import (
	"context"
	"errors"
	"testing"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
)

var ctx = context.Background()

func (f *fixture) medicine(t *testing.T, in port.CreateMedicineInput) *model.Medicine {
	t.Helper()
	if in.Unit == "" {
		in.Unit = "tablet"
	}
	m, err := f.medicines.CreateMedicine(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	return m
}

func (f *fixture) receive(t *testing.T, medicineID int64, number string, expiry model.Date, qty int) *model.Batch {
	t.Helper()
	b, err := f.inventory.ReceiveBatch(ctx, port.ReceiveBatchInput{
		MedicineID: medicineID, BatchNumber: number, ExpiryDate: expiry, Quantity: qty, ReceivedBy: "kho",
	})
	if err != nil {
		t.Fatal(err)
	}
	return b
}

func TestCreateMedicineNormalizesAndValidates(t *testing.T) {
	f := newFixture()
	m := f.medicine(t, port.CreateMedicineInput{Code: " amox500 ", Name: "Amoxicillin 500mg", Ingredients: []string{" Amoxicillin ", "amoxicillin"}})
	if m.Code != "AMOX500" || len(m.Ingredients) != 1 || m.Ingredients[0] != "amoxicillin" {
		t.Errorf("not normalized: %+v", m)
	}

	_, err := f.medicines.CreateMedicine(ctx, port.CreateMedicineInput{Code: "X", Name: "X", Unit: "tablet"})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("missing ingredients: got %v", err)
	}
}

func TestReceiveBatchRejectsExpiredAndMergesSameLot(t *testing.T) {
	f := newFixture()
	m := f.medicine(t, port.CreateMedicineInput{Code: "P", Name: "Panadol", Ingredients: []string{"paracetamol"}})

	_, err := f.inventory.ReceiveBatch(ctx, port.ReceiveBatchInput{MedicineID: m.ID, BatchNumber: "L1", ExpiryDate: "2026-09-28", Quantity: 1, ReceivedBy: "kho"})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Fatalf("expired batch: got %v", err)
	}

	f.receive(t, m.ID, "L1", "2027-01-01", 10)
	b := f.receive(t, m.ID, "L1", "2027-01-01", 5)
	if b.Quantity != 15 || b.ReceivedQuantity != 15 || len(f.store.batches) != 1 {
		t.Errorf("same lot not merged: %+v", b)
	}

	_, err = f.inventory.ReceiveBatch(ctx, port.ReceiveBatchInput{MedicineID: m.ID, BatchNumber: "L1", ExpiryDate: "2027-02-01", Quantity: 1, ReceivedBy: "kho"})
	if !errors.Is(err, port.ErrConflict) {
		t.Errorf("same lot, different expiry: got %v", err)
	}
}

func TestDispenseUsesFEFOAndRecordsMovements(t *testing.T) {
	f := newFixture()
	m := f.medicine(t, port.CreateMedicineInput{Code: "P", Name: "Panadol", Ingredients: []string{"paracetamol"}})
	late := f.receive(t, m.ID, "LATE", "2027-06-01", 100)
	early := f.receive(t, m.ID, "EARLY", "2026-11-01", 4)

	d, err := f.dispenses.Dispense(ctx, port.DispenseInput{
		PatientRef: "BN-1", DispensedBy: "ds.lan",
		Items: []port.DispenseItemInput{{MedicineID: m.ID, Quantity: 10}},
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Items) != 2 || d.Items[0].BatchID != early.ID || d.Items[0].Quantity != 4 || d.Items[1].Quantity != 6 {
		t.Errorf("not FEFO: %+v", d.Items)
	}
	if f.store.batches[early.ID].Quantity != 0 || f.store.batches[late.ID].Quantity != 94 {
		t.Errorf("stock not decremented: %+v", f.store.batches)
	}
	var out int
	for _, mv := range f.store.movements {
		if mv.Type == model.MovementDispense {
			out += mv.Quantity
		}
	}
	if out != -10 {
		t.Errorf("dispense movements total %d, want -10", out)
	}
}

func TestDispenseIsAllOrNothing(t *testing.T) {
	f := newFixture()
	a := f.medicine(t, port.CreateMedicineInput{Code: "A", Name: "A", Ingredients: []string{"a"}})
	b := f.medicine(t, port.CreateMedicineInput{Code: "B", Name: "B", Ingredients: []string{"b"}})
	batchA := f.receive(t, a.ID, "LA", "2027-01-01", 10)
	f.receive(t, b.ID, "LB", "2027-01-01", 1)

	_, err := f.dispenses.Dispense(ctx, port.DispenseInput{
		PatientRef: "BN-1", DispensedBy: "ds.lan",
		Items: []port.DispenseItemInput{{MedicineID: a.ID, Quantity: 5}, {MedicineID: b.ID, Quantity: 2}},
	})
	if !errors.Is(err, port.ErrInsufficientStock) {
		t.Fatalf("got %v, want ErrInsufficientStock", err)
	}
	if f.store.batches[batchA.ID].Quantity != 10 || len(f.store.dispenses) != 0 {
		t.Error("partial dispense was not rolled back")
	}
}

func TestDispenseIgnoresExpiredStock(t *testing.T) {
	f := newFixture()
	m := f.medicine(t, port.CreateMedicineInput{Code: "P", Name: "P", Ingredients: []string{"p"}})
	f.store.batches[99] = model.Batch{ID: 99, MedicineID: m.ID, BatchNumber: "OLD", ExpiryDate: "2026-09-01", Quantity: 50}

	_, err := f.dispenses.Dispense(ctx, port.DispenseInput{
		PatientRef: "BN-1", DispensedBy: "ds.lan", Items: []port.DispenseItemInput{{MedicineID: m.ID, Quantity: 1}},
	})
	if !errors.Is(err, port.ErrInsufficientStock) {
		t.Errorf("got %v, want ErrInsufficientStock", err)
	}
}

func TestDispensePrescriptionAndSafetyGates(t *testing.T) {
	f := newFixture()
	amox := f.medicine(t, port.CreateMedicineInput{Code: "AMOX", Name: "Amoxicillin", Ingredients: []string{"amoxicillin"}, AllergenGroups: []string{"penicillin"}, RequiresPrescription: true})
	panadol := f.medicine(t, port.CreateMedicineInput{Code: "PANADOL", Name: "Panadol", Ingredients: []string{"paracetamol"}})
	effer := f.medicine(t, port.CreateMedicineInput{Code: "EFFER", Name: "Efferalgan", Ingredients: []string{"paracetamol"}})
	f.receive(t, amox.ID, "L1", "2027-01-01", 10)
	f.receive(t, panadol.ID, "L2", "2027-01-01", 10)

	base := port.DispenseInput{PatientRef: "BN-1", DispensedBy: "ds.lan"}

	in := base
	in.Items = []port.DispenseItemInput{{MedicineID: amox.ID, Quantity: 1}}
	if _, err := f.dispenses.Dispense(ctx, in); !errors.Is(err, port.ErrPrescriptionRequired) {
		t.Errorf("no prescription: got %v", err)
	}

	in.PrescriptionRef = "DT-9"
	in.Allergies = []string{"Penicillin"}
	in.OverrideReason = "doctor approved"
	_, err := f.dispenses.Dispense(ctx, in)
	var safety *port.SafetyError
	if !errors.Is(err, port.ErrSafetyBlocked) || !errors.As(err, &safety) || len(safety.Report.Warnings) != 1 {
		t.Errorf("allergy must block even with override: got %v", err)
	}

	in = base
	in.Items = []port.DispenseItemInput{{MedicineID: panadol.ID, Quantity: 1}}
	in.CurrentMedicineIDs = []int64{effer.ID}
	if _, err := f.dispenses.Dispense(ctx, in); !errors.Is(err, port.ErrOverrideRequired) {
		t.Errorf("duplicate paracetamol without reason: got %v", err)
	}
	in.OverrideReason = "patient stopped Efferalgan"
	d, err := f.dispenses.Dispense(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(d.Warnings) != 1 || d.OverrideReason == "" {
		t.Errorf("override not recorded: %+v", d)
	}
}

func TestAdjustAndDisposeExpired(t *testing.T) {
	f := newFixture()
	m := f.medicine(t, port.CreateMedicineInput{Code: "P", Name: "P", Ingredients: []string{"p"}})
	b := f.receive(t, m.ID, "L1", "2027-01-01", 10)
	f.store.batches[50] = model.Batch{ID: 50, MedicineID: m.ID, BatchNumber: "OLD", ExpiryDate: "2026-09-28", Quantity: 7}

	adjusted, err := f.inventory.AdjustBatch(ctx, port.AdjustBatchInput{BatchID: b.ID, NewQuantity: 8, Reason: "kiểm kê: vỡ 2 viên", AdjustedBy: "kho"})
	if err != nil || adjusted.Quantity != 8 {
		t.Fatalf("adjust: %+v %v", adjusted, err)
	}

	res, err := f.inventory.DisposeExpired(ctx, "kho")
	if err != nil {
		t.Fatal(err)
	}
	if res.Batches != 1 || res.Units != 7 || f.store.batches[50].Quantity != 0 || f.store.batches[b.ID].Quantity != 8 {
		t.Errorf("dispose = %+v, batches = %+v", res, f.store.batches)
	}

	stock, err := f.inventory.GetStock(ctx, m.ID)
	if err != nil || stock.Available != 8 || stock.Expired != 0 {
		t.Errorf("stock = %+v %v", stock, err)
	}
}

func TestListExpiringValidatesWindow(t *testing.T) {
	f := newFixture()
	if _, err := f.inventory.ListExpiring(ctx, 0); !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("got %v", err)
	}
}

func TestCreateInteractionRuleValidates(t *testing.T) {
	f := newFixture()
	rule, err := f.interactions.CreateRule(ctx, port.CreateInteractionInput{IngredientA: "Warfarin", IngredientB: "Aspirin", Severity: model.SeverityMajor})
	if err != nil {
		t.Fatal(err)
	}
	if rule.IngredientA != "aspirin" || rule.IngredientB != "warfarin" {
		t.Errorf("pair not normalized/ordered: %+v", rule)
	}
	_, err = f.interactions.CreateRule(ctx, port.CreateInteractionInput{IngredientA: "a", IngredientB: "A", Severity: model.SeverityMajor})
	if !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("same ingredient: got %v", err)
	}
}

func TestTerminologyAliasRules(t *testing.T) {
	f := newFixture()
	if _, err := f.terminology.AddAlias(ctx, "Aspirin", "acetylsalicylic acid"); err != nil {
		t.Fatal(err)
	}

	if _, err := f.terminology.AddAlias(ctx, "ASA", "aspirin"); !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("alias to an alias: got %v", err)
	}

	if _, err := f.terminology.AddAlias(ctx, "acetylsalicylic acid", "x"); !errors.Is(err, port.ErrInvalidInput) {
		t.Errorf("canonical turned into alias: got %v", err)
	}
}

func TestSynonymsAndReferenceGroupsFlowThroughDispensing(t *testing.T) {
	f := newFixture()
	f.terms.aliases = []model.IngredientAlias{{Alias: "acetaminophen", Canonical: "paracetamol"}}
	f.terms.groups = []model.IngredientAllergenGroup{{Ingredient: "cefalexin", AllergenGroup: "beta-lactam"}}

	tylenol := f.medicine(t, port.CreateMedicineInput{Code: "TYL", Name: "Tylenol", Ingredients: []string{"Acetaminophen"}})
	if tylenol.Ingredients[0] != "paracetamol" {
		t.Errorf("ingredient not stored canonically: %v", tylenol.Ingredients)
	}
	cef := f.medicine(t, port.CreateMedicineInput{Code: "CEF", Name: "Cefalexin", Ingredients: []string{"cefalexin"}})
	f.receive(t, cef.ID, "L1", "2027-01-01", 10)

	_, err := f.dispenses.Dispense(ctx, port.DispenseInput{
		PatientRef: "BN-1", DispensedBy: "ds", Allergies: []string{"beta-lactam"},
		Items: []port.DispenseItemInput{{MedicineID: cef.ID, Quantity: 1}},
	})
	if !errors.Is(err, port.ErrSafetyBlocked) {
		t.Errorf("reference allergen group not applied when dispensing: got %v", err)
	}

	rule, err := f.interactions.CreateRule(ctx, port.CreateInteractionInput{IngredientA: "acetaminophen", IngredientB: "warfarin", Severity: model.SeverityModerate})
	if err != nil || rule.IngredientA != "paracetamol" || rule.Source != model.SourceLocal {
		t.Errorf("rule not canonical/local: %+v %v", rule, err)
	}
}
