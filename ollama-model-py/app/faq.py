"""Mandatory scripted FAQ: questions that must return a fixed answer instead
of going to the LLM. Real questions rarely match the script word-for-word,
so matching is semantic (Ollama embeddings + cosine similarity), not exact.

load() embeds every FAQ question once at startup (see main.py's lifespan).
match() embeds the incoming question and returns the closest FAQ answer if
it scores at or above FAQ_SIMILARITY_THRESHOLD, else None — callers fall
through to the real backend in that case.
"""
import json
import math
from dataclasses import dataclass

from . import config
from .proxy import client

cfg = config.load()


@dataclass
class _Entry:
    question: str
    answer: str
    embedding: list[float]


_entries: list[_Entry] = []


def _cosine_similarity(a: list[float], b: list[float]) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    norm_a = math.sqrt(sum(x * x for x in a))
    norm_b = math.sqrt(sum(y * y for y in b))
    if norm_a == 0 or norm_b == 0:
        return 0.0
    return dot / (norm_a * norm_b)


async def load() -> None:
    with open(cfg.faq_file, "r", encoding="utf-8") as f:
        raw = json.load(f).get("faqs", [])

    if not raw:
        _entries.clear()
        return

    embeddings = await client.embed([e["question"] for e in raw])
    if len(embeddings) != len(raw):
        raise RuntimeError(
            f"faq: embedder returned {len(embeddings)} embeddings for {len(raw)} entries"
        )

    _entries.clear()
    _entries.extend(
        _Entry(e["question"], e["answer"], emb) for e, emb in zip(raw, embeddings)
    )


async def match(question: str) -> str | None:
    if not _entries:
        return None

    try:
        embeddings = await client.embed([question])
    except client.BackendError:
        return None
    if not embeddings:
        return None
    query = embeddings[0]

    best_entry: _Entry | None = None
    best_score = 0.0
    for entry in _entries:
        score = _cosine_similarity(query, entry.embedding)
        if score > best_score:
            best_score = score
            best_entry = entry

    if best_entry is not None and best_score >= cfg.faq_similarity_threshold:
        return best_entry.answer
    return None
