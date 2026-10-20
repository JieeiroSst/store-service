from __future__ import annotations

from typing import Optional

from app.domain.ports import TTSPort


class SpeechService:
    def __init__(self, tts: TTSPort) -> None:
        self._tts = tts

    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        return await self._tts.synthesize(text, language)
