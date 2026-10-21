# ai-agent-system

A Go gateway that sits between web/mobile clients and an Ollama-wire-compatible backend, with a MySQL audit log of every turn.

```
web/mobile  --WebSocket-->  ai-agent-system  --HTTP (Ollama API)-->  Ollama-compatible backend (real Ollama, or ../ollama-model-py)
                                  |
                                  +-- every turn logged to MySQL

other Ollama-style clients --HTTP--> ai-agent-system's own /api/... facade --(same path above)
```

Every entry point funnels through one shared use case (`application.ChatService`): relay the question to the configured Ollama-compatible backend and log the turn. FAQ short-circuiting (scripted, mandatory Q&A that must bypass the LLM) is **not** handled here — it lives in [`../ollama-model-py`](../ollama-model-py/), which checks its own JSON FAQ file before ever calling a real LLM. This server has no opinion on whether a given answer came from a script or a model; both arrive the same way over `/api/chat`.

## Architecture (hexagonal)

```
cmd/server/main.go                 composition root: config -> adapters -> use cases -> HTTP server
internal/
  domain/
    model/                         Question, Answer, Completion, Turn, HistoryEntry, Error{Kind}
    port/
      driving.go                   ChatService, HistoryService        (what the app offers)
      driven.go                    LLMClient, DeltaStream, HistoryRepository (what the app needs)
  application/                     use cases, depend only on domain/port
    chat.go                        Answer / AnswerStream, async best-effort history logging
    history.go                     List with default/max limits
  adapter/
    primary/                       driving adapters (call application through port.*Service)
      httpapi/                     /v1/chat/ws, /v1/chat/history, /healthz, /readyz
      ollamaapi/                   Ollama wire-compatible facade (/api/...)
    secondary/                     driven adapters (implement port.LLMClient / HistoryRepository)
      ollama/                      HTTP client for the backend's /api/chat, retries, error mapping
      mysql/                       chat_messages table, schema migration
  infrastructure/
    database/                      MySQL pool + ping
    server/                        http.Server, logging/recover middleware, graceful shutdown
  config/                          env-var configuration, validated at startup
```

```
 httpapi ─┐                                   ┌─> ollama  ──> Ollama-compatible backend
          ├─> port.ChatService ─> application ┤
ollamaapi ┘   port.HistoryService             └─> mysql   ──> MySQL (chat_messages)
                                   (via port.LLMClient / port.HistoryRepository)
```

**Dependency rule:** `domain` imports nothing from the project; `application` imports only `domain`; adapters import `domain` (and never each other); only `cmd/server` knows the concrete types and wires them together. Adapters translate their own failures into `model.Error{Kind}` (`bad_request`, `rate_limited`, `upstream_unavailable`, `internal`), so each primary adapter maps one error type to its own wire format.

`../ollama-model-py` is a separate, standalone service that speaks Ollama's wire format and owns the scripted FAQ check — a drop-in alternative to a real Ollama install; see its own README for details.

## Setup

```bash
docker compose up -d   # local MySQL for chat history, on host port 3307
cp .env.example .env
# edit .env: point OLLAMA_BASE_URL/OLLAMA_MODEL at your Ollama-compatible backend
```

