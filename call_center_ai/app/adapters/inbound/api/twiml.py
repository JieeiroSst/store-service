from __future__ import annotations

from fastapi.responses import Response
from twilio.twiml.voice_response import Gather, VoiceResponse

from app.domain.models import VoiceReply

SPEECH_ACTION = "/voice/process-speech"


def render(reply: VoiceReply, language: str, timeout: int) -> Response:
    response = VoiceResponse()
    if reply.transfer_to:
        response.say(reply.text, language=language)
        response.dial(reply.transfer_to)
        return Response(content=str(response), media_type="application/xml")
    if reply.listen:
        gather = Gather(
            input="speech",
            action=SPEECH_ACTION,
            language=language,
            timeout=timeout,
            speech_timeout="auto",
        )
        gather.say(reply.text, language=language)
        response.append(gather)
    else:
        response.say(reply.text, language=language)
    if reply.closing:
        response.say(reply.closing, language=language)
    if not reply.listen or reply.closing:
        response.hangup()
    return Response(content=str(response), media_type="application/xml")
