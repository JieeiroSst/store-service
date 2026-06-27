# ollama-model-py

A standalone Python service that speaks Ollama's own wire format (`/api/generate`, `/api/chat`, `/api/tags`, `/api/show`) — a drop-in alternative to a real Ollama install. The sibling Go service (`../ai-agent-system`) doesn't know or care whether `OLLAMA_BASE_URL` points at a real Ollama server or at this one; both speak the same HTTP API.

**This is a scaffold.** It answers every question with an obvious placeholder string. The one place to wire in a real backend is `app/proxy/client.py:answer()` — nothing else needs to change.

## Run

```bash
pip install -r requirements.txt
cp .env.example .env   # PORT (default 11434), MODEL_NAME (default custom-model)
uvicorn app.main:app --port "${PORT:-11434}"
```

## Try it

```bash
curl -s localhost:11434/api/tags | jq
curl -s localhost:11434/api/show -d '{"model":"custom-model"}' | jq

curl -s localhost:11434/api/chat -d '{"model":"custom-model","messages":[{"role":"user","content":"hi"}],"stream":false}' | jq
curl -s -N localhost:11434/api/chat -d '{"model":"custom-model","messages":[{"role":"user","content":"hi"}],"stream":true}'
```

## Point ai-agent-system (Go) at this instead of real Ollama

```bash
OLLAMA_BASE_URL=http://localhost:11434
OLLAMA_MODEL=custom-model   # must match MODEL_NAME above
```

## Wiring in a real answer

Edit `app/proxy/client.py`:

```python
async def answer(question: str) -> str:
    ...  # call whatever should actually answer the question
```

`routes.py` calls this for both `/api/generate` and `/api/chat`, streaming or not — it's the only function that needs real logic.
