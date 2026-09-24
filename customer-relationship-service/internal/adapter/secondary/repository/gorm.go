package repository

import (
	"context"
	"errors"
	"strings"

	"github.com/JIeeiroSst/customer-relationship-service/common"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/model"
	"github.com/JIeeiroSst/customer-relationship-service/internal/domain/port"
	"gorm.io/gorm"
)

type txKey struct{}

func conn(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx.WithContext(ctx)
	}
	return db.WithContext(ctx)
}

type gormRepository[T any] struct {
	db *gorm.DB
}

func New[T any](db *gorm.DB) port.Repository[T] {
	return &gormRepository[T]{db: db}
}

func (r *gormRepository[T]) Create(ctx context.Context, entity *T) error {
	if err := conn(ctx, r.db).Create(entity).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *gormRepository[T]) GetByID(ctx context.Context, id uint) (*T, error) {
	var entity T
	if err := conn(ctx, r.db).First(&entity, "id = ?", id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.ErrNotFound
		}
		return nil, common.ErrDBFailed
	}
	return &entity, nil
}

func (r *gormRepository[T]) List(ctx context.Context, q port.ListQuery) ([]T, int64, error) {
	tx := conn(ctx, r.db).Model(new(T))

	if len(q.Equals) > 0 {
		tx = tx.Where(q.Equals)
	}
	for col, vals := range q.Exclude {
		if len(vals) > 0 {
			tx = tx.Where(col+" NOT IN ?", vals)
		}
	}
	if q.Search != "" {
		if s, ok := any(new(T)).(model.Searchable); ok {
			var conds []string
			var args []any
			for _, col := range s.SearchColumns() {
				conds = append(conds, col+" LIKE ?")
				args = append(args, "%"+q.Search+"%")
			}
			tx = tx.Where("("+strings.Join(conds, " OR ")+")", args...)
		}
	}

	var total int64
	if err := tx.Count(&total).Error; err != nil {
		return nil, 0, common.ErrDBFailed
	}

	entities := []T{}
	if err := tx.Order("id").Offset(q.Offset).Limit(q.Limit).Find(&entities).Error; err != nil {
		return nil, 0, common.ErrDBFailed
	}
	return entities, total, nil
}

func (r *gormRepository[T]) Update(ctx context.Context, id uint, entity *T) error {
	err := conn(ctx, r.db).Model(new(T)).Where("id = ?", id).
		Select("*").Omit("id", "created_at").Updates(entity).Error
	if err != nil {
		return common.ErrDBFailed
	}
	return nil
}

func (r *gormRepository[T]) Delete(ctx context.Context, id uint) error {
	if err := conn(ctx, r.db).Delete(new(T), "id = ?", id).Error; err != nil {
		return common.ErrDBFailed
	}
	return nil
}
