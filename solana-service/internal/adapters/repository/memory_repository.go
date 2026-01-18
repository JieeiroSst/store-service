package repository

import (
	"context"
	"sync"

	"github.com/JIeeiroSst/solana-service/internal/core/domain"
)

type MemoryTransactionRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.Transaction
}

func NewMemoryTransactionRepository() *MemoryTransactionRepository {
	return &MemoryTransactionRepository{
		data: make(map[string]*domain.Transaction),
	}
}

func (r *MemoryTransactionRepository) Save(ctx context.Context, tx *domain.Transaction) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[tx.Signature] = tx
	return nil
}

func (r *MemoryTransactionRepository) FindBySignature(ctx context.Context, signature string) (*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	tx, ok := r.data[signature]
	if !ok {
		return nil, domain.ErrTransactionFailed
	}
	return tx, nil
}

func (r *MemoryTransactionRepository) FindByAddress(ctx context.Context, address string) ([]*domain.Transaction, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var txs []*domain.Transaction
	for _, tx := range r.data {
		txs = append(txs, tx)
	}
	return txs, nil
}

type MemoryAccountRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.Account
}

func NewMemoryAccountRepository() *MemoryAccountRepository {
	return &MemoryAccountRepository{
		data: make(map[string]*domain.Account),
	}
}

func (r *MemoryAccountRepository) Save(ctx context.Context, account *domain.Account) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.data[account.PublicKey.String()] = account
	return nil
}

func (r *MemoryAccountRepository) FindByPublicKey(ctx context.Context, pubkey string) (*domain.Account, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	account, ok := r.data[pubkey]
	if !ok {
		return nil, domain.ErrAccountNotFound
	}
	return account, nil
}

type MemoryBridgeRepository struct {
	mu   sync.RWMutex
	data map[string]*domain.BridgeTransfer
}

func NewMemoryBridgeRepository() *MemoryBridgeRepository {
	return &MemoryBridgeRepository{
		data: make(map[string]*domain.BridgeTransfer),
	}
}

func (m *MemoryBridgeRepository) Save(ctx context.Context, transfer *domain.BridgeTransfer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	// Use transfer ID as the key
	m.data[transfer.ID] = transfer
	return nil
}

func (m *MemoryBridgeRepository) Update(ctx context.Context, transfer *domain.BridgeTransfer) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, exists := m.data[transfer.ID]; !exists {
		return domain.ErrBridgeNotFound
	}

	m.data[transfer.ID] = transfer
	return nil
}

func (m *MemoryBridgeRepository) FindByID(ctx context.Context, id string) (*domain.BridgeTransfer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	transfer, exists := m.data[id]
	if !exists {
		return nil, domain.ErrBridgeNotFound
	}

	return transfer, nil
}

func (m *MemoryBridgeRepository) FindBySolanaAddress(ctx context.Context, address string) ([]*domain.BridgeTransfer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var transfers []*domain.BridgeTransfer

	for _, transfer := range m.data {
		if transfer.SolanaAddress == address {
			transfers = append(transfers, transfer)
		}
	}

	return transfers, nil
}

func (m *MemoryBridgeRepository) FindByCircleWallet(ctx context.Context, walletID string) ([]*domain.BridgeTransfer, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	var transfers []*domain.BridgeTransfer

	for _, transfer := range m.data {
		if transfer.CircleWalletID == walletID {
			transfers = append(transfers, transfer)
		}
	}

	return transfers, nil
}
