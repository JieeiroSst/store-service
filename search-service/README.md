# search-service

Search over every table that `sync-mysql-elasticsearch-service` and `sync-postgresql-with-elasticsearch-service` copy into Elasticsearch 7.17 from the store-service databases (index = MySQL table name, or `<connector>.<schema>.<table>` for Postgres).

## Architecture

Hexagonal (ports and adapters), wired with [uber-go/fx](https://github.com/uber-go/fx):

```
cmd/main.go                          fx.New(infrastructure.Module).Run()
config/                              defaults <- Consul KV <- environment, validated
internal/domain/                     queries, documents, sensitive-field policy, errors (no I/O)
internal/port/                       DocumentSearcher, SchemaManager, HealthChecker
internal/application/                DocumentService, SchemaService (use cases)
internal/adapter/primary/http/       net/http handlers, router, auth
internal/adapter/secondary/elasticsearch/
                                     port implementations, query builders, index template (schema/*.json)
internal/infrastructure/             fx module graph and the HTTP server lifecycle
```

The application layer validates queries, rejects sensitive fields and redacts every document; the Elasticsearch adapter turns domain queries into ES requests and maps ES errors to domain errors (400 -> invalid, 404 -> not found, 429/503/connection -> unavailable).

## Configuration

`config.Load` starts from defaults, merges the Consul key `CONSUL_KEY` (default `search-service`) from `CONSUL_ADDR` when set, then applies environment overrides. The Consul document is [consul.json](consul.json) (kept identical to `chart/search-service/files/consul.json`; CI checks it).

| Env | Overrides |
|---|---|
| `CONSUL_ADDR`, `CONSUL_KEY`, `CONSUL_HTTP_TOKEN` | where to read config |
| `PORT_HTTP_SERVER` | `server.server_port` (8080) |
| `AUTHORIZE_KEY` | `secret.authorize_key` (required) |
| `ELASTICSEARCH_URL`, `ELASTICSEARCH_USERNAME`, `ELASTICSEARCH_PASSWORD` | `elasticsearch.dns`, credentials |
| `ELASTICSEARCH_INDICES` | `elasticsearch.indices` (comma separated) |
| `SCHEMA_APPLY_ON_STARTUP` | `elasticsearch.schema.apply_on_startup` |

API requests need `Authorization: Bearer <authorize_key>`; the older `Authorization: base64(<authorize_key>)` is still accepted. `GET /health` pings Elasticsearch (readiness), `GET /livez` only checks the process. Successful responses are the JSON body itself; errors are `{"error": "..."}` with 400/401/404/503/504/500.

## Development

```
make up        # Elasticsearch 7.17.5 + Kibana via docker-compose
AUTHORIZE_KEY=dev ELASTICSEARCH_URL=http://localhost:9200 make run
make test      # unit tests; set SEARCH_TEST_ES_URL=http://localhost:9200 to also run the ES integration test
make docker
```

CI: `.github/workflows/go.yml` (job `search_service`, with an Elasticsearch service container) and `.github/workflows/search-service-cicd.yml` (manual release image to ghcr.io). Deploy: `chart/search-service`, rolled out to store-dev by Argo CD and built by `deploy/ci`.

## Search API

Default indices come from `elasticsearch.indices` (`["*", "-.*"]`: everything except system indices).

### Schema

On startup the service installs (idempotently) the component template `store-sync-schema` and the index template `store-sync` (pattern `*`, priority 10). Source: [internal/repository/schema](internal/repository/schema). Every index the sync services create from then on gets:

| Feature | How |
|---|---|
| Vietnamese with/without diacritics | `text_folding` analyzer: `standard` + `lowercase` + `asciifolding(preserve_original)` |
| One search field | strings and numbers are `copy_to: all_text` |
| Autocomplete / prefix | `all_text.autocomplete` (edge n-gram 1-20) |
| "Did you mean" | `all_text.shingle` + phrase suggester |
| Synonyms | `schema.synonyms` → search-time `synonym_graph` |
| Exact filter / facet / case-insensitive sort | `<field>.keyword`, `<field>.sort` (lowercase + folding normalizer) |
| Ids and codes | `id`, `*_id` → `long`; `*_uuid`, `*_code`, `*_no`, `sku`, `slug` → `keyword` |
| Dates | auto-detected `yyyy-MM-dd HH:mm:ss`, ISO 8601, `yyyy-MM-dd`; queries accept all of them and `epoch_millis` |
| Secrets | names matching `sensitive_pattern` (password, token, secret, api_key, otp, salt, pin, card_number, cvv...) are stored but not indexed, and removed from every response |
| Sync safety | `index.mapping.ignore_malformed: true`, so a type mismatch never rejects a synced row |

Consul options under `elasticsearch.schema`: `name`, `index_patterns`, `priority`, `number_of_shards`, `number_of_replicas`, `synonyms`, `apply_on_startup` (default true). Templates only apply to new indices, so after changing the schema migrate the existing ones (below).

### Endpoints

| Method | Path | Purpose |
|---|---|---|
| GET | `/api/v1/search` | keyword search (`q` required) |
| GET | `/api/v1/documents` | list / browse (`q` optional) |
| GET | `/api/v1/autocomplete?q=&index=&size=` | search-as-you-type (max 20) |
| GET | `/api/v1/documents/:index/:id` | one document |
| GET | `/api/v1/documents/:index/:id/similar?size=` | more-like-this |
| GET | `/api/v1/indices` | indices, doc counts, `managed` flag |
| GET | `/api/v1/indices/:index/fields` | fields usable for filter / sort / facets |
| GET | `/api/v1/admin/schema` | template status and which indices still need migrating |
| PUT | `/api/v1/admin/schema` | re-apply the templates |
| POST | `/api/v1/admin/reindex?index=a&index=b` or `?unmanaged=true` | migrate indices to the current schema |

Query parameters for `/search` and `/documents`:

| Param | Example | Notes |
|---|---|---|
| `q` | `q=dien thoai` | exact, phrase, prefix and fuzzy matching on `all_text` |
| `index` | `index=products&index=book-store.*` | repeatable or comma separated, wildcards allowed |
| `filter[field]` | `filter[status]=active,pending` | values OR-ed, fields AND-ed |
| `gt/gte/lt/lte[field]` | `gte[price]=100000&lt[created_at]=2026-04-01` | numbers or dates |
| `sort` | `sort=price:desc,name` | max 3; text fields sort by `.sort`; mixed numeric types across indices are allowed |
| `facets` | `facets=status,category_id` | terms counts (max 5); `_index` is always returned |
| `size` | `size=20` | 1-100 |
| `cursor` | `cursor=<next_cursor>` | `search_after` pagination |

Facets and the suggestion are computed on the first page only. Pass `next_cursor` back as `cursor` with the same parameters; `has_more=false` marks the last page, and a cursor from different parameters returns 400. A field with conflicting types across indices (e.g. `date` in one, `text` in another) returns 400: narrow `index` or migrate the legacy index.

```
curl -G -H "Authorization: Bearer no_name" http://localhost:8080/api/v1/search \
  --data-urlencode 'q=dien thoai' --data-urlencode 'filter[status]=active' \
  --data-urlencode 'gte[price]=16000000' --data-urlencode 'sort=price:desc' --data-urlencode 'facets=status'
```

```json
{
    "items": [
      {
        "index": "products",
        "id": "2",
        "score": 12.4,
        "source": {"id": 2, "name": "Điện thoại Samsung Galaxy S24", "price": 22000000, "status": "active"},
        "highlight": {"name": ["<em>Điện</em> <em>thoại</em> Samsung Galaxy S24"]}
      }
    ],
    "total": 2,
    "next_cursor": "eyJrIjoic2VhcmNoIi...",
    "has_more": true,
    "facets": {"_index": [{"value": "products", "count": 2}], "status": [{"value": "active", "count": 2}]},
    "suggestion": "xiaomi 14"
}
```

### Migrating existing indices

`POST /api/v1/admin/reindex` migrates an index online:

1. Block writes on `X`.
2. Create `X__v<unix>` from the template.
3. `_reindex`.
4. In one atomic `_aliases` call, delete `X` and point the alias `X` at the new index.

The sync services keep writing to `X` through the alias; responses always show `X`. On failure the new index is deleted and the write block removed.

While writes are blocked, Elasticsearch rejects them with `403 cluster_block_exception`:

- **mysql-river** retries these items with backoff for about 30s. If it still fails, it exits without saving the binlog position and replays after restart.
- **Kafka Connect Elasticsearch sink** fails its task. Offsets for that batch are not committed. Afterwards, restart it with `POST /connectors/<name>/tasks/0/restart`.
