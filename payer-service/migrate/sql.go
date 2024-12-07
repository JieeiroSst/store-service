package migrate

import (
	"context"

	"github.com/JIeeiroSst/utils/logger"
	"github.com/gocql/gocql"
)

type Migrate struct {
	session *gocql.Session
}

var (
	CreatePayersTable = `
	CREATE TABLE IF NOT EXISTS payer.payers (
		payer_id PRIMARY KEY,
		name text,
		email text,
		phone_number text
	);

	`
	CreateBuyersTable = `
		CREATE TABLE IF NOT EXISTS payer.Buyers (
			buyer_id PRIMARY KEY,
			name text,
			email text,
			phone_number text
		);
	`
	CreateTransactionsTable = `
		CREATE TABLE IF NOT EXISTS payer.payers (
			transaction_id int PRIMARY KEY,
			payer_id int,
			buyer_id int,
			amount float,
			transaction_date timestamp,
			transaction_type int,
			description text,
			status int
		);
	`

	CreateKeyspace = `
		CREATE KEYSPACE IF NOT EXISTS payer WITH REPLICATION = {'class' : 'NetworkTopologyStrategy', 'replication_factor' : 1};
	`
)

func (m *Migrate) Run() {
	// create keyspaces
	err := m.session.Query(CreateKeyspace).Exec()
	if err != nil {
		logger.Error(context.Background(), "err")
	}

	// create table
	err = m.session.Query(CreatePayersTable).Exec()
	if err != nil {

	}

	// create table
	err = m.session.Query(CreateBuyersTable).Exec()
	if err != nil {

	}

	// create table
	err = m.session.Query(CreateTransactionsTable).Exec()
	if err != nil {

	}
}
