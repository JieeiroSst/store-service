from __future__ import annotations

from abc import ABC, abstractmethod
from typing import Optional


class TTSPort(ABC):
    @abstractmethod
    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        raise NotImplementedError
