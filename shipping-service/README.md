# shipping-service

Domestic shipping (Vietnam only) on top of [Giao Hàng Nhanh (GHN)](https://api.ghn.vn): route selection
(fastest / cheapest / balanced), an API other services use to book deliveries, the full lifecycle of a
shipment, and status notifications through `notification_service` plus signed callbacks to the caller.

Hexagonal architecture wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                            fx.New(infrastructure.Module)
internal/domain/model                  shipment, lifecycle state machine, route ranking, VN address rules
internal/domain/port                   driving (use case) and driven (carrier, repositories, notifier...) interfaces
internal/application                   quote, create/place, cancel, sync, GHN webhook, outbox dispatcher
internal/adapter/primary/http          gin API with API-key auth, GHN webhook endpoint
internal/adapter/secondary/ghn         GHN public API client, cached province/district/ward master data
internal/adapter/secondary/notification  notification_service client
internal/adapter/secondary/callback    signed HTTP callbacks to the calling service
internal/adapter/secondary/repository  MySQL (gorm)
internal/infrastructure                config, database, HTTP server, outbox worker, fx module
```

## Route selection

A route is one pickup warehouse × one GHN service available between that warehouse's district and the
recipient's. For every route the service asks GHN for the fee (`/v2/shipping-order/fee`) and the expected
delivery time (`/v2/shipping-order/leadtime`), then ranks them:

| Strategy | Picks |
|---|---|
| `CHEAPEST` | lowest fee, ties → earliest delivery |
| `FASTEST` | earliest delivery, ties → lowest fee |
| `BALANCED` (default) | lowest 50/50 mix of fee and delivery time, each normalised to 0..1 across the options |

`POST /shipments/quote` returns every option, the best one per strategy, and the routes GHN refused with
the reason. `POST /shipments` either takes a `strategy` or a fixed `warehouse_id` + `service_id`.

## Vietnam only

- Recipient and warehouse addresses must resolve in GHN master data: the `ward_code` must belong to
  `district_id`, and `district_id` to `province_id` when given. GHN master data covers Vietnam only.
  Use `/locations/*` to build address pickers.
- Phone numbers must be Vietnamese mobile (`0[35789]xxxxxxxx`) or landline (`02xxxxxxxxx`) numbers;
  `+84`/`84` prefixes, spaces, dots and dashes are normalised.
- COD is capped by `MAX_COD_AMOUNT` (VND, default 50,000,000).

## Lifecycle

```
PENDING ──► CREATED ──► PICKING ──► PICKED ──► IN_TRANSIT ──► DELIVERING ──► DELIVERED
   │  ▲        │           │                        ▲   │           │  ▲
   ▼  │        ▼           ▼                        │   ▼           ▼  │
REJECTED    CANCELLED   CANCELLED               DELIVERY_FAILED ──► RETURNING ──► RETURNED
                        (any active state) ──► EXCEPTION (lost, damaged) ──► resumes or ends
```

- `PENDING`: stored, not yet accepted by GHN. `REJECTED`: GHN refused it (reason in `last_error`).
  Re-posting the same `client_order_code` or `POST /shipments/:id/place` retries.
- Placement uses the shipment code as GHN `client_order_code`; a retry first looks the order up at GHN
  (`detail-by-client-code`), so a timeout never produces two carrier orders. A lease prevents two
  placements of the same shipment at once.
- GHN statuses are mapped (`ready_to_pick` → CREATED, `storing`/`sorting`/`transporting` → IN_TRANSIT,
  `delivery_fail`/`waiting_to_return` → DELIVERY_FAILED, `return*` → RETURNING, `lost`/`damage` → EXCEPTION, ...).
- Every carrier update is stored in `shipment_events`. Duplicates are dropped; an update that is not a
  valid transition (for example a late `picked` after `delivering`) is recorded with `applied=false` and
  does not move the shipment backwards.
- Cancel: locally while `PENDING`/`REJECTED`; through GHN while `CREATED`/`PICKING`; refused after pickup.

## Notifications and callbacks

Each applied transition writes outbox jobs in the same transaction; a background worker delivers them
with exponential backoff (5s doubling, capped at 1h, 10 attempts) and `FOR UPDATE SKIP LOCKED`, so
several replicas can run it. Delivery is at-least-once.

- **notification_service** (`POST /api/v1/notifications`): Vietnamese messages to the customer for
  CREATED, PICKED, DELIVERING, DELIVERED, DELIVERY_FAILED, RETURNING, RETURNED, CANCELLED and EXCEPTION,
  by `email` if `customer.email` is set and `push` if `customer.user_id` is set.
- **Callback** to the caller's `callback_url` on every applied transition, with headers
  `X-Shipping-Event-Id` (deduplicate on it) and `X-Shipping-Signature: sha256=<hex HMAC-SHA256 of the body
  with CALLBACK_SECRET>`. The host must match `CALLBACK_ALLOWED_HOSTS` (exact host, or a suffix starting
  with `.`); redirects are not followed.

## API

All `/api/v1` routes except the webhook need `X-API-Key`. Each key belongs to one calling service
(`API_KEYS=order-service:<key>,shop-service:<key>`); a service only sees its own shipments, and
`client_order_code` is unique per service (re-posting it returns the existing shipment, `200`).

| Method | Path | |
|---|---|---|
| POST | `/api/v1/shipments/quote` | `recipient`, `parcel`, `cod_amount`, `insurance_value`, `warehouse_ids` |
| POST | `/api/v1/shipments` | quote fields + `client_order_code`, `strategy`, or `warehouse_id`+`service_id`, `payment_type` (`SENDER`/`RECIPIENT`), `required_note` (`CHOTHUHANG`/`CHOXEMHANGKHONGTHU`/`KHONGCHOXEMHANG`), `note`, `customer {user_id, email}`, `callback_url` |
| GET | `/api/v1/shipments?status=&limit=&offset=` | |
| GET | `/api/v1/shipments/:id` | shipment and its event history |
| GET | `/api/v1/shipments/by-client-code/:code` | |
| POST | `/api/v1/shipments/:id/place` | retry placement with GHN |
| POST | `/api/v1/shipments/:id/cancel` | `reason` |
| POST | `/api/v1/shipments/:id/sync` | pull current status from GHN |
| POST / GET | `/api/v1/warehouses` | pickup warehouses: `code`, `name`, `phone`, `street`, `ward_code`, `district_id`, `province_id`, `carrier_shop_id` |
| PATCH | `/api/v1/warehouses/:id` | `active` |
| GET | `/api/v1/locations/provinces`, `/districts?province_id=`, `/wards?district_id=` | GHN master data |
| POST | `/api/v1/webhooks/ghn?token=<GHN_WEBHOOK_TOKEN>` | GHN status callback |

Recipient: `name`, `phone`, `street`, `ward_code`, `district_id`, `province_id`.
Parcel: `weight_gram`, `length_cm`, `width_cm`, `height_cm`, `content`, `items[{name, code, quantity, weight_gram, price}]`.

Errors: 400 invalid input, 401 bad API key / webhook token, 404 not found (or another service's shipment),
409 conflict / invalid transition, 422 no route or GHN rejected, 502 GHN unavailable. When placement fails,
the body also carries the stored `shipment`.

## GHN setup

1. Create a GHN account (use `https://dev-online-gateway.ghn.vn` for testing) and get the API token.
2. Create one GHN shop per pickup warehouse; register each warehouse here with its `carrier_shop_id`.
   GHN picks up from the shop's registered address, so keep the warehouse address identical to it.
3. Set the GHN webhook URL to `https://<public host>/api/v1/webhooks/ghn?token=<GHN_WEBHOOK_TOKEN>`.
   GHN does not sign webhooks, so the secret token in the URL is what authenticates them.

## Configuration

| Env | Default |
|---|---|
| `PORT_HTTP_SERVER` | `8094` |
| `MYSQL_HOST` / `MYSQL_PORT` / `MYSQL_USER` / `MYSQL_PASSWORD` / `MYSQL_DBNAME` | `localhost` / `3306` / `root` / empty / `shipping_service` |
| `GHN_BASE_URL` | `https://dev-online-gateway.ghn.vn/shiip/public-api` (production: `https://online-gateway.ghn.vn/shiip/public-api`) |
| `GHN_TOKEN`, `GHN_WEBHOOK_TOKEN` | required; webhooks are rejected while the token is empty |
| `GHN_TIMEOUT`, `GHN_LOCATION_CACHE_TTL` | `10s`, `24h` |
| `NOTIFICATION_BASE_URL` | empty = notifications off (in the chart: `http://notification-service-svc`) |
| `CALLBACK_ALLOWED_HOSTS`, `CALLBACK_SECRET` | empty = callbacks refused |
| `API_KEYS` | empty = every API call is refused |
| `MAX_COD_AMOUNT`, `OUTBOX_INTERVAL` | `50000000`, `2s` |

```
make run
make test
```
