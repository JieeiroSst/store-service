# livestream_service

A self-built livestreaming backend: RTMP ingest → ABR transcode → HLS → CDN/object storage, plus the application layer around it (rooms, stream keys, realtime chat, viewer counts, VOD metadata) and a Redis-backed multi-node scheduler.

Two independently-scalable roles, both built from this same module (see "Two roles" below):

```
                              ── node role (StatefulSet, 1 SRS+ffmpeg per pod) ──
Streamer (OBS) ── RTMP ──▶ SRS (sidecar) ── localhost ──▶ ffmpeg (ABR/HLS)
                              │ on_publish/on_unpublish           │
                              ▼                          local /var/hls scratch
                     livestream-service-node                      │
                          │        │                     fsnotify watcher
                    Publish-     Node                              │
                    Usecase    Scheduler                           ▼
                          │        │                    MinIO/S3 (segments, playlists,
                          ▼        ▼                     correct Cache-Control headers)
                       Postgres  Redis (heartbeat/                 │
                      (rooms,    capacity/assignment)               ▼
                       streams,                                CDN ──▶ viewers (hls.js)
                       vod)
                              ── edge role (stateless Deployment, N pods) ──
viewer/streamer client ── HTTP/WS ──▶ livestream-service-edge
                                        │      │        │        │
                                   RoomUsecase │   ViewerUsecase ChatUsecase
                                        │   IngestUsecase   │        │
                                        ▼      │            ▼        ▼
                                     Postgres  ▼         Redis    Redis pub/sub
                                            Redis (node   (viewer  (one subscription
                                            assignment)   heartbeat per room per pod,
                                                           sliding   fanned out to all
                                                           window)   local WS conns)
```

## Two roles

The transcode-node role (SRS + ffmpeg + node scheduler) and the viewer-facing role (rooms/chat/viewer reads) scale on completely different axes: node capacity is bounded by CPU-per-active-stream and the node pool's core count, while edge traffic is stateless and bounded by connection count. Splitting them means viewer/chat capacity can scale (via HPA) without paying for idle SRS+ffmpeg sidecars, and vice versa - this is the main lever for handling a large number of concurrent viewers without over-provisioning transcode nodes.

Both roles are built from the same Go module and Docker image (`livestream_service/Dockerfile` builds all three binaries); only the fx wiring and entrypoint differ:

| Binary | fx module | Chart workload | Depends on |
|---|---|---|---|
| `livestream-service-node` | `infrastructure.NodeModule` | StatefulSet (`chart/livestream-service/templates/livestream-service.yaml`) - needs a headless Service for stable per-pod RTMP addressing | `transcode.Module` (ffmpeg), `schedulerAdapter.Module` (heartbeat/watchdog) |
| `livestream-service-edge` | `infrastructure.EdgeModule` | Deployment + HPA (`chart/livestream-service/templates/edge-deployment.yaml`) | Nothing ffmpeg/SRS-related - fx never constructs `TranscodeRunner` on this role |
| `livestream-service` | `infrastructure.Module` | Not used by the chart - monolithic dev/docker-compose convenience, mounts both roles' routes on one port | Everything (all-in-one) |

`port.StreamLifecycleUsecase` doesn't exist as a single interface - it's split into `port.IngestUsecase` (edge: `RequestIngestEndpoint`, `GetActiveStream`, `ListRecordings`, no `TranscodeRunner` dependency) and `port.PublishUsecase` (node-only: `HandleOnPublish`, `HandleOnUnpublish`, needs `TranscodeRunner`). Both share node-assignment logic via the unexported `internal/application/node_assignment.go:nodeAssigner`.

## Architecture

