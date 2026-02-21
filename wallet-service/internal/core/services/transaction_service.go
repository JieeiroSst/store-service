package services

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
	"go.uber.org/zap"

	"github.com/JIeeiroSst/wallet-service/internal/core/domain"
	"github.com/JIeeiroSst/wallet-service/internal/core/ports"
	"github.com/JIeeiroSst/wallet-service/pkg/algorithms/batch"
)

const (
	txCacheKey     = "tx:%s"
	txCacheTTL     = 30 * time.Minute
	idempotencyTTL = 24 * time.Hour
	idempotencyPfx = "idempotency:%s"
)

type transactionService struct {
	walletRepo   ports.WalletRepository
	txRepo       ports.TransactionRepository
	cardRepo     ports.CardRepository
	merchantRepo ports.MerchantRepository
	bankRepo     ports.BankRepository
	batchRepo    ports.SettlementBatchRepository
	clearRepo    ports.ClearingRepository
	cache        ports.CacheRepository
	feeCalc      ports.FeeCalculator
	batchProc    *batch.Processor[*domain.Transaction]
	logger       *zap.Logger
}

func NewTransactionService(
	walletRepo ports.WalletRepository,
	txRepo ports.TransactionRepository,
	cardRepo ports.CardRepository,
	merchantRepo ports.MerchantRepository,
	bankRepo ports.BankRepository,
	batchRepo ports.SettlementBatchRepository,
	clearRepo ports.ClearingRepository,
	cache ports.CacheRepository,
	feeCalc ports.FeeCalculator,
	logger *zap.Logger,
) ports.TransactionUseCase {
	batchProc := batch.NewProcessor[*domain.Transaction](
		batch.WithSize[*domain.Transaction](100),
		batch.WithFlushInterval[*domain.Transaction](5*time.Second),
	)
	return &transactionService{
		walletRepo: walletRepo, txRepo: txRepo, cardRepo: cardRepo,
		merchantRepo: merchantRepo, bankRepo: bankRepo, batchRepo: batchRepo,
		clearRepo: clearRepo, cache: cache, feeCalc: feeCalc,
		batchProc: batchProc, logger: logger,
	}
}

// Authorize implements the VISA Authorization Flow (steps 1-4.3):
//  1. Validate card at POS terminal
//  2. POS sends to Acquirer
//  3. Acquirer sends to Card Network
//  4. Card Network routes to Issuing Bank
//  4.1. Issuer freezes funds if approved
//  4.2-4.3. Approval/rejection returned to acquirer and POS
func (s *transactionService) Authorize(ctx context.Context, req ports.AuthorizeRequest) (*domain.Transaction, error) {
	if tx, err := s.checkIdempotency(ctx, req.IdempotencyKey); err == nil {
		return tx, nil
	}

	// Step 1: Validate card (POS swipe)
	card, err := s.cardRepo.GetByID(ctx, req.CardID)
	if err != nil {
		return nil, domain.ErrCardNotFound
	}
	if !card.IsUsable() {
		if card.IsExpired() {
			return nil, domain.ErrCardExpired
		}
		return nil, domain.ErrCardBlocked
	}

	merchant, err := s.merchantRepo.GetByID(ctx, req.MerchantID)
	if err != nil {
		return nil, fmt.Errorf("merchant not found: %w", err)
	}

	issuerBank, err := s.bankRepo.GetByID(ctx, card.IssuerBankID)
	if err != nil {
		return nil, fmt.Errorf("issuer bank not found: %w", err)
	}

	fee := s.feeCalc.Calculate(req.Amount, card.Network)

	wallet, err := s.walletRepo.GetByID(ctx, card.WalletID)
	if err != nil {
		return nil, err
	}
	if err := wallet.FreezeAmount(req.Amount); err != nil {
		s.logger.Warn("auth declined: insufficient funds", zap.String("wallet_id", wallet.ID.String()))
		return nil, err
	}
	if err := s.walletRepo.UpdateWithVersion(ctx, wallet); err != nil {
		if err == domain.ErrVersionConflict {
			return s.Authorize(ctx, req) 
		}
		return nil, err
	}
	s.cache.Delete(ctx,
		fmt.Sprintf(walletByIDKey, wallet.ID),
		fmt.Sprintf(walletByUserKey, wallet.UserID),
	)

	now := time.Now()
	tx := &domain.Transaction{
		ID: uuid.New(), IdempotencyKey: req.IdempotencyKey,
		WalletID: card.WalletID, MerchantID: req.MerchantID,
		AcquirerBankID: merchant.AcquirerBankID, IssuerBankID: issuerBank.ID,
		CardID: req.CardID, CardNetwork: card.Network,
		Amount: req.Amount, Currency: req.Currency, Fee: fee,
		Type: domain.TransactionTypeAuthorization, Status: domain.TransactionStatusPending,
		Description: req.Description, Metadata: req.Metadata,
		CreatedAt: now, UpdatedAt: now,
	}
	tx.Authorize(s.generateAuthCode())

	if err := s.txRepo.Create(ctx, tx); err != nil {
		wallet.UnfreezeAmount(req.Amount)
		s.walletRepo.UpdateWithVersion(ctx, wallet)
		return nil, err
	}

	s.cacheTx(ctx, tx)
	s.setIdempotency(ctx, req.IdempotencyKey, tx)
	s.logger.Info("authorized", zap.String("tx_id", tx.ID.String()), zap.String("auth", tx.AuthorizationCode))
	return tx, nil
}

