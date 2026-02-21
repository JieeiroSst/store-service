# wallet-service

A production-ready payment wallet service implementing the VISA card payment flow using **Hexagonal (Clean) Architecture** in Go.

## Architecture

```
wallet-service/
├── cmd/                         # Entry point
│   └── main.go
├── configs/
│   └── config.yaml
├── internal/
│   ├── core/                    # ← Domain (innermost layer, no dependencies)
│   │   ├── domain/              # Entities: Card, Wallet, Transaction, Batch
│   │   ├── ports/
│   │   │   ├── input/           # Driving ports: AuthorizationService, SettlementService
│   │   │   └── output/          # Driven ports: Repository, Cache interfaces
│   │   └── services/            # Business logic (implements input ports)
│   ├── adapters/
│   │   ├── primary/http/        # HTTP handlers (driving adapter)
│   │   └── secondary/
│   │       ├── postgres/        # PostgreSQL repositories (driven adapter)
│   │       └── redis/           # Redis cache repository (driven adapter)
│   └── infrastructure/
│       ├── config/              # Configuration loading (Viper)
│       ├── database/            # PostgreSQL connection pool
│       ├── cache/               # Redis client
│       └── fx/                  # Uber FX dependency injection modules
├── migrations/
│   └── 001_init.sql             # Full database schema
├── pkg/
│   └── algorithm/               # Reusable algorithms
│       ├── token_bucket.go      # Rate limiting
│       ├── lru_cache.go         # Generic LRU cache (in-process)
│       ├── consistent_hash.go   # Distributed cache node selection
│       ├── batch_processor.go   # Chunked/parallel batch processing
│       ├── bloom_filter.go      # Duplicate transaction detection
│       ├── netting.go           # Bilateral netting for clearing
│       └── sliding_window.go    # Fraud detection rate limiting
└── docker-compose.yaml
```

## VISA Payment Flows Implemented

### Authorization Flow (real-time)
```
Step 0:  Issuing bank issues card to cardholder
Step 1:  Cardholder swipes card at POS terminal
Step 2:  POS → Acquiring bank (sends transaction)
Step 3:  Acquiring bank → Card Network (VISA)
Step 4:  Card Network → Issuing bank (approval request)
Step 4.1 Issuing bank freezes funds if approved
Step 4.2 Approval/rejection → Acquirer
Step 4.3 Approval/rejection → POS terminal
```

### Capture & Settlement Flow (end-of-day)
```
Step 1:  Merchant hits "capture" on POS → batch of transactions
Step 2:  Acquirer sends batch file to card network
Step 3:  Card network performs clearing (bilateral netting)
Step 4:  Issuing banks confirm & transfer to acquiring banks
Step 5:  Acquiring bank transfers to merchant bank
```

## Algorithms

| Algorithm | Location | Purpose |
|-----------|----------|---------|
| Token Bucket | `pkg/algorithm/token_bucket.go` | Rate limiting (1000 TPS) |
| LRU Cache | `pkg/algorithm/lru_cache.go` | In-process hot-path caching |
| Consistent Hash | `pkg/algorithm/consistent_hash.go` | Redis node distribution |
| Batch Processor | `pkg/algorithm/batch_processor.go` | End-of-day batch settlement |
| Bloom Filter | `pkg/algorithm/bloom_filter.go` | Duplicate transaction detection |
| Bilateral Netting | `pkg/algorithm/netting.go` | Card network clearing |
| Sliding Window | `pkg/algorithm/sliding_window.go` | Fraud detection |

## Cache Strategy (Cache-Aside with Invalidation)

- **Read**: Check Redis first → on miss, read from PostgreSQL and populate cache
- **Write**: Update PostgreSQL → **delete** cached keys (invalidation)
- **Key naming**: `wallet:card:{id}`, `wallet:tx:{id}`, `wallet:wallet:{id}`, etc.

## Quick Start

```bash
# Start dependencies
docker compose up -d

# Run database migrations
make migrate

# Start the service
make run
```

## API Endpoints

### Authorization
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/authorization/authorize` | Authorize a transaction (Steps 1-4.3) |
| POST | `/api/v1/authorization/void` | Void an authorization |
| GET  | `/api/v1/authorization/:id` | Get transaction details |
| POST | `/api/v1/authorization/cards` | Issue a card (Step 0) |

### Settlement
| Method | Path | Description |
|--------|------|-------------|
| POST | `/api/v1/settlement/capture` | Capture a single authorization |
| POST | `/api/v1/settlement/batch-capture` | End-of-day batch capture (Steps 1-2) |
| POST | `/api/v1/settlement/clearing/:batch_id` | Run clearing (Step 3) |
| POST | `/api/v1/settlement/confirm/:batch_id` | Confirm settlement (Step 4) |
| POST | `/api/v1/settlement/transfer/:batch_id` | Transfer to merchant bank (Step 5) |
| GET  | `/api/v1/settlement/batch/:batch_id` | Get batch details |
| GET  | `/api/v1/settlement/pending/:merchant_id` | Get pending captures |

## Go Module

```
github.com/JIeeiroSst/wallet-service
```

```
https://bytebytego.com/guides/how-does-visa-work-when-we-swipe-a-credit-card-at-a-merchants-shop/

VISA, Mastercard, and American Express act as card networks for clearing and settling funds. The card acquiring bank and the card issuing bank can be – and often are – different. If banks were to settle transactions one by one without an intermediary, each bank would have to settle the transactions with all the other banks. This is quite inefficient.

The diagram shows VISA’s role in the credit card payment process. There are two flows involved. Authorization flow happens when the customer swipes the credit card. Capture and settlement flow occurs when the merchant wants to get the money at the end of the day.
Authorization Flow

    Step 0: The card issuing bank issues credit cards to its customers.

    Step 1: The cardholder wants to buy a product and swipes the credit card at the Point of Sale (POS) terminal in the merchant’s shop.

    Step 2: The POS terminal sends the transaction to the acquiring bank, which has provided the POS terminal.

    Steps 3 and 4: The acquiring bank sends the transaction to the card network, also called the card scheme. The card network sends the transaction to the issuing bank for approval.

    Steps 4.1, 4.2, and 4.3: The issuing bank freezes the money if the transaction is approved. The approval or rejection is sent back to the acquirer, as well as the POS terminal.

Capture and Settlement Flow

    Steps 1 and 2: The merchant wants to collect the money at the end of the day, so they hit ”capture” on the POS terminal. The transactions are sent to the acquirer in batches. The acquirer sends the batch file with transactions to the card network.

    Step 3: The card network performs clearing for the transactions collected from different acquirers, and sends the clearing files to different issuing banks.

    Step 4: The issuing banks confirm the correctness of the clearing files, and transfer money to the relevant acquiring banks.

    Step 5: The acquiring bank then transfers money to the merchant’s bank.

    Step 4: The card network clears the transactions from different acquiring banks. Clearing is a process in which mutual offset transactions are netted, so the number of total transactions is reduced.

In the process, the card network takes on the burden of talking to each bank and receives service fees in return.
```