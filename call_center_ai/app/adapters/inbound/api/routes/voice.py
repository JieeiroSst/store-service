from __future__ import annotations

import logging

from fastapi import APIRouter, Depends, Request
from fastapi.responses import Response

from app.adapters.inbound.api import twiml
from app.adapters.inbound.api.dependencies import container, voice_service
from app.adapters.inbound.api.twilio_auth import verify_twilio
from app.application.voice import VoiceCallService
from app.domain.models import CallStatus, VoiceReply

log = logging.getLogger(__name__)

router = APIRouter(prefix="/voice", dependencies=[Depends(verify_twilio)])

TWILIO_STATUSES = {
    "queued": CallStatus.INITIATED,
    "initiated": CallStatus.INITIATED,
    "ringing": CallStatus.INITIATED,
    "in-progress": CallStatus.IN_PROGRESS,
    "answered": CallStatus.IN_PROGRESS,
    "completed": CallStatus.COMPLETED,
    "busy": CallStatus.BUSY,
    "no-answer": CallStatus.NO_ANSWER,
    "failed": CallStatus.FAILED,
    "canceled": CallStatus.FAILED,
}


def _render(request: Request, reply: VoiceReply) -> Response:
    settings = container(request).settings
    return twiml.render(reply, settings.voice_language, settings.voice_gather_timeout)


@router.post("/incoming")
async def incoming_call(request: Request, service: VoiceCallService = Depends(voice_service)) -> Response:
    form = await request.form()
    reply = service.start_call(form.get("CallSid"), form.get("From"), form.get("To"))
    return _render(request, reply)


@router.post("/process-speech")
async def process_speech(request: Request, service: VoiceCallService = Depends(voice_service)) -> Response:
    form = await request.form()
    reply = await service.handle_speech(form.get("CallSid"), form.get("SpeechResult", ""))
    return _render(request, reply)


@router.post("/status")
async def status_callback(request: Request, service: VoiceCallService = Depends(voice_service)) -> dict:
    form = await request.form()
    status = TWILIO_STATUSES.get(form.get("CallStatus"))
    if status is None:
        log.warning("ignoring unknown Twilio status %r for %s", form.get("CallStatus"), form.get("CallSid"))
        return {"status": "ignored"}
    service.update_status(form.get("CallSid"), status, int(form.get("CallDuration", 0) or 0))
    return {"status": "ok"}
