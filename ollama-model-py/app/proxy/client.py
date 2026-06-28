"""The one place to wire in a real backend.

answer() is called for both /api/generate and /api/chat whenever the FAQ
check (see ../faq.py) finds no match — replace its body with a real call (an
LLM API, an internal service, a database lookup, whatever should actually
answer the question) and nothing else in this project needs to change.

embed() is used by ../faq.py to semantically match incoming questions
against the scripted FAQ.

This implementation forwards both to a real Ollama server (OLLAMA_BASE_URL /
OLLAMA_TARGET_MODEL / OLLAMA_EMBED_MODEL) and relays the response.
"""
import httpx

from .. import bank_context, config

cfg = config.load()


class BackendError(Exception):
    """Raised when an upstream service (Ollama, or an internal API — see
    ../internal_api/client.py) can't answer the question."""

    def __init__(self, message: str, status_code: int = 502):
        super().__init__(message)
        self.status_code = status_code


async def request_json(
    base_url: str,
    method: str,
    path: str,
    *,
    json: dict | None = None,
    timeout: float = 120.0,
    error_label: str = "backend",
) -> dict:
    """Shared request/error-handling for any JSON-speaking upstream — used
    below for Ollama itself, and by ../internal_api/client.py for internal
    services. Raises BackendError on a network failure or an error response,
    so every caller gets the same error shape for free.
    """
    try:
        async with httpx.AsyncClient(base_url=base_url, timeout=timeout) as http_client:
            resp = await http_client.request(method, path, json=json)
    except httpx.RequestError as exc:
        raise BackendError(f"can't reach {error_label} at {base_url}: {exc}") from exc

    if resp.is_error:
        try:
            detail = resp.json().get("error", resp.text)
        except ValueError:
            detail = resp.text
        raise BackendError(detail, status_code=resp.status_code)

    return resp.json()


async def answer(messages: list[dict]) -> str:
    ctx = bank_context.get()
    payload = [{"role": "system", "content": ctx}] + messages if ctx else messages
    data = await request_json(
        cfg.ollama_base_url,
        "POST",
        "/api/chat",
        json={
            "model": cfg.ollama_target_model,
            "messages": payload,
            "stream": False,
        },
        error_label="Ollama",
    )
    return data["message"]["content"]


async def answer_with_context(question: str, context: str) -> str:
    """Answer a question using only the provided context (from PDF chunks).
    Instructs the model to be concise and not fabricate beyond the context.
    """
    prompt = (
        "Dựa trên thông tin tài liệu dưới đây, hãy trả lời ngắn gọn câu hỏi của khách. "
        "Chỉ dùng thông tin được cung cấp; nếu tài liệu không đủ để trả lời, hãy nói rõ.\n\n"
        f"Tài liệu:\n{context}\n\n"
        f"Câu hỏi: {question}"
    )
    data = await request_json(
        cfg.ollama_base_url,
        "POST",
        "/api/chat",
        json={
            "model": cfg.ollama_target_model,
            "messages": [{"role": "user", "content": prompt}],
            "stream": False,
        },
        error_label="Ollama",
    )
    return data["message"]["content"]


async def embed(inputs: list[str]) -> list[list[float]]:
    data = await request_json(
        cfg.ollama_base_url,
        "POST",
        "/api/embed",
        json={"model": cfg.ollama_embed_model, "input": inputs},
        error_label="Ollama",
    )
    return data["embeddings"]
