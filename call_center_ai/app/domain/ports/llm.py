from __future__ import annotations

from abc import ABC, abstractmethod
from typing import List, Optional

from app.domain.models import ChatTurn


class LLMPort(ABC):
    @abstractmethod
    async def generate(self, history: List[ChatTurn], system_prompt: Optional[str] = None) -> Optional[str]:
        raise NotImplementedError
