package repository

import "github.com/gocql/gocql"

type Buyers interface {
}

type BuyersRepository struct {
	session *gocql.Session
}

func NewBuyersRepository(session *gocql.Session) *BuyersRepository {
	return &BuyersRepository{
		session: session,
	}
}
