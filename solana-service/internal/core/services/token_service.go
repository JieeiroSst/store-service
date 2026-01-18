package services

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/JIeeiroSst/solana-service/internal/core/ports"
)

type CircleTokenService struct {
	token ports.CircleTokenPort
}

func NewCircleTokenService(token ports.CircleTokenPort) *CircleTokenService {
	return &CircleTokenService{
		token: token,
	}
}

func (s *CircleTokenService) GetToken(ctx context.Context, tokenID string) (*domain.Token, error) {
	return s.token.GetToken(ctx, tokenID)
}

func (s *CircleTokenService) ListTokens(ctx context.Context, blockchain string) ([]*domain.Token, error) {
	return s.token.ListTokens(ctx, blockchain)
}

func (s *CircleTokenService) ImportToken(ctx context.Context, metadata domain.TokenMetadata) (*domain.Token, error) {
	return s.token.ImportToken(ctx, metadata)
}

func (s *CircleTokenService) GetNFTToken(ctx context.Context, tokenID string) (*domain.NFTToken, error) {
	return s.token.GetNFTToken(ctx, tokenID)
}

func (s *CircleTokenService) ListNFTTokens(ctx context.Context, blockchain string) ([]*domain.NFTToken, error) {
	return s.token.ListNFTTokens(ctx, blockchain)
}

func (s *CircleTokenService) GetSolanaTokens(ctx context.Context) ([]*domain.Token, error) {
	return s.token.ListTokens(ctx, "SOL")
}
