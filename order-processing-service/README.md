# Order Processing Service

A Temporal-based order processing workflow service built with **Go Clean Architecture**.

## Architecture

```
cmd/
├── worker/       # Temporal Worker entry point
├── starter/      # Starts an Order Processing Workflow Execution
└── cron/         # Starts the Cron Cleanup Workflow

internal/
├── domain/
│   ├── entity/       # Domain entities (Order, Payment, Shipping, etc.)
│   ├── repository/   # Repository interfaces
│   └── usecase/      # Business logic interfaces & implementation
├── activity/         # Temporal Activity Definitions
├── workflow/         # Temporal Workflow Definitions (main, child, cron)
├── worker/           # Worker configuration & registration
├── proxy/            # HTTP proxy layer for internal API calls
├── config/           # Application configuration
└── infrastructure/
    ├── temporal/     # Temporal client factory
    └── http/         # Repository implementations

pkg/
├── constants/    # Shared constants (task queues, statuses, timeouts)
├── errors/       # Custom error types
└── logger/       # Structured logging (zap)
```

## Temporal Features Used

| Feature | Location | Description |
|---------|----------|-------------|
| **Workflow Definition** | `internal/workflow/order_workflow.go` | Main order processing orchestration |
| **Workflow Execution** | `cmd/starter/main.go` | Starts workflow via `ExecuteWorkflow` |
| **Temporal Activity** | `internal/activity/order_activity.go` | Activities with heartbeats & retries |
| **Child Workflows** | `internal/workflow/payment_child_workflow.go`, `shipping_child_workflow.go` | Isolated payment & shipping flows |
| **Temporal Cron Job** | `internal/workflow/cron_workflow.go`, `cmd/cron/main.go` | Periodic stale order cleanup |
| **Temporal Worker** | `internal/worker/worker.go`, `cmd/worker/main.go` | Hosts workflow & activity code |
| **Events & Event History** | Used throughout | Saga compensation, query handlers |
| **Retry Policies** | All workflows & activities | Configurable retry with backoff |
| **Heartbeats** | Long-running activities | Progress tracking & failure detection |
| **Query Handlers** | `order_workflow.go` | Query current order status |
| **Saga Pattern** | `order_workflow.go` | Compensation/rollback on failure |

## Workflow Flow

```
OrderProcessingWorkflow
    │
    ├── 1. ValidateOrderActivity
    ├── 2. ReserveInventoryActivity
    ├── 3. PaymentChildWorkflow ──► ProcessPaymentActivity
    ├── 4. ShippingChildWorkflow ──► CreateShipmentActivity + SendNotificationActivity
    ├── 5. SendNotificationActivity (email)
    └── 6. UpdateOrderStatusActivity (COMPLETED)
    
    On failure at any step → Run compensations in reverse (Saga pattern)
```

## Paid shop orders (Keiki Shop)

basket-service starts `PaidOrderWorkflow` (task queue `ORDER_PROCESSING_TASK_QUEUE`, workflow id `paid-order-basket-<id>`) after the wallet payment succeeds:

```
PaidOrderWorkflow
    ├── 1. RecordPaidOrder      order_order + order_line (idempotent per basket, number KS<8 digits>)
    ├── 2. NotifyOrderPlaced    push via notification-service (failure is logged, not fatal)
    ├── 3. SetPaidOrderStatus   Processing
    └── 4. CreateShipment       GHN shipment through shipping-service
```

The GHN shipment uses the delivery address that basket-service snapshots at checkout with the buyer's own token
(`shipping_address` in the workflow input: name, phone, street, `ward_code`, `district_id`, `province_id`; the app stores
the GHN codes in the address `postcode` as `ghn:<province>:<district>:<ward>`).

- The buyer pays the shipping fee at checkout: basket-service quotes it through shipping-service (`POST /api/v1/baskets/{id}/shipping-quote`,
  same logic as checkout), charges goods + fee to the wallet only if the fee is not above what the buyer confirmed
  (`shipping_fee` in the checkout body, otherwise 409), and passes `shipping_fee` in the workflow input. It is stored in
  `order_order.shipping_incl_tax`; the shipment's insurance value is the goods total without the fee.
