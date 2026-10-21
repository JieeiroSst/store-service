# wallet-service

A user e-wallet ("ví điện tử") service: multi-currency wallets, top-up / withdraw against linked bank accounts, P2P transfers, merchant payments with refunds, payment requests, transaction PIN, tier limits, a double-entry style ledger with statements, plus the original VISA card flow (authorize → capture → settlement batch → clearing → settle).

```
mobile / web --GraphQL--> gateway-api --REST--> wallet-service --+--> PostgreSQL  (wallets, transactions, ledger, ...)
                                                                 +--> Redis       (read cache for wallets / transactions)
                                                                 +--> BankGateway (port; simulated adapter today)
```

## Architecture (hexagonal)

```
cmd/server/main.go                 fx.New(di.App)
di/module.go                       composition root: config -> adapters -> use cases -> HTTP server
internal/
  domain/
    model/                         Wallet, Transaction, LedgerEntry, Card, BankAccount, PaymentRequest,
                                   SettlementBatch, Limits, FeeSchedule, Error{Kind, Code}
    port/
      driving.go                   WalletService, FundingService, TransferService, PaymentService,
                                   CardPaymentService, SettlementService, HistoryService, CardService,
                                   BankAccountService, PaymentRequestService (+ command structs)
      driven.go                    TxManager, *Repository, Cache, PinHasher, BankGateway
  application/                     use cases; depend only on domain/model + domain/port
  adapter/
    primary/httpapi/               gin router, handlers, error mapping, rate limiter
    secondary/
      postgres/                    repositories + TxManager (tx carried in context, SELECT ... FOR UPDATE)
      rediscache/                  port.Cache
      pinhash/                     bcrypt port.PinHasher
      bankgateway/                 simulated port.BankGateway
  infrastructure/                  postgres / redis clients, logger, http.Server lifecycle
  config/                          Consul KV (via .env) or configs/config.yaml, env-var overrides
migrations/                        001_init.sql, 002_wallet_features.sql, seed/dev_seed.sql (dev only)
consul.json                        reference value for the Consul KV key
```

### Money-movement guarantees

- Every balance change runs in one DB transaction: wallet rows are locked with `SELECT ... FOR UPDATE` (in a stable order for transfers, so no deadlocks), and the transaction row, ledger entry and wallet update commit together.
- Each balance change appends a `ledger_entries` row with `balance_after`; statements and daily-limit usage come from the ledger.
- Money-out endpoints need an `Idempotency-Key` header (or an `idempotency_key` body field). The key is unique in the DB, so retries and concurrent duplicates return the original transaction. Reusing a key for another wallet or operation returns `409 IDEMPOTENCY_KEY_REUSED`.
- Transfers, withdrawals, merchant payments and paying a payment request need the wallet's 6-digit PIN (bcrypt). After 5 wrong attempts the PIN locks for 30 minutes. Failed attempts are saved even though the operation is rejected.
- Bank calls happen outside DB transactions. A top-up is saved as `PENDING`, the bank collects, then the wallet is credited. A withdrawal is debited as `PENDING`, the bank pays out, then it is marked `COMPLETED`; if the payout fails, it is reversed with a credit entry and marked `FAILED`.

### Limits and fees

| Tier | Per transaction (USD) | Daily outgoing (USD) |
|---|---|---|
| BASIC (default) | 2,000 | 5,000 |
| VERIFIED | 10,000 | 25,000 |
| PREMIUM | 50,000 | 100,000 |

Limits scale by currency (VND ×25,000). Daily usage is the sum of today's ledger debits (UTC). Withdrawal fee is `fees.withdrawal_rate` (0.5%), with a minimum of `fees.withdrawal_min_usd` (0.5 USD, scaled by currency). Transfers are free unless `fees.transfer_rate` is set. Card interchange (VISA 1.5%, MC 1.6%, AMEX 2.5%) is charged to the merchant at clearing. Amounts must fit the currency's precision (USD/EUR 2 decimals, VND 0).

## HTTP API (`/api/v1`)

Errors: `{"error": "...", "code": "INSUFFICIENT_FUNDS"}` with 400 / 403 / 404 / 409 / 422 / 423 (PIN locked) / 502 (bank declined).
Lists: `{"items": [...], "total": n, "page": p, "page_size": s}` (`page`, `page_size` ≤ 100).
`user_id` in request bodies is the acting user; gateway-api injects it from the session.

