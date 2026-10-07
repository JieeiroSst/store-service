# sard-card-service

The sARD card processor. Its model follows the Pismo platform: a **program**
defines a card product, an **account** holds a customer's money or credit line
inside a program, and **cards** are issued on that account. **Authorizations**
are identified by processing codes, and each one is either confirmed by
clearing or cancelled.

It works with the other sARD services:

```
user-service ──► sard-customer-info-service ──► sard-card-service ◄──► sard-auth-service
 (login)            (profile + CCCD eKYC)          (this service)      (OTP payment authentication)
```

- An account can only be opened for a customer that `sard-customer-info-service`
  reports as `card_eligibility.eligible`: the customer is ACTIVE and their KYC
  is VERIFIED and not expired. The account stores the customer's `user_id`.
- An ECOM purchase at or above the program's `step_up_threshold` needs an
  `authentication_id` from `sard-auth-service`. The card service consumes it
  through `POST /internal/v1/payment-authentications/{id}/consume` on that
  service. Each id is single use, and it is bound to the card, the currency and
  a maximum amount.
- `sard-auth-service` finds the card and its owner through
  `POST /internal/v1/cards/resolve` on this service.

## Programs

| Program          | Mode    | BIN    | Card types                   | Default limits (tx / day, VND) | Step-up from |
|------------------|---------|--------|------------------------------|--------------------------------|--------------|
| `DEBIT_CLASSIC`  | DEBIT   | 730100 | PLASTIC, VIRTUAL, TEMPORARY  | 20M / 50M                      | 2M           |
| `DEBIT_PLATINUM` | DEBIT   | 730200 | PLASTIC, VIRTUAL, TEMPORARY  | 100M / 300M                    | 5M           |
| `CREDIT_GOLD`    | CREDIT  | 735100 | PLASTIC, VIRTUAL, TEMPORARY  | 50M / 100M, credit line 50M (max 200M) | 1M  |
| `PREPAID_ONLINE` | PREPAID | 739900 | VIRTUAL, TEMPORARY           | 5M / 10M                       | every ECOM purchase |

## Business rules

- **Account.** A customer can have one open account per program. The account
  keeps three amounts:
  - `balance`: posted money. For a debit or prepaid account it is the funds;
    for a credit account it goes negative as the customer spends.
  - `credit_limit`: only set on CREDIT accounts.
  - `held`: the total of open authorizations.

  `available = balance + credit_limit - held`.

  Account statuses are NORMAL, BLOCKED and CANCELLED. An account can only be
  cancelled when it has no holds and no debt, and cancelling it also cancels
  its cards. Payments (processing code `28`) credit the balance: a top-up for
  debit or prepaid accounts, a bill payment for credit accounts.
- **Cards.** An account can have at most 1 live PLASTIC card, `max_active_virtual`
  VIRTUAL cards and 1 live TEMPORARY card.
  - PLASTIC cards start PENDING and are activated with their CVV.
  - VIRTUAL cards start NORMAL and work for ECOM only.
  - TEMPORARY cards are virtual and expire 24 hours after issue (`valid_until`).

  Card statuses follow Pismo: PENDING, NORMAL, BLOCKED (temporary), and the
  terminal CANCELLED, LOST, ROBBED and DAMAGED. A report with `reissue: true`
  issues a replacement card on the same account, with the same type, limits
  and controls.
- **PAN and CVV.** The PAN is the BIN, 9 digits from `crypto/rand` and a Luhn
  check digit. It is stored AES-256-GCM encrypted, plus an HMAC for lookups.
  The CVV is never stored: it is recomputed as HMAC(PAN | YYMM | service code).
  PAN and CVV are returned only once, when the card is issued.
- **PIN.** Only PLASTIC cards have a PIN. It is stored as PBKDF2-SHA256, and 3
  wrong tries lock it (code `75`) until a new PIN is set.
- **Authorization.** Processing codes:
  - `00` purchase: places a hold.
  - `01` cash withdrawal: ATM only, PIN required, places a hold.
  - `20` refund: credited and posted immediately.

  A hold becomes **CONFIRMED** (clearing) for an amount up to the authorized
  one, which releases the hold and posts the debit. It can instead be
  **CANCELLED**, which only releases the hold.

  Checks run in this order (ISO 8583 codes):
  1. account BLOCKED `62`, CANCELLED `46`
  2. card PENDING `78`, BLOCKED `62`, LOST `41`, ROBBED `43`, CANCELLED or DAMAGED `46`
  3. expiry `54`
  4. CVV `N7`, required for ECOM
  5. PIN `55` or `75`
  6. channel or withdrawal rule `57`
  7. currency `12`
  8. per-transaction or daily limit `61`
  9. step-up authentication `1A`
  10. available funds `51`

  If every check passes, the authorization is approved with `00` and a 6-digit
  auth code.

  Every authorization runs under an account lock and then a card lock
  (Postgres advisory locks), so cards sharing an account can never overspend
  it. The daily limit counts authorized and confirmed debits since midnight in
  `limits.timezone`.

