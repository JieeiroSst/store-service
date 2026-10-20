from __future__ import annotations

import asyncio
import io
from typing import Optional

from app.domain.ports import TTSPort


class GTTS(TTSPort):
    def __init__(self) -> None:
        from gtts import gTTS

        self._gtts = gTTS

    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        return await asyncio.to_thread(self._synthesize, text, language)

    def _synthesize(self, text: str, language: str) -> bytes:
        buffer = io.BytesIO()
        self._gtts(text=text, lang=language.split("-")[0], slow=False).write_to_fp(buffer)
        return buffer.getvalue()
