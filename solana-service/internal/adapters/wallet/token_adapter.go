package wallet

import (
	"context"
	"fmt"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
)

type TokenAdapter struct {
	client *Client
}

func NewTokenAdapter(apiKey string) *TokenAdapter {
	return &TokenAdapter{
		client: NewClient(apiKey),
	}
}

func (a *TokenAdapter) GetToken(ctx context.Context, tokenID string) (*domain.Token, error) {
	var token domain.Token
	err := a.client.Do(ctx, "GET", fmt.Sprintf("/w3s/tokens/%s", tokenID), nil, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (a *TokenAdapter) ListTokens(ctx context.Context, blockchain string) ([]*domain.Token, error) {
	params := map[string]string{
		"blockchain": blockchain,
	}

	endpoint := "/w3s/tokens" + BuildQueryParams(params)

	var response struct {
		Tokens []domain.Token `json:"tokens"`
	}

	err := a.client.Do(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.Token, len(response.Tokens))
	for i := range response.Tokens {
		result[i] = &response.Tokens[i]
	}

	return result, nil
}

func (a *TokenAdapter) ImportToken(ctx context.Context, metadata domain.TokenMetadata) (*domain.Token, error) {
	var token domain.Token
	err := a.client.Do(ctx, "POST", "/w3s/tokens", metadata, &token)
	if err != nil {
		return nil, err
	}

	return &token, nil
}

func (a *TokenAdapter) GetNFTToken(ctx context.Context, tokenID string) (*domain.NFTToken, error) {
	var nft domain.NFTToken
	err := a.client.Do(ctx, "GET", fmt.Sprintf("/w3s/nfts/tokens/%s", tokenID), nil, &nft)
	if err != nil {
		return nil, err
	}

	return &nft, nil
}

func (a *TokenAdapter) ListNFTTokens(ctx context.Context, blockchain string) ([]*domain.NFTToken, error) {
	params := map[string]string{
		"blockchain": blockchain,
	}

	endpoint := "/w3s/nfts/tokens" + BuildQueryParams(params)

	var response struct {
		NFTs []domain.NFTToken `json:"nfts"`
	}

	err := a.client.Do(ctx, "GET", endpoint, nil, &response)
	if err != nil {
		return nil, err
	}

	result := make([]*domain.NFTToken, len(response.NFTs))
	for i := range response.NFTs {
		result[i] = &response.NFTs[i]
	}

	return result, nil
}
