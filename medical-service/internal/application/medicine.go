package application

import (
	"context"
	"strings"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
)

type medicineService struct {
	medicines   port.MedicineRepository
	terminology port.TerminologyRepository
}

func NewMedicineService(medicines port.MedicineRepository, terminology port.TerminologyRepository) port.MedicineUsecase {
	return &medicineService{medicines: medicines, terminology: terminology}
}

func (s *medicineService) CreateMedicine(ctx context.Context, in port.CreateMedicineInput) (*model.Medicine, error) {
	term, err := s.terminology.Load(ctx)
	if err != nil {
		return nil, err
	}
	m := &model.Medicine{
		Code:                 strings.ToUpper(strings.TrimSpace(in.Code)),
		Name:                 strings.TrimSpace(in.Name),
		Ingredients:          term.CanonicalTerms(in.Ingredients),
		AllergenGroups:       term.CanonicalTerms(in.AllergenGroups),
		DosageForm:           strings.TrimSpace(in.DosageForm),
		Strength:             strings.TrimSpace(in.Strength),
		Unit:                 strings.TrimSpace(in.Unit),
		RequiresPrescription: in.RequiresPrescription,
		ReorderLevel:         in.ReorderLevel,
	}
	switch {
	case m.Code == "":
		return nil, invalid("code is required")
	case m.Name == "":
		return nil, invalid("name is required")
	case len(m.Ingredients) == 0:
		return nil, invalid("at least one active ingredient is required")
	case m.Unit == "":
		return nil, invalid("unit is required")
	case m.ReorderLevel < 0:
		return nil, invalid("reorder_level must not be negative")
	}

	if err := s.medicines.Create(ctx, m); err != nil {
		return nil, err
	}
	return m, nil
}

func (s *medicineService) GetMedicine(ctx context.Context, id int64) (*model.Medicine, error) {
	return s.medicines.GetByID(ctx, id)
}

func (s *medicineService) ListMedicines(ctx context.Context, query string, in port.ListInput) (*port.MedicineList, error) {
	limit, offset := normalizeList(in)
	items, total, err := s.medicines.List(ctx, strings.TrimSpace(query), limit, offset)
	if err != nil {
		return nil, err
	}
	return &port.MedicineList{Items: items, Total: total}, nil
}
