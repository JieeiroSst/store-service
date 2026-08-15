# Voucher Platform (voucher-service)

A UrBox-style Voucher Platform in Go: B2C retail + B2B corporate bulk gifting, vouchers sourced from three kinds of providers (real-time merchant API, pre-imported code stock, self-issued), built with strict Hexagonal Architecture and `go.uber.org/fx` dependency injection.

## Architecture

```
/cmd/api/main.go                 fx.New(app.Module).Run()
/internal/
  /domain/                       pure business logic — zero external imports
    voucher/  order/  merchant/  user/  corporate/  wallet/  shared/
  /application/                  use cases + orchestration, one package per bounded context
    <context>/port.go            driving (inbound) + driven (outbound) interfaces
    <context>/service.go         implements the driving port
    <context>/module.go          fx.Module binding service -> interface
  /adapters/
    /inbound/http/                gin handlers, router, middleware, error_mapper
    /inbound/consumer/             Kafka consumers
    /outbound/postgres/            gorm repositories, one file per aggregate
    /outbound/redis/                locker, rate limiter
    /outbound/provider/            MerchantProvider strategies: api / stock / self
    /outbound/payment/              VNPay, Momo gateways
    /outbound/notifier/             email (real SMTP), SMS (logging stand-in)
    /outbound/publisher/            Kafka producer for the outbox relay
    /outbound/authtoken/            JWT issuer
    /outbound/internalgateway/      cross-context translator adapters (see below)
  /platform/                     config, logger, db, redis, kafka, tracing, txmanager,
                                  lock, idempotency, outbox, consul, server, scheduler
  /app/module.go                 root fx.Options composing every module
/migrations/                     golang-migrate SQL (000001_init.up.sql / .down.sql)
```

**Dependency direction always points inward.** Domain imports nothing. Application imports only domain + platform port interfaces. Adapters implement those interfaces and are the only layer allowed to import frameworks (gin, gorm, kafka-go, redis).

**Cross-bounded-context calls never call another context's concrete service directly.** Each consuming context declares its own narrow interface in its own `port.go` (e.g. `order.VoucherIssuer`, `distribution.BudgetChecker`, `scheduler.VoucherExpirer`) and `adapters/outbound/internalgateway/` provides the one adapter file allowed to import both sides and translate between them. The single exception is `Payment -> Order`: rather than a same-process call (which would create an `Order -> Payment -> Order` cycle in the fx dependency graph — see below), a settled payment is published as a `payment.settled` outbox event and consumed by `adapters/inbound/consumer`, which calls `Order.ConfirmPayment` directly, the same way an HTTP handler would.

## The Voucher lifecycle

```
CREATED → ISSUED → ACTIVE → REDEEMED
                          → EXPIRED
                          → REVOKED
CREATED → REVOKED
ISSUED  → REVOKED
```

Every transition is a method on the `Voucher` aggregate (`internal/domain/voucher/voucher.go`) that validates the current state and returns a domain error on an illegal transition — callers never set `Status` directly. `Version` is bumped on every transition for optimistic locking.

**Redeem is the platform's most safety-critical path** (`internal/application/voucher/redeem.go`):
1. Redis lock (`SETNX`+`PX`, bounded retries, fail-closed on Redis unavailability) serializes concurrent redeem attempts on the same voucher.
2. An idempotency claim (Postgres-backed) makes retries replay-safe: a completed request replays its cached response; a domain rejection (already redeemed, expired, wrong PIN) is cached as a terminal, replayable failure; a transient failure (provider timeout) releases the claim so the caller can retry.
3. `SELECT ... FOR UPDATE` row-locks the voucher inside the transaction.
4. For API-sourced vouchers, the merchant is called **before** any local state changes, so a rejected/timed-out call never leaves local and remote state disagreeing.
5. The domain transition, the row update (`WHERE version = <version as loaded>`), and an outbox event write all commit in one transaction.

Only `adapters/inbound/http/error_mapper.go` translates domain/application errors to HTTP status codes — no driver/DB error text ever reaches a client.

## Service review — what was added beyond the initial 8 and why

The initial contexts (Auth, Catalog/Merchant, Order, Voucher, Wallet, Distribution, Corporate/B2B, Reconciliation) leave several cross-cutting needs unmet. All ten candidate services from the spec were evaluated; all ten were added:

