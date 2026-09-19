package repository

import (
	"fmt"

	"github.com/JIeeiroSst/threads-service/internal/domain/model"
	"gorm.io/gorm"
)

// applyDescCursor adds the keyset WHERE clause for a (created_at, id)
// DESC ordered query - see model.Cursor's doc comment for why this beats
// OFFSET pagination. A no-op (first page) when cursor is empty or fails
// to decode.
//
// createdAtCol/idCol take the column names to compare against - just
// "created_at"/"id" for a single-table query, but table-qualified
// ("posts.created_at"/"posts.id") wherever the query joins another table
// that also has columns with those names, to avoid an ambiguous-column
// error from Postgres.
func applyDescCursor(tx *gorm.DB, cursor string, createdAtCol, idCol string) *gorm.DB {
	c, ok := model.DecodeCursor(cursor)
	if !ok {
		return tx
	}
	cond := fmt.Sprintf("(%s < ?) OR (%s = ? AND %s < ?)", createdAtCol, createdAtCol, idCol)
	return tx.Where(cond, c.CreatedAt, c.CreatedAt, c.ID)
}

// applyAscCursor is applyDescCursor's mirror for ASC-ordered queries -
// currently just comments (see model.Comment.CursorKey's doc comment).
func applyAscCursor(tx *gorm.DB, cursor string, createdAtCol, idCol string) *gorm.DB {
	c, ok := model.DecodeCursor(cursor)
	if !ok {
		return tx
	}
	cond := fmt.Sprintf("(%s > ?) OR (%s = ? AND %s > ?)", createdAtCol, createdAtCol, idCol)
	return tx.Where(cond, c.CreatedAt, c.CreatedAt, c.ID)
}
