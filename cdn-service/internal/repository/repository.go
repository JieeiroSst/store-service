package repository

import "database/sql"

type Repositories struct {
	CDN
}

func NewRepositories(db *sql.DB) *Repositories {
	return &Repositories{
		CDN: NewCdnRepository(db),
	}
}
