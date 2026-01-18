package blockchain

import (
	"context"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/gagliardetto/solana-go"
	"github.com/gagliardetto/solana-go/rpc"
)

type SolanaAdapter struct {
	client *rpc.Client
}

func NewSolanaAdapter(rpcURL string) *SolanaAdapter {
	return &SolanaAdapter{
		client: rpc.New(rpcURL),
	}
}

func (a *SolanaAdapter) GetAccount(ctx context.Context, pubkey solana.PublicKey) (*domain.Account, error) {
	info, err := a.client.GetAccountInfo(ctx, pubkey)
	if err != nil {
		return nil, domain.ErrAccountNotFound
	}

	if info.Value == nil {
		return nil, domain.ErrAccountNotFound
	}

	return &domain.Account{
		PublicKey:  pubkey,
		Lamports:   info.Value.Lamports,
		Owner:      info.Value.Owner,
		Data:       info.Value.Data.GetBinary(),
		Executable: info.Value.Executable,
		RentEpoch:  info.Value.RentEpoch.Uint64(),
	}, nil
}

func (a *SolanaAdapter) GetBalance(ctx context.Context, pubkey solana.PublicKey) (uint64, error) {
	balance, err := a.client.GetBalance(ctx, pubkey, rpc.CommitmentFinalized)
	if err != nil {
		return 0, err
	}
	return balance.Value, nil
}

func (a *SolanaAdapter) SendTransaction(ctx context.Context, tx *solana.Transaction) (string, error) {
	sig, err := a.client.SendTransactionWithOpts(ctx, tx, rpc.TransactionOpts{
		SkipPreflight:       false,
		PreflightCommitment: rpc.CommitmentFinalized,
	})
	if err != nil {
		return "", domain.ErrTransactionFailed
	}
	return sig.String(), nil
}

func (a *SolanaAdapter) GetTransaction(ctx context.Context, signature string) (*domain.Transaction, error) {
	sig, err := solana.SignatureFromBase58(signature)
	if err != nil {
		return nil, domain.ErrInvalidTransaction
	}

	tx, err := a.client.GetTransaction(ctx, sig, &rpc.GetTransactionOpts{
		Encoding: solana.EncodingBase64,
	})
	if err != nil {
		return nil, domain.ErrTransactionFailed
	}

	var status domain.TransactionStatus
	if tx.Meta.Err != nil {
		status = domain.StatusFailed
	} else {
		status = domain.StatusFinalized
	}

	var blockTime *int64
	if tx.BlockTime != nil {
		t := int64(*tx.BlockTime)
		blockTime = &t
	}

	return &domain.Transaction{
		Signature: signature,
		BlockTime: blockTime,
		Slot:      tx.Slot,
		Fee:       tx.Meta.Fee,
		Status:    status,
	}, nil
}

func (a *SolanaAdapter) GetRecentBlockhash(ctx context.Context) (solana.Hash, error) {
	recent, err := a.client.GetLatestBlockhash(ctx, rpc.CommitmentFinalized)
	if err != nil {
		return solana.Hash{}, err
	}
	return recent.Value.Blockhash, nil
}

func (a *SolanaAdapter) GetFeeForMessage(ctx context.Context, msg solana.Message) (*domain.FeeInfo, error) {
	fee, err := a.client.GetFeeForMessage(ctx, msg.ToBase64(), rpc.CommitmentFinalized)
	if err != nil {
		return nil, err
	}

	return &domain.FeeInfo{
		Fee: *fee.Value,
		FeeCalculator: domain.FeeCalculator{
			LamportsPerSignature: 5000, // Default value
		},
	}, nil
}

func (a *SolanaAdapter) GetProgram(ctx context.Context, programID solana.PublicKey) (*domain.Program, error) {
	account, err := a.GetAccount(ctx, programID)
	if err != nil {
		return nil, domain.ErrProgramNotFound
	}

	return &domain.Program{
		ID:         programID,
		Executable: account.Executable,
		Owner:      account.Owner,
	}, nil
}

func (a *SolanaAdapter) FindProgramAddress(seeds [][]byte, programID solana.PublicKey) (*domain.PDA, error) {
	addr, bump, err := solana.FindProgramAddress(seeds, programID)
	if err != nil {
		return nil, domain.ErrInvalidPDA
	}

	return &domain.PDA{
		Address: addr,
		Bump:    bump,
		Seeds:   seeds,
	}, nil
}

func (a *SolanaAdapter) ExecuteCPI(ctx context.Context, req domain.CPIRequest) error {
	// CPI execution would be done within a program on-chain
	// This is a placeholder for the adapter interface
	return domain.ErrCPIFailed
}
