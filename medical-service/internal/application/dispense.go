package application

import (
	"context"
	"fmt"
	"sort"
	"strings"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
)

type dispenseService struct {
	medicines    port.MedicineRepository
	interactions port.InteractionRepository
	batches      port.BatchRepository
	movements    port.MovementRepository
	dispenses    port.DispenseRepository
	terminology  port.TerminologyRepository
	tx           port.TxManager
	clock        port.Clock
}

func NewDispenseService(
	medicines port.MedicineRepository,
	interactions port.InteractionRepository,
	batches port.BatchRepository,
	movements port.MovementRepository,
	dispenses port.DispenseRepository,
	terminology port.TerminologyRepository,
	tx port.TxManager,
	clock port.Clock,
) port.DispenseUsecase {
	return &dispenseService{
		medicines:    medicines,
		interactions: interactions,
		batches:      batches,
		movements:    movements,
		dispenses:    dispenses,
		terminology:  terminology,
		tx:           tx,
		clock:        clock,
	}
}

func (s *dispenseService) Dispense(ctx context.Context, in port.DispenseInput) (*model.Dispense, error) {
	in.PatientRef = strings.TrimSpace(in.PatientRef)
	in.PrescriptionRef = strings.TrimSpace(in.PrescriptionRef)
	in.DispensedBy = strings.TrimSpace(in.DispensedBy)
	in.OverrideReason = strings.TrimSpace(in.OverrideReason)
	if err := validateDispense(in); err != nil {
		return nil, err
	}

	ids := make([]int64, len(in.Items))
	for i, it := range in.Items {
		ids[i] = it.MedicineID
	}
	medicines, err := s.medicines.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	for _, m := range medicines {
		if m.RequiresPrescription && in.PrescriptionRef == "" {
			return nil, fmt.Errorf("%w: %s (%s)", port.ErrPrescriptionRequired, m.Name, m.Code)
		}
	}

	report, err := evaluate(ctx, s.medicines, s.interactions, s.terminology, ids, in.CurrentMedicineIDs, in.Allergies)
	if err != nil {
		return nil, err
	}
	if report.Blocked() {
		return nil, &port.SafetyError{Reason: port.ErrSafetyBlocked, Report: *report}
	}
	if report.NeedsOverride() && in.OverrideReason == "" {
		return nil, &port.SafetyError{Reason: port.ErrOverrideRequired, Report: *report}
	}

	codes := make(map[int64]string, len(medicines))
	for _, m := range medicines {
		codes[m.ID] = m.Code
	}
	items := append([]port.DispenseItemInput{}, in.Items...)
	sort.Slice(items, func(i, j int) bool { return items[i].MedicineID < items[j].MedicineID })

	dispense := &model.Dispense{
		PatientRef:      in.PatientRef,
		PrescriptionRef: in.PrescriptionRef,
		DispensedBy:     in.DispensedBy,
		OverrideReason:  in.OverrideReason,
		Warnings:        report.Warnings,
	}
	today := s.clock.Today()

	err = s.tx.WithinTx(ctx, func(ctx context.Context) error {
		var touched []*model.Batch
		for _, it := range items {
			batches, err := s.batches.ListUsableForUpdate(ctx, it.MedicineID, today)
			if err != nil {
				return err
			}
			allocs, shortfall := model.AllocateFEFO(batches, it.Quantity, today)
			if shortfall > 0 {
				return fmt.Errorf("%w: %s needs %d, only %d available",
					port.ErrInsufficientStock, codes[it.MedicineID], it.Quantity, it.Quantity-shortfall)
			}
			for _, a := range allocs {
				b := a.Batch
				b.Quantity -= a.Quantity
				touched = append(touched, &b)
				dispense.Items = append(dispense.Items, model.DispenseItem{
					MedicineID:  it.MedicineID,
					BatchID:     b.ID,
					BatchNumber: b.BatchNumber,
					ExpiryDate:  b.ExpiryDate,
					Quantity:    a.Quantity,
				})
			}
		}

		for _, b := range touched {
			if err := s.batches.Update(ctx, b); err != nil {
				return err
			}
		}
		if err := s.dispenses.Create(ctx, dispense); err != nil {
			return err
		}

		movements := make([]model.StockMovement, 0, len(dispense.Items))
		for _, item := range dispense.Items {
			movements = append(movements, model.StockMovement{
				MedicineID: item.MedicineID,
				BatchID:    item.BatchID,
				Type:       model.MovementDispense,
				Quantity:   -item.Quantity,
				Reference:  fmt.Sprintf("dispense:%d", dispense.ID),
				CreatedBy:  in.DispensedBy,
			})
		}
		return s.movements.Create(ctx, movements)
	})
	if err != nil {
		return nil, err
	}
	return dispense, nil
}

func validateDispense(in port.DispenseInput) error {
	switch {
	case in.PatientRef == "":
		return invalid("patient_ref is required")
	case in.DispensedBy == "":
		return invalid("dispensed_by is required")
	case len(in.Items) == 0:
		return invalid("items is required")
	}
	seen := make(map[int64]bool, len(in.Items))
	for _, it := range in.Items {
		if it.Quantity <= 0 {
			return invalid("quantity must be greater than zero (medicine %d)", it.MedicineID)
		}
		if seen[it.MedicineID] {
			return invalid("medicine %d is listed more than once", it.MedicineID)
		}
		seen[it.MedicineID] = true
	}
	return nil
}

func (s *dispenseService) GetDispense(ctx context.Context, id int64) (*model.Dispense, error) {
	return s.dispenses.GetByID(ctx, id)
}

func (s *dispenseService) ListByPatient(ctx context.Context, patientRef string, in port.ListInput) ([]model.Dispense, error) {
	patientRef = strings.TrimSpace(patientRef)
	if patientRef == "" {
		return nil, invalid("patient_ref is required")
	}
	limit, offset := normalizeList(in)
	return s.dispenses.ListByPatient(ctx, patientRef, limit, offset)
}
