"""Recognizes questions that must be answered by calling an internal API
(see ./client.py) instead of the scripted FAQ (../faq.py) or the LLM
(../proxy/client.py:answer). Real questions are rarely worded exactly like
the examples below, so matching is semantic (Ollama embeddings + cosine
similarity) — the same approach ../faq.py uses, against a per-intent set of
example questions instead of a single canned answer.

load() embeds every intent's example questions once at startup (see
main.py's lifespan). match() embeds the incoming question and, if it scores
at or above INTERNAL_API_SIMILARITY_THRESHOLD against some intent's
examples, calls that intent's handler and returns its answer — callers fall
through to the LLM otherwise.

Only the *recognition* step is best-effort: if the embedder is briefly
unreachable, matching is skipped for that question and it falls through
(same as ../faq.py). Once a question IS recognized as e.g. "asking for my
points", a failure to actually reach the internal API is not swallowed —
that BackendError propagates up to routes.py, which already returns a
proper error response for it. Guessing a point balance from the LLM would
be worse than a clear error.
"""
import math
from dataclasses import dataclass
from typing import Awaitable, Callable

from .. import config
from ..proxy import client as ollama_client
from . import client as internal_client

cfg = config.load()

# ai-agent-system forwards the caller's X-User-Id header as user_id on
# ChatRequest/GenerateRequest (see ../wire.py and routes.py's _resolve()).
# This is only the fallback for callers that don't send one (e.g. a direct
# curl against this service, or an ai-agent-system request with no
# X-User-Id) — real callers should always pass a real user_id.
PLACEHOLDER_USER_ID = "demo-user"

_Handler = Callable[[str], Awaitable[str]]  # takes a user_id, returns the answer


async def _handle_points_balance(user_id: str) -> str:
    data = await internal_client.get_points_balance(user_id)
    return (
        f"You currently have {data.get('points_active', 0)} active points "
        f"({data.get('points_pending', 0)} pending, "
        f"{data.get('total_points', 0)} total)."
    )


@dataclass
class _Intent:
    name: str
    examples: list[str]
    handler: _Handler
    keywords: frozenset[str]  # at least one must appear in the question (lowercased)


_REGISTRY: list[_Intent] = [
    _Intent(
        name="points_balance",
        examples=[
            "How many points do I have?",
            "What's my current point balance?",
            "Check my reward points",
            "How many reward points are in my account?",
            "Tôi còn bao nhiêu điểm?",
            "Số điểm hiện tại của tôi là bao nhiêu?",
            "Kiểm tra điểm thưởng của tôi",
        ],
        handler=_handle_points_balance,
        keywords=frozenset({"điểm", "points", "reward", "point"}),
    ),
]


@dataclass
class _Example:
    text: str
    intent: _Intent
    embedding: list[float]


_examples: list[_Example] = []


def _cosine_similarity(a: list[float], b: list[float]) -> float:
    dot = sum(x * y for x, y in zip(a, b))
    norm_a = math.sqrt(sum(x * x for x in a))
    norm_b = math.sqrt(sum(y * y for y in b))
    if norm_a == 0 or norm_b == 0:
        return 0.0
    return dot / (norm_a * norm_b)


async def load() -> None:
    all_examples = [ex for intent in _REGISTRY for ex in intent.examples]
    if not all_examples:
        _examples.clear()
        return

    embeddings = await ollama_client.embed(all_examples)
    if len(embeddings) != len(all_examples):
        raise RuntimeError(
            f"internal_api: embedder returned {len(embeddings)} embeddings "
            f"for {len(all_examples)} example questions"
        )

    _examples.clear()
    i = 0
    for intent in _REGISTRY:
        for ex in intent.examples:
            _examples.append(_Example(ex, intent, embeddings[i]))
            i += 1


async def match(question: str, user_id: str | None = None) -> str | None:
    if not _examples:
        return None

    # Fast keyword pre-filter: skip cosine similarity if the question contains
    # none of any registered intent's required keywords. Prevents false
    # positives from the embedding model conflating unrelated banking questions.
    q_lower = question.lower()
    eligible = [ex for ex in _examples if ex.intent.keywords & set(q_lower.split())]
    if not eligible:
        return None

    try:
        embeddings = await ollama_client.embed([question])
    except ollama_client.BackendError:
        return None
    if not embeddings:
        return None
    query = embeddings[0]

    best_example: _Example | None = None
    best_score = 0.0
    for ex in eligible:
        score = _cosine_similarity(query, ex.embedding)
        if score > best_score:
            best_score = score
            best_example = ex

    if best_example is not None and best_score >= cfg.internal_api_similarity_threshold:
        return await best_example.intent.handler(user_id or PLACEHOLDER_USER_ID)
    return None