## Architecture

Hexagonal (ports and adapters), wired with [uber-go/fx](https://github.com/uber-go/fx).

```
cmd/main.go                          fx.New(infrastructure.Module)
config/                              Consul KV + env configuration
internal/
  domain/                            Program catalog, Account, Card, PAN/Luhn, Expiry, Authorization + Decide, errors
  port/                              repositories, Locker, CustomerDirectory, PaymentAuthenticator, Vault, CardSecurity, Clock, Metrics
  application/                       AccountService, CardService, AuthorizationService
  adapter/
    primary/http/                    REST API + internal resolve endpoint
    secondary/postgres/              pgx repositories, advisory-lock transactions, embedded schema
    secondary/memory/                in-memory repositories + keyed locks
    secondary/storage/               picks postgres or memory
    secondary/customerinfo/          sard-customer-info-service client
    secondary/authservice/           sard-auth-service client
    secondary/crypto/                AES-GCM PAN vault, HMAC CVV, PBKDF2 PIN
    secondary/metrics/  clock/
  infrastructure/                    fx module graph + HTTP server lifecycle
```

## API

`/api/*` routes need `Authorization: Bearer <token>` when `ACCESS_TOKENS` is
set, and `/internal/*` routes need one of `INTERNAL_TOKENS`.

| Method | Path | Body |
|---|---|---|
| GET  | `/api/v1/programs` | |
| POST | `/api/v1/accounts` | `{"customer_id","program_code","credit_limit"?}` |
| GET  | `/api/v1/accounts/{id}` · `/api/v1/customers/{customerId}/accounts` | |
| POST | `/api/v1/accounts/{id}/block` · `/unblock` · `/cancel` | `{"reason"}` (optional) |
| PUT  | `/api/v1/accounts/{id}/credit-limit` | `{"credit_limit"}` |
| POST | `/api/v1/accounts/{id}/payments` | `{"amount","description"}` |
| GET  | `/api/v1/accounts/{id}/transactions?limit=` | |
| POST | `/api/v1/accounts/{id}/cards` | `{"type":"PLASTIC|VIRTUAL|TEMPORARY","printed_name"?}` returns `{card, pan, cvv}` once |
| GET  | `/api/v1/accounts/{id}/cards` · `/api/v1/cards/{id}` (with `spent_today`) | |
| POST | `/api/v1/cards/{id}/activate` | `{"cvv"}` |
| POST | `/api/v1/cards/{id}/block` · `/unblock` · `/cancel` | `{"reason"}` (optional) |
| POST | `/api/v1/cards/{id}/report` | `{"reason":"LOST|ROBBED|DAMAGED","note","reissue"}` |
| PUT  | `/api/v1/cards/{id}/limits` · `/controls` · `/pin` | |
| GET  | `/api/v1/cards/{id}/authorizations?limit=` | |
| POST | `/api/v1/authorizations` | `{"pan","expiry":"MM/YY","cvv","pin","amount","currency","channel","processing_code","merchant","mcc","authentication_id"}` |
| GET  | `/api/v1/authorizations/{id}` | |
| POST | `/api/v1/authorizations/{id}/confirm` | `{"amount"}` (optional, at most the authorized amount) |
| POST | `/api/v1/authorizations/{id}/cancel` | |
| POST | `/internal/v1/cards/resolve` | `{"pan","expiry"}` returns `{card_id, account_id, customer_id, user_id, masked_pan, status}` |

A declined authorization still returns `200`, with `approved: false` and the
response code.

## Configuration

Non-secret settings come from the Consul key `card_service` (seeded from
`consul.json`). Environment variables override them.

| Env | |
|---|---|
| `CARD_MASTER_KEY` | required, at least 32 characters; derives the PAN encryption, PAN hash and CVV keys |
| `ACCESS_TOKENS` / `INTERNAL_TOKENS` | comma-separated |
| `STORAGE_BACKEND`, `POSTGRES_*`, `LIMITS_TIMEZONE`, `PORT_HTTP_SERVER` | |
| `CUSTOMER_INFO_URL` / `CUSTOMER_INFO_TOKEN` | if empty, every customer is treated as eligible (development only) |
| `AUTH_SERVICE_URL` / `AUTH_SERVICE_TOKEN` | if empty, step-up authentication is not enforced |

## Run and test

```sh
CARD_MASTER_KEY=local-dev-master-key-0123456789abcdef make run
make test
CARD_TEST_POSTGRES_DSN=postgres://card:card@localhost:5432/card?sslmode=disable make test
```