- Route lock: basket-service also passes the quoted `shipping_route` (`warehouse_id`, `service_id`, `fee`). The shipment is
  created on exactly that GHN service so the buyer gets what they paid for; if shipping-service answers "no route" for it
  (warehouse disabled, service gone), the request is retried once with `strategy=CHEAPEST` and the result is flagged
  `route_changed`. The real GHN fee is stored in `order_order.shipping_carrier_fee` next to the paid `shipping_incl_tax`.
- Request: `POST /api/v1/shipments` with `X-API-Key`, `client_order_code` = order number (re-posting returns the same
  shipment, and re-places a `PENDING`/`REJECTED` one), `strategy=CHEAPEST`, `payment_type=SENDER`, `cod_amount=0` (already
  paid), `insurance_value` = order total capped at 5,000,000, `customer.user_id` so shipping-service notifies the buyer.
- Parcel: one item per order line, `PARCEL_ITEM_WEIGHT_GRAM` (default 500) per unit, 30 × 20 × (10 + 5 per unit, at most 6) cm.
- Outcomes are written to `order_order`: created → `shipping_method=GHN`, `shipping_code=<shipment code>`, status
  `Awaiting pickup`; skipped (no GHN address, shipping not configured) or rejected (no route, invalid phone, bad API key) →
  status `Needs shipping` so staff can ship it by hand (admin console "Vận chuyển" in the app).
- shipping-service outages (5xx, timeouts) are retried by Temporal (12 attempts, 5 s doubling up to 5 min); if they still
  fail the order stays `Processing`/`Needs shipping` and the paid order is never failed because of shipping.
- The step is guarded by `workflow.GetVersion("create-ghn-shipment")`, so workflows started before this change replay
  without it.

| Env | |
|---|---|
| `ORDER_DB_DSN` | MySQL DSN of the `basket_service` schema; without it `PaidOrderWorkflow` is not registered |
| `SHIPPING_SERVICE_URL` | default `http://localhost:8083` (chart: `http://shipping-service-svc`) |
| `SHIPPING_SERVICE_API_KEY` | key registered in shipping-service `API_KEYS` as `order-service:<key>`; empty = no shipments, orders go to `Needs shipping` |
| `PARCEL_ITEM_WEIGHT_GRAM` | default `500` (basket-service quotes with the same 500 g per unit) |

shipping-service scopes shipments by API key. To let staff see and re-place these shipments from the app's admin
console (which goes through gateway-api), give gateway-api the same key for `shipping-service` (`X-API-Key` in
gateway-api `consul.json`) as `SHIPPING_SERVICE_API_KEY` here.

basket-service needs its own key too (`basket-service:<key>` in shipping-service `API_KEYS`, and
`SHIPPING_SERVICE_API_KEY=<key>` in the basket-service env file). Without it basket-service does not charge shipping
and checkout behaves as before (the shop pays GHN).

## Proxy Layer

The proxy layer (`internal/proxy/`) handles all internal API calls to microservices:

- **PaymentProxy** → Payment Service (`/api/v1/payments/...`)
- **InventoryProxy** → Inventory Service (`/api/v1/inventory/...`)
- **ShippingProxy** → Shipping Service (`/api/v1/shipments/...`) for the demo `OrderProcessingWorkflow`
- **ShipmentProxy** → shipping-service (`POST /api/v1/shipments`, `X-API-Key`) for `PaidOrderWorkflow`
- **NotificationProxy** → Notification Service (`/api/v1/notifications/...`)

Each proxy uses standard `net/http` clients with timeouts, API key authentication, and structured error handling.

## Quick Start

```bash
# 1. Start Temporal (choose one)
docker compose up -d                    # Docker
temporal server start-dev --ui-port 8233  # CLI

# 2. Copy env
cp .env.example .env

# 3. Start the Worker
make worker

# 4. Start a Workflow (new terminal)
make starter

# 5. Start Cron Workflow (optional)
make cron
```

## Temporal UI

Open http://localhost:8233 to view workflow executions, event history, and query results.
