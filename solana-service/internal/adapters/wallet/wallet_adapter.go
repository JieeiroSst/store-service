package wallet

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
)

type WalletAdapter struct {
	client *Client
}

func NewWalletAdapter(apiKey string) *WalletAdapter {
	return &WalletAdapter{
		client: NewClient(apiKey),
	}
}

func (a *WalletAdapter) CreateWalletSet(ctx context.Context, name string, custodyType domain.CustodyType) (*domain.CircleWalletSet, error) {
	reqBody := map[string]interface{}{
		"name": name,
	}

	var walletSet domain.CircleWalletSet
	err := a.client.Do(ctx, "POST", "/w3s/developer/walletSets", reqBody, &walletSet)
	if err != nil {
		return nil, err
	}

	return &walletSet, nil
}

func (a *WalletAdapter) GetWalletSet(ctx context.Context, walletSetID string) (*domain.CircleWalletSet, error) {
	var walletSet domain.CircleWalletSet
	err := a.client.Do(ctx, "GET", fmt.Sprintf("/w3s/walletSets/%s", walletSetID), nil, &walletSet)
	if err != nil {
		return nil, err
	}

	return &walletSet, nil
}

func (a *WalletAdapter) ListWalletSets(ctx context.Context) ([]*domain.CircleWalletSet, error) {
	var walletSets []domain.CircleWalletSet
	err := a.client.Do(ctx, "GET", "/w3s/walletSets", nil, &walletSets)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.CircleWalletSet, len(walletSets))
	for i := range walletSets {
		result[i] = &walletSets[i]
	}

	return result, nil
}

func (a *WalletAdapter) UpdateWalletSet(ctx context.Context, walletSetID, name string) (*domain.CircleWalletSet, error) {
	reqBody := map[string]interface{}{
		"name": name,
	}

	var walletSet domain.CircleWalletSet
	err := a.client.Do(ctx, "PUT", fmt.Sprintf("/w3s/walletSets/%s", walletSetID), reqBody, &walletSet)
	if err != nil {
		return nil, err
	}

	return &walletSet, nil
}

func (a *WalletAdapter) CreateWallets(ctx context.Context, req domain.CreateWalletRequest) ([]*domain.CircleWallet, error) {
	var response struct {
		Wallets []domain.CircleWallet `json:"wallets"`
	}

	err := a.client.Do(ctx, "POST", "/w3s/developer/wallets", req, &response)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.CircleWallet, len(response.Wallets))
	for i := range response.Wallets {
		result[i] = &response.Wallets[i]
	}

	return result, nil
}

func (a *WalletAdapter) GetWallet(ctx context.Context, walletID string) (*domain.CircleWallet, error) {
	var wallet domain.CircleWallet
	err := a.client.Do(ctx, "GET", fmt.Sprintf("/w3s/wallets/%s", walletID), nil, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (a *WalletAdapter) ListWallets(ctx context.Context, walletSetID string, blockchain string) ([]*domain.CircleWallet, error) {
	params := map[string]string{
		"walletSetId": walletSetID,
		"blockchain":  blockchain,
	}

	endpoint := "/w3s/wallets" + BuildQueryParams(params)

	var wallets []domain.CircleWallet
	err := a.client.Do(ctx, "GET", endpoint, nil, &wallets)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.CircleWallet, len(wallets))
	for i := range wallets {
		result[i] = &wallets[i]
	}

	return result, nil
}

func (a *WalletAdapter) UpdateWallet(ctx context.Context, walletID string, req domain.UpdateWalletRequest) (*domain.CircleWallet, error) {
	var wallet domain.CircleWallet
	err := a.client.Do(ctx, "PUT", fmt.Sprintf("/w3s/wallets/%s", walletID), req, &wallet)
	if err != nil {
		return nil, err
	}

	return &wallet, nil
}

func (a *WalletAdapter) GetWalletBalance(ctx context.Context, walletID string) ([]*domain.CircleBalance, error) {
	var response struct {
		TokenBalances []domain.CircleBalance `json:"tokenBalances"`
	}

	endpoint := fmt.Sprintf("/w3s/wallets/%s/balances", walletID)
	err := a.client.Do(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.CircleBalance, len(response.TokenBalances))
	for i := range response.TokenBalances {
		result[i] = &response.TokenBalances[i]
	}

	return result, nil
}

func (a *WalletAdapter) GetWalletNFTs(ctx context.Context, walletID string) ([]*domain.NFTBalance, error) {
	var response struct {
		NFTBalances []domain.NFTBalance `json:"nfts"`
	}

	endpoint := fmt.Sprintf("/w3s/wallets/%s/nfts", walletID)
	err := a.client.Do(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NFTBalance, len(response.NFTBalances))
	for i := range response.NFTBalances {
		result[i] = &response.NFTBalances[i]
	}

	return result, nil
}
