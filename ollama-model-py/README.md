# ollama-model-py

A standalone Python service that speaks Ollama's own wire format (`/api/generate`, `/api/chat`, `/api/tags`, `/api/show`, `/api/embed`) — a drop-in alternative to a real Ollama install. The sibling Go service (`../ai-agent-system`) doesn't know or care whether `OLLAMA_BASE_URL` points at a real Ollama server or at this one; both speak the same HTTP API.

**This implementation forwards to a real Ollama server.** `app/proxy/client.py` relays `answer()`/`embed()` to a real Ollama install (`OLLAMA_BASE_URL`), using `OLLAMA_TARGET_MODEL` for chat and `OLLAMA_EMBED_MODEL` for embeddings — useful as a starting point for plugging in a custom (non-Ollama) backend instead, without touching `routes.py`.

## Run

```bash
pip install -r requirements.txt
cp .env.example .env   # PORT (default 11435), MODEL_NAME, OLLAMA_BASE_URL/OLLAMA_TARGET_MODEL/OLLAMA_EMBED_MODEL
uvicorn app.main:app --port "${PORT:-11435}"
```

## Try it

```bash
curl -s localhost:11435/api/tags | jq
curl -s localhost:11435/api/show -d '{"model":"custom-model"}' | jq

curl -s localhost:11435/api/chat -H 'Content-Type: application/json' -d '{"model":"custom-model","messages":[{"role":"user","content":"hi"}],"stream":false}' | jq
curl -s -N localhost:11435/api/chat -H 'Content-Type: application/json' -d '{"model":"custom-model","messages":[{"role":"user","content":"hi"}],"stream":true}'

curl -s localhost:11435/api/embed -H 'Content-Type: application/json' -d '{"model":"custom-model","input":["hello world"]}' | jq
```

## Point ai-agent-system (Go) at this instead of real Ollama

```bash
OLLAMA_BASE_URL=http://localhost:11435
OLLAMA_MODEL=custom-model   # must match MODEL_NAME above
OLLAMA_EMBED_MODEL=custom-model   # same — model name isn't validated for /api/chat or /api/embed
```

## Wiring in a real answer

Edit `app/proxy/client.py`:

```python
async def answer(question: str) -> str: ...      # called for /api/generate and /api/chat
async def embed(inputs: list[str]) -> list[list[float]]: ...  # called for /api/embed
```

`routes.py` calls `answer()` for both `/api/generate` and `/api/chat` (streaming or not), and `embed()` for `/api/embed` — these are the only functions that need real logic.
