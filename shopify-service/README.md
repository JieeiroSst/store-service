# shopify-service

Syncs products and orders from a [Shopify](https://www.shopify.com/vn) store into MySQL through the
GraphQL Admin API, creates products in the store, and keeps local data current via webhooks.

Hexagonal architecture wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                          fx.New(infrastructure.Module)
internal/domain/model                Product, ProductVariant, Order, LineItem, WebhookEvent
internal/domain/port                 driving (use case) and driven (repository, Shopify) interfaces
internal/application                 use cases: products, orders, webhooks
internal/adapter/primary/http        gin handlers and routes
internal/adapter/secondary/repository  MySQL (gorm)
internal/adapter/secondary/shopify   GraphQL Admin API client, webhook HMAC verifier
internal/infrastructure              config, logger, database, HTTP server, fx module
```

## API

| Method | Path | |
|---|---|---|
| GET | `/health` | |
| POST | `/api/v1/products` | create in Shopify, then store locally. Body: `title` (required), `description_html`, `vendor`, `product_type`, `tags`, `status` (`ACTIVE`/`DRAFT`/`ARCHIVED`) |
| GET | `/api/v1/products?limit=&offset=` | local products with variants |
| GET | `/api/v1/products/:id` | |
| POST | `/api/v1/products/sync` | pull every product from Shopify (upsert) |
| GET | `/api/v1/orders?limit=&offset=` | local orders with line items |
| GET | `/api/v1/orders/:id` | |
| POST | `/api/v1/orders/sync` | pull every order from Shopify (upsert) |
| POST | `/api/v1/webhooks` | Shopify webhook endpoint |

Webhooks are verified with `X-Shopify-Hmac-Sha256`, de-duplicated by `X-Shopify-Webhook-Id`, and the
resource is re-fetched from the GraphQL API before it is stored. Handled topics: `products/create`,
`products/update`, `products/delete`, `orders/create`, `orders/updated`, `orders/paid`,
`orders/cancelled`, `orders/fulfilled`. Other topics are acknowledged and ignored.

## Shopify setup

1. Shopify admin → Settings → Apps and sales channels → Develop apps → create a custom app.
2. Admin API scopes: `read_products`, `write_products`, `read_inventory`, `read_orders`
   (`read_all_orders` to sync orders older than 60 days).
3. Install the app and copy the Admin API access token (`shpat_...`) and the client secret.
4. Register webhooks for the topics above pointing to `https://<public host>/api/v1/webhooks` (JSON format).

## Configuration

| Env | Default | |
|---|---|---|
| `PORT_HTTP_SERVER` | `8092` | |
| `MYSQL_HOST` / `MYSQL_PORT` / `MYSQL_USER` / `MYSQL_PASSWORD` / `MYSQL_DBNAME` | `localhost` / `3306` / `root` / empty / `shopify_service` | schema in `database.sql` is applied at startup |
| `SHOPIFY_SHOP_DOMAIN` | | e.g. `my-store.myshopify.com` |
| `SHOPIFY_ACCESS_TOKEN` | | Admin API access token |
| `SHOPIFY_API_SECRET` | | client secret; webhooks are rejected while empty |
| `SHOPIFY_API_VERSION` | `2026-07` | quarterly version, keep within Shopify's supported window |
| `SHOPIFY_TIMEOUT` | `10s` | |

Setting `HostConsul`, `KeyConsul` and `ServiceConsul` in `.env` loads the same config from Consul instead.

Variants and line items are read 50 at a time inside list queries to stay under Shopify's query cost
limit; any beyond that are fetched with follow-up queries (250 per request), so nothing is truncated.
Throttled requests are retried up to 3 times.

Without the `read_all_orders` scope Shopify only returns orders from the last 60 days, silently: the
sync succeeds but older orders are missing. For a custom app, enable the scope and reinstall; a public
app must request it from Shopify first.

```
make run
make test
```
