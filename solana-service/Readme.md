## README.md - Complete Circle Integration Guide

```
https://solana.com
https://console.circle.com
```

```markdown
# Complete Solana + Circle Wallets Integration

Full-featured integration between Solana blockchain and Circle Programmable Wallets with all Circle API features.

## 🌟 Features

### Circle Wallet Management
- ✅ Wallet Sets (create, list, update)
- ✅ Multi-blockchain wallets (Solana, Ethereum, Polygon, etc.)
- ✅ Wallet freezing/unfreezing
- ✅ Balance tracking (tokens + NFTs)
- ✅ Metadata management

### Transaction Operations
- ✅ Token transfers
- ✅ NFT transfers
- ✅ Smart contract execution
- ✅ Fee estimation (LOW/MEDIUM/HIGH)
- ✅ Transaction acceleration
- ✅ Transaction cancellation
- ✅ Address validation

### Token Management
- ✅ List supported tokens
- ✅ Import custom tokens
- ✅ NFT token support
- ✅ Multi-chain token support

### Solana ↔ Circle Bridge
- ✅ Bidirectional transfers
- ✅ Status tracking
- ✅ Automatic monitoring

## 🚀 Quick Start

```bash
# Environment variables
export SOLANA_RPC_URL="https://api.devnet.solana.com"
export CIRCLE_API_KEY="your-circle-api-key"
export PORT="8080"

# Run
go run cmd/api/main.go
```

## 📚 Complete API Reference

### Wallet Operations

```bash
# Create wallet set
POST /api/v1/circle/wallet-sets
{
  "name": "Production Wallets",
  "custodyType": "DEVELOPER"
}

# Create Solana wallet
POST /api/v1/circle/wallets
{
  "walletSetId": "wallet-set-id",
  "blockchain": "SOL",
  "accountType": "EOA"
}

# Get wallet balance
GET /api/v1/circle/wallets/{walletId}/balance

# Get wallet NFTs
GET /api/v1/circle/wallets/{walletId}/nfts

# Freeze wallet
PUT /api/v1/circle/wallets/{walletId}/freeze

# Unfreeze wallet
PUT /api/v1/circle/wallets/{walletId}/unfreeze
```

### Transaction Operations

```bash
# Create transfer
POST /api/v1/circle/transactions/transfer
{
  "walletId": "wallet-id",
  "destinationAddress": "destination-address",
  "tokenId": "SOL",
  "amount": "0.001",
  "fee": {
    "type": "level",
    "config": {
      "feeLevel": "MEDIUM"
    }
  }
}

# Transfer NFT
POST /api/v1/circle/transactions/nft-transfer
{
  "walletId": "wallet-id",
  "destinationAddress": "destination-address",
  "tokenId": "nft-token-id",
  "nftTokenIds": ["token-1", "token-2"]
}

# Execute smart contract
POST /api/v1/circle/transactions/contract-execution
{
  "walletId": "wallet-id",
  "contractAddress": "contract-address",
  "abiParameters": [
    {
      "name": "param1",
      "type": "uint256",
      "value": "1000"
    }
  ]
}

# Estimate fee
GET /api/v1/circle/transactions/estimate-fee?walletId=xxx&destinationAddress=yyy&tokenId=SOL&amount=0.001

# Get transaction
GET /api/v1/circle/transactions/{txId}
```

### Token Operations

```bash
# List Solana tokens
GET /api/v1/circle/tokens?blockchain=SOL

# Get token details
GET /api/v1/circle/tokens/{tokenId}

# Import custom token
POST /api/v1/circle/tokens/import
{
  "name": "My Token",
  "symbol": "MTK",
  "decimals": 9,
  "blockchain": "SOL",
  "tokenAddress": "token-address",
  "standard": "SPL"
}
```

### Bridge Operations

```bash
# Solana → Circle
POST /api/v1/bridge/solana-to-circle
{
  "fromSolanaAddress": "solana-address",
  "toCircleWalletId": "circle-wallet-id",
  "amount": 1000000,
  "privateKey": "base58-private-key"
}

# Circle → Solana
POST /api/v1/bridge/circle-to-solana
{
  "fromCircleWalletId": "circle-wallet-id",
  "toSolanaAddress": "solana-address",
  "amount": "0.001"
}

# Get bridge transfer status
GET /api/v1/bridge/transfers/{transferId}
```

## 💡 Usage Examples

### Complete Workflow

```go
// 1. Create wallet set
walletSet := circleWalletService.CreateWalletSet(ctx, "My Wallets", domain.CustodyTypeDeveloper)

// 2. Create Solana wallet
wallet := circleWalletService.CreateSolanaWallet(ctx, walletSet.ID)

// 3. Get balances
balances := circleWalletService.GetWalletBalance(ctx, wallet.ID)

// 4. Estimate fee
feeEstimate := circleTransactionService.EstimateFee(
    ctx,
    wallet.ID,
    "destination-address",
    "SOL",
    "0.001",
)

// 5. Create transfer with custom fee
transferReq := domain.TransferRequest{
    WalletID:           wallet.ID,
    DestinationAddress: "destination-address",
    TokenID:            "SOL",
    Amount:             "0.001",
    Fee: &domain.FeeConfiguration{
        Type: domain.FeeTypeLevel,
        Config: domain.FeeConfig{
            FeeLevel: "MEDIUM",
        },
    },
}

tx := circleTransactionService.CreateTransfer(ctx, transferReq)

// 6. Monitor transaction
status := circleTransactionService.GetTransaction(ctx, tx.ID)

// 7. Accelerate if needed
accelerated := circleTransactionService.AccelerateTransaction(ctx, tx.ID, "HIGH")
```

### NFT Transfer

```go
// Transfer NFT
nftReq := domain.NFTTransferRequest{
    WalletID:           wallet.ID,
    DestinationAddress: "recipient-address",
    TokenID:            "nft-collection-id",
    NFTTokenIDs:        []string{"nft-1", "nft-2"},
}

nftTx := circleTransactionService.CreateNFTTransfer(ctx, nftReq)
```

### Smart Contract Execution

```go
// Execute contract
contractReq := domain.ContractExecutionRequest{
    WalletID:        wallet.ID,
    ContractAddress: "contract-address",
    ABIParameters: []domain.ABIParameter{
        {
            Name:  "recipient",
            Type:  "address",
            Value: "0x123...",
        },
        {
            Name:  "amount",
            Type:  "uint256",
            Value: "1000000",
        },
    },
}

contractTx := circleTransactionService.ExecuteContract(ctx, contractReq)
```

## 🏗️ Architecture

- **Hexagonal/Clean Architecture**
- **Domain-Driven Design**
- **Port & Adapter Pattern**
- **Dependency Injection**
- **Separation of Concerns**

## 🔐 Security Features

- Wallet freezing capability
- Transaction validation
- Address validation
- Idempotency keys
- Fee estimation before execution

## 📊 Monitoring

- Transaction status tracking
- Bridge transfer monitoring
- Automatic status updates
- Error handling and reporting

Complete Circle Wallets integration with full API coverage! 🎉
```