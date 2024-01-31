package repository

import "github.com/jackc/pgx/v4"

type Repositories struct {
}

func NewRepositories(db *pgx.Conn) *Repositories {
	return &Repositories{
		
	}
}
