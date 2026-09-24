package repository

import (
	"errors"

	"github.com/JIeeiroSst/catalogues-service/internal/domain/model"
	"gorm.io/gorm"
)

func translate(err, fkErr error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return model.ErrNotFound
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return model.ErrConflict
	case errors.Is(err, gorm.ErrForeignKeyViolated):
		return fkErr
	}
	return err
}

var errBadReference = model.Invalid("a referenced record does not exist")

func writeErr(err error) error  { return translate(err, errBadReference) }
func deleteErr(err error) error { return translate(err, model.ErrConflict) }

func deleteByID[T any](db *gorm.DB, id int64) error {
	res := db.Delete(new(T), id)
	if res.Error != nil {
		return deleteErr(res.Error)
	}
	if res.RowsAffected == 0 {
		return model.ErrNotFound
	}
	return nil
}
