package repository

import (
	"context"
	"errors"
	"time"

	"github.com/JIeeiroSst/payer-service/common"
	"github.com/JIeeiroSst/payer-service/model"
	"github.com/JIeeiroSst/utils/logger"
	"github.com/gocql/gocql"
	"golang.org/x/sync/errgroup"
)

type Transactions interface {
	Transactions(ctx context.Context, payer model.Payers, buyers model.Buyers, amount float64, transactionType, transactionID int, description string) error
	GetTransactions(ctx context.Context, transactionID int) (*model.Buyers, *model.Payers, *model.Transactions, error)
}

type TransactionsRepository struct {
	session *gocql.Session
}

func NewTransactionsRepository(session *gocql.Session) *TransactionsRepository {
	return &TransactionsRepository{
		session: session,
	}
}

func (r *TransactionsRepository) Transactions(ctx context.Context, payer model.Payers, buyers model.Buyers,
	amount float64, transactionType, transactionID int, description string) error {

	batch := r.session.NewBatch(gocql.LoggedBatch)

	batch.Query("INSERT INTO payer.payers (payer_id, name, email, phone_number) VALUES (?, ?, ?, ?)", payer.PayerID, payer.Name, payer.Email, payer.PhoneNumber)
	batch.Query("INSERT INTO payer.buyers (buyer_id, name, email, phone_number) VALUES (?, ?, ?, ?)", buyers.BuyerID, buyers.Name, buyers.Email, buyers.PhoneNumber)
	batch.Query("INSERT INTO payer.transactions (transaction_id, payer_id, buyer_id, amount, transaction_date, transaction_type, description, status) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
		transactionID, payer.PayerID, buyers.BuyerID, amount, time.Now(), transactionType, common.Peding.Value())

	if err := r.session.ExecuteBatch(batch); err != nil {
		err = r.session.Query("UPDATE payer.transactions set status = ? where transaction_id = ?", transactionID, common.Reject.Value()).Exec()
		if err != nil {
			logger.Error(ctx, "error %v", err)
			return err
		}
		return nil
	} else {
		err = r.session.Query("UPDATE payer.transactions set status = ? where transaction_id = ?", transactionID, common.Susscess.Value()).Exec()
		if err != nil {
			logger.Error(ctx, "error %v", err)
			return err
		}
		return nil
	}
}

func (r *TransactionsRepository) GetTransactions(ctx context.Context, transactionID int) (*model.Buyers, *model.Payers, *model.Transactions, error) {
	var (
		transaction model.Transactions
		payers      model.Payers
		buyers      model.Buyers
	)

	query := r.session.Query("SELECT * FROM payer.transactions where transaction_id = ?", transactionID).Iter()
	result := query.Scan(&transaction)
	if !result {
		return nil, nil, nil, errors.New("not found")
	}

	g, _ := errgroup.WithContext(ctx)

	g.Go(func() error {
		var payersResult model.Payers
		query := r.session.Query("SELECT * FROM payer.payers where payer_id = ?", transaction.PayerID).Iter()
		result := query.Scan(&payersResult)
		if !result {
			return errors.New("not found")
		}
		payers = payersResult
		return nil
	})

	g.Go(func() error {
		var buyersResult model.Buyers
		query := r.session.Query("SELECT * FROM payer.buyers where buyer_id = ?", transaction.BuyerID).Iter()
		result := query.Scan(&buyersResult)
		if !result {
			return errors.New("not found")
		}
		buyers = buyersResult
		return nil
	})

	if err := g.Wait(); err != nil {
		return nil, nil, nil, errors.New("not found")
	}

	return &buyers, &payers, &transaction, nil
}
