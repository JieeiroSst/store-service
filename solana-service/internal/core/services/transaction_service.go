package services

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/JIeeiroSst/solana-service/internal/core/ports"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/programs/system"
)

type TransactionService struct {
	blockchain ports.BlockchainPort
	repo       ports.TransactionRepository
}

func NewTransactionService(blockchain ports.BlockchainPort, repo ports.TransactionRepository) *TransactionService {
	return &TransactionService{
		blockchain: blockchain,
		repo:       repo,
	}
}

func (s *TransactionService) CreateTransfer(ctx context.Context, from, to solana.PublicKey, amount uint64, signer solana.PrivateKey) (string, error) {
	// Get recent blockhash
	recentBlockhash, err := s.blockchain.GetRecentBlockhash(ctx)
	if err != nil {
		return "", err
	}

	// Create transfer instruction
	instruction := system.NewTransferInstruction(
		amount,
		from,
		to,
	).Build()

	// Build transaction
	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recentBlockhash,
		solana.TransactionPayer(from),
	)
	if err != nil {
		return "", domain.ErrInvalidTransaction
	}

	// Sign transaction
	_, err = tx.Sign(func(key solana.PublicKey) *solana.PrivateKey {
		if key.Equals(from) {
			return &signer
		}
		return nil
	})
	if err != nil {
		return "", err
	}

	// Send transaction
	signature, err := s.blockchain.SendTransaction(ctx, tx)
	if err != nil {
		return "", domain.ErrTransactionFailed
	}

	return signature, nil
}

func (s *TransactionService) GetTransaction(ctx context.Context, signature string) (*domain.Transaction, error) {
	tx, err := s.repo.FindBySignature(ctx, signature)
	if err == nil {
		return tx, nil
	}

	tx, err = s.blockchain.GetTransaction(ctx, signature)
	if err != nil {
		return nil, err
	}

	_ = s.repo.Save(ctx, tx)

	return tx, nil
}

func (s *TransactionService) EstimateFee(ctx context.Context, from, to solana.PublicKey, amount uint64) (*domain.FeeInfo, error) {
	instruction := system.NewTransferInstruction(amount, from, to).Build()

	recentBlockhash, err := s.blockchain.GetRecentBlockhash(ctx)
	if err != nil {
		return nil, err
	}

	tx, err := solana.NewTransaction(
		[]solana.Instruction{instruction},
		recentBlockhash,
		solana.TransactionPayer(from),
	)
	if err != nil {
		return nil, err
	}

	return s.blockchain.GetFeeForMessage(ctx, tx.Message)
}
