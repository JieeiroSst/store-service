# room-service

Group chat rooms. Users come from **user-service** (login and session validation over gRPC); room-service stores rooms, members and messages in MySQL and pushes new messages to members over WebSockets.

- `BE/` – Go, hexagonal architecture, wired with [uber-go/fx](https://github.com/uber-go/fx)
- `FE/` – React (create-react-app)

## Backend layout

```
internal/domain/model        entities (Room, Member, Message, Event)
internal/domain/port         driving ports (services) and driven ports (repositories, user directory, event publisher)
internal/application         use cases: auth, rooms + membership, chat
internal/adapter/primary     http (gin) and ws (hub/clients)
internal/adapter/secondary   mysql, userservice (gRPC), eventbus (Redis pub/sub), memory (tests)
internal/infrastructure      fx module and HTTP server lifecycle
```

## API (all under `/api`)

| | |
|---|---|
| `POST /auth/login` `{username,password}` | tokens from user-service |
| `POST /auth/refresh` `{refresh_token}` | new tokens |
| `GET /me` | current user |
| `GET /rooms` · `POST /rooms` `{name}` | my rooms · create (caller becomes owner) |
| `GET /rooms/:id` | room details (members only) |
| `GET /rooms/:id/members` · `POST` `{username}` · `DELETE /:username` | list · add (owner) · remove (owner, or yourself to leave) |
| `GET /rooms/:id/messages?before=<id>&limit=` | history, oldest first |
| `GET /rooms/:id/ws?token=<access token>` | WebSocket: send `{"content":"…"}`; receive `message`, `member_added`, `member_removed` events |

Authenticated routes take `Authorization: Bearer <access token>`.

## Notes

- **Members are keyed by username.** user-service has no lookup-by-username, so adding a member doesn't check the account exists; a typo is just an invitation nobody can claim.
- **Sessions:** tokens are validated by user-service (`ValidateSession`), so logout is honoured; results are cached for `AUTH_CACHE_TTL_SECONDS` (default 30).
- **Multiple replicas:** set `REDIS_ADDR` so chat events reach members connected to other replicas. Without it, run a single replica.

## Run locally

```sh
docker compose up --build        # MySQL + BE (:8081) + FE (:3000)
# needs user-service gRPC reachable at USER_SERVICE_GRPC_ADDR (default host.docker.internal:1236)

cd BE && go test ./...            # ROOM_TEST_MYSQL_DSN=… also runs the MySQL adapter test
cd FE && npm start                # dev server, proxies /api to :8081
```

Config is environment variables; see `BE/.env.example`. Deployment: `.github/workflows/room-service-cicd.yml` and `room-web-cicd.yml`, charts in `chart/room-service` and `chart/room-web`.
