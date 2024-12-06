package repository

import "github.com/gocql/gocql"

type Payers interface {
}

type PayersRepository struct {
	session *gocql.Session
}

func NewPayersRepository(session *gocql.Session) *PayersRepository {
	return &PayersRepository{
		session: session,
	}
}
