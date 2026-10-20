from __future__ import annotations

import logging
from typing import Any, Dict, List, Optional

from app.domain.models import ChatTurn
from app.domain.ports import LLMPort

log = logging.getLogger(__name__)

DEFAULT_SYSTEM_PROMPT = "Bạn là trợ lý AI thông minh và hữu ích."
CALL_CONNECTED = "[Khách hàng vừa gọi đến tổng đài]"
FALLBACK_BETA = "server-side-fallback-2026-07-01"


class AnthropicLLM(LLMPort):
    def __init__(
        self,
        api_key: str,
        model: str,
        max_tokens: int = 2048,
        effort: str = "low",
        timeout: float = 10.0,
        max_retries: int = 0,
        refusal_fallback: bool = True,
        client: Any = None,
    ) -> None:
        if client is None:
            from anthropic import AsyncAnthropic

            client = AsyncAnthropic(api_key=api_key, timeout=timeout, max_retries=max_retries)
        self._client = client
        self._model = model
        self._max_tokens = max_tokens
        self._effort = effort
        self._refusal_fallback = refusal_fallback

    async def generate(self, history: List[ChatTurn], system_prompt: Optional[str] = None) -> Optional[str]:
        request: Dict[str, Any] = dict(
            model=self._model,
            max_tokens=self._max_tokens,
            system=system_prompt or DEFAULT_SYSTEM_PROMPT,
            messages=to_messages(history),
            output_config={"effort": self._effort},
        )
        if self._refusal_fallback:
            response = await self._client.beta.messages.create(
                betas=[FALLBACK_BETA], fallbacks="default", **request
            )
        else:
            response = await self._client.messages.create(**request)

        if response.stop_reason == "refusal":
            log.warning("Claude refused the request: %s", getattr(response, "stop_details", None))
            return None
        text = " ".join(block.text for block in response.content if block.type == "text").strip()
        return text or None


def to_messages(history: List[ChatTurn]) -> List[Dict[str, str]]:
    messages = [{"role": t.role, "content": t.content} for t in history if t.role in ("user", "assistant")]
    if not messages or messages[0]["role"] != "user":
        messages.insert(0, {"role": "user", "content": CALL_CONNECTED})
    return messages
