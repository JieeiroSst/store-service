"""The one place to wire in a real backend.

routes.py calls answer() for both /api/generate and /api/chat — replace the
body below with a real call (an LLM API, an internal service, a database
lookup, whatever should actually answer the question) and nothing else in
this project needs to change.

This implementation forwards the question to a real Ollama server
(OLLAMA_BASE_URL / OLLAMA_TARGET_MODEL) and relays its answer.
"""
import httpx

from .. import config

cfg = config.load()


class BackendError(Exception):
    """Raised when the upstream Ollama server can't answer the question."""

    def __init__(self, message: str, status_code: int = 502):
        super().__init__(message)
        self.status_code = status_code


async def answer(question: str) -> str:
    try:
        async with httpx.AsyncClient(base_url=cfg.ollama_base_url, timeout=120.0) as client:
            resp = await client.post(
                "/api/chat",
                json={
                    "model": cfg.ollama_target_model,
                    "messages": [{"role": "user", "content": question}],
                    "stream": False,
                },
            )
    except httpx.RequestError as exc:
        raise BackendError(f"can't reach Ollama at {cfg.ollama_base_url}: {exc}") from exc

    if resp.is_error:
        try:
            detail = resp.json().get("error", resp.text)
        except ValueError:
            detail = resp.text
        raise BackendError(detail, status_code=resp.status_code)

    return resp.json()["message"]["content"]
