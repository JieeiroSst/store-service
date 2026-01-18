package wallet

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
)

type TransactionAdapter struct {
	client *Client
}

func NewTransactionAdapter(apiKey string) *TransactionAdapter {
	return &TransactionAdapter{
		client: NewClient(apiKey),
	}
}

func (a *TransactionAdapter) CreateTransfer(ctx context.Context, req domain.TransferRequest) (*domain.CircleTransaction, error) {
	var tx domain.CircleTransaction
	err := a.client.Do(ctx, "POST", "/w3s/developer/transactions/transfer", req, &tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (a *TransactionAdapter) CreateNFTTransfer(ctx context.Context, req domain.NFTTransferRequest) (*domain.CircleTransaction, error) {
	var tx domain.CircleTransaction
	err := a.client.Do(ctx, "POST", "/w3s/developer/transactions/nftTransfer", req, &tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (a *TransactionAdapter) ExecuteContract(ctx context.Context, req domain.ContractExecutionRequest) (*domain.CircleTransaction, error) {
	var tx domain.CircleTransaction
	err := a.client.Do(ctx, "POST", "/w3s/developer/transactions/contractExecution", req, &tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (a *TransactionAdapter) GetTransaction(ctx context.Context, txID string) (*domain.CircleTransaction, error) {
	var tx domain.CircleTransaction
	err := a.client.Do(ctx, "GET", fmt.Sprintf("/w3s/transactions/%s", txID), nil, &tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (a *TransactionAdapter) ListTransactions(ctx context.Context, filter domain.TransactionFilter) ([]*domain.CircleTransaction, error) {
	params := make(map[string]string)

	if filter.Blockchain != "" {
		params["blockchain"] = filter.Blockchain
	}
	if filter.WalletIDs != nil && len(filter.WalletIDs) > 0 {
		params["walletIds"] = filter.WalletIDs[0] // Simplified
	}
	if filter.TokenID != "" {
		params["tokenId"] = filter.TokenID
	}

	endpoint := "/w3s/transactions" + BuildQueryParams(params)

	var response struct {
		Transactions []domain.CircleTransaction `json:"transactions"`
	}

	err := a.client.Do(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.CircleTransaction, len(response.Transactions))
	for i := range response.Transactions {
		result[i] = &response.Transactions[i]
	}

	return result, nil
}

func (a *TransactionAdapter) ValidateAddress(ctx context.Context, blockchain, address string) (bool, error) {
	reqBody := map[string]interface{}{
		"blockchain": blockchain,
		"address":    address,
	}

	var response struct {
		IsValid bool `json:"isValid"`
	}

	err := a.client.Do(ctx, "POST", "/w3s/transactions/validateAddress", reqBody, &response)
	if err != nil {
		return false, err
	}

	return response.IsValid, nil
}

func (a *TransactionAdapter) EstimateFee(ctx context.Context, walletID, destinationAddress, tokenID, amount string) (*domain.FeeEstimate, error) {
	reqBody := map[string]interface{}{
		"walletId":           walletID,
		"destinationAddress": destinationAddress,
		"tokenId":            tokenID,
		"amount":             amount,
	}

	var feeEstimate domain.FeeEstimate
	err := a.client.Do(ctx, "POST", "/w3s/transactions/transfer/estimateFee", reqBody, &feeEstimate)
	if err != nil {
		return nil, err
	}

	return &feeEstimate, nil
}

func (a *TransactionAdapter) AccelerateTransaction(ctx context.Context, txID string, req domain.AccelerateTransactionRequest) (*domain.CircleTransaction, error) {
	var tx domain.CircleTransaction
	err := a.client.Do(ctx, "PUT", fmt.Sprintf("/w3s/transactions/%s/accelerate", txID), req, &tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}

func (a *TransactionAdapter) CancelTransaction(ctx context.Context, txID string, req domain.CancelTransactionRequest) (*domain.CircleTransaction, error) {
	var tx domain.CircleTransaction
	err := a.client.Do(ctx, "PUT", fmt.Sprintf("/w3s/transactions/%s/cancel", txID), req, &tx)
	if err != nil {
		return nil, err
	}

	return &tx, nil
}
