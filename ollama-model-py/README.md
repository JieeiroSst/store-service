# ollama-model-py

A standalone Python service that speaks Ollama's own wire format (`/api/generate`, `/api/chat`, `/api/tags`, `/api/show`) — a drop-in alternative to a real Ollama install. The sibling Go service (`../ai-agent-system`) doesn't know or care whether `OLLAMA_BASE_URL` points at a real Ollama server or at this one; both speak the same HTTP API.

**Domain:** banking & fintech customer support (Vietnamese). Every incoming question is resolved through a four-layer pipeline before reaching the LLM:

```
Question
   │
   ▼
1. Scripted FAQ         (app/faq.json)                  → exact scripted answer, no LLM call
   │ no match
   ▼
2. PDF RAG              (app/pdf_rag.py)                → answer grounded in uploaded docs
   │ no match
   ▼
3. Internal API intents (app/internal_api/intents.py)   → live data from internal services
   │ no match
   ▼
4. LLM fallback         (app/proxy/client.py:answer())  → bank-assistant persona (Modelfile)
```

---

## Quick start

```bash
# 1. Install dependencies
make install

# 2. Copy env and configure
cp .env.example .env   # edit OLLAMA_BASE_URL if Ollama runs elsewhere

# 3. Pull required Ollama models
ollama pull nomic-embed-text   # embed model (FAQ + intent + PDF matching)
ollama pull qwen2.5:14b        # base model for the bank-assistant persona

# 4. Build the bank-assistant persona
make create-model

# 5. (Optional) Embed your PDF documents — see PDF RAG section below
make ingest

# 6. Start the service
make dev
```

---

## Environment variables

| Variable | Default | Description |
|---|---|---|
| `PORT` | `11436` | Port this service listens on |
| `MODEL_NAME` | `custom-model` | Model name advertised on `/api/tags` |
| `OLLAMA_BASE_URL` | `http://localhost:11434` | Real Ollama server for chat + embed |
| `OLLAMA_TARGET_MODEL` | `bank-assistant` | Chat/completion model |
| `OLLAMA_EMBED_MODEL` | `nomic-embed-text` | Embedding model |
| `FAQ_SIMILARITY_THRESHOLD` | `0.6` | Cosine threshold for FAQ matching |
| `INTERNAL_API_SIMILARITY_THRESHOLD` | `0.6` | Cosine threshold for intent matching |
| `POINTS_SERVICE_BASE_URL` | `http://localhost:8089` | Internal points/account service |
| `PDF_VECTOR_FILE` | `data/pdf_vectors.json` | Pre-built PDF vector store |
| `PDF_SIMILARITY_THRESHOLD` | `0.65` | Cosine threshold for PDF RAG matching |

---

## Try it

### GET /api/tags — list available models

```bash
curl -s localhost:11436/api/tags | jq
```

### POST /api/show — model info

```bash
curl -s localhost:11436/api/show \
  -H 'Content-Type: application/json' \
  -d '{"model":"custom-model"}' | jq
```

### POST /api/chat — single turn (stream off)

```bash
# Layer 1: FAQ match — scripted answer, no LLM call
curl -s localhost:11436/api/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "custom-model",
    "messages": [{"role":"user","content":"giờ hỗ trợ khách hàng là mấy giờ?"}],
    "stream": false
  }' | jq

# Layer 2: PDF RAG — answer grounded in uploaded docs (requires make ingest)
curl -s localhost:11436/api/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "custom-model",
    "messages": [{"role":"user","content":"lãi suất tiết kiệm 12 tháng là bao nhiêu?"}],
    "stream": false
  }' | jq

# Layer 3: Internal API intent — live data, requires user_id
curl -s localhost:11436/api/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "custom-model",
    "messages": [{"role":"user","content":"số dư tài khoản của tôi là bao nhiêu?"}],
    "stream": false,
    "user_id": "user-123"
  }' | jq

# Layer 4: LLM fallback — no match in any layer, forwarded to bank-assistant
curl -s localhost:11436/api/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "custom-model",
    "messages": [{"role":"user","content":"xin chào"}],
    "stream": false
  }' | jq
```

### POST /api/chat — streaming (Layer 4)

```bash
curl -s -N localhost:11436/api/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "custom-model",
    "messages": [{"role":"user","content":"thẻ tín dụng có lợi gì so với thẻ ghi nợ?"}],
    "stream": true
  }'
```

### POST /api/chat — multi-turn conversation

Truyền toàn bộ lịch sử hội thoại để model nhớ context:

```bash
curl -s localhost:11436/api/chat \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "custom-model",
    "messages": [
      {"role":"user",     "content":"tôi muốn mở thẻ tín dụng"},
      {"role":"assistant","content":"Anh/chị cho mình biết thu nhập hàng tháng khoảng bao nhiêu để mình tư vấn hạn mức phù hợp nhé?"},
      {"role":"user",     "content":"khoảng 20 triệu"}
    ],
    "stream": false
  }' | jq
```

### POST /api/generate — generate endpoint