// Capture is called by merchant at end-of-day (Settlement Flow Step 1)
func (s *transactionService) Capture(ctx context.Context, txID uuid.UUID) (*domain.Transaction, error) {
	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, domain.ErrTransactionNotFound
	}
	if err := tx.Capture(); err != nil {
		return nil, err
	}
	if err := s.txRepo.Update(ctx, tx); err != nil {
		return nil, err
	}
	s.cache.Delete(ctx, fmt.Sprintf(txCacheKey, txID))
	return tx, nil
}

func (s *transactionService) Void(ctx context.Context, txID uuid.UUID) (*domain.Transaction, error) {
	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, domain.ErrTransactionNotFound
	}
	if err := tx.Void(); err != nil {
		return nil, err
	}
	wallet, err := s.walletRepo.GetByID(ctx, tx.WalletID)
	if err == nil {
		wallet.UnfreezeAmount(tx.Amount)
		s.walletRepo.UpdateWithVersion(ctx, wallet)
		s.cache.Delete(ctx, fmt.Sprintf(walletByIDKey, wallet.ID))
	}
	if err := s.txRepo.Update(ctx, tx); err != nil {
		return nil, err
	}
	s.cache.Delete(ctx, fmt.Sprintf(txCacheKey, txID))
	return tx, nil
}

// CreateSettlementBatch groups captured transactions (Settlement Flow steps 1-2)
func (s *transactionService) CreateSettlementBatch(ctx context.Context, merchantID uuid.UUID) (*domain.SettlementBatch, error) {
	since := time.Now().Add(-24 * time.Hour)
	txns, err := s.txRepo.ListCapturedByMerchant(ctx, merchantID, since)
	if err != nil {
		return nil, err
	}
	if len(txns) == 0 {
		return nil, fmt.Errorf("no captured transactions for merchant %s", merchantID)
	}

	merchant, err := s.merchantRepo.GetByID(ctx, merchantID)
	if err != nil {
		return nil, err
	}

	var total, totalFee decimal.Decimal
	txIDs := make([]uuid.UUID, 0, len(txns))
	for _, tx := range txns {
		total = total.Add(tx.Amount)
		totalFee = totalFee.Add(tx.Fee)
		txIDs = append(txIDs, tx.ID)
	}

	batchID := uuid.New()
	now := time.Now()
	sb := &domain.SettlementBatch{
		ID: batchID, AcquirerID: merchant.AcquirerBankID, MerchantID: merchantID,
		TotalAmount: total, TotalFee: totalFee, TxnCount: len(txns),
		Status: domain.TransactionStatusCaptured, CreatedAt: now,
		Transactions: make([]domain.Transaction, 0, len(txns)),
	}
	for _, tx := range txns {
		sb.Transactions = append(sb.Transactions, *tx)
	}

	if err := s.batchRepo.Create(ctx, sb); err != nil {
		return nil, err
	}
	if err := s.txRepo.UpdateBatch(ctx, txIDs, domain.TransactionStatusCaptured, batchID); err != nil {
		return nil, err
	}
	return sb, nil
}

// ProcessClearing performs card network netting (Settlement Flow step 3).
// Mutual offset transactions between the same issuer-acquirer pair are netted,
// reducing the total number of inter-bank transfers required.
func (s *transactionService) ProcessClearing(ctx context.Context, batchID uuid.UUID) ([]*domain.ClearingRecord, error) {
	sb, err := s.batchRepo.GetByID(ctx, batchID)
	if err != nil {
		return nil, domain.ErrBatchNotFound
	}

	type netKey struct{ issuerID, acquirerID uuid.UUID }
	netAmounts := make(map[netKey]decimal.Decimal)
	for _, tx := range sb.Transactions {
		k := netKey{issuerID: tx.IssuerBankID, acquirerID: tx.AcquirerBankID}
		netAmounts[k] = netAmounts[k].Add(tx.Amount.Sub(tx.Fee))
	}

	now := time.Now()
	records := make([]*domain.ClearingRecord, 0, len(netAmounts))
	for k, netAmt := range netAmounts {
		records = append(records, &domain.ClearingRecord{
			ID: uuid.New(), BatchID: batchID, NetAmount: netAmt,
			CardNetwork: domain.CardNetworkVisa,
			AcquirerID: k.acquirerID, IssuerID: k.issuerID, ClearedAt: now,
		})
	}

	if err := s.clearRepo.CreateBulk(ctx, records); err != nil {
		return nil, err
	}

	processedAt := now
	sb.Status = domain.TransactionStatusSettled
	sb.ProcessedAt = &processedAt
	s.batchRepo.Update(ctx, sb)

	s.logger.Info("clearing done",
		zap.String("batch_id", batchID.String()),
		zap.Int("netted_records", len(records)),
		zap.Int("original_txns", sb.TxnCount),
	)
	return records, nil
}