| Service | Why it was needed |
|---|---|
| **Payment** | Order needed a real redirect/webhook/refund flow, not an inline `if provider == "vnpay"` branch buried in Order. Separated so gateway-specific signing (VNPay HMAC-SHA512, Momo HMAC-SHA256) stays out of the order domain. |
| **Notification** | Every other service (voucher issued, low-stock, order fulfilled) needs to notify someone; centralizing it behind one `Notifier` port avoids each context reimplementing its own SMTP/SMS glue. |
| **Inventory/Stock** | `StockProvider` needs a code pool to claim from (`SELECT ... FOR UPDATE SKIP LOCKED`), and someone has to own bulk import + low-stock visibility — that's a distinct concern from Voucher's redeem logic. |
| **Idempotency** | Both a generic HTTP-middleware response cache (`Idempotency-Key` header) and domain-level idempotency (vouchers/orders/payments) are needed — the former protects the transport, the latter protects the use case even when invoked from a Kafka consumer retry. Both are implemented; see `internal/platform/idempotency`. |
| **Audit/Event Log** | Every domain state change flows through the outbox already (for the relay); a dedicated consumer persisting those events to `audit_log` gives a durable, queryable trail independent of the operational tables, without adding write-path latency to the services that emit the events. |
| **Reporting/Analytics** | Read-only aggregate queries (redemption rate, corporate spend) don't belong mixed into Voucher/Corporate's write-path ports — kept as its own read-model service. |
| **Scheduler/Cron** | Voucher expiry, reconciliation, and low-stock alerts all need to run on a timer; centralizing job registration avoids each context managing its own goroutine. |
| **File** | CSV import of stock codes and CSV export of reports is an I/O concern orthogonal to Inventory/Reporting's own ports — kept separate so those ports stay framework-agnostic. |
| **API-Key/Partner Gateway** | Merchant/POS integrations calling `redeem` directly need a different trust boundary (HMAC-signed requests, per-partner rate limiting) than a logged-in end user's JWT — this is genuinely a separate concern from `Auth`. |
| **Outbox** | Required for reliable event delivery: domain writes and the event that describes them must commit atomically, and delivery to Kafka must be decoupled from that commit so a broker outage can never roll back business state. |

## Honesty notes — what's real vs. simplified

- **Payment gateways** (`adapters/outbound/payment/{vnpay,momo}.go`) implement the real request-signing/webhook-verification schemes, but cannot complete a live transaction without real VNPay/Momo merchant credentials (`VNPAY_*` / `MOMO_*` env vars) — that's inherent to not having a merchant account, not a shortcut.
- **SMS notification** is a structured-logging stand-in behind the real `Notifier` port (no SMS provider credentials available). Swapping in Twilio/ESMS is a single new adapter file plus a module.go change.
- **OpenTelemetry** tracing is wired end-to-end (context propagation via a gin middleware, spans per request) with a stdout exporter by default — pointing it at a real collector is a one-line change in `platform/tracing`.
- **Consul self-registration** matches this monorepo's existing `.env` convention but is optional (`CONSUL_ENABLED`) and never blocks startup.
- **Reconciliation** compares internal payment records for internal consistency (every settled payment has a `provider_txn_ref`) since there's no external provider statement feed available in this environment; the port (`PaymentRecordSource`) is shaped so a real statement-file comparison can be dropped in without changing callers.
- **Order's checkout** funds itself via wallet (synchronous) or a payment gateway redirect (async, confirmed via the `payment.settled` outbox event). The wallet has an admin/top-up HTTP endpoint (`POST /wallets/:ownerType/:ownerId/credit`) since without one, wallet-funded checkout would be untestable — a real deployment would gate this behind an actual funding flow.

## Two correctness bugs found and fixed during live end-to-end testing

Both were caught by actually booting the service against real Postgres/Redis/Kafka and exercising the checkout flow, not just by unit tests:

1. **A deadlock in nested cross-context transactions.** `Order.Checkout`'s transaction row-locks the order, then (via the `VoucherIssuer` driven port) calls into `Voucher.IssueVouchers`, which — before the fix — opened a **second, independent** Postgres transaction. That second transaction's `INSERT INTO vouchers ... order_id = ...` triggers a foreign-key check that blocks on the still-open outer transaction's lock, while the outer transaction is itself blocked waiting for the Go callback to return. Postgres's deadlock detector can't see this cycle (the outer side is blocked in application code, not on a DB lock), so it hung forever instead of erroring. Fixed by making `TxManager.WithinTx` re-entrant: it now joins an already-active transaction from context instead of opening a nested one (`internal/platform/txmanager/gorm_tx_manager.go`).
2. **A broken optimistic-lock guard on multi-transition saves.** The `Save` methods inferred the row's previous version as `currentVersion - 1`. That's only true when exactly one domain transition happens between load and save — true for Voucher's flows, false for `Order.Checkout`'s wallet path (`MarkPaid` → `MarkFulfilling` → `Complete`, three transitions before one `Save`), which always produced a spurious `version_conflict`. Fixed by adding a `PersistedVersion` field to every aggregate, set only by the repository at load time and never touched by domain mutators, so `Save` guards against the version as actually read rather than a guess.

