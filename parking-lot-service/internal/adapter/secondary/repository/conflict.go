package repository

import "gorm.io/gorm/clause"

// onConflictDoNothing guards a Create against a duplicate-key race (e.g.
// two check-ins for a plate/id that doesn't exist yet both passing a
// preceding SELECT before either INSERT commits).
func onConflictDoNothing(column string) clause.OnConflict {
	return clause.OnConflict{
		Columns:   []clause.Column{{Name: column}},
		DoNothing: true,
	}
}
