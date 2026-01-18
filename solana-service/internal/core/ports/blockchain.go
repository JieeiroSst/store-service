package ports

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/gagliardetto/solana-go"
)

type BlockchainPort interface {
	// Account operations
	GetAccount(ctx context.Context, pubkey solana.PublicKey) (*domain.Account, error)
	GetBalance(ctx context.Context, pubkey solana.PublicKey) (uint64, error)

	// Transaction operations
	SendTransaction(ctx context.Context, tx *solana.Transaction) (string, error)
	GetTransaction(ctx context.Context, signature string) (*domain.Transaction, error)
	GetRecentBlockhash(ctx context.Context) (solana.Hash, error)
	GetFeeForMessage(ctx context.Context, msg solana.Message) (*domain.FeeInfo, error)

	// Program operations
	GetProgram(ctx context.Context, programID solana.PublicKey) (*domain.Program, error)
	FindProgramAddress(seeds [][]byte, programID solana.PublicKey) (*domain.PDA, error)

	// CPI operations
	ExecuteCPI(ctx context.Context, req domain.CPIRequest) error
}
