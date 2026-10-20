from __future__ import annotations

import logging
from typing import Optional

import httpx

from app.domain.ports import TTSPort

log = logging.getLogger(__name__)


class ElevenLabsTTS(TTSPort):
    def __init__(self, api_key: str, voice_id: str, timeout: float = 30.0) -> None:
        self._api_key = api_key
        self._voice_id = voice_id
        self._timeout = timeout

    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        async with httpx.AsyncClient(timeout=self._timeout) as client:
            response = await client.post(
                f"https://api.elevenlabs.io/v1/text-to-speech/{self._voice_id}",
                headers={
                    "Accept": "audio/mpeg",
                    "Content-Type": "application/json",
                    "xi-api-key": self._api_key,
                },
                json={
                    "text": text,
                    "model_id": "eleven_multilingual_v2",
                    "voice_settings": {"stability": 0.5, "similarity_boost": 0.5},
                },
            )
        if response.status_code != 200:
            log.error("ElevenLabs TTS error: %s - %s", response.status_code, response.text)
            return None
        return response.content
