from __future__ import annotations

import logging
from typing import List, Optional, Sequence

from app.domain.models import ChatTurn
from app.domain.ports import LLMPort

log = logging.getLogger(__name__)


class FallbackLLM(LLMPort):
    def __init__(self, providers: Sequence[LLMPort]) -> None:
        self._providers = list(providers)

    def __len__(self) -> int:
        return len(self._providers)

    async def generate(self, history: List[ChatTurn], system_prompt: Optional[str] = None) -> Optional[str]:
        for provider in self._providers:
            try:
                answer = await provider.generate(history, system_prompt)
            except Exception as e:
                log.warning("%s failed: %s, trying next provider", type(provider).__name__, e)
                continue
            if answer:
                return answer
        log.error("all LLM providers failed")
        return None
