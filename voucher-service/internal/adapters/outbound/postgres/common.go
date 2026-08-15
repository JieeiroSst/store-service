package postgres

import (
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func lockForUpdate(db *gorm.DB) *gorm.DB {
	return db.Clauses(clause.Locking{Strength: "UPDATE"})
}

func newUUID() string {
	return uuid.New().String()
}