// ProcessSettlement finalizes money movement (Settlement Flow steps 4-5):
// Issuing bank confirms clearing → transfers to acquirer → acquirer → merchant bank
func (s *transactionService) ProcessSettlement(ctx context.Context, batchID uuid.UUID) error {
	records, err := s.clearRepo.GetByBatchID(ctx, batchID)
	if err != nil {
		return err
	}
	sb, err := s.batchRepo.GetByID(ctx, batchID)
	if err != nil {
		return err
	}

	for _, rec := range records {
		s.logger.Info("settling record",
			zap.String("clearing_id", rec.ID.String()),
			zap.String("net_amount", rec.NetAmount.String()),
		)
	}

	txIDs := make([]uuid.UUID, 0, len(sb.Transactions))
	for _, tx := range sb.Transactions {
		txIDs = append(txIDs, tx.ID)
	}
	if err := s.txRepo.UpdateBatch(ctx, txIDs, domain.TransactionStatusSettled, batchID); err != nil {
		return err
	}
	for _, id := range txIDs {
		s.cache.Delete(ctx, fmt.Sprintf(txCacheKey, id))
	}

	s.logger.Info("settlement complete", zap.String("batch_id", batchID.String()))
	return nil
}

func (s *transactionService) GetTransaction(ctx context.Context, txID uuid.UUID) (*domain.Transaction, error) {
	key := fmt.Sprintf(txCacheKey, txID)
	if val, err := s.cache.Get(ctx, key); err == nil && val != "" {
		var tx domain.Transaction
		if err := json.Unmarshal([]byte(val), &tx); err == nil {
			return &tx, nil
		}
	}
	tx, err := s.txRepo.GetByID(ctx, txID)
	if err != nil {
		return nil, err
	}
	s.cacheTx(ctx, tx)
	return tx, nil
}

func (s *transactionService) ListTransactionsByWallet(ctx context.Context, walletID uuid.UUID, page, pageSize int) ([]*domain.Transaction, int64, error) {
	return s.txRepo.ListByWalletID(ctx, walletID, page, pageSize)
}

func (s *transactionService) generateAuthCode() string {
	b := make([]byte, 3)
	rand.Read(b)
	return hex.EncodeToString(b)
}

func (s *transactionService) cacheTx(ctx context.Context, tx *domain.Transaction) {
	if data, err := json.Marshal(tx); err == nil {
		s.cache.Set(ctx, fmt.Sprintf(txCacheKey, tx.ID), string(data), txCacheTTL)
	}
}

func (s *transactionService) checkIdempotency(ctx context.Context, key string) (*domain.Transaction, error) {
	val, err := s.cache.Get(ctx, fmt.Sprintf(idempotencyPfx, key))
	if err != nil || val == "" {
		return nil, fmt.Errorf("not found")
	}
	var tx domain.Transaction
	if err := json.Unmarshal([]byte(val), &tx); err != nil {
		return nil, err
	}
	return &tx, nil
}

func (s *transactionService) setIdempotency(ctx context.Context, key string, tx *domain.Transaction) {
	if data, err := json.Marshal(tx); err == nil {
		s.cache.Set(ctx, fmt.Sprintf(idempotencyPfx, key), string(data), idempotencyTTL)
	}
}

// FeeCalculatorImpl computes card network interchange fees
type FeeCalculatorImpl struct{}

func NewFeeCalculator() ports.FeeCalculator { return &FeeCalculatorImpl{} }

func (f *FeeCalculatorImpl) Calculate(amount decimal.Decimal, network domain.CardNetwork) decimal.Decimal {
	ratemap := map[domain.CardNetwork]float64{
		domain.CardNetworkVisa: 0.015, domain.CardNetworkMastercard: 0.016, domain.CardNetworkAmex: 0.025,
	}
	rate, ok := ratemap[network]
	if !ok {
		rate = 0.02
	}
	return amount.Mul(decimal.NewFromFloat(rate)).Round(2)
}
