# video-service

YouTube-style video service. Uploads land in MinIO, a worker transcodes them to adaptive-bitrate HLS, and viewers stream from a horizontally scalable, buffered API.

- `BE/` – Go backend (hexagonal architecture, [uber-go/fx](https://github.com/uber-go/fx)). One binary, two roles: `api` (default) and `ROLE=worker`.
- `FE/` – React app: home grid, search, trending, watch page with quality menu and related videos, upload.

```
            ┌────────── browser (hls.js) ──────────┐
            │  /            /api/videos/:id/hls/…  │
            ▼                       ▼              │
       video-web (nginx)     video-service (API xN, HPA) ── Redis ── video-worker (ffmpeg xN)
                                    │  chunk buffer                        │
                                    └──────────────── MinIO ◄──────────────┘
                                        videos/  hls/<id>/  thumbs/  meta/
```

## Upload → watch

1. `POST /api/videos` streams the file into MinIO (`videos/<id>`), saves metadata (`status: processing`) and enqueues a job in Redis.
2. The original is playable immediately over range requests (the watch page uses it while processing).
3. A worker downloads it, runs ffmpeg once to produce 1080p/720p/480p/360p renditions (never upscaling) as 4-second HLS segments plus a thumbnail, uploads them under `hls/<id>/` and `thumbs/`, and flips the status to `ready`. Redis Streams give at-least-once delivery: a job whose worker dies is reclaimed by another worker; an undecodable file is marked `failed` (not retried) and hidden from lists.
4. The player loads `master.m3u8` (hls.js; native on Safari) and switches quality per segment based on bandwidth; users can also pick a quality manually.

DASH is not produced: HLS plays everywhere (hls.js on desktop/Android, native on Apple).

## API

| Method | Path | |
|---|---|---|
| GET | `/api/videos?q=&sort=newest\|popular&page=1&page_size=20` | list/search (failed videos hidden) |
| POST | `/api/videos` | multipart: `title`, `description`, **then** `file` (`video/*`) |
| GET | `/api/videos/:id` | metadata incl. `status`, `duration`, `views` |
| GET | `/api/videos/:id/related?limit=` | other ready videos, most viewed first |
| POST | `/api/videos/:id/view` | count a view (once per client IP per 30 min) |
| GET | `/api/videos/:id/thumbnail` | JPEG, once ready |
| GET | `/api/videos/:id/hls/master.m3u8` (and `v<N>/…`) | adaptive stream, once ready |
| GET/HEAD | `/api/videos/:id/stream` | the original file, supports `Range` |
| DELETE | `/api/videos/:id` | removes original, renditions, thumbnail, views |
| GET | `/healthz` | |

## How it scales

Stateless API pods behind the ingress, scaled by the HPA (`chart/video-service`); workers scale independently. Per pod:

1. **Chunk buffer** (`adapter/secondary/buffer`) – reads go through fixed-size chunks (default 1 MiB) in a byte-bounded, sharded LRU (`CACHE_SIZE_MB`). A chunk is fetched from MinIO once, then served from memory to every viewer; concurrent misses collapse into one MinIO request (singleflight); the next `PREFETCH_CHUNKS` chunks are warmed in the background. HLS segments, playlists and thumbnails use the same cache.
2. **Metadata cache** – video metadata and the list are cached in memory. Every write publishes the id on Redis pub/sub and **every replica invalidates immediately**, so a new upload or a finished transcode shows up everywhere at once. The TTLs (`META_TTL_SECONDS`, `LIST_TTL_SECONDS`) are only a safety net if a message is missed.
3. **Cacheable responses** – HLS files, thumbnails and the original carry `Cache-Control: immutable` (+ `ETag` for the original), so a CDN in front absorbs most traffic. For very large audiences put one in front: caches and HPA alone still funnel every miss to MinIO and every byte through your pods' network.

Views are counted in one Redis hash shared by all replicas. `sort=popular` reads the counts for every matching video, which is fine for thousands of videos; for millions, keep a sorted set instead.

## Layout (BE)

```
cmd/main.go                          ROLE=worker → WorkerModule, else APIModule
config/                              env config
internal/domain/model                Video, Status, Asset
internal/domain/port                 driving: VideoUsecase, TranscodeUsecase
                                     driven:  VideoStorage, VideoRepository, ViewCounter,
                                              JobQueue, InvalidationBus, Transcoder
internal/application                 VideoService (API), TranscodeService (worker)
internal/adapter/primary/http        gin handlers + routes
internal/adapter/secondary/minio     storage + metadata repository (meta/<id>.json)
internal/adapter/secondary/redis     views + transcode queue (Streams) + invalidation pub/sub
internal/adapter/secondary/ffmpeg    ffprobe/ffmpeg transcoder
internal/adapter/secondary/buffer    chunk buffer + metadata cache (fx.Decorate over the adapters)
internal/infrastructure              fx wiring, http server + worker lifecycles
```

## FE layout

```
src/api/            fetch/XHR client
src/hooks/          useInfiniteVideos, useOnVisible, useTheme
src/components/     Header, Sidebar, VideoCard/CompactCard, Player (hls.js + quality menu + shortcuts), UploadModal
src/pages/          HomePage, WatchPage
src/styles/theme.css  dark/light tokens
```

Watch-page shortcuts: `space`/`k` play-pause, `j`/`l` ±10 s, `←`/`→` ±5 s, `m` mute, `f` fullscreen. Playback position is remembered per video.

## Config (env)

| Var | Default |
|---|---|
| `ROLE` | `api` (or `worker`) |
| `PORT_HTTP_SERVER` | `8080` |
| `MINIO_ENDPOINT` / `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | `localhost:9000` / `minioadmin` / `minioadmin123` |
| `MINIO_BUCKET` / `MINIO_USE_SSL` | `video-service` / `false` |
| `REDIS_ADDR` / `REDIS_PASSWORD` | `localhost:6379` / – |
| `ALLOWED_ORIGINS` | `*` (comma-separated) |
| `MAX_UPLOAD_MB` | `2048` |
| `CHUNK_SIZE_KB` / `CACHE_SIZE_MB` / `PREFETCH_CHUNKS` | `1024` / `512` / `2` |
| `META_TTL_SECONDS` / `LIST_TTL_SECONDS` | `300` / `30` |
| `WORKER_CONCURRENCY` / `TRANSCODE_TIMEOUT_MINUTES` | `1` / `60` |

## Run locally

```sh
docker run -d -p 9000:9000 -e MINIO_ROOT_USER=minioadmin -e MINIO_ROOT_PASSWORD=minioadmin123 minio/minio server /data
docker run -d -p 6379:6379 redis:7-alpine

cd BE && make run                  # API on :8080
cd BE && ROLE=worker make run      # worker (needs ffmpeg + ffprobe on PATH)
cd FE && npm install && npm start  # :3000, proxies /api to :8080
```

Tests: `cd BE && make test` (ffmpeg isn't needed; the worker use case is tested with a fake transcoder).

## Deploy

- API + worker image: `BE/Dockerfile` (includes ffmpeg) → `ghcr.io/jieeirosst/video-service`, `.github/workflows/video-service-cicd.yml`, chart `chart/video-service` (API Deployment + HPA + Ingress at `/api`, and the `video-worker` Deployment).
- Web image: `FE/Dockerfile` (nginx) → `ghcr.io/jieeirosst/video-web`, `.github/workflows/video-web-cicd.yml`, chart `chart/video-web` (Ingress at `/`, same host as the API so no CORS).
- Both charts are dependencies of the umbrella `chart/`. They use the `video-service` bucket of `chart/minio` and `redis-svc` from `chart/redis`.
