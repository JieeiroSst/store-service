package repository

import (
	"context"
	"errors"

	"github.com/JIeeroSst/hospital-patient-management-service/internal/domain/model"
	"github.com/jackc/pgx/v5/pgconn"
	"gorm.io/gorm"
)

// PostgreSQL error classes we translate into domain errors.
const (
	pgUniqueViolation     = "23505"
	pgForeignKeyViolation = "23503"
)

// crud implements port.Store for any entity with an int32 primary key.
type crud[T any] struct {
	db *gorm.DB
	pk string // primary key column, e.g. "department_id"
}

func (r crud[T]) Create(ctx context.Context, v *T) error {
	return translate(r.db.WithContext(ctx).Create(v).Error, false)
}

func (r crud[T]) Get(ctx context.Context, id int32) (*T, error) {
	var v T
	if err := r.db.WithContext(ctx).Where(r.pk+" = ?", id).First(&v).Error; err != nil {
		return nil, translate(err, false)
	}
	return &v, nil
}

func (r crud[T]) Update(ctx context.Context, v *T) error {
	// Select("*") writes zero values too (clearing a field is a valid edit);
	// created_at is the one column an update must never touch.
	res := r.db.WithContext(ctx).Model(v).Select("*").Omit("created_at").Updates(v)
	if err := translate(res.Error, false); err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

func (r crud[T]) Delete(ctx context.Context, id int32) error {
	var v T
	res := r.db.WithContext(ctx).Where(r.pk+" = ?", id).Delete(&v)
	if err := translate(res.Error, true); err != nil {
		return err
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}

// translate maps driver errors to domain errors. A foreign key violation on
// write means the caller referenced something missing; on delete it means
// other rows still depend on this one.
func translate(err error, deleting bool) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return model.ErrNotFound
	}
	var pg *pgconn.PgError
	if errors.As(err, &pg) {
		switch pg.Code {
		case pgUniqueViolation:
			return model.Conflict("already exists (%s)", pg.ConstraintName)
		case pgForeignKeyViolation:
			if deleting {
				return model.Conflict("still referenced by other records")
			}
			return model.Invalid("referenced record does not exist")
		}
	}
	return err
}

func paginate[T any](q *gorm.DB, page model.Page, order string) ([]T, int64, error) {
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var out []T
	err := q.Order(order).Offset(page.Offset()).Limit(page.Limit()).Find(&out).Error
	return out, total, err
}