```bash
# Non-streaming
curl -s localhost:11436/api/generate \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "custom-model",
    "prompt": "vay tiêu dùng tín chấp cần hồ sơ gì?",
    "stream": false
  }' | jq

# Streaming
curl -s -N localhost:11436/api/generate \
  -H 'Content-Type: application/json' \
  -d '{
    "model": "custom-model",
    "prompt": "vay tiêu dùng tín chấp cần hồ sơ gì?",
    "stream": true
  }'
```

---

## Layer 1 — Scripted FAQ (`app/faq.json`)

Questions that must always return a fixed answer (support hours, fee policies, hotline numbers, etc.). Matching is semantic, not exact string matching — embed the incoming question, compare cosine similarity against all FAQ embeddings, return the scripted answer if score ≥ `FAQ_SIMILARITY_THRESHOLD`.

```json
{
  "faqs": [
    {
      "question": "Hotline hỗ trợ khách hàng là số nào?",
      "answer": "Hotline của chúng tôi là 1800-XXXX, miễn phí 24/7."
    },
    {
      "question": "Phí chuyển khoản ngoại mạng là bao nhiêu?",
      "answer": "Phí chuyển khoản ngoại mạng là 11.000đ/giao dịch."
    }
  ]
}
```

FAQ questions are embedded once at startup (`app/faq.py:load()`). If the embed backend is unreachable mid-request, the FAQ check is skipped (not a hard failure) and the question falls through. Edit `app/faq.json` and restart to update.

**Tuning `FAQ_SIMILARITY_THRESHOLD`:** lower = catch more loosely-worded paraphrases (risk: wrong match); higher = stricter (risk: FAQ miss falls to LLM). `0.6` is a starting point — tune against real traffic or swap `OLLAMA_EMBED_MODEL` for a stronger model.

---

## Layer 2 — PDF RAG (`app/pdf_rag.py`)

Answers grounded in your uploaded banking documents (rate sheets, product brochures, NHNN circulars, internal procedures). When a question matches a PDF chunk above `PDF_SIMILARITY_THRESHOLD`, the matching chunks are sent as context to the LLM which generates a concise answer — it cannot fabricate beyond what the document says.

---

## Layer 3 — Internal API intents (`app/internal_api/`)

Questions the LLM can't answer correctly because they need live data (e.g. "số dư tài khoản của tôi là bao nhiêu?", "tra cứu lịch sử giao dịch"). These are matched the same way as FAQs (embed + cosine), then dispatched to a real internal service call.

- `intents.py` — maps example phrases → async handler, dispatched at `INTERNAL_API_SIMILARITY_THRESHOLD`
- `client.py` — one function per internal service endpoint (e.g. `get_points_balance()`)

**Adding a new intent:** add an entry to `_REGISTRY` in `intents.py` with example phrasings + handler, and a matching fetch function in `client.py`.

**User identity** flows from the sibling Go service via the `user_id` field on `/api/chat` requests (populated from `X-User-Id` header). Handlers fall back to `PLACEHOLDER_USER_ID` when absent.

### Folder structure

```
docs/
  banking-products/   ← rate cards, product sheets, brochures
  regulations/        ← NHNN circulars, compliance documents
  procedures/         ← internal guides, user manuals
```

### Workflow

```bash
# 1. Drop PDF files into the appropriate subfolder under docs/
# 2. Run the ingest script (chunks PDFs, embeds, saves data/pdf_vectors.json)
make ingest

# 3. Restart the service to load the new vectors
make dev
```

Re-run `make ingest` every time you add, remove, or update PDFs. The service loads `data/pdf_vectors.json` at startup — if the file doesn't exist (ingest hasn't been run yet), the PDF layer is silently skipped.

**Tuning `PDF_SIMILARITY_THRESHOLD`:** `0.65` is slightly stricter than the FAQ threshold because PDF chunks are longer and less precisely phrased than FAQ questions — lower it if relevant document questions are falling through to the LLM.

---

## Layer 4 — LLM fallback persona (`Modelfile`)

Reached only when all three layers above find no match. Defined by `Modelfile` on top of `qwen2.5:14b`:

- Always replies in Vietnamese (regardless of input language)
- Professional banking tone — friendly but not overly formal
- Concise — a few sentences, no unnecessary lists
- Honest about what it doesn't know (no live account data) → directs to `support@example.com`
- Never fabricates balances, transaction status, or interest rates

```bash
make create-model              # ollama create bank-assistant -f Modelfile
make create-model OLLAMA_TARGET_MODEL=my-persona   # custom name
```

Drop back to `FROM qwen2.5:7b` in `Modelfile` if `14b` is too slow (~10GB RAM needed).

---

## Pointing ai-agent-system (Go) at this service

```bash
# In ai-agent-system's .env
OLLAMA_BASE_URL=http://localhost:11436
OLLAMA_MODEL=custom-model   # must match MODEL_NAME in this service's .env
```

---

## Adding a real answer backend

Edit `app/proxy/client.py:answer()` — it's called for both `/api/generate` and `/api/chat` as the last fallback. Everything above it in the pipeline (FAQ, internal API, PDF RAG) is already handled; `answer()` only needs to handle what genuinely requires open-ended generation.
