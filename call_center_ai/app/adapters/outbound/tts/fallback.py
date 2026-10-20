from __future__ import annotations

import logging
from typing import Optional, Sequence

from app.domain.ports import TTSPort

log = logging.getLogger(__name__)


class FallbackTTS(TTSPort):
    def __init__(self, providers: Sequence[TTSPort]) -> None:
        self._providers = list(providers)

    def __len__(self) -> int:
        return len(self._providers)

    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        for provider in self._providers:
            try:
                audio = await provider.synthesize(text, language)
            except Exception as e:
                log.warning("%s failed: %s, trying next provider", type(provider).__name__, e)
                continue
            if audio:
                return audio
        log.error("all TTS providers failed")
        return None
