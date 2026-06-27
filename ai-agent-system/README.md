# ai-agent-system

A Go gateway that sits between web/mobile clients and an Ollama-wire-compatible backend, with a MySQL audit log of every turn.

```
web/mobile  --WebSocket-->  ai-agent-system  --HTTP (Ollama API)-->  Ollama-compatible backend (real Ollama, or ../ollama-model-py)
                                  |
                                  +-- every turn logged to MySQL

other Ollama-style clients --HTTP--> ai-agent-system's own /api/... facade --(same path above)
```

Every entry point funnels through one shared core (`internal/chat.Service`): relay the question to the configured Ollama-compatible backend and log the turn. FAQ short-circuiting (scripted, mandatory Q&A that must bypass the LLM) is **not** handled here — it lives in [`../ollama-model-py`](../ollama-model-py/), which checks its own JSON FAQ file before ever calling a real LLM. This server has no opinion on whether a given answer came from a script or a model; both arrive the same way over `/api/chat`.

## Architecture

```
cmd/server/main.go        wiring + graceful shutdown
internal/
  config/                 env-var configuration, validated at startup
  proxy/                  ALL outbound calls to external services live here
    ollama_client.go        HTTP client for the Ollama-compatible backend's /api/chat
    retry.go                backoff helper
    errors.go                normalizes Ollama errors into one Error{Kind}
  chat/                   the ONE shared "answer a question" path
    service.go               AnswerQuestion / AnswerQuestionStream
  history/                chat history audit log, persisted to MySQL
  httpapi/                WebSocket + history-lookup surface (/v1/...)
  ollamaapi/              Ollama wire-compatible facade (/api/...) for other systems
docker-compose.yml        local MySQL for development
```

`../ollama-model-py` is a separate, standalone service (own repo path, own Python source tree) that speaks Ollama's wire format and owns the scripted FAQ check — a drop-in alternative to a real Ollama install; see its own README for details.

**Dependency direction:** `httpapi`, `ollamaapi` → `chat` → `proxy`. Neither HTTP surface imports `proxy` directly — they only know `chat` types, so every request path shares exactly one relay-then-log path.

### Why a `proxy` package

The Ollama-compatible backend has no Go SDK, so `internal/proxy/ollama_client.go` talks to it over raw HTTP using its own wire format (the same shapes `internal/ollamaapi` implements as a server, redefined locally to avoid a cross-package refactor for a handful of small structs). Timeouts, retries, and request construction are centralized in `internal/proxy` so no other package needs to know how to talk to an external service directly.

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
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Base URL of the Ollama-compatible backend |
| `OLLAMA_MODEL` | — (required) | Model name to request from that backend (must already be pulled there, if it's a real Ollama server) |
| `MYSQL_DSN` | — (required) | Go MySQL DSN, e.g. `app:app@tcp(localhost:3307)/aiagent?parseTime=true&loc=Local` |
| `PORT` | `8080` | HTTP listen port |

The server fails fast at startup if `OLLAMA_MODEL` or `MYSQL_DSN` is missing, or if MySQL is unreachable — all treated as hard requirements, not best-effort features. It does **not** check at startup that the backend can serve `OLLAMA_MODEL` — that's only discovered on the first question (surfaced as an `upstream_unavailable` error), since the backend being temporarily down shouldn't block the whole process from starting.

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

> `CheckOrigin` is currently permissive (accepts any origin) — tighten this in `internal/httpapi/ws_handler.go` before exposing this endpoint outside a trusted network.

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

## Development

```bash
go build ./...
go vet ./...
gofmt -l .
```
