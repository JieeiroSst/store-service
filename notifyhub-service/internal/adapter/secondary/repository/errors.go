package repository

import (
	"errors"
	"fmt"

	"github.com/JIeeiroSst/notifyhub-service/internal/domain/port"
	"gorm.io/gorm"
)

func translate(err error, what string) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, gorm.ErrRecordNotFound):
		return fmt.Errorf("%w: %s", port.ErrNotFound, what)
	case errors.Is(err, gorm.ErrDuplicatedKey):
		return fmt.Errorf("%w: %s already exists", port.ErrConflict, what)
	default:
		return err
	}
}

func affected(res *gorm.DB, what string) error {
	if res.Error != nil {
		return translate(res.Error, what)
	}
	if res.RowsAffected == 0 {
		return fmt.Errorf("%w: %s", port.ErrNotFound, what)
	}
	return nil
}

func offset(page, size int) int {
	if page < 1 {
		page = 1
	}
	return (page - 1) * size
}
