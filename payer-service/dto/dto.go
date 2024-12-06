package dto

import (
	"github.com/JIeeiroSst/payer-service/model"
	"github.com/JIeeiroSst/utils/geared_id"
	"github.com/JIeeiroSst/utils/time_custom"
)

type CreateTransactionsRequest struct {
	Payers          Payers  `json:"payer"`
	Buyers          Buyers  `json:"buyer"`
	TransactionID   int     `json:"transaction_id"`
	Amount          float64 `json:"amount"`
	TransactionDate int     `json:"transaction_date"`
	TransactionType int     `json:"transaction_type"`
	Description     string  `json:"description"`
}

type Payers struct {
	PayerID     int    `json:"payer_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

type Buyers struct {
	BuyerID     int    `json:"buyer_id"`
	Name        string `json:"name"`
	Email       string `json:"email"`
	PhoneNumber string `json:"phone_number"`
}

func (p Payers) Build() model.Payers {
	return model.Payers{
		PayerID:     p.PayerID,
		Name:        p.Name,
		Email:       p.Email,
		PhoneNumber: p.PhoneNumber,
	}
}

func (p Buyers) Build() model.Buyers {
	return model.Buyers{
		BuyerID:     p.BuyerID,
		Name:        p.Name,
		Email:       p.Email,
		PhoneNumber: p.PhoneNumber,
	}
}

func (p CreateTransactionsRequest) Build() model.Transactions {
	return model.Transactions{
		TransactionID:   geared_id.GearedIntID(),
		PayerID:         p.Payers.PayerID,
		BuyerID:         p.Buyers.BuyerID,
		Amount:          p.Amount,
		TransactionDate: time_custom.FormatUnixTime(p.TransactionDate),
		Description:     p.Description,
	}
}