| Method | Path | Notes |
|---|---|---|
| POST | `/wallets` | `{user_id, currency}`; one wallet per user per currency |
| GET | `/wallets/:id` · `/wallets/:id/balance` · `/wallets/:id/limits` | balance shows `balance`, `held`, `available` |
| POST | `/wallets/:id/credit` | admin adjustment `{amount, description}` |
| POST | `/wallets/:id/freeze` · `/unfreeze` · `/close` | close needs `{user_id}` and zero balance |
| PUT | `/wallets/:id/tier` · `/wallets/:id/pin` | pin: `{user_id, current_pin?, new_pin}` |
| POST | `/wallets/:id/topup` | `{user_id, bank_account_id, amount}` + Idempotency-Key |
| POST | `/wallets/:id/withdraw` | `{user_id, bank_account_id, amount, pin}` + Idempotency-Key |
| GET | `/wallets/:id/transactions` | `?type=TRANSFER_OUT,PAYMENT&status=&from=&to=` (RFC3339 or YYYY-MM-DD) |
| GET | `/wallets/:id/ledger` · `/wallets/:id/statement?from=&to=` | statement: opening/closing balance, totals |
| POST/GET | `/wallets/:id/cards` | issue `{issuer_bank_id, network, holder_name}` / list |
| GET/POST | `/cards/:id` · `/cards/:id/block` · `/cards/:id/unblock` | |
| GET | `/users/:user_id/wallets` · `/users/:user_id/wallets/:currency` | |
| POST/GET | `/users/:user_id/bank-accounts` | link `{bank_id, account_number, account_name}` (stored masked + hashed) |
| DELETE/POST | `/users/:user_id/bank-accounts/:account_id` · `.../default` | |
| GET | `/users/:user_id/payment-requests?role=incoming\|outgoing&status=` | |
| POST | `/transfers` | `{user_id, from_wallet_id, to_wallet_id \| to_user_id, amount, pin}` + Idempotency-Key |
| POST | `/payments` | `{user_id, wallet_id, merchant_id, amount, pin, metadata}` + Idempotency-Key |
| POST | `/transactions/:id/refund` | `{amount?, reason}` (empty amount = refund the rest) + Idempotency-Key |
| POST | `/payment-requests` | `{user_id, wallet_id, payer_user_id, amount, note}`; expires after `payment_requests.ttl_hours` |
| GET/POST | `/payment-requests/:id` · `/pay` · `/decline` · `/cancel` | pay `{user_id, wallet_id, pin}` is idempotent per request |
| POST | `/transactions/authorize` | card authorization (holds funds) + Idempotency-Key |
| POST | `/transactions/:id/capture` · `/void` | capture debits the hold; void releases it |
| GET | `/transactions/:id` | |
| POST | `/settlements/batches` · `/batches/:batch_id/clear` · `/batches/:batch_id/settle` | OPEN → CLEARED → SETTLED; clear/settle are idempotent |

## Run

```bash
docker compose up -d --build   # postgres, redis, service on :8080 (migrates and seeds demo banks/merchants on boot)
make test                      # domain, application (in-memory fakes), HTTP and config tests
```

On startup the service applies `migrations/*.sql` (embedded) under a Postgres advisory lock and records them in `schema_migrations`. Databases created earlier by hand or via `docker-entrypoint-initdb.d` are baselined from their existing tables, then upgraded. `make migrate` still runs the same files with `psql --single-transaction`. Set `database.auto_migrate: false` to skip.

Migration 002 changes existing data: `user_id` uniqueness becomes `(user_id, currency)` for wallets that aren't closed, card-only transaction columns become nullable, the zero-UUID `batch_id` sentinel becomes `NULL`, and `ledger_entries`, `bank_accounts` and `payment_requests` are added. Ledger history starts at the migration, so statements for earlier periods open at 0.

## Deploy

[`chart/wallet-service`](../chart/wallet-service) uses `chart/common`. Argo CD picks it up through `values-dev.yaml` (store-dev), and the umbrella chart includes it as `walletService`.

- `consulSeed` writes [`consul.json`](consul.json) to Consul key `wallet-service` if the key is missing. The chart copy is `chart/wallet-service/files/consul.json`; keep the two in sync.
- The postgres chart (`serviceDatabases`) creates the database `wallet_service` and role `wallet_service_svc`, writes the credentials to Secret `wallet-service-db-credentials`, and patches `database.user/password` into the Consul key. The pod also gets the credentials as `DATABASE_USER` / `DATABASE_PASSWORD` from that Secret, so it works even when the key was seeded after the provisioner ran.
- dev/staging copy the Secret from `default` with an ExternalSecret. dev also sets `seedDevData: true` (demo banks and merchants). prod reads `store/prod/wallet-service/postgres` and `store/prod/wallet-service/env-file` from the secret manager.
- Probes: `/health` for startup and liveness (startup allows 3 minutes for migrations), `/health/ready` for readiness (it pings Postgres).
- CI: `deploy/ci/services.yaml` (in-cluster build to store-dev) and `.github/workflows/wallet-service-cicd.yml` (manual GHCR release, then promote.yml).

## Configuration

Order of precedence (highest first): environment variables → Consul KV **or** `configs/config.yaml` → built-in defaults.

- **Consul** (deployed environments, same convention as the other services): `HostConsul`, `KeyConsul`, `ServiceConsul` come from the environment or from `.env` in the working directory (the chart's `envFile` mounts it at `/app/.env`). The service reads key `KeyConsul` as JSON in the shape of [`consul.json`](consul.json), and refuses to start if Consul can't be reached or the key is missing or empty. Seed the key from `consul.json` (`consulSeed.key: wallet-service` once the service has a chart).
- **Local file**: if `HostConsul` is not set, `configs/config.yaml` is used. Copy `.env.example` to `.env` to switch local runs to Consul.
- **Env overrides**: config keys map to env vars with `.` → `_`, e.g. `DATABASE_DSN`, `REDIS_ADDR`, `FEES_WITHDRAWAL_RATE`, or `BANK_GATEWAY_DECLINE_ABOVE` to make the simulated bank decline large amounts.

The startup log line `configuration loaded` shows which source was used (`consul:<key>`, `file:<path>` or `defaults`).
