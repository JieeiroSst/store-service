package repository

import (
	"fmt"

	"github.com/JIeeiroSst/post-service/model"
	"gorm.io/gorm"
)

// applyDescCursor adds the keyset WHERE clause for a (created_at, id)
// DESC ordered query - see model.Cursor's doc comment for why this beats
// OFFSET pagination. A no-op (first page) when cursor is empty or fails
// to decode.
func applyDescCursor(tx *gorm.DB, cursor string, createdAtCol, idCol string) *gorm.DB {
	c, ok := model.DecodeCursor(cursor)
	if !ok {
		return tx
	}
	cond := fmt.Sprintf("(%s < ?) OR (%s = ? AND %s < ?)", createdAtCol, createdAtCol, idCol)
	return tx.Where(cond, c.CreatedAt, c.CreatedAt, c.ID)
}
