from __future__ import annotations

from typing import List, Optional

from app.domain.models import ChatTurn
from app.domain.ports import LLMPort


class OpenAILLM(LLMPort):
    def __init__(
        self,
        api_key: str,
        model: str,
        max_tokens: int = 500,
        temperature: float = 0.7,
        timeout: float = 10.0,
        max_retries: int = 0,
    ) -> None:
        from openai import AsyncOpenAI

        self._client = AsyncOpenAI(api_key=api_key, timeout=timeout, max_retries=max_retries)
        self._model = model
        self._max_tokens = max_tokens
        self._temperature = temperature

    async def generate(self, history: List[ChatTurn], system_prompt: Optional[str] = None) -> Optional[str]:
        messages = [{"role": "system", "content": system_prompt}] if system_prompt else []
        messages.extend({"role": t.role, "content": t.content} for t in history)
        response = await self._client.chat.completions.create(
            model=self._model,
            messages=messages,
            temperature=self._temperature,
            max_tokens=self._max_tokens,
        )
        return response.choices[0].message.content
