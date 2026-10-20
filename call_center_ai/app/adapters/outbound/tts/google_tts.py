from __future__ import annotations

import asyncio
from typing import Optional

from app.domain.ports import TTSPort


class GoogleTTS(TTSPort):
    VOICES = {
        "vi-VN": "vi-VN-Standard-A",
        "en-US": "en-US-Standard-C",
    }

    def __init__(self) -> None:
        from google.cloud import texttospeech

        self._tts = texttospeech
        self._client = texttospeech.TextToSpeechClient()

    async def synthesize(self, text: str, language: str = "vi-VN") -> Optional[bytes]:
        return await asyncio.to_thread(self._synthesize, text, language)

    def _synthesize(self, text: str, language: str) -> bytes:
        tts = self._tts
        response = self._client.synthesize_speech(
            input=tts.SynthesisInput(text=text),
            voice=tts.VoiceSelectionParams(
                language_code=language,
                name=self.VOICES.get(language, self.VOICES["vi-VN"]),
                ssml_gender=tts.SsmlVoiceGender.FEMALE,
            ),
            audio_config=tts.AudioConfig(
                audio_encoding=tts.AudioEncoding.MP3,
                speaking_rate=1.0,
                pitch=0.0,
            ),
        )
        return response.audio_content
