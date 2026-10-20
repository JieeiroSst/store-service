# lucky-service

Mini games that reward users with points: lucky wheel, scratch card, gift box, shake, flip card, slot machine, capsule and daily check-in.

Go 1.25, hexagonal architecture, dependency injection with [uber-go/fx](https://github.com/uber-go/fx), PostgreSQL, Redis, RabbitMQ.

## Games

The catalog follows what Vietnamese commerce and wallet apps (Shopee, Lazada, MoMo, Tiki, Grab) run for loyalty:

| `game_type` | Game | Mechanic |
| --- | --- | --- |
| `wheel` | Vòng quay may mắn | instant win |
| `scratch_card` | Cào thẻ | instant win |
| `gift_box` | Mở hộp quà, user picks a box (`choice` 1–24) | instant win |
| `shake` | Lắc lì xì / lắc xu | instant win |
| `flip_card` | Lật thẻ, user picks a card (`choice` 1–24) | instant win |
| `slot_machine` | Máy quay slot | instant win |
| `capsule` | Máy gắp / gacha capsule | instant win |
| `checkin` | Điểm danh hằng ngày | streak |

Instant-win games share one engine; they differ only in presentation, which the client drives from the campaign's free-form `settings` JSON (theme, number of boxes, reel symbols, ...). The result is always drawn on the server, `choice` is recorded but does not change the odds.

### Instant-win rules

- A **campaign** sets `start_at`/`end_at`, `timezone`, `win_rate` (percent chance to win, two decimals), `daily_spins` (free plays per user per day, default 1), `max_spins_per_day` (cap including extra plays, 0 = none) and `spin_rules` (events that earn extra plays).
- **One campaign per game at a time**: a PostgreSQL exclusion constraint (`btree_gist`) rejects activating a campaign whose window overlaps another active campaign of the same game (`409 campaign_overlap`). Different games run side by side.
- Prizes: 2–24 slots, **exactly one `none` slot** and at least one `points` slot. `weight` splits the winning chance between points prizes, `quantity` is optional stock.
- A play rolls `win_rate`, then picks a points prize by weight among prizes in stock; a loss or a sold-out race lands on the `none` slot.
- Extra plays: another service calls `POST /api/v1/lucky/credits` with `{user_id, event, reference_id}`. Every running game with a rule for that event grants its `spins`, at most `max_per_day` grants per event per day, idempotent per `reference_id`. Unused extra plays carry over within the campaign.
- Plays of one user in one campaign are serialized by a PostgreSQL advisory lock and stock is claimed with a conditional update, so quotas and stock hold under concurrency. `Idempotency-Key` makes retries safe.

### Daily check-in

A `checkin` campaign lists `checkin_rewards`, the points for streak day 1..N (N ≤ 31), e.g. `[10, 20, 30, 40, 50, 60, 200]`. One check-in per day in the campaign timezone; consecutive days grow the streak, a missed day resets it, and after day N the cycle starts again.

## Anti-fraud

What the game itself guarantees:

- Results are drawn on the server with `crypto/rand`; the client only sends which box or card it opened.
- Quotas, prize stock and the campaign budget are enforced inside one PostgreSQL transaction per player (advisory lock, conditional updates, a savepoint that rolls the stock back when the budget refuses), so parallel requests cannot play twice or overspend.
- The day is computed on the server in the campaign timezone, so changing the phone clock does nothing.
- Extra plays are granted only with a service token or by an admin, idempotently per `reference_id`.
- Rewards leave through the outbox with a unique `event_id`; point-service credits each `event_id` once.

What the guard adds on every play and check-in (`internal/application/fraud.go`):

| Check | Signal | Action |
| --- | --- | --- |
| Blocked account | `lucky_user_blocks` | refuse, `403 account_restricted`; blocking also holds the user's unpaid rewards |
| Account farming on one device | `X-Device-Id`, accounts seen per device in Redis over 30 days | refuse above `FRAUD_MAX_USERS_PER_DEVICE` (3), `403 device_limit` |
| Many accounts behind one IP | `X-Real-IP`, accounts per IP in Redis over 24h | play allowed, reward **held** above `FRAUD_MAX_USERS_PER_IP` (50); not refused because carrier NAT shares IPs |
| Missing device id | `FRAUD_REQUIRE_DEVICE=true` | refuse, `400 device_required` |
| Large reward | campaign `review_points` | reward **held** for review |
| Farming one campaign | campaign `max_points_per_user_day` | a win above the cap becomes the no-prize slot |
| Exploit draining points | campaign `budget_points` (tracked in `points_awarded`) | wins stop once the budget is spent |
| Bots | `RATE_LIMIT_PLAYS_PER_MINUTE` | `429 rate_limited` |

Held rewards stay in `lucky_reward_outbox` with `status = held` and are never relayed until an admin approves them; a rejected reward is never paid and its points return to the campaign budget. IP, device id and user agent are stored on every play and check-in for investigations. Device and IP counters live in Redis and fail open when it is down; the block list is in PostgreSQL and always enforced.

gateway-api forwards the app's `X-Device-Id` header together with the client IP and user agent. Turn on `FRAUD_REQUIRE_DEVICE` once the mobile and web clients send a stable device id.

## Redis

Optional (`REDIS_ADDR`); every use fails open, so a Redis outage never blocks a play.

- **Campaign cache**: running campaigns are read on every request; an `fx.Decorate` wraps the PostgreSQL repository with a cache (`CACHE_CAMPAIGN_TTL`, default 15s) that admin changes invalidate. A newly started window can take up to the TTL to appear.
- **Rate limit**: `RATE_LIMIT_PLAYS_PER_MINUTE` (default 30) per user, separately for plays and check-ins.
- **Winner feed**: the latest `WINNER_FEED_SIZE` wins per campaign for the "who just won" ticker, with user ids masked; it falls back to PostgreSQL when Redis is empty or down.

## Points payout (outbox)

Every winning play and every check-in writes a row to `lucky_reward_outbox` in the same transaction. A relay publishes pending rows to RabbitMQ (exchange `lucky`, routing key `lucky.reward.granted`) and marks them published. point-service consumes the queue `point-service.rewards`, credits the user's balance once per `event_id`, and acks.

```json
{"event_id":"lucky-play-10","source":"lucky-wheel","game_type":"wheel","user_id":"7","campaign_id":1,"reference_id":10,"title":"500 điểm","points":500,"occurred_at":"2026-10-20T09:19:03Z"}
```

With `RABBITMQ_ENABLED=false` the relay does not run and rewards stay queued.

## API

Authentication matches the other Go services: `Authorization: Bearer <session_token>` is validated against user-service, `X-Internal-Token` identifies a service. Admin endpoints need the `admin`/`super_admin` role or an internal token. Through gateway-api the same operations are the `lucky*` GraphQL fields (`gateway-api/graph/lucky.graphqls`), and point balances are `pointMyBalance` / `pointMyTransactions`.

| Method | Path | Who |
| --- | --- | --- |
| GET | `/api/v1/lucky/games` | anyone, running games for the lobby |
| GET | `/api/v1/lucky/games/{game}` | anyone (admins also see weights, stock, win rate) |
| GET | `/api/v1/lucky/games/{game}/winners?limit=` | anyone |
| GET | `/api/v1/lucky/games/{game}/me/status` | user |
| POST | `/api/v1/lucky/games/{game}/me/plays` `{"choice": 3}` | user, optional `Idempotency-Key` |
| GET | `/api/v1/lucky/me/plays?game_type=&campaign_id=&limit=&offset=` | user |
| GET, POST | `/api/v1/lucky/checkin/me` | user (status, check in) |
| GET | `/api/v1/lucky/checkin/me/history` | user |
| POST | `/api/v1/lucky/credits` | service / admin |
| POST, GET | `/api/v1/lucky/campaigns?game_type=&status=` | admin |
| GET, PUT, DELETE | `/api/v1/lucky/campaigns/{id}` | admin (edit needs draft/paused, delete needs draft) |
| POST | `/api/v1/lucky/campaigns/{id}/activate`, `/pause` | admin |
| GET | `/api/v1/lucky/campaigns/{id}/plays?user_id=` | admin |
| POST, GET | `/api/v1/lucky/fraud/blocks` `{"user_id","reason"}` | admin |
| DELETE | `/api/v1/lucky/fraud/blocks/{userId}` | admin |
| GET | `/api/v1/lucky/fraud/rewards?status=held&user_id=` | admin |
| POST | `/api/v1/lucky/fraud/rewards/{id}/approve`, `/reject` | admin |
| GET | `/api/v1/lucky/fraud/plays?user_id=&device_id=&ip=` | admin |
| GET | `/livez`, `/health` | probes |

Errors are `{"code": "...", "error": "..."}`: `no_active_campaign` (404), `no_spins_left` (429), `rate_limited` (429), `credit_limit_reached` (429), `already_checked_in` (409), `campaign_overlap` (409), `account_restricted` (403), `device_limit` (403), `device_required` (400).

Gift box campaign:

```json
{
  "game_type": "gift_box",
  "name": "Mở hộp quà tháng 10",
  "start_at": "2026-10-20T00:00:00+07:00",
  "end_at": "2026-11-01T00:00:00+07:00",
  "win_rate": 35,
  "daily_spins": 1,
  "max_spins_per_day": 5,
  "max_points_per_user_day": 2000,
  "budget_points": 500000,
  "review_points": 1000,
  "settings": {"boxes": 6, "theme": "autumn"},
  "spin_rules": [
    {"event": "order_completed", "spins": 1, "max_per_day": 3},
    {"event": "referral_signup", "spins": 2}
  ],
  "prizes": [
    {"name": "50 điểm", "type": "points", "points": 50, "weight": 70},
    {"name": "200 điểm", "type": "points", "points": 200, "weight": 25},
    {"name": "1000 điểm", "type": "points", "points": 1000, "weight": 5, "quantity": 100},
    {"name": "Chúc bạn may mắn lần sau", "type": "none"}
  ]
}
```

Check-in campaign:

```json
{
  "game_type": "checkin",
  "name": "Điểm danh nhận điểm",
  "start_at": "2026-10-20T00:00:00+07:00",
  "end_at": "2026-12-31T00:00:00+07:00",
  "checkin_rewards": [10, 20, 30, 40, 50, 60, 200]
}
```

`game_type` cannot change after creation. To edit prizes of a paused campaign send their `id`; prizes left out are removed, which is refused once someone has won them.

## Layout

```
cmd/main.go                          fx.New(infrastructure.Options()...).Run()
internal/domain/model                games, campaign, draw, quota, streak rules (pure Go)
internal/domain/port                 driving (use cases) and driven (repositories, outbox, publisher, cache, clock, random) ports
internal/application                 CampaignService, GameService, CheckinService, Guard, FraudService, RewardRelay
internal/adapter/primary/http        REST handlers
internal/adapter/secondary/postgres  repositories, schema.sql (applied on start under an advisory lock)
internal/adapter/secondary/redis     campaign cache decorator, rate limiter, winner feed, fraud signals
internal/adapter/secondary/rabbitmq  reward publisher
internal/adapter/secondary/system    clock, crypto/rand
internal/infrastructure              fx modules, database, redis client, HTTP server lifecycle
```

## Configuration

See `.env.example`.

## Development

```
make build vet test
LUCKY_TEST_DB_HOST=localhost LUCKY_TEST_DB_PASSWORD=... make test   # also runs the PostgreSQL tests
make docker
```

The PostgreSQL role needs to be allowed to `CREATE EXTENSION btree_gist` (a trusted extension; the database owner can, which is how chart/postgres provisions it).
