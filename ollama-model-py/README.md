# ollama-model-py

A standalone Python service that speaks Ollama's own wire format (`/api/generate`, `/api/chat`, `/api/tags`, `/api/show`) — a drop-in alternative to a real Ollama install. The sibling Go service (`../ai-agent-system`) doesn't know or care whether `OLLAMA_BASE_URL` points at a real Ollama server or at this one; both speak the same HTTP API.

**This is also where the mandatory scripted FAQ lives.** Every question to `/api/generate`/`/api/chat` is checked against `app/faq.json` first. Real questions are rarely worded exactly like the script, so matching is semantic (Ollama embeddings + cosine similarity), not exact string matching — a close-enough match returns that scripted answer verbatim, with zero calls to the LLM. Only questions with no close-enough match fall through to a real Ollama server (`app/proxy/client.py`, using `OLLAMA_BASE_URL`/`OLLAMA_TARGET_MODEL`).

## Run

```bash
pip install -r requirements.txt
cp .env.example .env   # PORT (default 11435), MODEL_NAME, OLLAMA_BASE_URL/OLLAMA_TARGET_MODEL/OLLAMA_EMBED_MODEL
ollama pull nomic-embed-text   # the default OLLAMA_EMBED_MODEL, used for FAQ matching
uvicorn app.main:app --port "${PORT:-11435}"
```

## Try it

```bash
curl -s localhost:11435/api/tags | jq
curl -s localhost:11435/api/show -d '{"model":"custom-model"}' | jq

# worded differently than the script, but close enough -> scripted answer, no LLM call
curl -s localhost:11435/api/chat -H 'Content-Type: application/json' -d '{"model":"custom-model","messages":[{"role":"user","content":"when can I reach support?"}],"stream":false}' | jq

# no FAQ match -> forwarded to the real Ollama server
curl -s -N localhost:11435/api/chat -H 'Content-Type: application/json' -d '{"model":"custom-model","messages":[{"role":"user","content":"hi"}],"stream":true}'
```

## The scripted FAQ (`app/faq.json`)

```json
{
  "faqs": [
    {
      "question": "What are your support hours?",
      "answer": "Our support team is available 9am-6pm ET, Monday through Friday."
    }
  ]
}
```

At startup (`app/main.py`'s lifespan), every FAQ question in `app/faq.json` is embedded once via `OLLAMA_EMBED_MODEL` (`app/faq.py:load()`) — this requires the real Ollama server at `OLLAMA_BASE_URL` to be reachable and that model already pulled, and fails startup otherwise (mandatory feature, not best-effort). On each incoming question, `app/faq.py:match()` embeds it and compares via cosine similarity against the cached FAQ vectors; the closest match at or above `FAQ_SIMILARITY_THRESHOLD` (default `0.6`) wins. If the embedding backend is temporarily unreachable mid-request, the FAQ check is skipped for that question (not a hard failure) and it falls through to `answer()`. Change `app/faq.json` and restart to update the script; point `FAQ_FILE` at a different path to use a JSON file elsewhere.

**Tuning `FAQ_SIMILARITY_THRESHOLD`:** `nomic-embed-text` on short sentences doesn't separate cleanly — in testing, real paraphrases of the FAQ scored anywhere from ~0.55 to ~0.90, while unrelated questions topped out around ~0.45-0.55. There's no threshold that's correct for every case: lower it to catch more loosely-worded paraphrases at the risk of occasionally matching an unrelated question to the wrong FAQ; raise it to be stricter at the risk of letting more real paraphrases fall through to the LLM. `0.6` is a starting point, not a guarantee — tune it against your own FAQ content and real traffic, or swap in a stronger embedding model via `OLLAMA_EMBED_MODEL`.

## Point ai-agent-system (Go) at this instead of real Ollama

```bash
OLLAMA_BASE_URL=http://localhost:11435
OLLAMA_MODEL=custom-model   # must match MODEL_NAME above
```

## Wiring in a real answer

Edit `app/proxy/client.py`:

```python
async def answer(question: str) -> str:
    ...  # called for /api/generate and /api/chat, after the FAQ check finds no match
```

`routes.py`'s `_resolve()` calls `faq.match()` first for both `/api/generate` and `/api/chat`, streaming or not, and falls back to `answer()` only when nothing scores above `FAQ_SIMILARITY_THRESHOLD` — `answer()` is the only function that needs real backend logic.
