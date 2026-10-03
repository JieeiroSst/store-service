# webrtc-service

WebRTC signaling server for small mesh video calls. Browsers join a room over
a WebSocket; the service relays SDP offers/answers and ICE candidates, room
chat and mute/screen-share state, and hands out STUN/TURN servers. Media flows
peer-to-peer (or through TURN) and never passes through this service.

A full call client is served at `/` (share `/?room=<id>` to invite people).

## Features

| Area | What it does |
|------|--------------|
| Mesh calls | The newcomer offers to every member; perfect negotiation resolves offer collisions; ICE candidates that arrive early are queued |
| NAT traversal | `GET /api/ice-servers` returns STUN plus short-lived TURN credentials (coturn `use-auth-secret`), so the TURN secret never reaches browsers |
| Resilience | ICE restart when a connection fails; the WebSocket reconnects with backoff (e.g. pod rollout) and rebuilds the mesh; a second tab with the same user ID replaces the first, which is told not to reconnect |
| Room limits | `MAX_PEERS_PER_ROOM`: a full room rejects joins with a `room_full` error |
| Lobby | Camera/mic preview with mute toggles and device pickers before joining; the same stream is reused in the call |
| Layout | Adaptive grid; spotlight + filmstrip when someone presents or a tile is pinned; active-speaker ring from audio levels; avatars when the camera is off; invite card when alone |
| Media controls | Mute / camera off, screen sharing (`replaceTrack`, no renegotiation), camera and microphone switching mid-call |
| Presence | `media_state` is broadcast and replayed to newcomers, so everyone sees who is muted or sharing; `GET /api/rooms/{room_id}` lists participants |
| Chat | Room-wide or private messages, timestamped by the server |
| Diagnostics | Per-peer stats overlay: RTT, in/out bitrate, packet loss, resolution/fps, and whether the path is host, srflx or relay (TURN) |
| Recording | Records the call in the browser (tiles composited on a canvas, all audio mixed) and downloads a WebM; frames are timed by a Worker so recording continues while the tab is in the background |
| Metrics | Prometheus `/metrics`: `webrtc_rooms`, `webrtc_peers`, `webrtc_messages_relayed_total{type}`, `webrtc_joins_rejected_total{reason}` |

## Architecture

Hexagonal (ports & adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                          fx.New(infrastructure.Module)
config/                              env config
web/                                 embedded React call client (index.html + static/vendor)
internal/
  domain/                            messages, media state, chat, room info, ICE servers, errors
  port/                              Peer, SignalingService + ICEService (inbound),
                                     RoomRegistry + Metrics (outbound)
  application/                       SignalingService (join/leave/relay/chat/media state),
                                     ICEService (TURN REST credentials)
  adapter/primary/http/              HTTP routes, WebSocket connection (implements port.Peer)
  adapter/secondary/memory/          in-memory RoomRegistry
  adapter/secondary/metrics/         Prometheus Metrics
  infrastructure/                    fx module + HTTP server lifecycle
```

## API

| Method | Path | Description |
|--------|------|-------------|
| GET | `/` | call client |
| GET | `/health` | liveness/readiness |
| GET | `/metrics` | Prometheus metrics |
| GET | `/api/ice-servers?user_id=..` | `{"ice_servers": [RTCIceServer...]}`, not cacheable |
| GET | `/api/rooms/{room_id}` | `{room_id, capacity, participants: [{user_id, media}]}` |
| GET | `/ws?user_id=..&room_id=..` | signaling WebSocket (IDs: 1-64 printable chars) |

### WebSocket messages

JSON `{type, room_id, user_id, to_user_id, data}`. The server always sets
`user_id`/`room_id` to the sender's own, whatever the client sends.

| Type | Direction | `to_user_id` | `data` |
|------|-----------|--------------|--------|
| `offer`, `answer` | client → one peer | required | RTCSessionDescription |
| `ice_candidate` | client → one peer | required | RTCIceCandidate |
| `chat` | client → room or one peer | optional (private) | `{text}` in, `{text, sent_at}` out; 1-2000 chars |
| `media_state` | client → room | - | `{audio, video, screen}` |
| `room_users` | server → newcomer | - | other user IDs, then one `media_state` per member |
| `user_joined`, `user_left` | server → room | - | - |
| `error` | server → client | - | `{code, message}`; codes `invalid_id`, `invalid_message`, `peer_not_found`, `peer_busy`, `room_full`, `replaced` |

After `room_full`, `replaced` or `invalid_id` the server closes the socket and
clients must not reconnect automatically.

## Frontend

`web/index.html` is a React 18 app with no build step: components are written
with [htm](https://github.com/developit/htm) tagged templates (JSX-like syntax
that runs directly in the browser), and React, ReactDOM and htm are vendored
in `web/static/vendor` and embedded in the binary, so no CDN or Node toolchain
is needed at build or run time.

- `CallEngine` owns the WebSocket, the `RTCPeerConnection`s and the media
  tracks, and publishes an immutable snapshot of the call.
- React components (`Lobby`/`PreviewTile`, `Room` with `TopBar`, `Stage`/
  `VideoTile`, `SidePanel` (`Chat`, `People`) and `ControlBar`) read that
  snapshot with `useSyncExternalStore` and call engine methods on user actions.

To upgrade a vendored library, replace the file in `web/static/vendor` (the
version is part of the file name, which `/static/` serves as immutable).

## Configuration

| Env | Default | Description |
|-----|---------|-------------|
| `PORT_HTTP_SERVER` | `8080` | listen port |
| `ALLOWED_ORIGINS` | empty | comma-separated origins for `/ws`; empty = same origin, `*` = any |
| `MAX_PEERS_PER_ROOM` | `8` | mesh room size limit |
| `STUN_URLS` | `stun:stun.l.google.com:19302` | comma-separated; set empty to disable |
| `TURN_URLS` | empty | e.g. `turn:turn.example.com:3478?transport=udp,turns:turn.example.com:5349` |
| `TURN_SECRET` | empty | coturn `static-auth-secret`; TURN is offered only when this and `TURN_URLS` are set |
| `TURN_TTL_SECONDS` | `86400` | lifetime of issued TURN credentials |
| `WS_PONG_WAIT_SECONDS` | `60` | drop a client silent for this long (pings every 90% of it) |
| `WS_MAX_MESSAGE_KB` | `64` | max inbound message size |
| `WS_SEND_BUFFER` | `256` | queued messages per client before it is disconnected as too slow |

A matching coturn config:

```
use-auth-secret
static-auth-secret=<same as TURN_SECRET>
realm=turn.example.com
```

In the chart, put the secret in a Secret and set `webrtcService.turnSecretRef.name`.

## Scaling

Rooms are held in each pod's memory, so all members of a room must reach the
same pod. The chart routes by `room_id` with ingress-nginx
`upstream-hash-by` and keeps a fixed replica count (no HPA). A rollout drops
the signaling sockets; clients reconnect and rebuild their calls on their own.
To scale freely, replace `adapter/secondary/memory` with a shared registry
plus cross-pod fan-out (e.g. Redis pub/sub) behind the same ports.

Mesh calls send every stream to every peer, so they suit rooms of up to
roughly 6-8 people; larger rooms need an SFU (e.g. built on pion).

## Run

```sh
go run ./cmd            # http://localhost:8080
go test -race ./...
docker build -t webrtc-service .
```
