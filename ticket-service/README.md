# ticket-service

Ticketing for live events, modelled on [ticketbox.vn](https://ticketbox.vn): organizers publish **events**
(concerts, shows, sports, workshops) with **ticket types** (VIP, GA, early bird, ...) — general admission or with a
**seat map** — and buyers hold, pay for and receive **e-tickets** (a QR code each) that are scanned once at the gate.
It owns its own database.

Accounts, login, sessions and roles belong to **user-service**: callers send its access token as an
`Authorization: Bearer` header, and `user_id` / `organizer_id` are user ids there. Money moves through
**payment-wallet-service** (wallets) and **payment_service** (paypal, stripe, ...). **notification-service** is
optional and only forwards notifications to devices.

## Who can do what

| Role | How you become one | Can |
| --- | --- | --- |
| Buyer | any signed-in user | search events, hold tickets, pay, cancel / refund their own orders (if the event allows it), see their tickets, wishlist, notifications |
| Organizer | create an event (`POST /events`) | manage *their own* events: details, ticket types, seats, promo codes; see orders, attendees and the sales report; **scan tickets at the gate**; cancel the event |
| Gate staff | named by an organizer (`POST /events/{id}/staff`) | scan and look up tickets of *that* event; nothing else |
| Platform admin | role `admin` in user-service (`AdminRole`) | approve / reject events, feature them, and manage anything |

## Event lifecycle

```
draft ──submit──▶ pending_review ──approve──▶ published ──cancel──▶ cancelled
  ▲                    │                          │
  └──── (fix, submit) ─┴─reject (with a reason)   └─ ends: sales stop by themselves
```

An event needs at least one ticket type with tickets (for a seated type: generate its seats) before it can be
submitted. Only `published` events are public; anyone else's draft answers `404`. Cancelling a published event needs a
reason; every pending and paid order of it is refunded in the background and the buyers are notified.

## Order lifecycle

```
pending ──pay──────▶ paid ──cancel──▶ refunded   (allowed for the buyer until refund_cutoff_hours before the event;
   │                                               the organizer / admin may always, unless a ticket was already used)
   ├─cancel─────────▶ cancelled
   └─hold ran out───▶ expired
```

`POST /orders` **holds** the tickets (default 10 minutes, `OrderHoldMinutes`) and returns the order with its total.
Paying (`POST /orders/{id}/pay`, `{"method": "wallet"}` or `{"method": "gateway", "provider": "stripe"}`) turns it
into `paid` and issues one ticket per admission. A hold that runs out puts the tickets back on sale.

## Ticket lifecycle

A ticket is one admission. It is issued when its order is paid, and from then on it has a state, a holder and a
history (`GET /tickets/{id}/history`: every change, who made it, when).

```
                 ┌────────── check in ──────────▶ used ──── revert (organizer) ──┐
                 │                                                                │
issued ──▶ valid ┼── offer ──▶ transferring ── accept ─▶ valid (new holder, new code)
                 │                  │  └── decline / cancel / expire ─▶ valid (unchanged)
                 ├── list for resale ──▶ listed ── sold ─▶ valid (new holder, new code)
                 │                          └── cancelled / event starts ─▶ valid
                 ├── event ends unused ──▶ expired
                 └── refund, cancel event, or revoke ──▶ void
```

| | Who | How |
| --- | --- | --- |
| Give the ticket an attendee name / e-mail | holder | `PUT /tickets/{id}/holder` |
| Replace the QR code (it leaked); the old code dies at once | holder | `POST /tickets/{id}/reissue` |
| Pass it on: offer by user id or e-mail, the recipient accepts or declines, the sender can cancel; unanswered offers lapse (`TransferTTLHours`, or when the event starts) | holder, recipient | `POST /tickets/{id}/transfer`, `GET /transfers?direction=incoming\|outgoing`, `POST /transfers/{id}/accept\|decline\|cancel` |
| Sell it to another buyer at up to face value (see *Resale*) | holder, buyer | `POST /tickets/{id}/resale`, `POST /resale/{id}/buy` |
| Revoke one ticket (fraud, a withdrawn invitation): it stops admitting, the order is untouched | organizer, admin | `POST /tickets/{id}/void` |
| Undo a mistaken scan | organizer, admin | `POST /tickets/{id}/revert-check-in` |
| Look at a ticket | holder, organizer, gate staff, admin | `GET /tickets/{id}`, `GET /tickets`, `GET /events/{id}/tickets/{code}` |

What keeps it honest: accepting a transfer gives the ticket a **new code**, so the sender's QR (a screenshot) is
dead and the recipient never sees the code before accepting; a ticket on offer cannot be scanned; every hand-over
counts against `MaxTicketTransfers` (a brake on scalping); an organizer can forbid transfers for an event
(`"transferable": false`); refunding an order voids its tickets and calls off their offers; a ticket is scanned
exactly once even if many gates read it at the same instant. The sweeper closes unanswered offers and expires the tickets
of ended events, a few hours after the end so a gate that was offline can still upload its scans.

## Resale

A holder who cannot go sells the ticket to someone else instead of losing the money. The organizer turns it on per event
with `resale_cap_percent` (0 = off, 100 = up to face value, 80 = at most 80% of what was paid); the price is checked against
what the *seller paid* for that ticket. The buyer pays the **seller's wallet** (payment-wallet-service); the platform keeps
`ResaleFeePercent` of the price, paid to `PlatformWalletID` (no fee without a wallet).

```
POST /tickets/{id}/resale {"price"}   ticket -> listed (cannot be scanned or transferred)
GET  /events/{id}/resale              open listings: type, seat, price, face price (never the code or the seller)
POST /resale/{id}/buy                 pays the seller, returns the ticket with a new code only the buyer has
DELETE /resale/{id}                   the seller takes it down (until a buyer is paying)
```

Money and ticket move in an order that a crash cannot break: (1) one buyer *claims* the listing (a single atomic update,
so of many buyers exactly one goes on), (2) the wallet transfers are made and recorded on the listing, (3) the ticket is
handed over. Any failure undoes the earlier steps (a buyer without funds, a fee that cannot be collected, a ticket revoked
while they paid). If a process dies half-way, the listing still says who paid and how, and the sweeper finishes the
hand-over, or refunds the buyer if the ticket can no longer be handed over. Listings come down when the event starts,
and when an order is refunded. A ticket a buyer has put on sale, or has already passed on, cannot be refunded by the
original buyer of the order (the organizer and a cancelled event still refund it).

## Venues and seat maps

Halls differ: a fan-shaped opera house, a lecture hall with a centre aisle, an arena with the stage in the middle, a black
box with two stages. A **venue** describes one once, as a blueprint of stages and sections; events made in it reuse it.

```
POST /venues/preview?seats=true   what a blueprint (or template) would make: seats per section, the drawing, every seat's place
POST /venues                      save one: {"name", "template": "arena", "params": {...}}  or  {"name", "map": {...}}
GET/PUT/DELETE /venues/{id}       read (?seats=true), edit, delete;  POST /venues/{id}/copy;  GET /venues?scope=mine|shared
POST /events/{id}/seat-map/apply  {"venue_id", "assignments": [{"section", "ticket_type_id"}], "replace"}
GET  /events/{id}/seat-map        the whole seating of a published event on one map
GET  /venue-templates             the templates and their parameters
```

**Templates** are parametric starting points: `theatre` (rows on arcs around the stage, each row longer than the one in
front, optional balcony), `hall` (straight rows, optional centre aisle, stage at one end) and `arena` (four blocks facing a stage
in the middle). **Blueprints** describe anything else with:

- **stages**: any number, `rect`, `ellipse` or `polygon`;
- **sections**, each with a `key` (`ORCH`, `A`, ...), name, colour and a layout:
  - `grid`: straight rows with `rows` × `seats_per_row` or a count per row (`row_seats: [8, 10, 10, 12]`), `aisle_after`,
    alignment (rows of different lengths centre on the widest), `rotation` to turn a block towards the stage;
  - `arc`: rows on concentric arcs round a focal point (`center`, `radius`, `angle_start`/`angle_end`); with no
    seat count a row holds as many as fit, so the wider rows at the back hold more;
  - for both: `skip` (a pillar: `"B:7"`), `accessible` (wheelchair places), `number_from: "right"`, `row_start`, an optional
    `outline` (otherwise a box round the seats is drawn).

The blueprint is validated as it is generated: a venue has at most 50,000 seats, and every seat, stage and outline must
lie inside the canvas, with a reason naming the section that does not.

**Using a venue for an event**: create the event's seated ticket types (VIP, Stalls, Balcony, ...), then *assign sections to
ticket types*: `S` and `N` sold as VIP, `E` and `W` as Standing. Each type gets the seats of its sections (its stock is their count);
a section left out is not sold (a closed balcony). The event keeps a snapshot of the drawing, so editing or deleting the venue later never
moves seats of an event on sale. A type that already has seats is only given another map with `replace`, and only while
nothing of it is sold or held.

**Drawing for clients**: `GET /events/{id}/seat-map` returns `layout` (canvas, stages, one outline per section with the
`ticket_type_id` that sells it, so a client can colour the map by tier), `ticket_types` (name, price, availability) and every
`seats` entry (`section`, `row`, `number`, `x`, `y`, `status`, `accessible`, `ticket_type_id`). Buyers pick seats from it and send their
ids in `POST /orders` (seats of different sections of one type may share an order). The map of a single ticket type
(`GET .../ticket-types/{type}/seats`) and the older hand-made way (`POST .../seats` on a grid, `PUT .../seat-map` to place seats)
still work; a seat is identified by section, row and number, so sections may each start at row A.

**The editor**: `GET /editor/seat-map` is an interactive editor (one embedded page, no build step). Paste your token, start from
a template or a blank canvas, add and drag sections and stages, edit rows, seat pitch, aisles, arcs, rotation and colour in the
form, click a seat to remove it (a pillar) or mark it accessible, and watch the seats as the server computes them (the page draws
`POST /venues/preview`, it never lays seats out itself). *Save* stores the venue; *Use for an event* loads an event's seated ticket
types (each shows the showtime it sells) and applies the venue to them. The page holds no secret and is served with a strict
Content-Security-Policy.

Venues are private to their owner; an admin can `share` one into a library that every organizer can read and copy (not edit).

## Documents: PDFs kept for the buyer

Once an order is paid its **invoice** and its **e-tickets** are made as PDFs and kept in **upload-service** (MinIO behind it),
under the buyer's id: receiver `ticket-user:<user id>`.

- The invoice is Unicode (Vietnamese prints correctly: an embedded DejaVu font) with the seller, buyer, event, lines, discount,
  total and VAT; a refunded order's says so. The e-tickets are one page per ticket with the event, seat, attendee and a **QR
  code of the ticket's code**, which a gate scans (the tests decode the QR and check it is the code).
- `GET /orders/{id}/documents` lists what is kept; `GET /orders/{id}/documents/invoice|tickets` downloads it. Downloads go
  **through this service**, which checks whose the order is (the buyer; the organizer and admins too), because upload-service
  only hands files to services.
- Drawn fresh, not stored: `GET /orders/{id}/invoice?format=pdf` (the invoice as it is now) and `GET /tickets/{id}/pdf`
  (one *valid* ticket, for its holder: after a transfer or a resale it carries the new code, and the previous holder can no longer get it).
- The stored copies are the payment-time ones; the QR of a ticket that later changed hands is dead by design.

Who makes them: the order's workflow (below) right after payment, and the sweeper for anything it missed or that was paid before
storage was switched on (it backfills, 20 orders per sweep). Making them is idempotent and safe on several replicas at once: a
replica that loses the race removes its copy. With no `UploadServiceURL` nothing is stored and the endpoints that draw on request
still work.

**upload-service** was changed for this, without breaking its other users (hospital-patient-management-service uses receivers
like `patient:<id>` with the caller's own token): it accepts a *service key* (`X-Service-Key`, `SERVICE_API_KEY`) in place of a
user's token, and `SERVICE_ONLY_RECEIVER_PREFIXES` (here `ticket-user:`) makes those receivers reachable only with that key: a
user's token can neither list, read, replace nor delete them (a file by id answers 404, so its existence is not revealed).
Without the setting every receiver stays open to any valid token, as before, which would let any signed-in user list another
buyer's PDFs by guessing an id: set it.

## The life of an order in Temporal

Each order is followed by a **Temporal workflow** (`order-<id>`, one per order, on the task queue `TemporalTaskQueue`), started
when tickets are held:

```
held ── the hold ends ──▶ expired: the tickets go back on sale
  │  └─ cancelled / refunded ──▶ over
  └─ paid ──▶ documents made and stored ──▶ a day before the event: the reminder ──▶ over
```

The database stays the truth about the order; the workflow decides *when to look*: on the dot when the hold ends (also while a
provider payment is still in flight, it looks again every 30 seconds), when it is **signalled** that the order was paid,
cancelled or refunded, and a day before the event. Every step is an idempotent activity (`AdvanceOrder`, `GenerateOrderDocuments`,
`RemindOrder`) that calls the same use cases the API does, so a retry, a replay or a lost signal cannot do harm; transient
failures are retried with backoff, permanent ones (not found, invalid) are not, and documents that keep failing (upload-service
down for long) are left to the sweeper without holding up the reminder.

It is switched on with `TemporalAddress`. Starting and signalling are best effort and never fail a request: with Temporal down the
service serves as before, and **the sweeper still runs as the safety net** (expiring holds, reminders, documents), so nothing
depends on Temporal being up. Every replica runs a worker; Temporal gives each workflow task to one of them. The workflow ID
carries the task queue, so give staging and production different queues if they share a Temporal namespace.

Tests: the workflow is tested with Temporal's test environment and virtual time (expiry, payment, cancellation, refund before the
reminder, a payment in flight, retries, a permanent failure), with real activities against the real database, and against a real
server: `TEST_TEMPORAL_ADDRESS=localhost:7233 go test ./internal/integration -run TemporalServer -v` (see the file).

## Invoices, showtimes, seat maps, e-mail

- **Invoices**: `GET /orders/{id}/invoice` is the receipt of a paid order, as JSON, a printable page (`?format=html`, escaped) or a PDF (`?format=pdf`; see *Documents*). Number `INV-<year>-<order id>`; prices include VAT at `VATPercent` and the VAT is shown.
  Issued in the name of `InvoiceSellerName` (with tax id and address) as the platform invoicing on the organizers' behalf.
  A refunded order's invoice says so.
- **Showtimes**: an event has one or more **sessions** (`starts_at`, `ends_at`, optional `label`). `POST /events/{id}/sessions`
  adds one (`copy_from` gives it the ticket types of another session, with fresh stock and seats); `PUT` moves it (its buyers are
  told), `POST .../cancel` cancels one session (its orders are refunded, the others are untouched, the event goes on),
  `DELETE` removes a session that has no ticket types. Every ticket type belongs to one session and an order is for
  one session: buying, the refund cutoff, transfers, resale, gate scanning, expiry and the reminder all use the
  ticket's own showtime, not the event's. The event's `starts_at`/`ends_at` are the first start and the last end of its
  sessions that are still on (an event with several sessions cannot have its dates edited directly). Search lists an event
  once, with `next_session_at`, and a date filter matches any showtime still to come. A copy of an event
  (`POST /events/{id}/duplicate`) repeats its sessions shifted by the difference in dates, and `GET /events/{id}/series`
  lists the other published events of the same show.
- **Seat maps**: see the next section.
- **E-mail** through notification-service (`NotificationServiceURL`): the receipt with the order, an offer of a ticket (also to
  an address that has no account), an invitation, a cancelled event and a reminder. E-mails are queued in the database
  next to the notification and sent by the sweeper, so they survive a restart and a down notification-service (tried up
  to 5 times); they are only queued when a notification-service is configured. The templates (`ticket_paid`,
  `ticket_transfer_offer`, `ticket_invited`, `event_cancelled`, `event_reminder`, `resale_sold`) ship in notification-service.
  Dates are written in `TimeZone`.
- **Transfers to a user id check that the user exists** in user-service. An offer by e-mail cannot be checked (user-service
  cannot look people up by e-mail): whoever signs in with that address may accept it.

## What is covered from ticketbox.vn

| ticketbox feature | here |
| --- | --- |
| Browse / search events by text, city, category, date range, price; sort by date, popularity, newest; featured events | `GET /events`, `GET /categories` |
| Event page with ticket tiers, price, availability, "sold out", sale windows | `GET /events/{id}` |
| Seat map, pick your own seats, for halls of any shape (stages, sections, aisles, wheelchair places) | `GET /events/{id}/seat-map`, `seat_ids` on the order; see *Venues and seat maps* |
| Checkout with a countdown hold, buyer details, per-order and per-account ticket limits | `POST /orders` (`expires_at` is the countdown) |
| Promo / voucher codes (percent or fixed, limited uses, min tickets, validity window) | `/events/{id}/promotions`, `promo_code` on the order |
| Pay by wallet or by payment provider; refunds | `POST /orders/{id}/pay`, `POST /orders/{id}/cancel` |
| E-tickets with a QR code, "my tickets" | `GET /my/tickets` (`code` is what the QR encodes) |
| Organizer tools: create events, tiers, seats, promos, orders, attendee list, sales report | `/events/{id}/...`, `GET /my/events` |
| Gate scanning: each ticket admitted once, a copied QR code is caught | `POST /events/{id}/check-in` |
| Event moderation by the platform | `/admin/events`, `approve`, `reject` |
| Give the ticket to a friend (transfer), attendee names, replace a leaked QR code | see *Ticket lifecycle* |
| Invitation tickets (free, from the same stock, outside the buyer's limits and before sales open) | `POST /events/{id}/invitations` |
| "Notify me when tickets are back" for a sold-out type | `PUT/DELETE /events/{id}/ticket-types/{type}/waitlist`, `GET /waitlist` |
| Gate staff who scan but do not manage; scanning offline and uploading later | `/events/{id}/staff`, `POST /events/{id}/check-in/batch` |
| Sales per day, attendee export (spreadsheet-safe CSV) | `GET /events/{id}/report`, `GET /events/{id}/attendees.csv` |
| Resale of tickets at up to face value, paid through wallets | see *Resale* |
| Invoices (VAT, printable), several showtimes, e-mails | see *Invoices, showtimes, seat maps, e-mail* |
| Favourite events | `/wishlist` |
| Notifications: order paid, event cancelled + refund, "starts tomorrow" reminder | `/notifications` (+ optional push) |

Not built: rendering the QR image (the API returns the code; any client draws it) and SMS (notification-service has no SMS
provider).

## API

Base path `/api/v1`. JSON in and out (unknown fields are rejected), amounts are integers in the currency's minor unit
(`150000` is 150,000 VND), times are RFC 3339. Lists that grow with use are paged with `?limit=` and
`?cursor=<next_cursor>`; event search uses `?limit=` and `?offset=`. Errors are `{"error": "..."}` with
`400` bad input, `401` no/invalid token, `403` not allowed, `404` not found (also for other people's things),
`409` conflict / **sold out**, `402` payment problem, `429` busy (retry after `Retry-After`), `503` overloaded.

| | Public | |
| --- | --- | --- |
| `GET /events` | `q`, `city`, `category`, `from`, `to`, `min_price`, `max_price`, `featured`, `sort` (`upcoming` `popular` `newest`) | search |
| `GET /categories` | | the categories |
| `GET /events/{id}` | | an event with its ticket types (a token shows drafts to their organizer) |
| `GET /events/{id}/ticket-types/{type}/seats` | | seat map: `available` / `held` / `sold` |

| | Organizer (the event's) / admin | |
| --- | --- | --- |
| `POST /events`, `PUT /events/{id}`, `GET /my/events` | | create, edit, list mine |
| `POST /events/{id}/submit`, `POST /events/{id}/cancel` | `{"reason"}` | send for review, cancel |
| `POST /events/{id}/ticket-types`, `PUT .../ticket-types/{type}` | | tiers; raising or lowering `total` moves the available count with it |
| `POST /events/{id}/ticket-types/{type}/seats` | `{"section","rows","seats_per_row","first_row"}` | add seats to a `seated` type (rows are lettered `A`, `B`, ...) |
| `GET/POST /events/{id}/promotions`, `PUT .../promotions/{promo}` | | promo codes |
| `GET /events/{id}/orders`, `/attendees`, `/report` | | orders, attendees (`?cursor`), sales report |
| `POST /events/{id}/check-in` | `{"code"}` | scan a ticket: `200`, or `409` with the ticket if already used |
| `POST /events/{id}/check-in/batch` | `{"scans":[{"code","scanned_at"}]}` | up to 500 scans from an offline gate; each gets `admitted` / `already_used` / `not_found` / `void` / `expired` / `transferring` / `event_over` |
| `GET /events/{id}/tickets/{code}`, `POST /tickets/{id}/void`, `POST /tickets/{id}/revert-check-in` | | look up, revoke, un-scan |
| `POST /events/{id}/invitations` | `{"ticket_type_id","quantity","user_id","name","email","note"}` | give free tickets |
| `POST /events/{id}/duplicate` | `{"title","starts_at","ends_at","copy_promotions"}` | a copy of the event (all its showtimes, shifted) |
| `GET /events/{id}/sessions`, `POST /events/{id}/sessions`, `PUT/DELETE .../sessions/{session}`, `POST .../sessions/{session}/cancel` | `{"starts_at","ends_at","label","copy_from"}`, `{"reason"}` | showtimes of an event |
| `PUT /events/{id}/ticket-types/{type}/seat-map` | `{"width","height","stage","sections","seats"}` | draw the map of one type by hand |
| `POST /venues`, `/venues/preview`, `GET/PUT/DELETE /venues/{id}`, `POST /venues/{id}/copy`, `POST /events/{id}/seat-map/apply` | see *Venues and seat maps* | venues, and seats made from them |
| `GET/POST /events/{id}/staff`, `DELETE .../staff/{user}`, `GET /my/staff-events` | `{"user_id"}` | gate staff |
| `GET /events/{id}/attendees.csv` | | the attendee list as a spreadsheet-safe CSV |
| `GET /admin/events`, `POST /events/{id}/approve`, `.../reject`, `PUT .../featured` | admin | moderation |

| | Buyer | |
| --- | --- | --- |
| `POST /orders` | `{"event_id","items":[{"ticket_type_id","quantity","seat_ids"}],"promo_code","buyer_name","buyer_email","buyer_phone","request_id"}` | hold tickets |
| `GET /orders`, `GET /orders/{id}` | | my orders (with tickets once paid) |
| `POST /orders/{id}/pay`, `POST /orders/{id}/cancel` | `{"method","provider"}` | pay, cancel or refund |
| `GET /tickets` (`?event_id`, `?status`), `GET /tickets/{id}`, `/history`, `PUT /tickets/{id}/holder`, `POST /tickets/{id}/reissue` | | my tickets and what I can do with one (`/my/tickets` is the same list) |
| `POST /tickets/{id}/transfer`, `GET /transfers`, `POST /transfers/{id}/accept`, `/decline`, `/cancel` | `{"to_user_id","to_email","message"}` | pass a ticket on |
| `GET /waitlist`, `PUT/DELETE /events/{id}/ticket-types/{type}/waitlist` | | tell me when tickets are back |
| `POST /tickets/{id}/resale`, `GET /resale`, `DELETE /resale/{id}`, `POST /resale/{id}/buy`, public `GET /events/{id}/resale` | `{"price"}` | resell a ticket, buy a resold one |
| `GET /orders/{id}/invoice` (`?format=html\|pdf`), `GET /orders/{id}/documents`, `GET /orders/{id}/documents/{invoice\|tickets}`, `GET /tickets/{id}/pdf` | | the receipt, and the PDFs of an order |
| public `GET /events/{id}/series` | | other showtimes |
| `GET /wishlist`, `PUT/DELETE /wishlist/{event_id}` | | favourites |
| `GET /notifications`, `.../unread-count`, `POST .../read-all`, `POST .../{id}/read` | | inbox |

`request_id` makes `POST /orders` idempotent: send the same one again (a double click, a retry after a timeout) and
you get the same order back instead of holding twice.

## The last ticket: a million buyers, one ticket

When a sale opens, everyone arrives at once and Postgres has to serialise them on the one row that counts the
tickets. Correctness and survival are separate problems here, and they are solved separately.

**Correctness — nobody is oversold, whatever the timing, with any number of replicas.** A ticket type keeps
`total`, `available` (nobody holds or owns it) and `sold`. Holding tickets is one statement:

```sql
update ticket_type set available = available - $n where id = $1 and active and available >= $n and <on sale now>
```

Postgres re-evaluates the `WHERE` on the current version of the row once it wins the row lock, so of any number of
buyers racing for the last ticket exactly one `UPDATE` matches; the rest match no row and are told `sold out`. The
whole purchase (seats, ticket types, promo use, the order) is one transaction, taken in a fixed order (seats
without waiting, ticket types by id) so two orders can never deadlock, and rolled back as a unit. The table has
`CHECK (available >= 0 AND available + sold <= total)`: even a bug in the application cannot persist an oversell.
The same holds for promo codes (`max_uses`), per-account limits (one account's requests are serialised by an advisory
lock, so parallel requests cannot slip past the cap), and seats (`FOR UPDATE SKIP LOCKED`).

**Survival — almost none of the million ever reaches that row.** In front of the database, cheapest first
(`internal/application/service/hotpath.go`):

1. `MaxInFlight` — a pod handles at most N requests at once; beyond that it answers `503` immediately.
2. **Sold-out check before authentication** — `POST /orders` first asks an in-memory sold-out cache (one shared
   lookup per ticket type per moment, via single-flight). Once the type has run out, every later request is refused
   from memory with `409`, without an identity lookup and without touching Postgres.
3. **Per-account rate limit** and a cap on **unpaid orders per account** (`MaxPendingPerUser`).
4. **Bulkhead** — only `ReserveConcurrency` requests per pod compete in the database; the rest wait
   `ReserveQueueWaitMillis`, then get `429` and retry. The connection pool is bounded too, and lock waits are capped
   (`lock_timeout`), so a rush never parks a thousand connections behind one row.
5. The public event page and seat map are served from a `PublicCacheTTLSeconds` in-memory cache, shared by
   concurrent misses.

The caches only ever *refuse*; a stale "available" just lets a request through to the database, which has the final
say, and tickets that come back (an expired hold) invalidate the cache at once on the pod that saw it, or after
`SoldOutTTLSeconds` on the others.

`internal/integration` proves it against a real Postgres with several independent pods. On a laptop:
1 ticket and **1,000,000 buyers** on 3 pods → exactly 1 order, 999,999 told sold out, in about 5 seconds; with every
in-memory protection switched off, 3,000 simultaneous buyers for 5 tickets → exactly 5 orders.

## Calling ticket-service from other services

Inside the cluster ticket-service is reachable at `http://ticket-service-svc:80`. Besides the public API (which needs a
user's token) it has an **internal API** under `/internal/v1` for services acting on their own:

| | |
| --- | --- |
| `GET /ticket-codes/{code}?event_id=` | find a ticket by the code its QR carries (any event when `event_id` is left out) |
| `GET /tickets/{id}`, `GET /tickets/{id}/history`, `GET /users/{user}/tickets?event_id=&status=&cursor=` | read tickets |
| `POST /tickets/{id}/void`, `POST /tickets/{id}/revert-check-in` | revoke, un-scan |
| `POST /events/{id}/check-in`, `POST /events/{id}/check-in/batch` | scan (a turnstile service, a mobile gateway) |
| `GET /orders/{id}`, `GET /orders/{id}/invoice`, `POST /orders/{id}/cancel` | read, invoice, cancel or refund |
| `GET /events/{id}`, `GET /events/{id}/report` | availability, sales |
| `POST /events/{id}/invitations` | give free tickets (a rewards or partner service) |

It is off unless `InternalAPIKey` is set. Callers send it as `X-Internal-Key` (not a user token) and act with admin
rights; `X-Acting-User: <id>` names the user they act for, and is recorded in ticket histories. The key is in the Secret
`ticket-service-secret` (`internalApiKey`), so a caller mounts it with `secretKeyRef`. The chart's ingress routes only
`/api/v1`; never route `/internal`.

Go services use the client, which depends on nothing but the standard library:

```go
import "github.com/JIeeiroSst/ticket-service/pkg/ticketclient"

c := ticketclient.New("http://ticket-service-svc:80", os.Getenv("TicketServiceKey"))
ctx = ticketclient.WithActingUser(ctx, adminID)

t, err := c.TicketByCode(ctx, code, 0)
_, err = c.CheckIn(ctx, eventID, code)          // ticketclient.IsAlreadyCheckedIn(err) carries who/when
_, err = c.Invite(ctx, eventID, ticketclient.Invitation{TicketTypeID: id, Quantity: 1, UserID: winner, Name: "...", Email: "..."})
page, err := c.UserTickets(ctx, userID, ticketclient.TicketFilter{Status: "valid"})
```

## Configuration

Environment variables (the Helm chart sets them):

| Variable | Default | |
| --- | --- | --- |
| `PORT` | `8080` | |
| `HostPostgres` `PortPostgres` `DatabasePostgres` `UserPostgres` `PasswordPostgres` `SSLModePostgres` | `localhost` `5432` `ticket_service` `postgres` (empty) `disable` | the schema is applied on startup (replicas take an advisory lock) |
| `PostgresMaxConns` | `10` | per pod; replicas × this must stay under Postgres `max_connections` |
| `UserServiceURL` | **required** | validates tokens (`UserServiceValidatePath`, `UserServiceUserPath`, `AuthCacheTTLSeconds`, `AdminRole`) |
| `WalletServiceURL`, `PaymentServiceURL` | at least one required | how orders are paid |
| `NotificationServiceURL` | empty | forwards notifications to devices and sends e-mails (none are queued without it) |
| `InternalAPIKey` | empty | the key of the internal API (above); empty turns it off |
| `TransferTTLHours`, `MaxTicketTransfers` | `48`, `3` | how long an offer of a ticket waits; how often a ticket may change hands (0 = unlimited) |
| `ResaleFeePercent`, `PlatformWalletID` | `10`, empty | the platform's share of a resale, and the wallet it is paid to (empty: no fee) |
| `InvoiceSellerName`, `InvoiceSellerTaxID`, `InvoiceSellerAddress`, `VATPercent` | `Ticket Service`, empty, empty, `0` | who invoices, and the VAT rate contained in prices |
| `TimeZone` | `Asia/Ho_Chi_Minh` | the zone of dates in messages, e-mails and PDFs |
| `UploadServiceURL`, `UploadServiceKey` | empty | where PDFs are kept, and the service key to call it (required together) |
| `TemporalAddress`, `TemporalNamespace`, `TemporalTaskQueue` | empty, `default`, `ticket-service-orders` | run each order as a workflow; empty leaves it to the sweeper |
| `OrderHoldMinutes` | `10` | how long unpaid tickets are held |
| `SweepIntervalSeconds` | `30` | release expired holds, refund cancelled events, close unanswered transfers, expire tickets of ended events, remind, push |
| `ReserveConcurrency` `ReserveQueueWaitMillis` `SoldOutTTLSeconds` `PublicCacheTTLSeconds` `UserRatePerSecond` `UserBurst` `MaxPendingPerUser` `MaxInFlight` | `16` `250` `2` `2` `2` `10` `3` `10000` | the rush protection above; `0` disables each |

## Architecture

Hexagonal (ports and adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                          fx.New(bootstrap.Module).Run()
internal/bootstrap                   the fx module: every provider and lifecycle hook (server, sweeper, pool)
internal/domain                      entities and rules, no dependencies: Event, TicketType, Order, Ticket, Promotion ...
internal/application/port/inbound    what the outside asks of the service: EventUseCase, OrderUseCase, GateUseCase ...
internal/application/port/outbound   what the service needs: repositories, IdentityProvider, WalletGateway, PaymentGateway ...
internal/application/service         the use cases, plus the rush protection (hotpath.go) — depend on ports only
internal/adapter/inbound/http        net/http handlers, DTOs, middleware
internal/adapter/outbound/postgres   pgx repositories (the atomic SQL lives here)
internal/adapter/outbound/*service   HTTP clients of user-, wallet-, payment- and notification-service
pkg/ticketclient                     the Go client for the internal API (standard library only)
schema.sql                           the database schema (embedded)
```

Dependencies point inward: `domain` imports nothing, services import `domain` and the ports, adapters import the
ports, and only `bootstrap` knows them all. Replace an adapter (say, Redis for the sold-out cache) by implementing a
port and changing one line in `bootstrap`.

Background work (`startSweeper`, every replica, no leader election — each step is safe to run concurrently because the
database serialises it): release expired holds, refund the orders of cancelled events, close unanswered transfers, expire
the tickets of ended events, remind buyers a day ahead, push notifications.

## Tests

```
go test ./...                                  # unit tests
TEST_DATABASE_URL='postgres://postgres:pw@localhost:5432/postgres?sslmode=disable' go test -race ./...
```

With `TEST_DATABASE_URL` the tests in `internal/integration` run against a real Postgres (each creates and drops its
own database); without it they are skipped. They cover the flash sale (pods, oversell, per-account caps, promo uses,
seats, churn of payments / cancellations / expiries with a balanced ledger), the whole event lifecycle, and the HTTP
API end to end through the real fx module. `STORM_USERS` (default 20000) and `STORM_WORKERS` (default 2000) size the
storms, e.g. `STORM_USERS=1000000 go test ./internal/integration -run LastTicket -v`.

```
docker build -t ticket-service .
```
