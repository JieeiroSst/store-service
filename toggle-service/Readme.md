# toggle-service

An enterprise feature-flag management service (Unleash-like) written in Go: projects,
environments, feature flags, activation strategies with constraints, RBAC, API tokens,
audit logging, and a client SDK API backed by a server-side evaluation engine.

## Stack

- Go 1.22+, [uber-go/fx](https://github.com/uber-go/fx) for dependency injection
- [go-chi/chi](https://github.com/go-chi/chi) for HTTP routing
- GORM (`gorm.io/driver/postgres`) + PostgreSQL
- [golang-migrate](https://github.com/golang-migrate/migrate) for versioned SQL migrations
- zap (structured logging), go-playground/validator (request validation)
- JWT (human/admin auth) + sha256-hashed API tokens (SDK/client + admin-API auth)

## Architecture — Hexagonal (Ports & Adapters)

```
internal/
  domain/
    model/     Plain Go structs — the core business entities (Project, FeatureFlag,
               ActivationStrategy, Constraint, User, Role, APIToken, AuditEvent, ...).
               No framework dependencies; these are what every other layer works with.
    port/
      driving.go  "Driving" (inbound) interfaces — what the application offers the
                  outside world (ProjectService, FeatureFlagService, ClientService, ...).
                  Implemented by internal/application, consumed by the HTTP adapter.
      driven.go   "Driven" (outbound) interfaces — what the application needs from
                  infrastructure (ProjectRepository, TokenRepository, ...).
                  Implemented by internal/adapter/secondary/repository.

  application/   The use cases / business logic. Each subpackage (project, environment,
                 featureflag, strategy, auth, rbac, token, audit, client) implements one
                 driving port, depending only on driven ports — never on GORM, chi, or fx
                 directly. `evaluation/` is the pure, side-effect-free flag evaluation
                 engine used by both the client API and (indirectly) by tests.

  adapter/
    primary/http/    The "driving" adapter: chi router, middleware (JWT auth, API-token
                      auth, RBAC permission checks), HTTP handlers, and request/response
                      DTOs. Translates HTTP <-> application service calls.
    secondary/repository/  The "driven" adapter: GORM implementations of every
                      port.XRepository interface, translating domain models <-> Postgres.

  infrastructure/  Composition root and cross-cutting concerns: config loading (Consul,
                    with a plain-env-var fallback), zap logger construction, the gorm.DB
                    connection pool, golang-migrate invocation, and the fx.Lifecycle-managed
                    HTTP server. internal/infrastructure/module.go wires every layer above
                    together into the single fx.App started from cmd/main.go.
```

The dependency direction always points inward: `adapter` depends on `domain/port` and
`application`; `application` depends on `domain/port` and `domain/model`; `domain` depends
on nothing else in the project. This is what lets the evaluation engine, for example, be
unit-tested with zero database or HTTP setup.

## Configuration

Config loading is hybrid, matching the Consul-based bootstrap convention already used by
other services in this monorepo, with a plain-env-var fallback for local development/tests
where Consul isn't running:

1. `.env` (or the process environment) is read via viper for `HostConsul`, `KeyConsul`,
   `ServiceConsul`.
2. If all three are set and Consul is reachable, the full `Config` JSON blob is fetched
   from Consul KV and unmarshaled directly.
3. Otherwise, the service falls back to plain environment variables:

   | Variable             | Default          |
   |-----------------------|-----------------|
   | `HTTP_PORT`           | `8080`          |
   | `APP_ENV`             | `development`   |
   | `DB_HOST`             | `localhost`     |
   | `DB_PORT`             | `5432`          |
   | `DB_USER`             | `postgres`      |
   | `DB_PASSWORD`         | (empty)         |
   | `DB_NAME`             | `toggle_service`|
   | `DB_SSLMODE`          | `disable`       |
   | `JWT_SECRET`          | `change-me`     |
   | `JWT_EXPIRY_MINUTES`  | `60`            |

When deploying via Consul, seed the KV value at `KeyConsul` with a JSON document shaped
like `internal/infrastructure/config.Config` (see `config.go`), e.g.:

```json
{
  "server": { "port": "8080", "env": "production" },
  "postgres": { "host": "postgres", "port": "5432", "user": "toggle", "password": "...", "dbName": "toggle_service", "sslMode": "disable" },
  "jwt": { "secret": "a-real-secret", "expiryMinutes": 60 }
}
```

## Running locally

```bash
# 1. Start Postgres, e.g.:
docker run --rm -p 5432:5432 -e POSTGRES_PASSWORD=postgres -e POSTGRES_DB=toggle_service postgres:16

# 2. Either point HostConsul/KeyConsul/ServiceConsul at a real Consul (see above),
#    or just unset them / leave Consul unreachable to use the DB_* env var fallback:
export DB_HOST=localhost DB_USER=postgres DB_PASSWORD=postgres DB_NAME=toggle_service

# 3. Run — migrations in ./migrations are applied automatically on startup.
go run ./cmd
```

The server listens on `HTTP_PORT` (default 8080). `GET /health` is unauthenticated.

### Bootstrapping the first admin user

`POST /api/admin/auth/register` always creates a non-admin user (there's no public
"become admin" endpoint, by design). Promote the first user manually:

```sql
UPDATE users SET is_admin = true WHERE email = 'you@example.com';
```

Instance admins bypass all per-project RBAC checks. Everyone else needs a
`ProjectMembership` (Owner/Member/Viewer) — creating a project via
`POST /api/admin/projects` automatically makes the creator that project's Owner.

## API surface

- **Admin/management API** (`/api/admin/*`) — JWT bearer auth (`POST /api/admin/auth/login`),
  per-project RBAC via `RequirePermission` middleware. Projects, environments (instance
  admin only), feature flags, activation strategies + constraints, project members/roles,
  API tokens (instance admin only), and audit log queries.
- **Client/SDK API** (`/api/client/*`) — API-token auth (`Authorization: <token>` or
  `Bearer <token>`, scoped to one project+environment):
  - `GET /api/client/features` — raw flags + strategies for local SDK evaluation.
  - `POST /api/client/metrics` — ingest per-flag yes/no evaluation counts.
  - `POST /api/client/evaluate` — runs the server-side evaluation engine for one flag,
    for SDKs that would rather not implement strategy evaluation themselves.

## Testing

```bash
go build ./...
go vet ./...
go test ./...
```

The evaluation engine (`internal/application/evaluation`) is fully unit-tested — default
strategy, flexibleRollout percentage boundaries + deterministic stickiness, userWithId,
remoteAddress (including CIDR), all three constraint operators (IN/NOT_IN/STR_CONTAINS,
case-insensitive), and OR-across-strategies / AND-across-constraints semantics — with no
database required. End-to-end testing against real Postgres/Consul is left to the reader
in this environment.