## Running locally

```bash
cp .env .env.local   # already has sane local defaults
make up               # docker-compose up -d postgres redis kafka
make run              # go run ./cmd/api — runs migrations automatically on boot
```

The server listens on `PORT` (default `3000`). Health check: `GET /health`.

### Example flow

```bash
# Register + login
curl -X POST localhost:3000/api/v1/auth/register -d '{"email":"a@b.com","password":"secret123"}'
TOKEN=$(curl -X POST localhost:3000/api/v1/auth/login -d '{"email":"a@b.com","password":"secret123"}' | jq -r .token)

# Register a self-issued merchant
MID=$(curl -X POST localhost:3000/api/v1/merchants -H "Authorization: Bearer $TOKEN" \
  -d '{"name":"Coffee Co","provider_type":"self","config":{}}' | jq -r .id)

# Issue a voucher (response includes the plaintext PIN — the only time it's ever returned)
curl -X POST localhost:3000/api/v1/vouchers/issue -H "Authorization: Bearer $TOKEN" \
  -d "{\"merchant_id\":\"$MID\",\"product_sku\":\"SKU1\",\"denomination_amount\":100000,\"currency\":\"VND\",\"quantity\":1,\"idempotency_key\":\"k1\"}"

# Activate (assigns an owner, ISSUED -> ACTIVE) then redeem
curl -X POST localhost:3000/api/v1/vouchers/$VID/activate -H "Authorization: Bearer $TOKEN" \
  -d '{"owner_type":"user","owner_id":"<user_id>"}'
curl -X POST localhost:3000/api/v1/vouchers/$VID/redeem -H "Authorization: Bearer $TOKEN" \
  -d '{"pin":"<pin from issue>","amount":100000,"currency":"VND","idempotency_key":"k2"}'
```

See `internal/adapters/inbound/http/router.go` for the full route table (orders/checkout, corporate budgets, B2B distribution jobs + claim tokens, payment webhooks, partner HMAC redeem, reporting).

## Testing

```bash
make test              # unit tests: domain state machines + voucher redeem orchestration (fakes, no Docker)
make test-integration  # + Postgres integration test via testcontainers-go (requires Docker):
                        #   optimistic-lock conflict rejection, and FOR UPDATE row-lock blocking a
                        #   concurrent reader — both verified against a real Postgres instance
```

## Configuration

All env vars are in `.env` with local defaults. Notable ones:

| Var | Purpose |
|---|---|
| `POSTGRES_*`, `REDIS_*`, `KAFKA_BROKERS` | Infra connection settings |
| `JWT_SECRET`, `JWT_EXPIRATION_MINUTES` | Auth token signing |
| `PARTNER_HMAC_ENC_KEY` | Encrypts partner API-key secrets at rest (AES-GCM) so HMAC verification can decrypt and recompute the signature — a one-way hash can't be used here, unlike password auth |
| `OUTBOX_RELAY_INTERVAL_MS` | How often the outbox relay polls for unpublished events |
| `VOUCHER_EXPIRY_SWEEP_MINUTES` | Not directly used by the scheduler's cron spec (fixed at `*/5 * * * *`) — reserved for a future configurable schedule |
| `VNPAY_*`, `MOMO_*`, `SMTP_*` | Real credentials for those adapters; blank by default in dev |

## What's stubbed vs. fully implemented (depth tiering)

**Full business logic + Postgres/Redis adapters + tests:** Merchant/Provider abstraction (api/stock/self), Voucher (full state machine, double-spend prevention), Order (checkout orchestration, wallet + gateway payment).

**Real domain entities + real ports + at least one working end-to-end use case, simpler internals:** Auth, Wallet, Corporate, Distribution, Reconciliation, Payment, Notification, Inventory, Audit, Reporting, Scheduler, File, API-Key/Partner Gateway.

**Solid, not stubbed, because the core depends on them for correctness:** TxManager, Locker, IdempotencyStore, Outbox (transactional outbox + Kafka relay).
