"""PDF RAG (Retrieval-Augmented Generation) layer.

load() reads the pre-built vector store (data/pdf_vectors.json) synchronously
at startup. Run scripts/ingest_pdfs.py first to build it.

match() embeds the incoming question, finds the top-k most similar PDF chunks,
and returns a concise LLM-generated answer grounded in those chunks.
Returns None when no chunk scores above PDF_SIMILARITY_THRESHOLD so the caller
can fall through to the next layer.
"""
import json
import math
from dataclasses import dataclass
from pathlib import Path

from . import config
from .proxy import client

cfg = config.load()


@dataclass
class _Chunk:
    source: str
    text: str
    embedding: list[float]


_chunks: list[_Chunk] = []


def _cosine(a: list[float], b: list[float]) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    na = math.sqrt(sum(x * x for x in a))
    nb = math.sqrt(sum(y * y for y in b))
    if na == 0 or nb == 0:
        return 0.0
    return dot / (na * nb)


def load() -> None:
    """Load pre-built PDF vectors from disk. No-op if the file doesn't exist."""
    vector_file = Path(cfg.pdf_vector_file)
    if not vector_file.exists():
        return

    with open(vector_file, "r", encoding="utf-8") as f:
        raw: list[dict] = json.load(f)

    _chunks.clear()
    _chunks.extend(_Chunk(r["source"], r["text"], r["embedding"]) for r in raw)


async def match(question: str, top_k: int = 3) -> str | None:
    """Return an LLM answer grounded in the best matching PDF chunks, or None."""
    if not _chunks:
        return None

    try:
        embeddings = await client.embed([question])
    except client.BackendError:
        return None
    if not embeddings:
        return None

    query = embeddings[0]
    scored = sorted(
        ((_cosine(query, c.embedding), c) for c in _chunks),
        key=lambda x: x[0],
        reverse=True,
    )

    best_score, _ = scored[0]
    if best_score < cfg.pdf_similarity_threshold:
        return None

    context_chunks = [
        c for score, c in scored[:top_k] if score >= cfg.pdf_similarity_threshold
    ]
    context = "\n\n".join(c.text for c in context_chunks)

    return await client.answer_with_context(question, context)
