package services

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/JIeeiroSst/solana-service/internal/core/ports"
	"github.com/gagliardetto/solana-go"
)

type ProgramService struct {
	blockchain ports.BlockchainPort
}

func NewProgramService(blockchain ports.BlockchainPort) *ProgramService {
	return &ProgramService{
		blockchain: blockchain,
	}
}

func (s *ProgramService) GetProgram(ctx context.Context, programID string) (*domain.Program, error) {
	pubkey, err := solana.PublicKeyFromBase58(programID)
	if err != nil {
		return nil, domain.ErrInvalidPublicKey
	}

	return s.blockchain.GetProgram(ctx, pubkey)
}

func (s *ProgramService) FindPDA(ctx context.Context, seeds [][]byte, programID string) (*domain.PDA, error) {
	pubkey, err := solana.PublicKeyFromBase58(programID)
	if err != nil {
		return nil, domain.ErrInvalidPublicKey
	}

	return s.blockchain.FindProgramAddress(seeds, pubkey)
}

func (s *ProgramService) ExecuteCrossProgram(ctx context.Context, req domain.CPIRequest) error {
	return s.blockchain.ExecuteCPI(ctx, req)
}

func (s *ProgramService) CreatePDAExample(ctx context.Context, programID string, userPubkey string, identifier string) (*domain.PDA, error) {
	programPubkey, err := solana.PublicKeyFromBase58(programID)
	if err != nil {
		return nil, domain.ErrInvalidPublicKey
	}

	userKey, err := solana.PublicKeyFromBase58(userPubkey)
	if err != nil {
		return nil, domain.ErrInvalidPublicKey
	}

	// Create seeds: [user_pubkey, identifier]
	seeds := [][]byte{
		userKey.Bytes(),
		[]byte(identifier),
	}

	return s.blockchain.FindProgramAddress(seeds, programPubkey)
}
