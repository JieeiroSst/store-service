"""Load bank-specific fixed facts from faq.json and expose them as a
system-message string that gets injected into every LLM call.

Keeps exact hotline numbers, interest rates, and fees accurate without
requiring the LLM to memorize bank-specific data. No embedding needed —
this is a synchronous read at startup, not a similarity search.
"""
import json

from . import config

cfg = config.load()
_context: str = ""


def load() -> None:
    global _context
    try:
        with open(cfg.faq_file, "r", encoding="utf-8") as f:
            facts: dict = json.load(f).get("bank_facts", {})
    except (FileNotFoundError, json.JSONDecodeError):
        _context = ""
        return

    if not facts:
        _context = ""
        return

    lines = [
        "Dữ liệu cố định của ngân hàng (luôn dùng chính xác các số liệu dưới đây, không tự suy đoán):"
    ]
    for key, value in facts.items():
        lines.append(f"- {key}: {value}")
    _context = "\n".join(lines)


def get() -> str:
    return _context
