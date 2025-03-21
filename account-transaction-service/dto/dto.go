package dto

import "time"

type Account struct {
	ID          string    `json:"id"`
	FirstName   string    `json:"first_name"`
	LastName    string    `json:"last_name"`
	DateCreated time.Time `json:"date_created"`
}

type Transaction struct {
	ID          string    `json:"id"`
	Type        string    `json:"type"`
	Amount      float64   `json:"amount"`
	DateCreated time.Time `json:"date_created"`
	// Only one of the following will be non-nil
	Deposit    *DepositTransactionDetails           `json:"deposit,omitempty"`
	Withdrawal *WithdrawalTransactionDetails        `json:"withdrawal,omitempty"`
	Transfer   *AccountToAccountTransactionDetails  `json:"transfer,omitempty"`
	Payment    *PaymentForServiceTransactionDetails `json:"payment,omitempty"`
}

type DepositTransactionDetails struct {
	AccountID string `json:"account_id"`
}

type WithdrawalTransactionDetails struct {
	AccountID string `json:"account_id"`
}

type AccountToAccountTransactionDetails struct {
	SenderID   string `json:"sender_id"`
	ReceiverID string `json:"receiver_id"`
}

type PaymentForServiceTransactionDetails struct {
	AccountID   string `json:"account_id"`
	ServiceName string `json:"service_name"`
}

type CreateAccountRequest struct {
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
}

type GetAccountRequest struct {
	ID string `json:"id"`
}

type ListAccountsRequest struct {
	PageSize  int32  `json:"page_size"`
	PageToken string `json:"page_token"`
}

type ListAccountsResponse struct {
	Accounts      []Account `json:"accounts"`
	NextPageToken string    `json:"next_page_token"`
}

type CreateDepositTransactionRequest struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

type CreateWithdrawalTransactionRequest struct {
	AccountID string  `json:"account_id"`
	Amount    float64 `json:"amount"`
}

type CreateAccountToAccountTransactionRequest struct {
	SenderID   string  `json:"sender_id"`
	ReceiverID string  `json:"receiver_id"`
	Amount     float64 `json:"amount"`
}

type CreatePaymentForServiceTransactionRequest struct {
	AccountID   string  `json:"account_id"`
	ServiceName string  `json:"service_name"`
	Amount      float64 `json:"amount"`
}

type GetTransactionRequest struct {
	ID string `json:"id"`
}

type ListTransactionsRequest struct {
	PageSize  int32  `json:"page_size"`
	PageToken string `json:"page_token"`
	Type      string `json:"type"`
}

type ListTransactionsResponse struct {
	Transactions  []Transaction `json:"transactions"`
	NextPageToken string        `json:"next_page_token"`
}

type GetAccountTransactionsRequest struct {
	AccountID string `json:"account_id"`
	PageSize  int32  `json:"page_size"`
	PageToken string `json:"page_token"`
}
