package services

import (
	"context"
	"fmt"
	"time"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
	"github.com/JIeeiroSst/solana-service/internal/core/ports"
	"github.com/gagliardetto/solana-go"
	"github.com/google/uuid"
)

type CircleTransactionService struct {
	transaction  ports.CircleTransactionPort
	blockchain   ports.BlockchainPort
	bridgeRepo   ports.BridgeRepository
	solTxService *TransactionService
}

func NewCircleTransactionService(
	transaction ports.CircleTransactionPort,
	blockchain ports.BlockchainPort,
	bridgeRepo ports.BridgeRepository,
	solTxService *TransactionService,
) *CircleTransactionService {
	return &CircleTransactionService{
		transaction:  transaction,
		blockchain:   blockchain,
		bridgeRepo:   bridgeRepo,
		solTxService: solTxService,
	}
}

func (s *CircleTransactionService) CreateTransfer(ctx context.Context, req domain.TransferRequest) (*domain.CircleTransaction, error) {
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = uuid.New().String()
	}

	return s.transaction.CreateTransfer(ctx, req)
}

func (s *CircleTransactionService) CreateNFTTransfer(ctx context.Context, req domain.NFTTransferRequest) (*domain.CircleTransaction, error) {
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = uuid.New().String()
	}

	return s.transaction.CreateNFTTransfer(ctx, req)
}

func (s *CircleTransactionService) ExecuteContract(ctx context.Context, req domain.ContractExecutionRequest) (*domain.CircleTransaction, error) {
	if req.IdempotencyKey == "" {
		req.IdempotencyKey = uuid.New().String()
	}

	return s.transaction.ExecuteContract(ctx, req)
}

func (s *CircleTransactionService) GetTransaction(ctx context.Context, txID string) (*domain.CircleTransaction, error) {
	return s.transaction.GetTransaction(ctx, txID)
}

func (s *CircleTransactionService) ListTransactions(ctx context.Context, filter domain.TransactionFilter) ([]*domain.CircleTransaction, error) {
	return s.transaction.ListTransactions(ctx, filter)
}

func (s *CircleTransactionService) ValidateAddress(ctx context.Context, blockchain, address string) (bool, error) {
	return s.transaction.ValidateAddress(ctx, blockchain, address)
}

func (s *CircleTransactionService) EstimateFee(ctx context.Context, walletID, destinationAddress, tokenID, amount string) (*domain.FeeEstimate, error) {
	return s.transaction.EstimateFee(ctx, walletID, destinationAddress, tokenID, amount)
}

func (s *CircleTransactionService) AccelerateTransaction(ctx context.Context, txID string, feeLevel string) (*domain.CircleTransaction, error) {
	req := domain.AccelerateTransactionRequest{
		Fee: domain.FeeConfiguration{
			Type: domain.FeeTypeLevel,
			Config: domain.FeeConfig{
				FeeLevel: feeLevel,
			},
		},
	}

	return s.transaction.AccelerateTransaction(ctx, txID, req)
}

func (s *CircleTransactionService) CancelTransaction(ctx context.Context, txID string, feeLevel string) (*domain.CircleTransaction, error) {
	req := domain.CancelTransactionRequest{
		Fee: domain.FeeConfiguration{
			Type: domain.FeeTypeLevel,
			Config: domain.FeeConfig{
				FeeLevel: feeLevel,
			},
		},
	}

	return s.transaction.CancelTransaction(ctx, txID, req)
}