Hexagonal (ports & adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
config/                          env-var config, with a Consul KV override (see below)
migrations/                      hand-applied numbered SQL (rooms, streams, vod_recordings)
cmd/                              all-in-one dev entrypoint
cmd/node/                         node role entrypoint
cmd/edge/                         edge role entrypoint
internal/
  domain/
    model/                       Room, Stream, Recording, TranscodeNode, ChatMessage
    port/
      driving.go                 RoomUsecase, IngestUsecase, PublishUsecase, ViewerUsecase,
                                  ChatUsecase, NodeSchedulerUsecase (primary ports)
      driven.go                  *Repository, NodeRegistry, ViewerCounter, ChatBroadcaster,
                                  ObjectStorage, TranscodeRunner (secondary ports)
  application/
    room.go                      create rooms, (re)generate stream keys
    ingest.go                    RequestIngestEndpoint, GetActiveStream, ListRecordings (edge)
    publish.go                   HandleOnPublish/HandleOnUnpublish, VOD finalization (node)
    node_assignment.go           shared node-assignment logic used by both ingest and scheduler
    viewer.go                    player heartbeat -> Redis sliding-window viewer count
    chat.go                      thin pass-through to the Redis-backed ChatBroadcaster
    scheduler.go                 node heartbeat, stale-job watchdog policy (node only)
  adapter/
    primary/                     drives the application
      http/                      NewNodeRouter (SRS hooks only), NewEdgeRouter (rooms/ingest/
                                  viewers/chat), NewAllInOneRouter (dev); WebSocket chat has
                                  ping/pong keepalive against zombie connections
      scheduler/                 cron loops: Heartbeat (push capacity to Redis),
                                  Watchdog (restart stale ffmpeg jobs) - node only
    secondary/                   driven by the application
      repository/                Postgres-backed Room/Stream/VOD repositories (gorm)
      redisstore/                NodeRegistry, ViewerCounter (sliding window), ChatBroadcaster
                                  (one Redis subscription per room per pod, fanned out
                                  in-process to local WS connections - see chat_broadcaster.go)
      storage/                   MinIO/S3 ObjectStorage
      transcode/                 ffmpeg argv builder, job manager (spawn/respawn/backoff,
                                  a Go port of a Node.js "hook-server" design), fsnotify-based
                                  segment uploader
  infrastructure/                Module (all-in-one), NodeModule, EdgeModule; config/logger
                                  bootstrap; lifecycle-managed HTTP + scheduler servers
```

## Setup

```bash
for f in migrations/*.sql; do psql "$DATABASE_URL" -f "$f"; done
go run ./cmd        # all-in-one, for local dev
# or, matching production:
go run ./cmd/node    # requires a local SRS instance forwarding to NODE_LOCAL_RTMP
go run ./cmd/edge
```

Requires `ffmpeg` on `PATH` (or set `FFMPEG_PATH`) for the node role, a reachable Postgres, Redis, and an S3-compatible bucket (MinIO or AWS S3) for both roles. The node role is meant to run alongside an SRS sidecar (see `chart/livestream-service`) - without SRS forwarding a real RTMP stream to `NODE_LOCAL_RTMP`, there's nothing for ffmpeg to read.

Schema changes live in `migrations/` as plain numbered SQL files, applied by hand — no migration runner or `AutoMigrate`, same as the rest of this repo's services.

### Configuration

Config resolves from **Consul KV** first (bootstrapped via `HostConsul`/`KeyConsul`/`ServiceConsul` env vars, JSON shape in `consul.json`), falling back to plain env vars if Consul is unset or unreachable — see `internal/infrastructure/config_provider.go`. The Helm chart (`chart/livestream-service`) doesn't set the Consul bootstrap vars, so in that deployment config always comes from the plain env vars below (`consul.json` is a reference shape for standalone/Consul-managed deployments, not something the chart consumes).

| Env var | Default | Used by | Description |
|---|---|---|---|
| `PORT_HTTP_SERVER` | `8080` | both | REST API / SRS webhook receiver / chat WebSocket |
| `POSTGRES_*` | `localhost:5432/livestream` | both | Room/stream/VOD metadata |
| `REDIS_ADDR`, `REDIS_PASSWORD`, `REDIS_DB` | `localhost:6379`, ``, `0` | both | Node scheduler, viewer counters, chat pub/sub |
| `S3_ENDPOINT`, `S3_BUCKET`, `S3_ACCESS_KEY`, `S3_SECRET_KEY`, `S3_USE_SSL` | `localhost:9000`, `livestream-hls`, ``, ``, `false` | node | HLS segment/playlist/VOD object storage |
| `NODE_ID` | hostname | node | This pod's identity in the Redis node registry - the chart sets it to the pod name |
| `NODE_RTMP_ADDR` | `` | node | This node's externally-addressable RTMP URL, handed out by the scheduler - the chart derives it from the StatefulSet's stable pod DNS |
| `NODE_LOCAL_RTMP` | `rtmp://127.0.0.1:1935/live` | node | Where ffmpeg reads from - the SRS sidecar in the same pod |
| `MAX_STREAMS` | `20` | node | Concurrent transcode jobs this node accepts before rejecting new publishes |
| `HLS_DIR` | `/var/hls` | node | Local scratch dir for segments/playlists before upload (see the segment uploader) |
| `FFMPEG_PATH` | `ffmpeg` | node | Binary invoked to transcode |
| `HLS_SEGMENT_TIME`, `HLS_LIST_SIZE` | `4`, `6` | node | HLS segment duration (seconds) and live playlist window |
| `TRANSCODE_MAX_RESTARTS`, `TRANSCODE_RESTART_WINDOW`, `TRANSCODE_RESTART_DELAY` | `5`, `60s`, `2s` | node | ffmpeg crash-loop backoff: at most N restarts per window, with a delay between attempts |
| `WATCHDOG_STALE_AFTER`, `WATCHDOG_CHECK_INTERVAL` | `30s`, `30s` | node | A job producing no new segments for this long is considered "clinically dead" and restarted; checked on this interval |
| `HEARTBEAT_INTERVAL`, `HEARTBEAT_TTL` | `5s`, `15s` | node | How often this node reports capacity to Redis, and how long that report is trusted before the node is considered dead |
| `VIEWER_HEARTBEAT_WINDOW` | `40s` | edge | Sliding-window size for the online-viewer count - should be roughly 2.5x the player's heartbeat interval (~15s recommended) |

## API

### Health

```
GET /health
```

### Rooms (`/api/v1/rooms`) - edge role

```
POST   /api/v1/rooms                            {ownerUserId, title, description} -> model.Room (includes streamKey)
GET    /api/v1/rooms?live=true                   list rooms, optionally filtered to live only
GET    /api/v1/rooms/:id                         get one room
POST   /api/v1/rooms/:id/stream-key/regenerate    -> {streamKey}
POST   /api/v1/rooms/:id/ingest                  assign a transcode node -> model.IngestEndpoint {rtmpURL, nodeId, streamKey}
GET    /api/v1/rooms/:id/stream                  the room's active model.Stream, 404 if offline
GET    /api/v1/rooms/:id/recordings              VOD recordings for the room
GET    /api/v1/rooms/:id/viewers                 {viewers: <count>}
POST   /api/v1/rooms/:id/viewers/heartbeat        {sessionId} -> 204
```

Stream keys are retrievable (not hashed), the same model Twitch/YouTube use for stream keys shown in account settings — a room's owner is expected to paste it into OBS as `<ingest rtmpURL>/<streamKey>`.

**Viewer counting is heartbeat-based, not SRS-based.** Viewers pull HLS from a CDN/object storage, never connecting to SRS directly - so SRS's `on_play`/`on_stop` hooks would never fire for a real viewer. Instead, the HLS player (hls.js) is expected to call `POST .../viewers/heartbeat` with a client-generated random `sessionId` roughly every 15s; the count is a Redis sorted-set sliding window (`ViewerCounter`) that self-heals when a player stops heartbeating (crashed tab, closed window) - no explicit "leave" call needed.

### Chat - edge role

```
GET /ws/rooms/:id/chat?userId=...&username=...   WebSocket upgrade
```
Send `{"body": "..."}` frames; receive `model.ChatMessage` frames broadcast to everyone in the room. Fanned out via Redis pub/sub with exactly one subscription per (pod, room) - see `chat_broadcaster.go` - not one per connection, so a busy room with thousands of local viewers on one pod doesn't open thousands of redundant Redis subscriptions. The connection is kept alive with a 30s ping / 60s pong-timeout so dead clients (network drop, killed tab) don't leak goroutines and hub registrations indefinitely.

### SRS webhooks (`/api/srs`) - node role only

Not meant to be called directly — this is SRS's `http_hooks` contract (see `chart/livestream-service/templates/configmap-srs.yaml`): HTTP 200 with `{"code": 0}` to allow, `{"code": 1}` to reject. Only mounted on the node role's router, reached exclusively by that node's own local SRS sidecar over `127.0.0.1`.

```
POST /api/srs/on_publish     validates the stream key, claims/confirms the node assignment, starts the ffmpeg job
POST /api/srs/on_unpublish   stops the ffmpeg job, closes the stream, records a VOD entry
```

## Development

```bash
go build ./...
go vet ./...
gofmt -l .
go test ./...
```

Redis-backed adapters (`internal/adapter/secondary/redisstore`) are tested against an in-memory Redis (`github.com/alicebob/miniredis/v2`), not a live server.

## Known limitations

- No auth of its own: `ownerUserId` on room creation is trusted as a plain string from whoever calls the API — enforce identity upstream if that matters.
- External RTMP ingest (a streamer pushing from outside the cluster to a *specific* assigned node) needs a TCP LoadBalancer/NodePort per replica or an L4 proxy that understands the scheduler's per-node routing — not solved by `chart/livestream-service` itself, see its `rtmpService` comments.
- Redis is a single logical instance for the node scheduler, viewer counts, and chat fan-out; no Cluster/Sentinel wiring is included. This is the next lever for HA/throughput beyond a single instance once one becomes a bottleneck or SPOF risk.
- Per-stream resource isolation relies on the pod's container-level `resources.limits` (via the chart) plus the per-pod `MAX_STREAMS` cap, not per-process cgroup limits — a deliberate simplification for running under Kubernetes rather than bare-metal systemd.
- The node role's HPA (disabled by default) scales on CPU%, a coarse proxy for transcode load; a custom metric off each node's reported Redis capacity would scale more precisely if Prometheus/KEDA are added later.
- VOD recordings only track the HLS master playlist's object key and duration; there's no automatic MP4 remux or thumbnail generation.
