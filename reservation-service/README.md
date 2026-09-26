# reservation-service

Reservations for **5-star hotels**: hotels, their room types and rooms, a nightly inventory and rate per
room type, extra services (airport transfer, spa, breakfast, ...), and reservations that are paid by
e-wallet or by a payment provider. It follows the high-level design in `diagram.png`, `table.png` and
`status.png` and owns its own database (`Reservation DB`).

Accounts, login, sessions and roles belong to **user-service**: callers send its access token as a
`Authorization: Bearer` header, and `guest_id` / `manager_id` are user ids there. Money moves through
**payment-wallet-service** (wallets) and **payment_service** (paypal, stripe, ...). **recompense-service**
keeps points and tiers.

## Who can do what

| Role | How you become one | Can |
| --- | --- | --- |
| Guest | any signed-in user | search hotels, reserve, pay, cancel/refund their own reservations, review after a stay |
| Hotel manager | register a hotel (`POST /hotels`) | manage *their own* hotels only: details, room types, rooms, inventory and rates, services; see and cancel/reject reservations at their hotels |
| Platform admin | role `admin` in user-service (`AdminRole`) | verify hotels (approve / reject), and manage anything |

Only **5-star** hotels are accepted (`stars` must be 5). A hotel registered by a manager is `pending` and
invisible to the public until an admin approves it; a rejected one carries the admin's reason and goes back
to the queue when its manager fixes it. A manager cannot reserve a room at their own hotel.

## Reservation lifecycle

```
pending ──pay──────▶ paid ──cancel──▶ refunded   (full refund, or a refund minus the hotel's cancellation fee)
   ├─────cancel────▶ canceled   (guest, or the hold ran out)
   └─────reject────▶ rejected   (the hotel)
```

- Creating a reservation **holds the rooms** (`room_type_inventory.total_reserved`) and prices the stay
  from the nightly rates, atomically: two guests racing for the last room cannot both get it.
- An unpaid reservation is released after `ReservationHoldMinutes` (default 15) by a background sweeper.
- `Idempotency-Key` (or `request_id`) makes a reservation call safe to retry; paying is idempotent too.
- Payment goes to the wallet named by the **hotel** (`wallet_id`, which must belong to its manager).
- Every step is written to the reservation's **history** (`GET /reservations/{id}/history`).

### Cancellation policy and refunds

Each hotel has a policy of tiers, e.g. `[{"hours_before":72,"fee_percent":0},{"hours_before":24,"fee_percent":50}]`:
free 72h or more before check-in, half the price between 72h and 24h, non-refundable after that (it defaults to
"free until `free_cancel_hours` before check-in"). `GET /reservations/{id}/cancellation-quote` says what cancelling
costs right now. A guest can cancel while the fee is under 100%; the hotel's manager and admins cancel without a
fee at any time. A wallet payment is reversed and the fee charged again to the guest for the hotel; a gateway
payment is refunded partially. A fee that cannot be collected (the guest spent the refund) is waived, not chased.

### Front desk

After payment the manager records `POST /reservations/{id}/check-in`, `/check-out` and `/no-show` (a no-show is
only possible after the first night has started, is not refunded, and cannot review). Guests are thanked and
reminded automatically (reminders go out for stays starting today or tomorrow).

## When a million users want the last room

Correctness never depends on caching: taking rooms is one atomic database update (rows locked in date order, the
availability test re-checked after each lock is won, and a `CHECK (total_reserved <= total_inventory)` behind it),
so however many guests race, at most as many succeed as there are rooms. Everything else only decides how cheaply
the rest are turned away. In order, for `POST /reservations`:

1. Concurrent requests in flight per pod are capped (`MaxInFlight`, 503 + `Retry-After`).
2. The body is read and **before the caller is identified**, the request is checked against "sold out": from memory,
   or from one plain inventory read shared by every request asking the same question (`SoldOutTTLSeconds`). No
   user-service call, no transaction. Freed rooms (cancellation, expired hold, new inventory) clear it at once.
3. One account is rate limited; a request id already on a reservation returns that reservation (a winner's retry
   is never told "sold out"); a guest may hold only `MaxPendingPerGuest` unpaid reservations at a time.
4. Only `ReserveConcurrency` requests per pod reach the database together; the rest wait `ReserveQueueWaitMillis`
   and get `429` + `Retry-After`. The pool is bounded (`PostgresMaxConns`) and lock waits are capped at 800 ms.

Measured on one laptop with a single pod (server, Postgres in Docker, user-service stand-in and load generator all on
it), one room left and every request from a different user, 3000 in flight: 1,000,000 requests -> exactly 1 reservation,
997,356 x 409 sold out and 2,643 x 429 "try again", 27,500 requests/s, 26 database connections, 42 MB of memory,
and user-service called about 3,100 times instead of a million. 10 rooms and 300,000 users -> exactly 10 winners.

