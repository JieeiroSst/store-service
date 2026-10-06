# serpapi-service

A gateway to every [SerpApi](https://serpapi.com) API: all 125 search engines
(Google, Bing, Baidu, Yahoo, Yandex, DuckDuckGo, Brave, eBay, Walmart, Home
Depot, Amazon, YouTube, Apple, Naver, Yelp, Tripadvisor, Airbnb, Zillow, ...),
and every non-search API: Search Archive, Account, Locations, Image (upload)
and Pixel Position.

| SerpApi API | SerpApi endpoint | API key |
|---|---|---|
| Search (125 engines, json / html / md) | `GET /search` | required |
| Pixel Position (google, google_ads) | `GET /search?output=json_with_pixel_position` | required |
| Search Archive | `GET /searches/{id}.{json,html,json_with_pixel_position}` | required |
| Account | `GET /account.json` | required |
| Image upload | `POST /image` (multipart, `api_key` form field) | required |
| Locations | `GET /locations.json` | not used |

The SerpApi key stays in this service; clients never see it. On top of SerpApi
the service adds parameter validation per engine, a shared result cache, a
concurrency cap, batch searches, access tokens and Prometheus metrics.

## Architecture

Hexagonal (ports and adapters), wired with [uber-go/fx](https://github.com/uber-go/fx).

```
cmd/main.go                          fx.New(infrastructure.Module)
config/                              env configuration
internal/
  domain/                            Engine catalog + validation, SearchRequest, Account, Location, errors
  port/                              outbound ports: SerpAPI, Cache, Metrics
  application/                       use cases: SearchService (search, batch, archive), AccountService, Limiter
  adapter/
    primary/http/                    REST API (driving adapter)
    secondary/serpapi/               serpapi.com HTTP client (retries 5xx, redacts the key from errors)
    secondary/cache/                 memory (LRU + TTL) | redis | none
    secondary/metrics/               Prometheus
  infrastructure/                    fx module graph + HTTP server lifecycle
```

The application layer depends only on `domain` and `port`; adapters are bound
to ports in their own fx modules, so swapping e.g. the cache backend touches
no use case. `infrastructure/module_test.go` validates the whole fx graph.

## API

All `/api/*` routes need `Authorization: Bearer <token>` (or `X-Api-Token`)
when `ACCESS_TOKENS` is set.

| Method | Path | |
|---|---|---|
| GET | `/api/v1/engines?group=Google` | Engine catalog with required parameters |
| GET | `/api/v1/engines/{engine}` | One engine |
| GET | `/api/v1/search?engine=google&q=coffee&...` | Any engine, SerpApi's own query interface |
| GET | `/api/v1/search/{engine}?q=coffee&...` | Same, engine in the path |
| POST | `/api/v1/search` | `{"engine":"google","params":{"q":"coffee","num":20},"no_cache":false,"async":false,"output":"json"}` |
| POST | `/api/v1/search/batch` | `{"searches":[{...},{...}]}`, run concurrently, one result per search |
| GET | `/api/v1/searches/{id}?output=json\|html\|json_with_pixel_position` | Search Archive (async results, past searches; 410 once expired after 31 days) |
| GET | `/api/v1/account` | Plan and usage (API key removed) |
| GET | `/api/v1/locations?q=Austin&limit=5` | Values for the `location` parameter (limit 1-10) |
| POST | `/api/v1/images` | Upload a jpg/png/webp up to 500 KB (multipart field `image`, or the raw body); returns `image_id` for `google_lens`, valid 10 minutes |
| GET | `/health`, `/metrics` | Probes and Prometheus |

Search responses are SerpApi's body unchanged (`output` = `json`, `html`,
`md`, or `json_with_pixel_position` for google and google_ads) with these headers: `X-Cache: HIT|MISS`, `X-Search-Id`,
`X-Search-Status`. Every other SerpApi parameter (`location`, `gl`, `hl`,
`start`, `num`, `zero_trace`, `json_restrictor`, ...) is passed through;
`api_key`, `engine`, `output`, `async` and `no_cache` are controlled by the
service.

`Accept: text/markdown` on a GET search is the same as `output=md`.
`async` cannot be combined with `no_cache`. Engines SerpApi has discontinued
(`google_scholar_profiles`, `google_lens_image_sources`) stay in the catalog
with a `discontinued` hint and are rejected with 400 instead of spending a call.

Errors are `{"error": "..."}`: 400 invalid parameters, 401 access token, 404
unknown search/engine, 410 search expired from the archive, 429 SerpApi quota or hourly limit, 502 SerpApi failure
(including a rejected API key), 504 timeout.

```sh
curl 'localhost:8080/api/v1/search/google_maps?q=pizza&ll=@40.7455096,-74.0083012,14z&type=search'
curl -XPOST localhost:8080/api/v1/search -d '{"engine":"youtube","params":{"search_query":"golang"}}'
curl 'localhost:8080/api/v1/search?engine=google&q=coffee&async=true'   # then:
curl localhost:8080/api/v1/searches/<search_metadata.id>
curl -F image=@cat.png localhost:8080/api/v1/images                   # then:
curl 'localhost:8080/api/v1/search?engine=google_lens&image_id=<image_id>'
curl 'localhost:8080/api/v1/search?engine=google&q=coffee&output=json_with_pixel_position'
```

### Caching

Successful searches are cached for `CACHE_TTL` (default 1h, the same as
SerpApi's own cache) keyed by engine, output and parameters in any order.
Not cached: `async` searches, `zero_trace=true` searches, results whose status
is not `Success`, and the Account API. `no_cache=true` skips the cache read and
asks SerpApi for a fresh search. Archive results are cached once `Success`;
Locations for `LOCATIONS_CACHE_TTL`.

### Engines

`GET /api/v1/engines` lists the catalog from `internal/domain/engines.go`.
Engines missing from it are passed through to SerpApi (so new SerpApi engines
work at once) unless `STRICT_ENGINES=true`.

## Configuration

Config is loaded in this order, each step overriding the previous one:

1. Defaults (`config.Defaults`).
2. Consul KV: the JSON at key `CONSUL_KEY` (default `serpapi_service`) on
   `CONSUL_ADDR`, with `CONSUL_HTTP_TOKEN` if Consul ACLs are on. See
   [consul.json](consul.json) for the format. A missing key falls back to
   defaults; an unreachable Consul is retried, then the pod fails to start.
   Without `CONSUL_ADDR`, this step is skipped.
3. Environment variables (below). Secrets come from here only and are never
   read from Consul.

In Kubernetes the chart seeds the Consul key from
`chart/serpapi-service/files/consul.json` once (it never overwrites an edited
key) and passes secrets through env. CI checks that both `consul.json` files
match.

| Consul (`consul.json`) | Env | Default | |
|---|---|---|---|
| - | `SERPAPI_API_KEY` | (required) | SerpApi private key, secret |
| - | `ACCESS_TOKENS` | empty (open) | Comma-separated tokens for `/api/*`, secret |
| - | `REDIS_PASSWORD` | - | Secret |
| `server.portHttpServer` | `PORT_HTTP_SERVER` | `8080` | |
| `serpapi.baseUrl` | `SERPAPI_BASE_URL` | `https://serpapi.com` | |
| `serpapi.timeout` | `SERPAPI_TIMEOUT` | `90s` | Per attempt |
| `serpapi.maxRetries` | `SERPAPI_MAX_RETRIES` | `2` | Retries on network errors and 5xx only |
| `search.strictEngines` | `STRICT_ENGINES` | `false` | Reject engines missing from the catalog |
| `search.maxConcurrent` | `MAX_CONCURRENT_SEARCHES` | `32` | In-flight SerpApi calls per pod |
| `search.maxBatchSize` | `MAX_BATCH_SIZE` | `20` | |
| `cache.backend` | `CACHE_BACKEND` | `memory` | `memory`, `redis` or `none` |
| `cache.ttl` / `cache.locationsTtl` | `CACHE_TTL` / `LOCATIONS_CACHE_TTL` | `1h` / `24h` | |
| `cache.maxEntries` | `CACHE_MAX_ENTRIES` | `1000` | Memory backend |
| `cache.redisAddr`, `cache.redisDb`, `cache.redisPrefix` | `REDIS_ADDR`, `REDIS_DB`, `REDIS_PREFIX` | `localhost:6379`, `0`, `serpapi:` | Redis backend |

## Metrics

`serpapi_upstream_requests_total{op,engine,outcome}`,
`serpapi_upstream_request_duration_seconds{op}`,
`serpapi_cache_lookups_total{op,hit}`, plus Go/process collectors.

## Development

```sh
make test     # go test -race ./...
SERPAPI_API_KEY=... make run
make docker
```
