# ai-agent-system

A Go gateway that sits between web/mobile clients and a real, locally-running [Ollama](https://ollama.com) server, with an FAQ short-circuit for common questions and a MySQL audit log of every turn.

```
web/mobile  --WebSocket-->  ai-agent-system  --HTTP (Ollama API)-->  Ollama server (real, local model)
                                  |
                                  +-- FAQ check (Voyage embeddings) before ever calling Ollama
                                  +-- every turn logged to MySQL

other Ollama-style clients --HTTP--> ai-agent-system's own /api/... facade --(same path above)
```

Every entry point funnels through one shared core (`internal/chat.Service`): check the FAQ store first, and only call the real Ollama server if nothing matches closely enough.

## Architecture

```
cmd/server/main.go        wiring + graceful shutdown
internal/
  config/                 env-var configuration, validated at startup
  proxy/                  ALL outbound calls to external services live here
    ollama_client.go        HTTP client for a real Ollama server's own /api/chat
    voyage_client.go        raw HTTP client for Voyage AI embeddings (no Go SDK exists)
    retry.go                backoff helper
    errors.go                normalizes Ollama/Voyage errors into one Error{Kind}
  chat/                   the ONE shared "answer a question" path
    service.go               AnswerQuestion / AnswerQuestionStream
  faq/                    FAQ store + semantic (embedding) matching
    store.go                 Load (embed at startup) / Match (cosine similarity)
  history/                chat history audit log, persisted to MySQL
  httpapi/                WebSocket + history-lookup surface (/v1/...)
  ollamaapi/              Ollama wire-compatible facade (/api/...) for other systems
configs/faq.yaml          FAQ question/answer pairs
docker-compose.yml        local MySQL for development
```

`../ollama-model-py` is a separate, standalone service (own repo path, own Python source tree) that speaks Ollama's wire format — a drop-in alternative to a real Ollama install; see its own README for details.

**Dependency direction:** `httpapi`, `ollamaapi` → `chat` → `proxy`, `faq`. Neither HTTP surface imports `proxy` directly — they only know `chat` types, so every request path shares exactly one FAQ-check-then-Ollama-call path.

### Why a `proxy` package

The real Ollama server has no Go SDK, so `internal/proxy/ollama_client.go` talks to it over raw HTTP using its own wire format (the same shapes `internal/ollamaapi` implements as a server, redefined locally to avoid a cross-package refactor for a handful of small structs). Voyage AI (used for FAQ embeddings) also has no Go SDK. Both outbound integrations — timeouts, retries, request construction — are centralized in `internal/proxy` so no other package needs to know how to talk to an external service directly.

## Setup

```bash
docker compose up -d   # local MySQL for chat history, on host port 3307
cp .env.example .env
# edit .env: set VOYAGE_API_KEY, and point OLLAMA_BASE_URL/OLLAMA_MODEL at your Ollama install
```

This expects a **real Ollama server** already running somewhere reachable — install it natively from [ollama.com](https://ollama.com) (simplest: it listens on `http://localhost:11434` by default) or run it yourself however you prefer. There's no bundled Ollama container here: if you already have one running locally (`ollama serve`, or the desktop app), just point `OLLAMA_BASE_URL` at it. Make sure the model named in `OLLAMA_MODEL` has actually been pulled:

```bash
ollama pull llama3.2
ollama list   # confirm it's there
```

`OLLAMA_BASE_URL` doesn't have to point at a real Ollama install — anything that speaks the same wire format works. [`../ollama-model-py`](../ollama-model-py/) is exactly that: a standalone Python service exposing the same `/api/generate`/`/api/chat`/`/api/tags`/`/api/show` endpoints, useful as a starting point for plugging in a custom (non-Ollama) answering backend without touching any Go code.

> The compose file maps MySQL to host port **3307**, not the default 3306, in case something else on your machine already owns 3306. Adjust `MYSQL_DSN` if you change the mapping.

| Env var | Default | Description |
|---|---|---|
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Base URL of the real Ollama server |
| `OLLAMA_MODEL` | — (required) | Model name to request from that Ollama server (must already be pulled there) |
| `VOYAGE_API_KEY` | — (required) | Voyage AI embeddings API key, used for FAQ matching |
| `MYSQL_DSN` | — (required) | Go MySQL DSN, e.g. `app:app@tcp(localhost:3307)/aiagent?parseTime=true&loc=Local` |
| `PORT` | `8080` | HTTP listen port |
| `FAQ_FILE` | `./configs/faq.yaml` | Path to the FAQ question/answer file |
| `FAQ_SIMILARITY_THRESHOLD` | `0.85` | Minimum cosine similarity to short-circuit to a canned FAQ answer instead of calling Ollama |

The server fails fast at startup if `OLLAMA_MODEL`, `VOYAGE_API_KEY`, or `MYSQL_DSN` is missing, if MySQL is unreachable, or if the FAQ file can't be loaded/embedded — all treated as hard requirements, not best-effort features. It does **not** check at startup that the real Ollama server is reachable — that's only discovered on the first question (surfaced as an `upstream_unavailable` error), since Ollama being temporarily down shouldn't block the whole process from starting.

## Run

```bash
go run ./cmd/server
```

## FAQ file format

`configs/faq.yaml`:

```yaml
faqs:
  - question: "What are your support hours?"
    answer: "Our support team is available 9am-6pm ET, Monday through Friday."
```

At startup, every FAQ question is embedded once (batched into a single Voyage call). On each incoming question, the question itself is embedded and compared via cosine similarity against the cached FAQ vectors; a match at or above `FAQ_SIMILARITY_THRESHOLD` returns the canned answer with zero Ollama calls. If Voyage is temporarily unreachable, the FAQ check is skipped (logged as a warning) and the question falls through to Ollama rather than failing the request.

## Chat history

Every answered question — from web/mobile (WebSocket) and from the Ollama-compatible facade (`/api/...`) alike — is logged to MySQL as a write-only audit trail. It is **not** read back as conversation memory; it has zero effect on how a question gets answered.

Callers identify themselves with an `X-User-Id` header (a free-form string). For the WebSocket endpoint this is read once at the handshake and applied to every turn sent over that connection; the Ollama-compatible facade reads it per request. If omitted, the turn is logged under `anonymous` and the question is still answered normally.

```bash
curl -s localhost:8080/v1/chat/history -H 'X-User-Id: alice' | jq
# {"messages": [{"role": "assistant", "content": "...", "source": "faq" | "ollama", "created_at": "..."}, ...]}
```

`GET /v1/chat/history` supports `?limit=N` (default 50, capped at 200), newest first. Saving runs in the background and is best-effort: if MySQL is temporarily unreachable, the question is still answered and the save failure is only logged server-side — chat history is an audit log, not a correctness requirement.

## WebSocket API (`/v1/chat/ws`) — for web/mobile

```js
const ws = new WebSocket("ws://localhost:8080/v1/chat/ws"); // X-User-Id header on the handshake if your client supports custom headers
ws.onmessage = (e) => console.log(JSON.parse(e.data));
ws.onopen = () => ws.send(JSON.stringify({ question: "What are your support hours?" }));
```

One connection handles many turns — send another `{"question": "..."}` frame any time after the previous turn's `done`/`error`. Every message the server sends is one JSON object with a `type`:

| `type` | Fields | Meaning |
|---|---|---|
| `delta` | `text` | One incremental chunk of the answer (only for non-FAQ answers; FAQ answers arrive as a single `delta`) |
| `done` | `answer`, `source` | The complete answer for this turn; `source` is `"faq"` or `"ollama"` |
| `error` | `message` | The turn failed; the connection stays open for the next question |

The server also sends periodic ping frames to keep the connection alive through proxies/browsers; respond to them automatically (every standard WebSocket client does this transparently).

> `CheckOrigin` is currently permissive (accepts any origin) — tighten this in `internal/httpapi/ws_handler.go` before exposing this endpoint outside a trusted network.

## Ollama-compatible facade (`/api/...`) — for other systems

A second, independent surface that emulates Ollama's own wire API (`/api/generate`, `/api/chat`, `/api/tags`, `/api/show`), so any existing Ollama client/SDK/tool can point at **this** server instead of a real one and transparently get the same FAQ-check-then-real-Ollama-relay-then-history-log behavior.

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
