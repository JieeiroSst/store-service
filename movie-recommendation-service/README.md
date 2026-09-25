# movie-recommendation-service

Recommendation engine for [video-service](../video-service). It answers three questions:

| Endpoint | Meaning |
|---|---|
| `GET /api/recommendations/users/:user_id?page=&page_size=` | "For you" - personalized, excludes videos the user already interacted with |
| `GET /api/recommendations/videos/:id/similar?page=&page_size=` | "Related" - shown next to a video |
| `GET /api/recommendations/trending?page=&page_size=` | Popular right now; also the cold-start fallback |
| `GET /api/recommendations/users/:user_id/home` | Ready-made home rows: continue watching, recommended for you, up to two "because you watched X", new releases, trending. No video appears twice; empty rows are omitted |
| `GET /api/recommendations/users/:user_id/continue-watching` | Videos the user is partway through (watched 3%-90%), newest first, with `progress` |
| `GET /api/recommendations/users/:user_id/history?page=&page_size=` | Distinct videos the user interacted with, newest first |
| `DELETE /api/recommendations/users/:user_id/history/:video_id` | Forget a video: deletes the user's signals for it so it stops shaping their recommendations |
| `GET /api/recommendations/new?page=&page_size=` | Newest videos (paginated with `snapshot`, like the other rankings) |
| `POST /api/recommendations/events` | Record a signal: `{"user_id","video_id","type","value"}` |
| `PUT /api/recommendations/videos/:id` / `DELETE` | Register/remove a video by hand (video-service sync normally does this) |
| `GET /healthz`, `GET /readyz` | Liveness / readiness (readiness pings Postgres) |

List endpoints are paginated: `page` (1-based, default 1) and `page_size` (default 20, max 50). The response is `{"items":[...],"total":N,"page":P,"page_size":S,"snapshot":"..."}`; `total` is the size of the ranked list, capped at the top 500 candidates. To keep pages consistent while a client scrolls, a ranking that spans more than one page comes back with an opaque `snapshot` token. Send it as `?snapshot=` on the next pages and you read the exact ranking page 1 was cut from: the ranked list is stored in a store shared by all replicas (TTL `SNAPSHOT_TTL_MINUTES`, default 15): Postgres (`ranking_snapshots`) by default, or Redis when `REDIS_ADDR` is set, so retrains, new likes, and a different replica answering all leave it unchanged. No session affinity is needed. If the token is unknown, expired or was issued for a different list, the request is served from a fresh ranking and returns a *different* `snapshot` (restart from page 1). Single-page results have an empty `snapshot`. Malformed tokens return 400. Redis expires snapshots itself; with Postgres the background loop purges them. Each multi-page feed load costs one write (a JSON list of up to 500 ids, tens of KB), so at high traffic set `REDIS_ADDR`.

Every recommendation carries `reason` (`because_you_watched`, `watched_together`, `similar_content`, `for_you`, `trending`) and, where relevant, the `because` video, so the UI can render "Because you watched X".

## Event types

| type | value | weight |
|---|---|---|
| `view` | - | +1 |
| `watch` | fraction watched, 0..1 (send the current position; the latest one drives continue-watching) | +3 x value |
| `like` | - | +4 |
| `dislike` | - | -4 |
| `rating` | 1..5 | (value - 3) x 2 |

## How it works

video-service has no user identity, so the front end (or gateway) posts events to this service with the user id it already knows.

```
video-service --GET /api/videos--> [sync]--> Postgres <--events-- front end / gateway
                                               |
                                            [train] every TRAIN_INTERVAL
                                               v
                                   immutable in-memory Model (atomic swap)
                                               |
                          /users/:id  /videos/:id/similar  /trending
```

Serving never touches Postgres except to load the requesting user's own history (so new likes affect their next request immediately). The model is rebuilt in the background and swapped atomically, so reads never wait on training. Each replica keeps its own model; there is no coordination between replicas.

Scoring (`internal/domain/engine`):

- **Item-item collaborative filtering** - cosine similarity from user co-occurrence, damped for heavy users, top 50 neighbours per item.
- **Content similarity** - TF-IDF over title (x2), tags (x3), description.
- **Trending** - interactions decayed by half-life (default 7 days) plus a small log(views) term from video-service.
- **Freshness** - a video's bonus is 2^(-age / half-life) (default 14 days), so new uploads with no interactions yet still get a chance (cold start for items). `new` ranks 80% freshness + 20% popularity.
- **Personalized** = 0.50 CF + 0.28 content (against a profile built from the user's history; dislikes subtract) + 0.12 popularity + 0.10 freshness. Videos with no positive history fall back to trending.
- **Diversity** - the top 100 candidates of `for you` and `similar` are re-ordered greedily (MMR): each pick's score is reduced by `DIVERSITY` (default 0.25) times its content similarity to what is already picked, so near-duplicates are spread out. `DIVERSITY=0` turns it off. Scores stay as computed, so the list is not strictly descending by score.
- **Similar** = 0.6 co-viewing + 0.4 content.

Limits to know about: the model holds the whole catalog in memory and content similarity scans it per request, which is fine for tens of thousands of videos; beyond that, precompute neighbours or move to an ANN index.

## Architecture (hexagonal + uber-go/fx)

```
cmd/main.go                          fx.New(infrastructure.Module).Run()
config/                              env-based config
internal/
  domain/
    model/                           Video, Interaction, Recommendation
    engine/                          pure algorithm, no I/O (unit tested)
    port/driving.go                  RecommendationUsecase, ModelMaintainer   (inbound ports)
    port/driven.go                   VideoRepository, InteractionRepository,
                                     CatalogSource, HealthChecker              (outbound ports)
  application/                       Service implements the driving ports using the driven ones
  adapter/
    primary/http/                    gin handlers + routes (drives the use case)
    secondary/postgres/              repositories + migrations (driven)
    secondary/videoclient/           video-service HTTP client (driven)
  infrastructure/
    module.go                        the fx graph
    server/                          HTTP server lifecycle
    jobs/                            background sync + retrain loops
```

Dependencies point inward: adapters import `domain`, `domain` imports nothing from outside. Each adapter package exposes an fx `Module` that binds its concrete type to the port with `fx.As`.

## Configuration

See [.env.example](.env.example). Postgres database and schema are created on startup (advisory-locked, safe with several replicas).

## Run / test

```
make run
make test
```
