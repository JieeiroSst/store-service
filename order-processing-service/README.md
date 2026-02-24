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

## Proxy Layer

The proxy layer (`internal/proxy/`) handles all internal API calls to microservices:

- **PaymentProxy** → Payment Service (`/api/v1/payments/...`)
- **InventoryProxy** → Inventory Service (`/api/v1/inventory/...`)
- **ShippingProxy** → Shipping Service (`/api/v1/shipments/...`)
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
