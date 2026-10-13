# wishlists-service

Customer wishlists (`wishlists_wishlist`) and their product lines (`wishlists_line`), following the ERD in `voucher_voucher_erd.webp`. The owner (`auth_user`) lives in user-service: `owner_id` stores the user-service user id, and the public profile (`id`, `username`, `name`, `active`) is fetched from user-service `GET /api/v1/users/{id}/profile` and returned as `owner` on every wishlist.

The service holds no token of its own for user-service. It forwards the caller's credentials: the access token the user got from `POST /api/v1/login` (`Authorization: Bearer <session_token>`), or the `X-Internal-Token` another service sent. Anonymous requests only see owners that are already cached in Redis.

## Architecture

Hexagonal (ports & adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                              fx.New(infrastructure.Options()...)
internal/
  domain/model                           Wishlist, Line, Visibility, Actor, domain errors, validation
  domain/port                            driving (WishlistUsecase) and driven (WishlistRepository, HealthChecker) ports
  application                            use cases: ownership + visibility rules
  adapter/primary/http                   net/http handlers, routes, auth (pkg/authn)
  adapter/secondary/postgres             GORM repository + embedded schema.sql (applied on start)
  adapter/secondary/userservice          auth_user lookups against user-service (UserDirectory port)
  adapter/secondary/redis                cache decorators (fx.Decorate) for the repository and UserDirectory, rate limiter
  adapter/secondary/memory               in-memory repository and user directory (tests)
  infrastructure                         fx module graph, logging
  infrastructure/database                *gorm.DB provider + lifecycle
  infrastructure/cache                   *redis.Client provider (nil when REDIS_ADDR is unset)
  infrastructure/server                  HTTP server lifecycle
pkg/authn                                shared Bearer / X-Internal-Token authentication
```

## Data model

| table | columns |
|---|---|
| `wishlists_wishlist` | `id`, `name`, `key` (unique share key), `visibility` (`Private`/`Shared`/`Public`), `date_created`, `owner_id` |
| `wishlists_line` | `id`, `wishlist_id` (FK, cascade), `product_id`, `title`, `quantity`; unique (`wishlist_id`, `product_id`) |

Adding a product that is already in the wishlist increases its quantity (capped at 10000). A wishlist holds at most 500 distinct products. A user's first wishlist is their default one; it is created on demand, guarded by a per-owner advisory lock so concurrent requests never create two.

Visibility:

- `Private`: only the owner (and admins / internal callers).
- `Shared`: anyone who has the key (`GET /api/v1/shared-wishlists/{key}`).
- `Public`: anyone, and listed under `GET /api/v1/users/{userId}/wishlists`.

## API

Responses are `{"data": ...}`, errors are `{"error": "..."}`. List endpoints return `{"data": {"items", "total", "limit", "offset"}}`.

| method | path | auth |
|---|---|---|
| GET | `/livez`, `/health` | none (`/health`: postgres is critical, redis only degrades) |
| POST | `/api/v1/wishlists` | user; `owner_id` only for admin / internal (owner is checked in user-service) |
| GET | `/api/v1/wishlists?owner_id=&limit=&offset=` | user (own lists; other owners: public only) |
| GET | `/api/v1/wishlists/{id}` | owner, or anyone if public |
| PUT/PATCH | `/api/v1/wishlists/{id}` `{name, visibility, regenerate_key}` | owner |
| DELETE | `/api/v1/wishlists/{id}` | owner |
| POST | `/api/v1/wishlists/{id}/lines` `{product_id, title, quantity}` | owner |
| DELETE | `/api/v1/wishlists/{id}/lines` | owner, removes every line |
| PUT/PATCH | `/api/v1/wishlists/{id}/lines/{lineId}` `{quantity}` | owner |
| DELETE | `/api/v1/wishlists/{id}/lines/{lineId}` | owner |
| POST | `/api/v1/wishlists/{id}/lines/{lineId}/move` `{target_wishlist_id}` | owner of both lists; merges into an existing line |
| GET | `/api/v1/shared-wishlists/{key}` | anyone with the key (not private) |
| POST | `/api/v1/shared-wishlists/{key}/copy` | user; copies into a new private list |
| GET | `/api/v1/users/{userId}/wishlists` | anyone (public lists) |
| GET | `/api/v1/me/wishlists/default` | user; created if missing |
| POST | `/api/v1/me/wishlist-items` `{product_id, title, quantity}` | user; adds to the default list |
| GET | `/api/v1/me/products/{productId}/wishlists` | user; `{in_wishlist, wishlists}` for a product page |
| GET | `/api/v1/wishlist-stats/popular-products?limit=` | anyone; most wished products |

Auth: `Authorization: Bearer <session>` is validated against user-service; service-to-service calls use `X-Internal-Token`. If `INTERNAL_API_TOKENS` is unset, anonymous requests are accepted as internal calls with admin rights (gradual rollout, same as the other services); set it in every shared environment.

The share `key` is only returned to the owner (and admins); other readers of a public wishlist get an empty `key`. Share keys are redacted from access logs.

## Redis

Redis is optional and fail-open: when `REDIS_ADDR` is unset or Redis is down, every request still works against PostgreSQL and user-service.

- wishlists by id and share key, invalidated on every write. Each wishlist has a generation counter; a read only fills the cache under the generation it started with, so a slow read can never put stale data back after a write.
- public user profiles from user-service (`CACHE_USER_TTL`), unknown users cached for 1 minute, concurrent lookups collapsed with singleflight. Lookups run in parallel (8 at a time, 3s budget) and a failed lookup only drops `owner` from the response.
- popular products (`CACHE_POPULAR_TTL`).
- per-user write rate limit (`RATE_LIMIT_WRITES_PER_MINUTE`, fixed one-minute window, 429 + `Retry-After`); internal callers are exempt.

## Configuration

| env | default |
|---|---|
| `HTTP_PORT` | `9000` |
| `LOG_LEVEL` | `info` |
| `DB_HOST` / `DB_PORT` | `localhost` / `5432` |
| `DB_USER` / `DB_PASSWORD` | `postgres` / empty |
| `DB_NAME` | `wishlists_service` |
| `DB_SSLMODE` | `disable` |
| `DB_MAX_OPEN_CONNS` / `DB_MAX_IDLE_CONNS` | `10` / `5` |
| `USER_SERVICE_URL` | `http://user-api-svc` |
| `USER_SERVICE_TIMEOUT` | `2s` |
| `INTERNAL_API_TOKENS` | empty (comma separated, accepted from callers) |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | empty / empty / `0` |
| `REDIS_PREFIX` | `wishlists:` |
| `CACHE_WISHLIST_TTL` / `CACHE_USER_TTL` / `CACHE_MISSING_TTL` / `CACHE_POPULAR_TTL` | `5m` / `10m` / `1m` / `5m` |
| `REDIS_OP_TIMEOUT` | `300ms` |
| `RATE_LIMIT_WRITES_PER_MINUTE` | `120` (`0` disables) |

## Development

```
make run
make test
WISHLISTS_TEST_DB_HOST=localhost WISHLISTS_TEST_DB_PASSWORD=pw go test ./internal/adapter/secondary/postgres/
make docker
```

Redis tests use miniredis, so no Redis server is needed. The PostgreSQL repository test is skipped unless `WISHLISTS_TEST_DB_HOST` is set; it drops and recreates the wishlist tables in the target database.