### One room, a million users, several pods

**The guarantee:** however many users ask at once, on however many pods, at most as many reservations succeed as
there are rooms; with one room, exactly one user gets it. Nothing here depends on a pod's memory: the rooms are taken
by the database statement described above, so a pod that is restarted, stale or dead cannot change the outcome.

What each pod adds is only speed at turning the rest away. Every pod is independent: its own connection pool,
sold-out cache, bulkhead and background loops. They share nothing but the database.

**What has been verified** (`internal/integration`, several pods in one process against a real Postgres 16;
1,000,000 users, 5,000 in flight, every request from a different user, released at the same instant):

| Scenario | Result |
| --- | --- |
| 1 room, 5 pods | exactly **1** reservation; 998,541 told "sold out", 1,458 told "busy, retry"; no other error; 0.9 s |
| 7 rooms, 4 pods, 2-night stays | exactly 7 |
| overlapping stays over 6 nights, 5 pods | no deadlock, no night oversold, counters equal the reservations covering each night |
| 1 room, a pod is killed mid-storm | never more than 1; counters match what is stored; only requests on the dead pod fail |
| the winner's answer is lost with the pod | retrying with the same `Idempotency-Key` on another pod returns that reservation: not "sold out", not a second room |
| a room comes back | the pod that freed it forgets "sold out" at once, the others within `SoldOutTTLSeconds` |
| one guest sends 40 requests for different nights over 5 pods | exactly `MaxPendingPerGuest` (3) held, 37 refused |
| last use of a promo code, 250 guests over 5 pods | exactly 1 gets it; `used_count` is 1 |
| every pod runs the sweeper, the waiting list and notification forwarding | each expired hold is released once, one waiting-list offer per room (and it is not cancelled by the second pod), each notification pushed once |

The tests were also run with each safeguard removed to make sure they fail without it (no per-guest lock, no lease on
notification claims, a second pod cancelling the waiting-list offer, no availability test in the locking statement). The one exception: removing the
explicit `order by date` from the locking statement does not fail the deadlock test, because Postgres already scans the
primary key in date order. That clause is insurance against a different query plan.

Run them with a Postgres server (they are skipped without one):

```
TEST_DATABASE_URL='postgres://postgres:pw@localhost:5432/postgres?sslmode=disable' go test ./internal/integration/ -v
MULTIPOD_USERS=1000000 MULTIPOD_WORKERS=5000 TEST_DATABASE_URL=... go test ./internal/integration/ -run OneRoom -v
```

CI runs them against a Postgres service container. `loadtest/README.md` has the same storm as a Kubernetes Job.

**What has not been verified**

- The million-user storm on a real cluster after the multi-pod fixes. A 500,000-user run on 5 pods (before those
  fixes) also produced exactly one reservation, but with 500s and 503s from the problems since fixed (below), and it has
  not been repeated. The manifests in `loadtest/k8s/` are for you to run on a cluster you can spare.
- Throughput and latency on real infrastructure. The numbers in this document come from one machine and mean
  nothing else: a pod limited to 1 CPU serves a fraction of what it does unlimited, and the network adds more.
- Anything in front of the service (load balancer, WAF, ingress limits) that may throttle before requests arrive.

**What a client must do:** send an `Idempotency-Key` (or `request_id`). If a pod dies after committing a reservation
and before answering, the user only sees an error; retrying with the same key returns their reservation. Without a
key they cannot tell whether they hold the room.

**Sizing for several pods**

- Every pod opens its own pool: replicas x `PostgresMaxConns` must stay under Postgres `max_connections` (100 by
  default), together with everything else on that server. The default is 10 per pod, so 5 replicas use half. A pod
  that cannot get a connection answers `429` instead of `500`, but it cannot serve until it does.
- A flooded pod answers health probes slowly. The chart's liveness probe is patient (5 s timeout, 6 failures) so a
  busy pod is taken out of rotation for a moment instead of being restarted, which would push its load onto the others.
- Under a rush the service does not write a log line per request (it would starve the pod): errors, slow requests and
  changes that succeeded are always logged, refusals and plain reads are sampled 1 in 100, health checks never.
- "Sold out" is remembered per pod (`SoldOutTTLSeconds`, default 2 s). A room that comes back on one pod can still be
  refused for up to that long on the others.

Guests who miss out can join the **waiting list** (`POST /waitlist`): when rooms come back the oldest guest whose stay
fits is given a pending reservation (a real hold, made through the same atomic path) and notified; if they do not pay
in time, it goes to the next one.

## Notifications

