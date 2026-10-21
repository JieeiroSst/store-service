package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type BatchStatus string

const (
	BatchStatusOpen    BatchStatus = "OPEN"
	BatchStatusCleared BatchStatus = "CLEARED"
	BatchStatusSettled BatchStatus = "SETTLED"
)

type SettlementBatch struct {
	ID           uuid.UUID       `json:"id"                     db:"id"`
	AcquirerID   uuid.UUID       `json:"acquirer_id"            db:"acquirer_id"`
	MerchantID   uuid.UUID       `json:"merchant_id"            db:"merchant_id"`
	TotalAmount  decimal.Decimal `json:"total_amount"          db:"total_amount"`
	TotalFee     decimal.Decimal `json:"total_fee"             db:"total_fee"`
	TxnCount     int             `json:"txn_count"              db:"txn_count"`
	Status       BatchStatus     `json:"status"                 db:"status"`
	Transactions []*Transaction  `json:"transactions,omitempty" db:"-"`
	ProcessedAt  *time.Time      `json:"processed_at,omitempty" db:"processed_at"`
	CreatedAt    time.Time       `json:"created_at"             db:"created_at"`
}

func NewSettlementBatch(merchant *Merchant, txns []*Transaction, now time.Time) (*SettlementBatch, error) {
	if len(txns) == 0 {
		return nil, ErrNothingToSettle
	}
	total, totalFee := decimal.Zero, decimal.Zero
	for _, tx := range txns {
		total = total.Add(tx.Amount)
		totalFee = totalFee.Add(tx.Fee)
	}
	return &SettlementBatch{
		ID:           uuid.New(),
		AcquirerID:   merchant.AcquirerBankID,
		MerchantID:   merchant.ID,
		TotalAmount:  total,
		TotalFee:     totalFee,
		TxnCount:     len(txns),
		Status:       BatchStatusOpen,
		Transactions: txns,
		CreatedAt:    now,
	}, nil
}

type ClearingRecord struct {
	ID          uuid.UUID       `json:"id"           db:"id"`
	BatchID     uuid.UUID       `json:"batch_id"     db:"batch_id"`
	CardNetwork CardNetwork     `json:"card_network" db:"card_network"`
	AcquirerID  uuid.UUID       `json:"acquirer_id"  db:"acquirer_id"`
	IssuerID    uuid.UUID       `json:"issuer_id"    db:"issuer_id"`
	NetAmount   decimal.Decimal `json:"net_amount"   db:"net_amount"`
	ClearedAt   time.Time       `json:"cleared_at"   db:"cleared_at"`
}

func NetClearing(batchID uuid.UUID, txns []*Transaction, now time.Time) []*ClearingRecord {
	type key struct {
		issuer, acquirer uuid.UUID
		network          CardNetwork
	}
	order := make([]key, 0)
	net := make(map[key]decimal.Decimal)
	for _, tx := range txns {
		k := key{issuer: tx.IssuerBankID.UUID, acquirer: tx.AcquirerBankID.UUID, network: tx.CardNetwork}
		if _, seen := net[k]; !seen {
			order = append(order, k)
		}
		net[k] = net[k].Add(tx.Amount.Sub(tx.Fee))
	}
	records := make([]*ClearingRecord, 0, len(order))
	for _, k := range order {
		records = append(records, &ClearingRecord{
			ID:          uuid.New(),
			BatchID:     batchID,
			CardNetwork: k.network,
			AcquirerID:  k.acquirer,
			IssuerID:    k.issuer,
			NetAmount:   net[k],
			ClearedAt:   now,
		})
	}
	return records
}
