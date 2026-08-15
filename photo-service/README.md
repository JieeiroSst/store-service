# photo-service

Image composition (collage) microservice. Merges multiple child images into a
single output image on a configurable layout — even grid, mosaic (one hero
image + smaller ones), or freeform (arbitrary, possibly overlapping,
placement) — and stores the result in MinIO. Built with Hexagonal
Architecture (Ports & Adapters) and [uber-go/fx](https://github.com/uber-go/fx)
for dependency injection.

## Architecture

```
cmd/main.go                          fx.New(...) bootstrap
internal/
  domain/                            entities, value objects, domain errors (no infra imports)
  application/
    ports/                           inbound.go (driving port) + outbound.go (driven ports)
    service/                         ComposeService — the use case, built only from ports
  adapters/
    inbound/http/                    gin HTTP handlers, DTOs, router
    outbound/
      storage/minio/                 ImageStorage adapter (fx module name: minio)
      persistence/pg/                JobRepository adapter, pgx + golang-migrate (fx module name: postgres)
      cache/redis/                   CacheRepository adapter (fx module name: redis)
      imaging/                       ImageComposer + ImageDecoder + ImageFetcher adapters (fx module name: imaging)
pkg/
  config/                            Consul KV-backed config (bootstrap .env -> Consul), fx.Module
  logger/                            zap logger + ports.Logger adapter, fx.Module
migrations/                          embedded SQL migrations
docker-compose.yml                   consul + postgres + redis + minio for local dev
Dockerfile                           multi-stage build for the CI/CD image
chart/photo-service/                 Helm chart deployed by CI/CD (in the monorepo root chart/)
.github/workflows/photo-service-cicd.yml   build/test -> build & push image -> deploy
```

Dependency direction is enforced inward: adapters depend on `application/ports`
and `domain`; `application` depends only on `domain` and the standard library
(including the stdlib `image` package, which the composer/decoder ports use as
a shared abstraction — same tier as `io.Reader`, not "infrastructure"); `domain`
imports nothing outside the standard library. Neither `domain` nor
`application` imports minio, pgx, redis, gin, or zap. Every infrastructure
package exposes its own `fx.Module`, and constructors are provided as the port
interface they satisfy (e.g. `fx.Provide(func(s *Storage) ports.ImageStorage { return s })`),
so the core code only ever sees abstractions.

## Layouts

`LayoutConfig.Type` selects one of three ways `Cells` (image placement
rectangles) are determined — `domain.LayoutConfig.ResolveCells` computes them,
shared by both validation and rendering:

| Type | Description | Cell source |
|------|-------------|-------------|
| `grid` | Even `Rows x Cols` grid (Shopee-style banners: 1x4, 2x2, 3x3, ...) | Derived from `Width/Height/Rows/Cols/Spacing/Padding` |
| `mosaic` | One large hero image + several smaller ones | Explicit `Cells` (must not overlap in intent, though not enforced) |
| `freeform` | Arbitrary placement, overlap allowed | Explicit `Cells`, painted in array order (later cells draw over earlier ones) |

The number of image sources must exactly match the number of resolved cells
(`Rows*Cols` for grid, `len(Cells)` for mosaic/freeform) — too few or too many
sources is rejected with `domain.ErrSourceCellMismatch` before any decoding
happens.

`CellFit` controls how each source is resized into its cell:
`cover` (crop to fill, default), `contain` (fit inside, centered, background
may show through), `stretch` (exact size, aspect ratio ignored).

## API

### `POST /v1/compositions`

The request is routed by `Content-Type`:

- **`multipart/form-data`** — the classic form described below (multiple
  files, URLs, MinIO object keys, plus a `layout` field).
- **anything else** — the request body itself is treated as raw image bytes
  (a single upload source), read into a `bytes.Buffer` and capped at 32 MiB.
  Use this to push an image straight from a byte buffer without building a
  multipart body. `layout`, when needed, is passed as JSON via the
  `X-Layout` header instead of a form field.

#### `multipart/form-data`

| field         | description                                                     |
|---------------|------------------------------------------------------------------|
| `layout`      | JSON `layoutDTO` — see below                                     |
| `images`      | zero or more uploaded image files                                 |
| `urls`        | zero or more URLs to fetch as additional sources                  |
| `object_keys` | zero or more existing MinIO object keys (this service's bucket) as sources |

Sources are ordered `images`, then `urls`, then `object_keys`, in submission
order; that final 0-based index is what `layout.cells[].image_index` refers to.

#### raw buffer body

| part                | description                                                        |
|---------------------|---------------------------------------------------------------------|
| body                | raw image bytes (the full request body), e.g. `Content-Type: image/jpeg` or `application/octet-stream` |
| `X-Layout` header   | optional JSON `layoutDTO` — see below                                |

This path always produces exactly one source (`image_index: 0`), so it only
makes sense with a single-cell layout (e.g. `mosaic`/`freeform` with one
cell, or a `1x1` grid).

`layout` JSON shape:

```jsonc
{
  "type": "grid",              // "grid" | "mosaic" | "freeform" (default: grid)
  "rows": 2, "cols": 2,          // grid only; omit both for an auto near-square grid
  "width": 1200, "height": 1200, // canvas size in px (default 1200x1200)
  "spacing": 8, "padding": 16,
  "background": "#FFFFFF",
  "cell_fit": "cover",           // "cover" | "contain" | "stretch" (default: cover)
  "cells": [                     // mosaic/freeform only
    {"x": 0, "y": 0, "width": 800, "height": 1200, "image_index": 0},
    {"x": 800, "y": 0, "width": 400, "height": 600, "image_index": 1},
    {"x": 800, "y": 600, "width": 400, "height": 600, "image_index": 2}
  ],
  "format": "jpeg",              // "jpeg" | "png" | "webp" (default: jpeg)
  "quality": 90
}
```

Header `Idempotency-Key` (optional): repeating the same key within its TTL
returns the original job instead of creating a duplicate composition.

Response: `201 Created` with the job (id, status, layout, sources, and — once
completed — `object_key`, `url`, `width`, `height`, `format`, `size_bytes`).

### `GET /v1/compositions/{id}`

Returns the job's current status and metadata, refreshing the presigned URL
via MinIO/Redis if the cached one has expired.

## Examples (curl)

Assumes the service is running locally on `:8080` (`go run ./cmd`).

**Grid (2x2) from 4 uploaded files:**

```sh
curl -X POST http://localhost:8080/v1/compositions \
  -F 'layout={"type":"grid","cols":2,"rows":2,"width":1200,"height":1200,"spacing":8,"padding":16,"background":"#FFFFFF","cell_fit":"cover","format":"jpeg","quality":90};type=application/json' \
  -F 'images=@photo1.jpg' \
  -F 'images=@photo2.jpg' \
  -F 'images=@photo3.jpg' \
  -F 'images=@photo4.jpg'
```

**Grid (auto near-square) from URLs, with an idempotency key:**

```sh
curl -X POST http://localhost:8080/v1/compositions \
  -H 'Idempotency-Key: 3f1c8e2a-b2e0-4a3d-9c7a-1a2b3c4d5e6f' \
  -F 'layout={"type":"grid","width":900,"height":900,"spacing":4,"format":"png"};type=application/json' \
  -F 'urls=https://example.com/a.jpg' \
  -F 'urls=https://example.com/b.jpg' \
  -F 'urls=https://example.com/c.jpg'
```

**Mosaic (1 hero + 2 smaller cells) mixing an upload with an existing MinIO object:**

```sh
curl -X POST http://localhost:8080/v1/compositions \
  -F 'layout={
        "type": "mosaic",
        "width": 1200, "height": 1200,
        "background": "#000000",
        "cell_fit": "cover",
        "format": "webp",
        "quality": 85,
        "cells": [
          {"x": 0,   "y": 0,   "width": 800, "height": 1200, "image_index": 0},
          {"x": 800, "y": 0,   "width": 400, "height": 600,  "image_index": 1},
          {"x": 800, "y": 600, "width": 400, "height": 600,  "image_index": 2}
        ]
      };type=application/json' \
  -F 'images=@hero.jpg' \
  -F 'urls=https://example.com/side1.jpg' \
  -F 'object_keys=uploads/existing-side2.jpg'
```

Source order is `images`, then `urls`, then `object_keys` — so `image_index`
0/1/2 above map to `hero.jpg` / `side1.jpg` / `uploads/existing-side2.jpg`.

**Freeform (overlapping cells):**

```sh
curl -X POST http://localhost:8080/v1/compositions \
  -F 'layout={
        "type": "freeform",
        "width": 400, "height": 400,
        "background": "#FFFFFF",
        "cell_fit": "cover",
        "format": "png",
        "cells": [
          {"x": 0,   "y": 0,   "width": 250, "height": 250, "image_index": 0},
          {"x": 150, "y": 150, "width": 250, "height": 250, "image_index": 1}
        ]
      };type=application/json' \
  -F 'images=@back.jpg' \
  -F 'images=@front.jpg'
```

**Raw buffer body (single image, no multipart):**

```sh
curl -X POST http://localhost:8080/v1/compositions \
  -H 'Content-Type: image/jpeg' \
  -H 'X-Layout: {"type":"grid","rows":1,"cols":1,"width":800,"height":600,"format":"jpeg"}' \
  --data-binary @photo.jpg
```

**Fetch a composition's metadata/URL:**

```sh
curl http://localhost:8080/v1/compositions/3f1c8e2a-b2e0-4a3d-9c7a-1a2b3c4d5e6f
```

## Configuration

Runtime config comes from **Consul KV**, not a local `.env` file. Only the
coordinates needed to reach Consul and locate the config are read from
`.env`/the environment (`pkg/config/config.go`):

| var | meaning |
|-----|---------|
| `HostConsul` | Consul HTTP API address, e.g. `http://localhost:8500` |
| `KeyConsul` | KV key holding the JSON config blob, e.g. `photo_service` |
| `ServiceConsul` | service name to check in Consul's catalog as a readiness probe |

See [.env.example](.env.example) for the bootstrap file and [consul.json](consul.json)
for the full JSON blob shape (`app`, `http`, `postgres`, `redis`, `minio`)
that must live at `KeyConsul`. Any field omitted from the blob falls back to
a local-dev default (see `Config.applyDefaults` in `pkg/config/config.go`),
so a KV entry only needs to specify what it overrides. Duration fields
(`shutdown_timeout`, `connect_timeout`, `presign_ttl`) are JSON strings like
`"10s"` / `"1h"`, not raw nanoseconds.

`postgres` is deliberately broken into `host`/`port`/`user`/`password`/`dbname`/`sslmode`
fields rather than one DSN string — `PostgresConfig.DSN()` builds the actual
connection string in Go. This matters in Kubernetes specifically: see below.

### In Kubernetes: this config is provisioned automatically from the chart

[consul.json](consul.json) at the service root is checked into
[chart/photo-service/files/consul.json](../chart/photo-service/files/consul.json)
(keep the two in sync) and wired into the deploy through two chart pieces
that already exist for every other service sharing this cluster's Postgres/MinIO:

1. **[chart/photo-service/templates/consul-seed-job.yaml](../chart/photo-service/templates/consul-seed-job.yaml)**
   — a post-install/post-upgrade hook (weight `1`) that `PUT`s
   `files/consul.json` into Consul KV at key `photo_service`, but only if
   that key doesn't exist yet (never clobbers credentials already patched in).
2. **[chart/postgres/templates/service-db-provisioner-job.yaml](../chart/postgres/templates/service-db-provisioner-job.yaml)**
   — runs right after (weight `2`), and because `photo-service` is registered
   in [chart/postgres/values.yaml](../chart/postgres/values.yaml)'s
   `serviceDatabases` list, it creates a dedicated `photo_service` database
   and `photo_service_svc` role, then PATCHes that role's generated
   username/password directly into the `.postgres.user` / `.postgres.password`
   fields of the `photo_service` Consul KV entry the seed job just created.
   This is exactly why `postgres` needed separate `user`/`password` fields
   instead of an opaque `dsn` string — there'd be nothing for this job to patch.

So `consul.json`'s checked-in `postgres.user`/`postgres.password`
(`testUser`/`testPassword`) are only the seed value — the real, per-service
credentials get patched in automatically on first deploy and are never
written back to this repo. The in-cluster hostnames in `consul.json`
(`postgresdb:80`, `redis-svc:80`, `minio-svc:9000`) come from those
services' own Helm Services — see
[chart/postgres/templates/service.yaml](../chart/postgres/templates/service.yaml),
[chart/redis/templates/service.yaml](../chart/redis/templates/service.yaml),
and [chart/minio/templates/service.yaml](../chart/minio/templates/service.yaml).
`photo-service` is also registered in
[chart/minio/values.yaml](../chart/minio/values.yaml)'s `buckets` list (in
addition to this service creating its own bucket at startup — belt and
suspenders, both are idempotent), and the bootstrap `.env` above is supplied
by mounting the cluster-wide `service-env-files` Secret at key
`photo-service.env` (see
[chart/templates/env-files-secret.yaml](../chart/templates/env-files-secret.yaml)
and [chart/photo-service/templates/photo-service.yaml](../chart/photo-service/templates/photo-service.yaml)).

## Running locally

Plain `docker compose` has no Helm chart to seed Consul for you, so that one
step stays manual:

```sh
docker compose up -d      # consul + postgres + redis + minio
cp .env.example .env
consul kv put -http-addr=http://localhost:8500 photo_service @consul.json
go run ./cmd
```

On startup: the MinIO bucket (`minio.bucket`) is created if missing, and
pending SQL migrations under `migrations/` are applied automatically.

## CI/CD

[.github/workflows/photo-service-cicd.yml](../.github/workflows/photo-service-cicd.yml)
runs on every push to `master` that touches `photo-service/**` or
`chart/photo-service/**` (or via manual `workflow_dispatch`), mirroring the
pattern the other services in this monorepo use (e.g. `cdn-service`):

1. **build-test** — `go build ./...` and `go test ./...` inside `photo-service/`.
2. **build-and-push** — multi-arch (`linux/amd64`, `linux/arm64`) Docker image
   built from [Dockerfile](Dockerfile) and pushed to
   `ghcr.io/jieeirosst/photo-service`, tagged with both the short commit SHA
   and `latest`.
3. **deploy** — on a self-hosted runner, `helm template chart/photo-service`
   (overriding the image name/tag from the previous job) piped into
   `kubectl apply`, then waits on the rollout. This renders and applies the
   subchart directly rather than `helm install/upgrade`, since
   `chart/photo-service`'s resources are owned by the monorepo's umbrella
   Helm release rather than deployed standalone.

Everything the deploy needs — the `service-env-files` Secret entry, the
Consul seed job, and the Postgres/MinIO provisioning registrations — is
checked into this repo (see [Configuration](#configuration) above), so a
plain `helm template | kubectl apply` is enough; there's no separate manual
cluster/ops step for this service.

## Notes / design choices

- Composition runs synchronously within the request goroutine after the job
  row is created. All infrastructure is reached exclusively through outbound
  ports, so moving heavy compositions onto a background worker (consuming
  from a Redis-backed queue) only requires changing what invokes
  `ComposeService.process`, not the use case itself.
- `ImageComposer.Compose` takes already-decoded `[]image.Image`, per the
  driven-port contract; a separate `ImageDecoder` port (implemented by the
  same imaging adapter) turns raw upload/URL/MinIO bytes into `image.Image`
  before composition, keeping codec libraries out of the application layer.
- Output supports JPEG, PNG, and WebP. WebP encode/decode uses
  [gen2brain/webp](https://github.com/gen2brain/webp), a CGo-free encoder
  (libwebp compiled to WASM, transpiled to Go) — no system `libwebp` needs to
  be installed on the build or runtime host.
- Idempotency is tracked in Redis only (`idempotency:<key> -> job id`, TTL
  24h), per "hỗ trợ idempotency theo request key" — `JobRepository` stays a
  minimal `Save`/`FindByID` pair rather than growing a query-by-idempotency-key
  method. The DB retains a unique index on `idempotency_key` as a secondary
  safety net, but it is not queried by the application.
- The URL fetcher refuses to dial loopback/private/link-local addresses to
  prevent SSRF via attacker-supplied source URLs.

## Tests

```sh
go test ./...
```

- `internal/application/service`: `ComposeService` covered with hand-written
  mocks for every outbound port (`JobRepository`, `ImageStorage`,
  `CacheRepository`, `ImageComposer`, `ImageDecoder`, `ImageFetcher`, `Logger`).
- `internal/adapters/outbound/imaging`: the real `Composer` exercised against
  all three layouts (grid/mosaic/freeform, including overlapping freeform
  cells) using images generated with `image.NewRGBA`, asserting output
  dimensions, per-cell pixel placement, and `ErrSourceCellMismatch` on
  too-few/too-many sources.
