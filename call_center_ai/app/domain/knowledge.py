from __future__ import annotations

import re
import unicodedata
from typing import Iterable, List, Set, Tuple

from app.domain.models import KnowledgeEntry

STOPWORDS = {
    "a", "ah", "ak", "ai", "anh", "ban", "bi", "cai", "can", "cho", "chi", "co", "cua", "da", "dang",
    "de", "di", "duoc", "em", "gi", "hay", "hoi", "khong", "la", "lam", "minh", "muon", "nao", "ne", "nhe",
    "nhu", "nay", "oi", "o", "ra", "roi", "sao", "thi", "toi", "tu", "va", "vay", "vang", "voi",
    "vui", "xin", "ve", "cac", "nhung", "mot",
}

TOKEN = re.compile(r"[a-z0-9]+")


def normalize(text: str) -> str:
    text = text.lower().replace("đ", "d")
    text = unicodedata.normalize("NFD", text)
    return "".join(ch for ch in text if unicodedata.category(ch) != "Mn")


def tokens(text: str) -> List[str]:
    return [t for t in TOKEN.findall(normalize(text)) if t not in STOPWORDS]


def _bigrams(words: List[str]) -> Set[Tuple[str, str]]:
    return set(zip(words, words[1:]))


def score(query: str, entry: KnowledgeEntry) -> float:
    q = tokens(query)
    if not q:
        return 0.0
    question = tokens(entry.question)
    keywords = tokens((entry.keywords or "").replace(",", " "))
    doc = set(question) | set(keywords)
    overlap = sum(1.0 for t in set(q) if t in doc)
    keyword_hits = sum(0.5 for t in set(q) if t in keywords)
    phrase_hits = 1.5 * len(_bigrams(q) & (_bigrams(question) | _bigrams(keywords)))
    return (overlap + keyword_hits + phrase_hits) / (len(set(q)) ** 0.5)


def search(
    query: str,
    entries: Iterable[KnowledgeEntry],
    limit: int = 3,
    min_score: float = 0.6,
    relative: float = 0.6,
) -> List[KnowledgeEntry]:
    ranked = sorted(((score(query, e), e) for e in entries), key=lambda pair: pair[0], reverse=True)
    if not ranked or ranked[0][0] < min_score:
        return []
    cutoff = max(min_score, ranked[0][0] * relative)
    return [entry for s, entry in ranked[:limit] if s >= cutoff]