func (s *CircleTransactionService) BridgeSolanaToCircle(
	ctx context.Context,
	fromSolanaAddress string,
	toCircleWalletID string,
	amount uint64,
	signer solana.PrivateKey,
) (*domain.BridgeTransfer, error) {
	bridgeTransfer := &domain.BridgeTransfer{
		ID:             uuid.New().String(),
		SolanaAddress:  fromSolanaAddress,
		CircleWalletID: toCircleWalletID,
		Direction:      domain.DirectionSolanaToCircle,
		Amount:         fmt.Sprintf("%d", amount),
		TokenID:        "SOL",
		Status:         domain.BridgeStatusInitiated,
		CreateDate:     time.Now(),
	}

	if err := s.bridgeRepo.Save(ctx, bridgeTransfer); err != nil {
		return nil, err
	}

	circleWalletAddress := "CircleSolanaAddress" 

	fromPubkey, _ := solana.PublicKeyFromBase58(fromSolanaAddress)
	toPubkey, _ := solana.PublicKeyFromBase58(circleWalletAddress)

	// Execute Solana transaction
	signature, err := s.solTxService.CreateTransfer(ctx, fromPubkey, toPubkey, amount, signer)
	if err != nil {
		bridgeTransfer.Status = domain.BridgeStatusFailed
		bridgeTransfer.ErrorMessage = err.Error()
		_ = s.bridgeRepo.Update(ctx, bridgeTransfer)
		return nil, err
	}

	// Update bridge transfer
	bridgeTransfer.SolanaTxSignature = signature
	bridgeTransfer.Status = domain.BridgeStatusCompleted
	now := time.Now()
	bridgeTransfer.CompleteDate = &now
	_ = s.bridgeRepo.Update(ctx, bridgeTransfer)

	return bridgeTransfer, nil
}

func (s *CircleTransactionService) BridgeCircleToSolana(
	ctx context.Context,
	fromCircleWalletID string,
	toSolanaAddress string,
	amount string,
) (*domain.BridgeTransfer, error) {
	bridgeTransfer := &domain.BridgeTransfer{
		ID:             uuid.New().String(),
		SolanaAddress:  toSolanaAddress,
		CircleWalletID: fromCircleWalletID,
		Direction:      domain.DirectionCircleToSolana,
		Amount:         amount,
		TokenID:        "SOL",
		Status:         domain.BridgeStatusInitiated,
		CreateDate:     time.Now(),
	}

	if err := s.bridgeRepo.Save(ctx, bridgeTransfer); err != nil {
		return nil, err
	}

	transferReq := domain.TransferRequest{
		WalletID:           fromCircleWalletID,
		DestinationAddress: toSolanaAddress,
		TokenID:            "SOL",
		Amount:             amount,
		IdempotencyKey:     uuid.New().String(),
	}

	circleTx, err := s.transaction.CreateTransfer(ctx, transferReq)
	if err != nil {
		bridgeTransfer.Status = domain.BridgeStatusFailed
		bridgeTransfer.ErrorMessage = err.Error()
		_ = s.bridgeRepo.Update(ctx, bridgeTransfer)
		return nil, err
	}

	bridgeTransfer.CircleTxID = circleTx.ID
	bridgeTransfer.Status = domain.BridgeStatusPending
	_ = s.bridgeRepo.Update(ctx, bridgeTransfer)

	go s.monitorCircleTransaction(context.Background(), circleTx.ID, bridgeTransfer)

	return bridgeTransfer, nil
}

func (s *CircleTransactionService) monitorCircleTransaction(ctx context.Context, txID string, bridgeTransfer *domain.BridgeTransfer) {
	maxAttempts := 20
	for i := 0; i < maxAttempts; i++ {
		time.Sleep(5 * time.Second)

		circleTx, err := s.transaction.GetTransaction(ctx, txID)
		if err != nil {
			continue
		}

		switch circleTx.State {
		case domain.CircleTxStateComplete, domain.CircleTxStateConfirmed:
			bridgeTransfer.Status = domain.BridgeStatusCompleted
			bridgeTransfer.SolanaTxSignature = circleTx.TxHash
			now := time.Now()
			bridgeTransfer.CompleteDate = &now
			_ = s.bridgeRepo.Update(ctx, bridgeTransfer)
			return

		case domain.CircleTxStateFailed, domain.CircleTxStateCancelled:
			bridgeTransfer.Status = domain.BridgeStatusFailed
			bridgeTransfer.ErrorMessage = circleTx.ErrorReason
			_ = s.bridgeRepo.Update(ctx, bridgeTransfer)
			return
		}
	}
}

func (s *CircleTransactionService) GetBridgeTransfer(ctx context.Context, transferID string) (*domain.BridgeTransfer, error) {
	return s.bridgeRepo.FindByID(ctx, transferID)
}
