package application

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
)

type fakeStore struct {
	medicines map[int64]model.Medicine
	batches   map[int64]model.Batch
	movements []model.StockMovement
	rules     []model.InteractionRule
	dispenses []model.Dispense
	nextID    int64
}

func newFakeStore() *fakeStore {
	return &fakeStore{medicines: map[int64]model.Medicine{}, batches: map[int64]model.Batch{}}
}

func (s *fakeStore) id() int64 { s.nextID++; return s.nextID }

type fakeMedicines struct{ s *fakeStore }

func (r fakeMedicines) Create(_ context.Context, m *model.Medicine) error {
	for _, existing := range r.s.medicines {
		if existing.Code == m.Code {
			return fmt.Errorf("%w: medicine %s already exists", port.ErrConflict, m.Code)
		}
	}
	m.ID = r.s.id()
	r.s.medicines[m.ID] = *m
	return nil
}

func (r fakeMedicines) GetByID(_ context.Context, id int64) (*model.Medicine, error) {
	m, ok := r.s.medicines[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &m, nil
}

func (r fakeMedicines) GetByIDs(ctx context.Context, ids []int64) ([]model.Medicine, error) {
	var out []model.Medicine
	for _, id := range ids {
		m, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		out = append(out, *m)
	}
	return out, nil
}

func (r fakeMedicines) List(context.Context, string, int, int) ([]model.Medicine, int64, error) {
	return nil, 0, nil
}

func (r fakeMedicines) ListLowStock(context.Context, model.Date) ([]model.LowStockItem, error) {
	return nil, nil
}

type fakeBatches struct{ s *fakeStore }

func (r fakeBatches) Create(_ context.Context, b *model.Batch) error {
	b.ID = r.s.id()
	r.s.batches[b.ID] = *b
	return nil
}

func (r fakeBatches) Update(_ context.Context, b *model.Batch) error {
	r.s.batches[b.ID] = *b
	return nil
}

func (r fakeBatches) GetByIDForUpdate(_ context.Context, id int64) (*model.Batch, error) {
	b, ok := r.s.batches[id]
	if !ok {
		return nil, port.ErrNotFound
	}
	return &b, nil
}

func (r fakeBatches) FindByNumberForUpdate(_ context.Context, medicineID int64, number string) (*model.Batch, error) {
	for _, b := range r.s.batches {
		if b.MedicineID == medicineID && b.BatchNumber == number {
			return &b, nil
		}
	}
	return nil, port.ErrNotFound
}

func (r fakeBatches) filter(keep func(model.Batch) bool) []model.Batch {
	var out []model.Batch
	for _, b := range r.s.batches {
		if keep(b) {
			out = append(out, b)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out
}

func (r fakeBatches) ListByMedicine(_ context.Context, medicineID int64) ([]model.Batch, error) {
	return r.filter(func(b model.Batch) bool { return b.MedicineID == medicineID }), nil
}

func (r fakeBatches) ListUsableForUpdate(_ context.Context, medicineID int64, today model.Date) ([]model.Batch, error) {
	return r.filter(func(b model.Batch) bool { return b.MedicineID == medicineID && b.UsableOn(today) }), nil
}

func (r fakeBatches) ListExpiredForUpdate(_ context.Context, today model.Date) ([]model.Batch, error) {
	return r.filter(func(b model.Batch) bool { return b.Quantity > 0 && b.ExpiryDate <= today }), nil
}

func (r fakeBatches) ListExpiring(_ context.Context, today, until model.Date) ([]model.Batch, error) {
	return r.filter(func(b model.Batch) bool {
		return b.Quantity > 0 && b.ExpiryDate > today && b.ExpiryDate <= until
	}), nil
}

type fakeMovements struct{ s *fakeStore }

func (r fakeMovements) Create(_ context.Context, m []model.StockMovement) error {
	r.s.movements = append(r.s.movements, m...)
	return nil
}

func (r fakeMovements) ListByMedicine(context.Context, int64, int, int) ([]model.StockMovement, error) {
	return r.s.movements, nil
}

type fakeInteractions struct{ s *fakeStore }

func (r fakeInteractions) Upsert(_ context.Context, rule *model.InteractionRule) error {
	rule.ID = r.s.id()
	r.s.rules = append(r.s.rules, *rule)
	return nil
}

func (r fakeInteractions) List(context.Context, int, int) ([]model.InteractionRule, error) {
	return r.s.rules, nil
}

func (r fakeInteractions) FindAmong(context.Context, []string) ([]model.InteractionRule, error) {
	return r.s.rules, nil
}

type fakeDispenses struct{ s *fakeStore }

func (r fakeDispenses) Create(_ context.Context, d *model.Dispense) error {
	d.ID = r.s.id()
	r.s.dispenses = append(r.s.dispenses, *d)
	return nil
}

func (r fakeDispenses) GetByID(context.Context, int64) (*model.Dispense, error) {
	return nil, port.ErrNotFound
}

func (r fakeDispenses) ListByPatient(context.Context, string, int, int) ([]model.Dispense, error) {
	return r.s.dispenses, nil
}

type fakeTx struct{ s *fakeStore }

func (t fakeTx) WithinTx(ctx context.Context, fn func(ctx context.Context) error) error {
	batches := make(map[int64]model.Batch, len(t.s.batches))
	for k, v := range t.s.batches {
		batches[k] = v
	}
	movements, dispenses := len(t.s.movements), len(t.s.dispenses)
	if err := fn(ctx); err != nil {
		t.s.batches = batches
		t.s.movements = t.s.movements[:movements]
		t.s.dispenses = t.s.dispenses[:dispenses]
		return err
	}
	return nil
}

type fakeTerminology struct {
	aliases []model.IngredientAlias
	groups  []model.IngredientAllergenGroup
}

func (r *fakeTerminology) Load(context.Context) (model.Terminology, error) {
	return model.NewTerminology(r.aliases, r.groups), nil
}

func (r *fakeTerminology) UpsertAlias(_ context.Context, a *model.IngredientAlias) error {
	r.aliases = append(r.aliases, *a)
	return nil
}

func (r *fakeTerminology) AddAllergenGroup(_ context.Context, g *model.IngredientAllergenGroup) error {
	r.groups = append(r.groups, *g)
	return nil
}

func (r *fakeTerminology) ListAliases(context.Context) ([]model.IngredientAlias, error) {
	return r.aliases, nil
}

func (r *fakeTerminology) ListAllergenGroups(context.Context) ([]model.IngredientAllergenGroup, error) {
	return r.groups, nil
}

type fixedClock struct{ today model.Date }

func (c fixedClock) Now() time.Time    { return time.Date(2026, 9, 28, 9, 0, 0, 0, time.UTC) }
func (c fixedClock) Today() model.Date { return c.today }

type fixture struct {
	store        *fakeStore
	terms        *fakeTerminology
	terminology  port.TerminologyUsecase
	medicines    port.MedicineUsecase
	inventory    port.InventoryUsecase
	interactions port.InteractionUsecase
	dispenses    port.DispenseUsecase
}

func newFixture() *fixture {
	s := newFakeStore()
	clock := fixedClock{today: "2026-09-28"}
	meds, batches, movs := fakeMedicines{s}, fakeBatches{s}, fakeMovements{s}
	ints, disp, tx := fakeInteractions{s}, fakeDispenses{s}, fakeTx{s}
	terms := &fakeTerminology{}
	return &fixture{
		store:        s,
		terms:        terms,
		terminology:  NewTerminologyService(terms),
		medicines:    NewMedicineService(meds, terms),
		inventory:    NewInventoryService(meds, batches, movs, tx, clock),
		interactions: NewInteractionService(meds, ints, terms),
		dispenses:    NewDispenseService(meds, ints, batches, movs, disp, terms, tx, clock),
	}
}
