# polymarket-service

A prediction-market exchange modelled on [polymarket.com](https://polymarket.com), built as a
hexagonal service (`ports & adapters`) wired with [uber-go/fx](https://github.com/uber-go/fx).
It calls other store-service services over REST:

| Service | Used for |
| --- | --- |
| `payment-wallet-service` | deposits into and withdrawals out of the exchange (`/wallets/user/:id`, `/transfers`, `/transfers/:id/reverse`) |
| `notification-service` | fills, deposits, resolutions, rewards, referrals (`/notifications`) |
| `user-service` | authenticates bearer tokens (`POST /api/v1/validate`) |
| `referral-service` | issues share codes (`/referral/generate`) and records who referred whom (`/referral/activate`) |

## How trading works

* **Events and markets.** An event groups binary yes/no markets (multi-outcome events are several markets).
* **Order book (CLOB).** Limit orders (`gtc`, `gtd`) and market orders (`fok`, `fak`) match with
  price-time priority, executing at the resting order's price. YES and NO share one YES-denominated
  book, so a *buy YES* can match a *buy NO* (**mint**: a new YES/NO pair is created) and a *sell YES* can
  match a *sell NO* (**merge**: a pair is redeemed), as on Polymarket.
* **Money.** Users deposit cash into a treasury wallet on `payment-wallet-service`; the exchange keeps
  their trading balance. Buy orders lock cash, sell orders lock shares. Every unit of cash is either in
  a balance or backs an outstanding YES/NO pair - the tests assert this identity after every scenario.
* **Prices** are integers in `[1, share_value-1]` (cents with the default `share_value` of 100). A
  winning share pays `share_value`.
* **Resolution.** An admin *proposes* an outcome, opening a dispute window. Anyone can *dispute*; an
  undisputed proposal can be *finalized* by anyone once the window closes; an admin can *resolve*
  directly (`yes`, `no` or `split` for 50-50). Settlement cancels resting orders and credits winners
  automatically.

## Fees, referrals and liquidity rewards

* **Taker fee.** `TAKER_FEE_BPS` of the cash the *taker* spends or receives on each fill (makers pay
  nothing). A buy reserves the worst-case fee up front and releases it once the order rests; a market
  buy's budget includes its fee; a sell's fee comes out of the proceeds.
* **Where the fee goes.** `MAKER_REBATE_BPS` of the fee goes to the maker, `REFERRAL_BPS` to the taker's
  referrer (if any), and the rest is exchange revenue. Every fill records the split on the trade.
* **Referrals.** `GET /users/:id/referral` asks referral-service for the user's share code (once) and
  reports referees and commissions earned. `POST /users/:id/referral/redeem` attributes a **new** user
  (no trades yet, never referred) to a code's owner through referral-service.
* **Liquidity rewards.** A market with a `reward_pool` (per day), `reward_max_spread` and
  `reward_min_size` pays makers every `REWARD_EPOCH`. Resting orders score quadratically the closer they
  are to the midpoint; quoting both sides scores the sum, one side a third, and one-sided quoting near
  the extremes (mid < 10% or > 90%) scores nothing. Rewards are paid **out of exchange revenue** (fees,
  forfeited bonds, and top-ups via `POST /admin/exchange/fund`) and never exceed it. Each epoch is claimed
  in the database, so any number of replicas can run the job and exactly one pays.

## Neg-risk events

Mark an event `neg_risk` when exactly one of its markets can win. Then:

* `POST /events/:id/convert` swaps NO shares in the markets you name for YES shares in every other market,
  plus `(k-1)` share values in cash per share for `k` markets named. Both sides are worth the same in
  every outcome, so no collateral moves and the event stays solvent (tests check this for every winner).
* Such events resolve as a whole: `POST /events/:id/propose|resolve` with a `winner_market_id`. Market-level
  propose/resolve is refused; dispute and finalize on any one market act on the whole event.

## Dispute bond

With `DISPUTE_BOND` set, disputing a proposal takes the bond from the disputer's balance. If the final
ruling differs from the proposal the bond is refunded; if the proposal stands it moves to exchange
revenue and shows as a loss in the disputer's profit.

## Authentication

`AUTH_MODE=token` (with `USER_SERVICE_BASE_URL`): every request that acts as a user must carry
`Authorization: Bearer <session token>`; the service validates it with user-service (successes cached
for 30s, failures never) and ignores `X-User-Id`. A token that is present but invalid is refused even on
public routes. `AUTH_MODE=gateway` (default): a gateway authenticates and sets `X-User-Id`. In both modes
the identity must match the user in the path or body.

## Realtime across replicas

`REDIS_ADDR` switches the SSE hub to Redis pub/sub: an event committed on one replica reaches clients
connected to any other. Without it the in-process hub is used (single replica only). A Redis outage never
blocks trading; missed live updates are repaired by the next book snapshot.

## API (`/api/v1`)

```
GET  /events?status=&category=&tag=&q=&featured=true&sort=volume|newest|ending
POST /events                       (admin)  create event with its markets
GET  /events/:id|slug              /related   /comments   POST /comments   POST|DELETE /watch
GET  /categories                   GET /leaderboard?metric=profit|volume&window=day|week|month|all
GET  /markets?q=                   GET /markets/:id|slug
GET  /markets/:id/book?outcome=    /quote?outcome=&side=&amount=   /trades   /holders
GET  /markets/:id/prices-history?outcome=&range=1h|6h|1d|1w|1m|max
GET  /markets/:id/stream           server-sent events: book, trade, status, resolved
POST /markets/:id/orders           place an order        DELETE /markets/:id/orders  cancel all (mine)
GET  /orders/:id                   DELETE /orders/:id    cancel
POST /markets/:id/propose|resolve  (admin)   POST /markets/:id/dispute|finalize
GET  /users/:id/balance|wallet|portfolio|activity|orders|watchlist|positions/closed|rewards|referral
POST /users/:id/referral/redeem
POST /events/:id/convert           swap NO shares for YES elsewhere (neg-risk)
POST /events/:id/propose|resolve   (admin) neg-risk resolution by winner_market_id
GET|POST /markets/:id/rewards      (POST is admin) liquidity reward settings
GET  /admin/exchange   POST /admin/exchange/fund    (admin) revenue, bonds held, fund rewards
POST /users/:id/deposit|withdraw   GET|PUT /users/:id/profile
DELETE /comments/:id   POST|DELETE /comments/:id/like
```

### Pagination

`GET /events`, `GET /events/:id/comments` and `GET /users/:id/orders` use cursor (keyset) pagination
and respond with:

```json
{"items": [...], "next_cursor": "eyJzIjoi...", "is_last_page": false}
```

Send `?limit=` and pass the previous `next_cursor` as `?cursor=` for the next page. Stop when
`is_last_page` is `true` (`next_cursor` is then empty). Cursors are opaque, bound to the `sort` they
were issued for, and a malformed or mismatched one returns `400`. Unlike offsets, pages stay stable when
new rows arrive mid-listing. There is no `total` - counting every page would defeat the point.

Placing an order:

```jsonc
{"user_id":"u1","outcome":"yes","side":"buy","price":62,"size":100}                       // limit
{"user_id":"u1","outcome":"yes","side":"buy","type":"market","amount":5000}               // spend a budget
{"user_id":"u1","outcome":"no","side":"sell","type":"market","size":40,"time_in_force":"fok"}
```

## Security model

There is no login here. The API gateway authenticates users and sets `X-User-Id`: when present it must
match the user in the path or body. Internal callers may omit it. Admin endpoints need `X-Admin-Token`
(`ADMIN_TOKEN`); with no token configured they are disabled.

## Configuration

Environment variables (or Consul, see `consul.json`): `PORT_HTTP_SERVER`, `MYSQL_*`, `WALLET_BASE_URL`,
`TREASURY_WALLET_ID` (**required** - create and fund the wallet first), `NOTIFICATION_BASE_URL`,
`EXCHANGE_CURRENCY`, `SHARE_VALUE`, `MIN_ORDER_SIZE`, `DISPUTE_WINDOW`, `DISPUTE_BOND`, `REWARD_EPOCH`,
`TAKER_FEE_BPS`, `MAKER_REBATE_BPS`, `REFERRAL_BPS`, `USER_SERVICE_BASE_URL`, `AUTH_MODE`,
`REFERRAL_BASE_URL`, `REDIS_ADDR`/`REDIS_PASSWORD`/`REDIS_DB`, `ADMIN_TOKEN`.

The schema (`database.sql`) only creates missing tables. Databases created by an earlier build of this
service need a fresh database or the new columns/tables added by hand.

## Development

```
go test ./...                       # unit tests, no infrastructure needed
MYSQL_TEST_DSN='root:pw@tcp(127.0.0.1:3306)/pm?charset=utf8mb4&parseTime=True&loc=Local' \
  go test ./internal/integration    # real MySQL: lifecycle, catalog/social, concurrent trading
```

## Not implemented

Referral rewards paid by referral-service itself (commissions here come from trading fees), a bond for
proposers (proposals come from admins), per-market fee schedules, and `total` counts on the paginated
listings. Profit is realised profit only (sells, settlements, fees, rebates, rewards); unrealised PnL is
in the portfolio.