This expects an **Ollama-compatible backend** already running somewhere reachable. That can be a real Ollama server — install it natively from [ollama.com](https://ollama.com) (simplest: it listens on `http://localhost:11434` by default) or run it yourself however you prefer; make sure the model named in `OLLAMA_MODEL` has actually been pulled (`ollama pull llama3.2 && ollama list`). Or it can be [`../ollama-model-py`](../ollama-model-py/), a standalone Python service exposing the same `/api/generate`/`/api/chat`/`/api/tags`/`/api/show` endpoints **plus the scripted FAQ check** — see its own README for the FAQ JSON format. `OLLAMA_BASE_URL` doesn't have to point at a real Ollama install; anything that speaks the same wire format works.

> The compose file maps MySQL to host port **3307**, not the default 3306, in case something else on your machine already owns 3306. Adjust `MYSQL_DSN` if you change the mapping.

| Env var | Default | Description |
|---|---|---|
| `OLLAMA_BASE_URL` | `http://localhost:11435` | Base URL of the Ollama-compatible backend |
| `OLLAMA_MODEL` | — (required) | Model name to request from that backend (must already be pulled there, if it's a real Ollama server) |
| `MYSQL_DSN` | — | Go MySQL DSN, e.g. `app:app@tcp(localhost:3307)/aiagent?parseTime=true&loc=Local`; wins over the `MYSQL_*` parts below |
| `MYSQL_HOST` / `MYSQL_PORT` / `MYSQL_USER` / `MYSQL_PASSWORD` / `MYSQL_DATABASE` | port `3306` | Used when `MYSQL_DSN` is empty (this is how the Helm chart passes the provisioned credentials Secret); `MYSQL_HOST` and `MYSQL_DATABASE` are required in that case |
| `PORT` | `8080` | HTTP listen port |

The server fails fast at startup if `OLLAMA_MODEL` or the MySQL settings are missing, or if MySQL is unreachable — all treated as hard requirements, not best-effort features. It does **not** check at startup that the backend can serve `OLLAMA_MODEL` — that's only discovered on the first question (surfaced as an `upstream_unavailable` error), since the backend being temporarily down shouldn't block the whole process from starting.

## Run

```bash
go run ./cmd/server
```

## Chat history

Every answered question — from web/mobile (WebSocket) and from the Ollama-compatible facade (`/api/...`) alike — is logged to MySQL as a write-only audit trail. It is **not** read back as conversation memory; it has zero effect on how a question gets answered.

Callers identify themselves with an `X-User-Id` header (a free-form string). For the WebSocket endpoint this is read once at the handshake and applied to every turn sent over that connection; the Ollama-compatible facade reads it per request. If omitted, the turn is logged under `anonymous` and the question is still answered normally.

```bash
curl -s localhost:8080/v1/chat/history -H 'X-User-Id: alice' | jq
# {"messages": [{"role": "assistant", "content": "...", "source": "ollama", "created_at": "..."}, ...]}
```

`GET /v1/chat/history` supports `?limit=N` (default 50, capped at 200), newest first. Saving runs in the background and is best-effort: if MySQL is temporarily unreachable, the question is still answered and the save failure is only logged server-side — chat history is an audit log, not a correctness requirement.

`source` is always `"ollama"` — this server can't tell whether the backend's answer came from a scripted FAQ or a real model call, since that decision happens inside the backend (e.g. `../ollama-model-py`'s own FAQ check), not here.

## WebSocket API (`/v1/chat/ws`) — for web/mobile

```js
const ws = new WebSocket("ws://localhost:8080/v1/chat/ws"); // X-User-Id header on the handshake if your client supports custom headers
ws.onmessage = (e) => console.log(JSON.parse(e.data));
ws.onopen = () => ws.send(JSON.stringify({ question: "What are your support hours?" }));
```

One connection handles many turns — send another `{"question": "..."}` frame any time after the previous turn's `done`/`error`. Every message the server sends is one JSON object with a `type`:

| `type` | Fields | Meaning |
|---|---|---|
| `delta` | `text` | One incremental chunk of the answer — real per-token streaming if the backend is a real Ollama server; a single chunk with the whole answer if it's `../ollama-model-py` (which doesn't stream) |
| `done` | `answer`, `source` | The complete answer for this turn; `source` is always `"ollama"` |
| `error` | `message` | The turn failed; the connection stays open for the next question |

The server also sends periodic ping frames to keep the connection alive through proxies/browsers; respond to them automatically (every standard WebSocket client does this transparently).

> `CheckOrigin` is currently permissive (accepts any origin) — tighten this in `internal/adapter/primary/httpapi/ws_handler.go` before exposing this endpoint outside a trusted network.

## Ollama-compatible facade (`/api/...`) — for other systems

A second, independent surface that emulates Ollama's own wire API (`/api/generate`, `/api/chat`, `/api/tags`, `/api/show`), so any existing Ollama client/SDK/tool can point at **this** server instead of a real one and transparently get the same relay-then-history-log behavior.

```bash
curl -s localhost:8080/api/tags | jq
curl -s localhost:8080/api/show -d '{"model":"llama3.2:latest"}' | jq

curl -s localhost:8080/api/generate -d '{"model":"llama3.2:latest","prompt":"why is the sky blue?","stream":false}' | jq
curl -s -N localhost:8080/api/generate -d '{"model":"llama3.2:latest","prompt":"why is the sky blue?","stream":true}'

curl -s localhost:8080/api/chat -H 'X-User-Id: alice' -d '{"model":"llama3.2:latest","messages":[{"role":"user","content":"hi"}],"stream":false}' | jq
```

Streaming responses are newline-delimited JSON (`application/x-ndjson`), matching Ollama's real wire format — not Server-Sent Events or WebSocket frames. `eval_count`/`prompt_eval_count` on the final line are the real Ollama server's own token usage, passed through.

There is only one model — this facade always answers as `OLLAMA_MODEL` regardless of what model name a client requests; `/api/show` returns 404 for any other name.

**Real client compatibility check** (stronger signal than curl, since SDKs enforce field-level expectations):

```bash
OLLAMA_HOST=http://localhost:8080 ollama run llama3.2:latest
# or, Python:
python -c "import ollama; c=ollama.Client(host='http://localhost:8080'); print(c.chat(model='llama3.2:latest', messages=[{'role':'user','content':'hi'}]))"
```

### Error shape

WebSocket errors arrive as an in-band `{"type":"error","message":"..."}` frame (the connection itself stays open). The `/api/...` facade and `/v1/chat/history` use Ollama-flat (`{"error":"..."}`) and nested (`{"error":{"message":"...","type":"..."}}`) JSON error bodies respectively, with `429`/`503`/`400`/`500` status codes for rate-limited/upstream-unavailable/bad-request/internal failures.

## Health

| Path | Meaning |
|---|---|
| `GET /healthz` | Process is up (liveness/startup probe) |
| `GET /readyz` | Also pings MySQL; `503` while the database is unreachable (readiness probe) |

## Development

```bash
make fmt-check build vet
make test                       # unit tests; the MySQL repository test is skipped
make db-up                      # local MySQL on :3307, then run the repository test too:
AI_AGENT_TEST_MYSQL_ADDR=127.0.0.1:3307 AI_AGENT_TEST_MYSQL_USER=app \
AI_AGENT_TEST_MYSQL_PASSWORD=app AI_AGENT_TEST_MYSQL_DATABASE=aiagent make test
```

## CI/CD

| Stage | Where | What |
|---|---|---|
| PR / push to `master` | `.github/workflows/go.yml` (job `ai_agent_system`) | gofmt, build, vet, `go test -race` with a MySQL service container |
| Dev deploy | in-cluster CI (`deploy/ci/services.yaml`, entry `ai-agent-system`) | every push to `master` is tested (MySQL sidecar), built to `localhost:5000/ai-agent-system`, and Argo CD Image Updater rolls it out to `store-dev` |
| Release image | `.github/workflows/ai-agent-system-cicd.yml` (manual) | same tests, then multi-arch image `ghcr.io/jieeirosst/ai-agent-system:<sha7>` with SBOM/provenance, blocked on critical CVEs (Trivy) |
| Staging / prod | `.github/workflows/promote.yml` (`service: ai-agent-system`) | opens a PR setting the tag in `chart/ai-agent-system/values-<env>.yaml`; staging auto-syncs, prod is synced by hand |

Database: `chart/mysql` provisions database `ai_agent_system` and user `ai_agent_system_svc`; the credentials Secret `ai-agent-system-db-credentials` is mirrored into `store-dev`/`store-staging` by ExternalSecret. Prod reads `store/prod/ai-agent-system/mysql` (`username`, `password`) from the secret manager — create that key before the first prod sync.
