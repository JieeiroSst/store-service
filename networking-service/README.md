# networking-service

A Consul-compatible service-networking server: service registry with health
checking, a KV store with sessions and locks, blocking queries, service-mesh
intentions, a DNS interface and a web UI. It speaks Consul's HTTP API, so the
official client (`github.com/hashicorp/consul/api`, which the other services in
this repo already use) works against it unchanged; the compatibility tests
drive it with that client.

State lives in one of two stores, picked with `STORE_BACKEND`:

- `redis` (chart default): every replica reads and writes the same Redis, so
  the service scales out; one replica, elected through a Redis lock, runs the
  health checks and reaps sessions.
- `memory`: a single replica keeps everything in process memory and
  snapshots it to disk.

There is no gossip, no multi-datacenter federation and no Envoy data plane;
see [Not implemented](#not-implemented).

The UI is served at `/ui/`.

## Features

| Area | What it does |
|------|--------------|
| Service registry | Agent API (`/v1/agent/service/register`) and catalog API (`/v1/catalog/register`, for external services); tags, meta, node meta filters |
| Health checks | HTTP (method, headers, body, TLS skip-verify), TCP and TTL checks with Consul's semantics: 2xx passing, 429 warning, anything else critical. Checks of *every* node are run, including external ones (what consul-esm does). `DeregisterCriticalServiceAfter` reaps dead instances. Maintenance mode per service or node |
| Discovery | `/v1/health/service/<name>?passing&tag=`, `/v1/catalog/service/<name>`, DNS `<service>.service.consul` (A/AAAA/SRV, RFC 2782 `_svc._tag` form, critical instances left out, answers shuffled) |
| KV store | Get/put/delete, `?recurse`, `?keys&separator=`, `?raw`, flags, check-and-set (`?cas=`), 512 KiB value limit |
| Sessions & locks | Sessions with TTL (invalidated after 2×TTL like Consul), `release`/`delete` behavior, lock-delay, bound to node checks (`serfHealth` by default); `?acquire=` / `?release=` on KV. `api.LockKey` leader election works as-is |
| Blocking queries | `?index=&wait=` on every read endpoint; `X-Consul-Index` per table, writes that change nothing do not wake watchers |
| Intentions | Allow/deny between services with wildcards and Consul's precedence (exact > `*`); `/check`, `/match`, exact and ID-based CRUD; configurable default |
| ACL | Management token, read-only tokens and a default policy (`X-Consul-Token`, `Authorization: Bearer` or `?token=`) |
| Storage | Redis backend: writes are `WATCH`/`MULTI` transactions serialized on one version key (CAS and locks stay exact across replicas), blocking queries on every replica are woken through pub/sub with an index poll as fallback, a Redis error is a 500/SERVFAIL rather than an empty answer |
| Persistence | Atomic JSON snapshot in `DATA_DIR` every `SNAPSHOT_INTERVAL` (when changed) and on shutdown; restored on start. `GET/PUT /v1/snapshot` for backups |
| Metrics | Prometheus `/metrics`: `networking_catalog_objects{kind}`, `networking_health_checks{status}`, `networking_check_runs_total{type,status}`, `networking_sessions_invalidated_total{reason}`, `networking_blocking_query_wakeups_total`, `networking_state_index` |

## Architecture

Hexagonal (ports & adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                          fx.New(infrastructure.Module)
config/                              env config
web/                                 embedded UI (single index.html)
internal/
  domain/                            nodes, services, checks, KV pairs, sessions,
                                     intentions, snapshot; validation, precedence and the
                                     merge/CAS/lock rules both stores share (rules.go)
  port/                              CatalogStore, KVStore, SessionStore, IntentionStore
                                     (+ StateStore), Leadership, SnapshotStorage, Prober,
                                     Metrics
  application/                       CatalogService (agent/catalog/health), KVService,
                                     SessionService, IntentionService, Authorizer (ACL),
                                     Blocker (blocking queries), HealthRunner (check
                                     scheduler + session reaper), Snapshotter
  adapter/primary/http/              Consul HTTP API (/v1/...) + UI route
  adapter/primary/dns/               Consul DNS interface (UDP + TCP)
  adapter/secondary/memory/          in-memory state store with per-table indexes
                                     and a watch channel (Consul's memdb + Raft index)
  adapter/secondary/redisstore/      Redis state store + leader Elector (Leadership)
  adapter/secondary/storetest/       contract tests every state store must pass
  adapter/secondary/probe/           HTTP/TCP check prober
  adapter/secondary/snapshot/        JSON snapshot file (or no-op without DATA_DIR)
  adapter/secondary/metrics/         Prometheus Metrics
  infrastructure/                    fx module, store selection (store.go) and lifecycle
                                     (restore -> HTTP -> DNS -> checks; stopped in
                                     reverse, final snapshot last)
```

## API

All paths match Consul's. Reads accept `?index`, `?wait`, `?dc` (only the
configured datacenter), `?node-meta=k:v` and `?pretty`.

| Method | Path | |
|--------|------|-|
| GET | `/v1/agent/self`, `/v1/agent/members` | this node |
| GET | `/v1/agent/services`, `/v1/agent/service/<id>` | services on this node |
| PUT | `/v1/agent/service/register[?replace-existing-checks]` | `api.AgentServiceRegistration` |
| PUT | `/v1/agent/service/deregister/<id>` | |
| PUT | `/v1/agent/service/maintenance/<id>?enable=&reason=`, `/v1/agent/maintenance?enable=` | maintenance mode |
| GET | `/v1/agent/checks` | |
| PUT | `/v1/agent/check/register`, `/v1/agent/check/deregister/<id>` | |
| PUT | `/v1/agent/check/{pass,warn,fail}/<id>?note=`, `/v1/agent/check/update/<id>` | TTL updates |
| GET | `/v1/agent/health/service/name/<name>` | aggregated status, also as HTTP 200/429/503 |
| PUT | `/v1/catalog/register`, `/v1/catalog/deregister` | external nodes/services/checks |
| GET | `/v1/catalog/{datacenters,nodes,services}`, `/v1/catalog/node/<n>`, `/v1/catalog/service/<s>[?tag=]` | |
| GET | `/v1/health/service/<s>[?passing&tag=]`, `/v1/health/checks/<s>`, `/v1/health/node/<n>`, `/v1/health/state/<any\|passing\|warning\|critical>` | |
| GET/PUT/DELETE | `/v1/kv/<key>` | `?recurse ?keys ?separator ?raw ?flags ?cas ?acquire ?release` |
| PUT | `/v1/session/{create,destroy/<id>,renew/<id>}` | |
| GET | `/v1/session/{info/<id>,list,node/<node>}` | |
| GET/POST | `/v1/connect/intentions` | list / create (returns ID) |
| GET/PUT/DELETE | `/v1/connect/intentions/exact?source=&destination=`, `/v1/connect/intentions/<id>` | |
| GET | `/v1/connect/intentions/check?source=&destination=`, `/v1/connect/intentions/match?by=&name=` | |
| GET | `/v1/status/leader`, `/v1/status/peers` | always this server |
| GET/PUT | `/v1/snapshot` | JSON export / restore (needs a management token when ACLs deny) |
| GET | `/ui/`, `/health`, `/metrics` | UI, probe, Prometheus |

### Examples

```sh
export CONSUL_HTTP_ADDR=http://networking-service-svc:8500

# Register an instance with an HTTP check. Always send Address: an agent
# registration without one resolves to this server's node address, exactly
# like a remote Consul agent would.
curl -X PUT $CONSUL_HTTP_ADDR/v1/agent/service/register -d '{
  "ID": "order-1", "Name": "order-service", "Tags": ["v2"],
  "Address": "10.244.1.17", "Port": 8080,
  "Check": {"HTTP": "http://10.244.1.17:8080/health", "Interval": "10s",
            "DeregisterCriticalServiceAfter": "5m"}}'

curl "$CONSUL_HTTP_ADDR/v1/health/service/order-service?passing"
dig @networking-service-svc -p 8600 order-service.service.consul SRV

# Config, watched with a blocking query
curl -X PUT $CONSUL_HTTP_ADDR/v1/kv/order_service -d @consul.json
curl "$CONSUL_HTTP_ADDR/v1/kv/order_service?index=42&wait=5m"
```

From Go, use the official client unchanged:

```go
client, _ := api.NewClient(&api.Config{Address: "networking-service-svc:8500"})
pair, _, _ := client.KV().Get("order_service", nil)
lock, _ := client.LockKey("service/order/leader") // leader election
```

## Configuration

| Env | Default | |
|-----|---------|-|
| `PORT_HTTP_SERVER` | `8500` | HTTP API + UI |
| `PORT_DNS_SERVER` | `8600` | DNS (UDP+TCP); empty disables it |
| `NODE_NAME` | hostname | this server's node; keep it stable across restarts |
| `NODE_ADDRESS` | `127.0.0.1` | node address (pod IP in k8s) |
| `DATACENTER` | `dc1` | |
| `DNS_DOMAIN` / `DNS_TTL` / `DNS_ONLY_PASSING` | `consul` / `0s` / `false` | |
| `ACL_DEFAULT_POLICY` | `allow` | `deny` requires a token for everything but `/health`, `/metrics`, the `/ui/` page, `/v1/status/*` and `/v1/catalog/datacenters` |
| `ACL_MANAGEMENT_TOKEN` / `ACL_READ_TOKENS` | | full-access token / comma-separated read-only tokens |
| `INTENTIONS_DEFAULT` | `allow` | decision when no intention matches |
| `KV_MAX_VALUE_KB` | `512` | |
| `BLOCKING_DEFAULT_WAIT` / `BLOCKING_MAX_WAIT` | `5m` / `10m` | |
| `CHECK_TICK_INTERVAL` / `CHECK_MAX_CONCURRENT` / `CHECK_DEFAULT_TIMEOUT` | `1s` / `64` / `10s` | check scheduler |
| `DATA_DIR` / `SNAPSHOT_INTERVAL` | (none) / `30s` | snapshot file; empty = none. A snapshot is only restored into an empty store, never over live Redis state |
| `STORE_BACKEND` | `memory` | `memory` or `redis` |
| `REDIS_ADDR` / `REDIS_PASSWORD` / `REDIS_DB` | `localhost:6379` / / `0` | |
| `REDIS_PREFIX` | `networking:` | every key, the pub/sub channel and the leader lock live under it |
| `REDIS_POLL_INTERVAL` / `REDIS_OP_TIMEOUT` | `1s` / `5s` | index poll backing up pub/sub; per-operation timeout |
| `LEADER_TTL` | `10s` | leader lock TTL; a dead leader is replaced within it |

## Not implemented

Rejected with `400` rather than silently ignored where a client could send them:
script, Docker, gRPC, UDP, H2PING, OS-service and alias checks; `TLSServerName`;
session `ServiceChecks`; L7 intention `Permissions`; `?filter=` expressions.
Not served at all: Raft/gossip (HA comes from Redis instead), WAN federation, Connect CA and
Envoy proxies, prepared queries, `/v1/txn`, config entries, namespaces and
partitions, the full ACL policy/role system. `/v1/snapshot` uses this server's
JSON format, not Consul's archive.

## Development

```sh
make run     # :8500 HTTP, :8600 DNS
make test    # go test -race ./...: hashicorp/consul/api compatibility tests and the
             # store contract run against both backends (Redis via miniredis)
make docker
```

## Deployment

Chart: `chart/networking-service`, Service on 8500 plus 8600/UDP+TCP for DNS.
By default it uses the Redis backend with 1/2/3 replicas in dev/staging/prod
and rolling updates, against a dedicated Redis the chart deploys
(`networking-service-redis`, a StatefulSet with a 1Gi PVC, AOF fsync every
second plus RDB, `maxmemory-policy noeviction`). It deliberately does not use
the umbrella's shared `redis-svc`: that one is a cache with no volume, and a
registry must neither lose nor evict keys. A single Redis is still a single
point of failure (state survives restarts, not node loss); for that, set
`redis.enabled: false` and point `store.redisAddr` at a managed or Sentinel
Redis. Set `redisPasswordSecretRef` to turn on AUTH for both sides. For
the memory backend set `store.backend: memory`, `env.nodeAddress: ""`,
`replicas: 1`, `strategy.type: Recreate` and `persistence.enabled: true`. CI: the
`networking_service` job in `.github/workflows/go.yml`; release images via
`.github/workflows/networking-service-cicd.yml`; local cluster builds via
`deploy/ci/services.yaml`.
