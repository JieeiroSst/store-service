# cdn-service

Internal CDN for company use across Vietnam only. This repo contains the
**origin API** (this Go service) plus the architecture design for the edge
layer that fronts it. The edge/cache/LB/DNS/monitoring stack described below
is a design doc, not runnable code in this repo — see [Scope](#scope).

## Scope

| Layer | Status |
|---|---|
| Origin API (Go, hexagonal, fx, Gin, GORM) | **Built** — this repo |
| Postgres (file metadata) | **Built** — `docker-compose.yml` for local dev |
| MinIO (S3-compatible object storage) | **Built** — `docker-compose.yml` for local dev |
| Edge PoPs (Nginx cache) | Design only — [Edge cache](#edge-cache-design) |
| Load balancing (HAProxy) | Design only — [Load balancing](#load-balancing-design) |
| GeoDNS / latency routing | Design only — [DNS routing](#dns-routing-design) |
| Monitoring (Prometheus/Grafana) | Origin metrics **built** (`/metrics`); edge metrics design only |

## Architecture overview

```
                     ┌───────────────────────────┐
  upload (auth'd) ─▶ │   cdn-service (origin)     │──▶ Postgres (metadata)
                     │   POST /files/presign      │
                     │   POST /files/:id/confirm  │──▶ MinIO (object bytes)
                     └───────────────────────────┘
                                  ▲
                                  │ cache miss only
                     ┌────────────┴────────────┐
                     │        HAProxy           │
                     └─────────────┬────────────┘
                          ┌────────┴────────┐
                          ▼                 ▼
                 ┌────────────────┐ ┌────────────────┐
                 │ Edge PoP        │ │ Edge PoP        │
                 │ Viettel IDC     │ │ VNPT/FPT DC     │
                 │ (Nginx cache)   │ │ (Nginx cache)   │
                 └────────────────┘ └────────────────┘
                          ▲                 ▲
                          └────── GeoDNS ────┘
                                     │
  download ───────────────────────▶ client (resolves to nearest PoP)
```

Upload and download are intentionally separate paths:

- **Upload** goes straight to the origin (this API + MinIO), never through
  the edge cache — reliability and auth matter more than latency here.
- **Download** goes through the edge cache first; the edge only calls back
  to the origin on a cache miss.

## Upload flow (no cache, direct to origin, authenticated)

1. `POST /api/v1/files/presign` (requires `X-API-Key`) — client asks for a
   place to upload. The API creates a `files` row (`status=pending`) and
   returns a MinIO **presigned PUT URL**.
2. Client `PUT`s the file bytes **directly to MinIO** using that URL — the
   bytes never pass through this API, so the origin isn't a bandwidth
   bottleneck.
3. `POST /api/v1/files/:id/confirm` (requires `X-API-Key`) — client tells
   the API the upload finished; the API `StatObject`s MinIO to verify it
   really exists and records the real size, flips `status=uploaded`.

## Download flow (through edge cache, origin only on miss)

- `GET /api/v1/files/:id/download` — 302-redirects to
  `{EDGE_BASE_URL}/{object_key}`, i.e. the edge PoP's public hostname. The
  edge layer (not built here) is responsible for caching that object by
  path/extension and only pulling from MinIO on a miss.
- `GET /api/v1/files/:id/download?direct=true` — bypasses the edge
  entirely and returns a short-lived presigned GET straight from MinIO.
  Useful for internal debugging or before the edge layer exists.
- `GET /api/v1/files/:id` — metadata only, no redirect.
- `GET /api/v1/files?limit=&offset=` — paginated metadata listing.
- `DELETE /api/v1/files/:id` (requires `X-API-Key`) — deletes the MinIO
  object + metadata row. Does **not** purge the edge cache (no edge nodes
  are deployed yet) — once the edge layer exists, wire this to an
  Nginx `proxy_cache_purge` call or equivalent.

## Edge cache design

Two PoPs, each at a different Vietnamese ISP's datacenter (e.g. Viettel
IDC and VNPT/FPT), each running Nginx with `proxy_cache` keyed by request
path, with rules tuned by extension:

```nginx
proxy_cache_path /var/cache/nginx/cdn levels=1:2 keys_zone=cdn_cache:100m
                  max_size=50g inactive=24h use_temp_path=off;

server {
    listen 443 ssl;
    server_name cdn.internal.company.vn;

    location / {
        proxy_pass http://origin_upstream;
        proxy_cache cdn_cache;
        proxy_cache_key $scheme$request_method$host$uri;
        proxy_cache_valid 200 206 24h;   # images/css/js
        proxy_cache_use_stale error timeout updating http_500 http_502 http_503;
        add_header X-Cache-Status $upstream_cache_status;
    }

    location ~* \.(?:html|json)$ {
        proxy_pass http://origin_upstream;
        proxy_cache cdn_cache;
        proxy_cache_valid 200 5m;         # short TTL for anything more dynamic
    }
}
```

Purge on origin change: either run with the (paid) `ngx_cache_purge`
module, or use short TTLs + a cache-busting query param
(`?v={updated_at}`) built from the `files.updated_at` column — cheaper to
operate than wiring a purge endpoint into every PoP.

## Load balancing design

HAProxy in front of the two edge PoPs, health-checking each and failing
over on outage:

```
frontend cdn_front
    bind *:443 ssl crt /etc/haproxy/certs/cdn.pem
    default_backend cdn_edges

backend cdn_edges
    balance leastconn
    option httpchk GET /healthz
    server edge-viettel 10.0.1.10:443 check ssl verify none
    server edge-vnpt    10.0.2.10:443 check ssl verify none
```

## DNS routing design

Full Anycast is unnecessary for a 2-3 PoP, Vietnam-only footprint. Options,
cheapest first:

1. **Split-horizon internal DNS** — resolve `cdn.internal.company.vn` to a
   different PoP IP depending on the resolver's source subnet (map
   corporate office/VPN egress IPs to their nearest PoP). Simplest, no new
   infra, works if client population is small and mostly known.
2. **GeoIP-aware DNS** (PowerDNS or CoreDNS + a GeoIP backend) — resolve
   based on the client resolver's IP ISP/ASN, approximating "nearest
   Vietnamese ISP PoP" without true RTT measurement.
3. **HAProxy/Nginx-level redirect** — client always hits one PoP first,
   which 302-redirects to the ISP-appropriate PoP based on `X-Forwarded-For`
   ASN lookup. No DNS changes needed, but adds a hop on first request.

Recommendation: start with (1), move to (2) once PoP count or client
diversity grows past what a static subnet map can handle.

## Monitoring

- **Origin**: `GET /metrics` (Prometheus format) — `cdn_service_http_requests_total`
  and `cdn_service_http_request_duration_seconds`, labeled by method/route/status.
- **Edge** (once deployed): `nginx-prometheus-exporter` per PoP exposes
  `nginx_http_requests_total` and cache status counters (parse
  `X-Cache-Status` via an Nginx log-based exporter, e.g. `mtail` or
  `nginx-log-exporter`, to derive cache hit ratio).
- **HAProxy**: built-in Prometheus exporter (`haproxy_backend_up`,
  `haproxy_backend_response_time_average_seconds`) for per-PoP latency.
- All scraped by a central Prometheus (already available as `chart/prometheus`
  in this monorepo) and visualized in Grafana (`chart/grafana`).

## Security

- **TLS** terminates at the edge PoPs / ingress, not in this Go process.
- **Upload auth**: static `X-API-Key` header, checked in constant time
  (`middleware.APIKeyAuth`). Rotate via the `secret.api_key` Consul KV
  value.
- **Rate limiting**: in-memory per-IP token bucket
  (`middleware.RateLimit`), single-instance only — if cdn-service is
  scaled horizontally, back it with Redis instead (already available as
  `chart/redis`).

## Local development

```bash
cd cdn-service
docker compose up -d          # postgres + minio

# consul.json targets the in-cluster ClusterIP (postgresdb:80, see
# chart/postgres) — override host/port for a local docker-compose Postgres
# before seeding:
jq '.postgres.postgresql_host="localhost" | .postgres.postgresql_port="5432"' \
  consul.json | consul kv put cdn_service -

go run ./cmd
```

### End-to-end smoke test

```bash
BASE=http://localhost:2232
API_KEY=change-me

# 1. presign — ask for an upload slot
resp=$(curl -s -X POST "$BASE/api/v1/files/presign" \
  -H "X-API-Key: $API_KEY" -H "Content-Type: application/json" \
  -d '{"file_name":"logo.png","content_type":"image/png","size_bytes":1024}')
echo "$resp"

file_id=$(echo "$resp" | jq -r .file_id)
upload_url=$(echo "$resp" | jq -r .upload_url)

# 2. PUT the bytes directly to MinIO (bypasses the API)
curl -s -X PUT "$upload_url" -H "Content-Type: image/png" --data-binary @logo.png

# 3. confirm — tell the API the upload finished
curl -s -X POST "$BASE/api/v1/files/$file_id/confirm" -H "X-API-Key: $API_KEY"

# 4. fetch metadata
curl -s "$BASE/api/v1/files/$file_id"

# 5. list (paginated)
curl -s "$BASE/api/v1/files?limit=20&offset=0"

# 6. download — 302 redirect to the edge-cache URL
curl -sI "$BASE/api/v1/files/$file_id/download"

# 7. download — bypass the edge, redirect straight to a presigned origin GET
curl -sI "$BASE/api/v1/files/$file_id/download?direct=true"

# 8. delete — removes the MinIO object + metadata row
curl -s -X DELETE "$BASE/api/v1/files/$file_id" -H "X-API-Key: $API_KEY" -o /dev/null -w "%{http_code}\n"
```

### API reference

| Method | Path | Auth | Description |
|---|---|---|---|
| POST | `/api/v1/files/presign` | `X-API-Key` | Create a `pending` file record + presigned MinIO PUT URL |
| POST | `/api/v1/files/:id/confirm` | `X-API-Key` | Verify the object landed in MinIO, flip status to `uploaded` |
| GET | `/api/v1/files/:id` | — | Fetch file metadata |
| GET | `/api/v1/files?limit=&offset=` | — | Paginated metadata listing |
| GET | `/api/v1/files/:id/download` | — | 302 to the edge-cache URL |
| GET | `/api/v1/files/:id/download?direct=true` | — | 302 to a short-lived presigned origin GET (bypasses edge) |
| DELETE | `/api/v1/files/:id` | `X-API-Key` | Delete the MinIO object + metadata row |
| GET | `/healthz` | — | Liveness probe |
| GET | `/metrics` | — | Prometheus metrics |

Error responses are JSON: `{"error": "..."}`, with `404` for an unknown
file ID, `400` for invalid/missing input, `413` for a file over
`minio.max_upload_size_bytes`, and `401` for a missing/incorrect
`X-API-Key`.

## Migrations

SQL files live in `/migrations`, versioned and applied via
[golang-migrate](https://github.com/golang-migrate/migrate) (`file://migrations`
source), run automatically on service startup and tracked in the
`schema_migrations` table.
