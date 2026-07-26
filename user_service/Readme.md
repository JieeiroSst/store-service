# user_service

User accounts, roles, and the login/session lifecycle for the platform:
sign-up, profile management, role-based access, and a JWT access + rotating
refresh token flow for login/refresh/logout. Exposed over gRPC with an
HTTP/JSON gateway (grpc-gateway) generated from the shared `lib-gateway`
proto definitions.

This service absorbed what used to be two separate services -
`session-cookie-auth-service` (cookie-based session demo) and
`oauth2-service` (OAuth2 authorization-server skeleton) - since neither had
a real, working token store and nothing else in the platform depended on
OAuth2 client/authorization-code flows. What was reusable from both (JWT
access tokens, refresh-token rotation) now lives here as one Redis-backed
implementation; both source services have been removed.

## Architecture

Hexagonal (ports & adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
main.go                                      entrypoint - cmd.Execute()
cmd/cmd.go                                   cobra "api" command -> fx.New(di.Module).Run()
config/config.go                             env/Consul-based configuration, incl. TokenPolicy
di/
  module.go                                  fx providers: config, db, cache, adapters, services, handler
  server.go                                  fx.Lifecycle: starts/stops the gRPC server + HTTP gateway

internal/
  domain/                                    plain structs, no framework deps
    user.go, role.go, roleitem.go            User, Role, RoleItem entities
    token.go                                 AccessClaims, Session, TokenPair
    errors.go                                sentinel errors + cache key format

  port/
    input/                                   driving ports - what the gRPC adapter calls INTO the app
      auth.go, user.go, role.go, roleitem.go
    output/                                  driven ports - what the app calls OUT to infra
      user_repository.go, role_repository.go, roleitem_repository.go
      token_generator.go, token_store.go, hasher.go

  application/                               business logic, implements the driving ports
    auth/auth_service.go                     login/logout/refresh/validate - see below
    user/user_service.go                     sign-up, profile, cache-through FindUser
    role/role_service.go, roleitem/roleitem_service.go

  adapter/
    inbound/grpcadapter/handler.go           driving adapter - implements the generated UserServiceServer
    outbound/
      pg/                                    driven adapter - GORM/Postgres repositories
      sessionstore/token_store.go            driven adapter - Redis-backed session store
      jwttoken/token_generator.go            driven adapter - JWT access tokens + opaque refresh tokens
      pwhash/hasher.go                       driven adapter - bcrypt
```

Dependencies only ever point inward (adapter -> application -> domain/port).
The application layer never imports gorm, redis, or jwt directly - only the
`port` interfaces - so any of those can be swapped without touching the
login logic itself.

## Login/session lifecycle

1. **Login** (`internal/application/auth`) checks the password (bcrypt) and
   issues an access token (JWT, short TTL) and a refresh token (opaque
   random value, long TTL), persisted together as one session in Redis.
2. **While the access token is valid**, `ValidateSession`/`Authentication`
   check it directly: JWT signature + expiry, plus a store lookup so a
   logout can revoke it before it naturally expires.
3. **Once the access token has expired**, the caller calls `RefreshToken`
   with the refresh token. If it's still valid, a brand new access+refresh
   pair is issued and persisted *before* the old pair is invalidated, so a
   transient store error never leaves the caller with zero valid sessions.
4. **If the refresh token itself is invalid or expired**, `RefreshToken`
   returns `Unauthenticated` - there's nothing left to rotate, so the caller
   has to `Login` again to get a fresh pair.
5. **Logout** deletes the session (both the access- and refresh-token
   entries) from the store - real revocation, not just a client-side
   forget.

TTLs are controlled by `config.TokenPolicy` (`Token.AccessTokenExpMinutes` /
`Token.RefreshTokenExpHours` in Consul/`.env`, defaulting to 15m/168h if
unset). Keep the access TTL well below the refresh TTL - the whole
refresh-then-relogin ordering above depends on the access token expiring
first.

## Features

Beyond the login flow, the service exposes (see
`internal/adapter/inbound/grpcadapter/handler.go` for the full RPC set):

- **Accounts** - `SignUp`, `UpdateProfile`, `FindUser` (Redis cache-through).
- **Roles** - `CreateRole`, `UpdateRole`, `DeleteRole`, `GetRole`, `ListRoles`.
- **Role assignment** - `AddRoleItem`, `UpdateItemRole`, `RemoveRoleItem`.

`LockAccount` and `Authentication` exist in the application/port layer but
aren't currently wired to a gRPC method.

## Running locally

```
# .env points at Consul (HostConsul/KeyConsul/ServiceConsul); adjust as needed
# seed Consul's KV with consul.json (or point HostConsul at your own KV)
go run . api
```

Postgres tables (`users`, `roles`, `user_roles`) are created automatically
via `AutoMigrate` on startup (see `di.newDB` in `di/module.go`).

- gRPC: `:1236` (`Server.PortGrpcServer`)
- HTTP/JSON gateway: `:1235` (`Server.PortHttpServer`)

### Docker

```
docker build -t user-service .
docker run -p 1235:1235 -p 1236:1236 user-service
```