Everything that matters raises a notification stored in the person's inbox (`GET /notifications`, unread count,
mark read): the manager hears about new and paid reservations and guest cancellations, the guest about
confirmation, rejection, hotel-side cancellation, refunds and fees, expiry, waiting-list offers and check-in
reminders. One event notifies a person once. With `NotificationServiceURL` set they are also forwarded to
notification-service (retried up to 5 times, so an outage delays them without losing them).

## API (`/api/v1`)

| Method | Path | Notes |
| --- | --- | --- |
| GET | `/hotels` | search: `city`, `q`, `amenities=spa,pool`, `min_rate`/`max_rate` (per night), `check_in`+`check_out`+`rooms`+`guests` (only hotels with a room type that fits), `sort` = recommended (default) / rating / newest, `limit`, `cursor` |
| GET | `/hotels/{id}` | unpublished hotels are visible to their manager and admins only |
| GET | `/hotels/{id}/room-types` · `/quotes?check_in&check_out&rooms` · `/services` · `/reviews` | public |
| GET | `/hotels/{id}/room-types/{type}/inventory?from&to` | rooms left and rate per night |
| POST | `/hotels` | register a 5-star hotel (becomes its manager) |
| PUT · DELETE | `/hotels/{id}` | update / pause (manager or admin) |
| GET | `/my/hotels` | the caller's hotels in every status |
| GET | `/admin/hotels?status=pending` · POST `/hotels/{id}/approve` · `/reject` | admin verification |
| POST · PUT | `/hotels/{id}/room-types[/{type}]`, `/rooms[/{room}]`, `/services[/{service}]` | manager |
| PUT | `/hotels/{id}/room-types/{type}/inventory` | `{from, to, total_inventory?, rate? \| rate_off}`; omitted total = number of available rooms |
| POST | `/reservations` | `{hotel_id, room_type_id, check_in, check_out, rooms, adults, children, special_requests, services:[{service_id, quantity}]}` |
| GET | `/reservations` (`as=manager`, `status`, `hotel_id`, `cursor`) · `/reservations/{id}` · `/{id}/history` · `/{id}/cancellation-quote` | polling `/{id}` also settles a captured gateway payment |
| POST | `/reservations/{id}/pay` | `{method: wallet\|gateway, provider}` |
| POST | `/reservations/{id}/cancel` · `/reject` · `/check-in` · `/check-out` · `/no-show` | cancel/refund · reject a pending one · front desk (manager) |
| GET · POST · PUT | `/hotels/{id}/promotions[/{promo}]` · `/hotels/{id}/report?from&to` | manager: promo codes (percent or fixed, min nights, max uses, dates; a reservation takes `promo_code`) · occupancy and revenue (up to 92 days) |
| POST · GET · DELETE | `/waitlist` | waiting list for sold-out rooms |
| GET · PUT · DELETE | `/wishlist` | favourite hotels |
| GET · POST | `/notifications`, `/notifications/unread-count`, `/{id}/read`, `/read-all` | the caller's inbox |
| POST | `/hotels/{id}/reviews` | `{reservation_id, rating, comment}`; only after a paid stay that has ended |
| GET · POST | `/wallet`, `/wallet/transactions`, `/loyalty` | the caller's own wallet and points |

Every list is paged with a **cursor**: repeat the request with `cursor=<next_cursor>`; the response is
`{items, has_more, next_cursor}` (a page never repeats or skips an item when others are added in between). `recommended` ranks hotels by rating (a few reviews count for less) plus a
boost for managers with a higher recompense tier, over the best 500 matches.

## Configuration (environment)

`PORT`, `HostPostgres`, `PortPostgres`, `DatabasePostgres` (`reservation_service`), `UserPostgres`,
`PasswordPostgres`, `UserServiceURL` (required), `UserServiceValidatePath`, `UserServiceUserPath`, `AdminRole`,
`AuthCacheTTLSeconds`, `WalletServiceURL` and/or `PaymentServiceURL` (at least one), `RecompenseServiceURL`
(optional), `NotificationServiceURL` (optional), `ReservationHoldMinutes`, `SweepIntervalSeconds`, and for a rush
on the last rooms `PostgresMaxConns`, `ReserveConcurrency`, `ReserveQueueWaitMillis`, `SoldOutTTLSeconds`,
`UserRatePerSecond`, `UserBurst`, `MaxPendingPerGuest`, `MaxInFlight`. The schema (`schema.sql`) is applied on startup.

## Layout

`internal/domain` (entities, no dependencies) · `internal/application/port/{inbound,outbound}` (interfaces) ·
`internal/application/service` (use cases) · `internal/adapter/inbound/http` · `internal/adapter/outbound/*`
(Postgres and the other services) · `internal/bootstrap` (the uber-go/fx wiring).
