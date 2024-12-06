package repository

import "github.com/gocql/gocql"

type Repository struct {
	Transactions
	Buyers
	Payers
}

func NewRepositories(session *gocql.Session) *Repository {
	return &Repository{
		Transactions: NewTransactionsRepository(session),
		Buyers:       NewBuyersRepository(session),
		Payers:       NewPayersRepository(session),
	}
}
