package repository

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/medical-service/internal/domain/model"
	"github.com/JIeeiroSst/medical-service/internal/domain/port"
	"gorm.io/gorm"
)

type medicineRepository struct {
	db *gorm.DB
}

func NewMedicineRepository(db *gorm.DB) port.MedicineRepository {
	return &medicineRepository{db: db}
}

func (r *medicineRepository) Create(ctx context.Context, m *model.Medicine) error {
	return translate(conn(ctx, r.db).Create(m).Error, "medicine "+m.Code)
}

func (r *medicineRepository) GetByID(ctx context.Context, id int64) (*model.Medicine, error) {
	var m model.Medicine
	if err := conn(ctx, r.db).First(&m, id).Error; err != nil {
		return nil, translate(err, fmt.Sprintf("medicine %d", id))
	}
	return &m, nil
}

func (r *medicineRepository) GetByIDs(ctx context.Context, ids []int64) ([]model.Medicine, error) {
	var found []model.Medicine
	if err := conn(ctx, r.db).Where("id IN ?", ids).Find(&found).Error; err != nil {
		return nil, err
	}
	byID := make(map[int64]model.Medicine, len(found))
	for _, m := range found {
		byID[m.ID] = m
	}

	out := make([]model.Medicine, 0, len(ids))
	seen := make(map[int64]bool, len(ids))
	for _, id := range ids {
		m, ok := byID[id]
		if !ok {
			return nil, fmt.Errorf("%w: medicine %d", port.ErrNotFound, id)
		}
		if !seen[id] {
			seen[id] = true
			out = append(out, m)
		}
	}
	return out, nil
}

func (r *medicineRepository) List(ctx context.Context, query string, limit, offset int) ([]model.Medicine, int64, error) {
	q := conn(ctx, r.db).Model(&model.Medicine{})
	if query != "" {
		like := "%" + query + "%"
		q = q.Where("name LIKE ? OR code LIKE ? OR ingredients LIKE ?", like, like, like)
	}

	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var items []model.Medicine
	if err := q.Order("name ASC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, 0, err
	}
	return items, total, nil
}

func (r *medicineRepository) ListLowStock(ctx context.Context, today model.Date) ([]model.LowStockItem, error) {
	var rows []struct {
		MedicineID int64
		Available  int
	}
	err := conn(ctx, r.db).Raw(`
		SELECT m.id AS medicine_id, COALESCE(SUM(b.quantity), 0) AS available
		FROM medicines m
		LEFT JOIN batches b ON b.medicine_id = m.id AND b.quantity > 0 AND b.expiry_date > ?
		GROUP BY m.id, m.reorder_level
		HAVING available <= m.reorder_level
		ORDER BY available ASC, m.id ASC`, today).Scan(&rows).Error
	if err != nil || len(rows) == 0 {
		return nil, err
	}

	ids := make([]int64, len(rows))
	for i, row := range rows {
		ids[i] = row.MedicineID
	}
	medicines, err := r.GetByIDs(ctx, ids)
	if err != nil {
		return nil, err
	}
	items := make([]model.LowStockItem, len(rows))
	for i, row := range rows {
		items[i] = model.LowStockItem{Medicine: medicines[i], Available: row.Available}
	}
	return items, nil
}
