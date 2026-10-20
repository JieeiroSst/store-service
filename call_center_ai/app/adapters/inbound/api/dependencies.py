from __future__ import annotations

from fastapi import Request

from app.application.calls import CallService
from app.application.knowledge import KnowledgeService
from app.application.scenarios import ScenarioService
from app.application.speech import SpeechService
from app.application.voice import VoiceCallService
from app.container import Container


def container(request: Request) -> Container:
    return request.app.state.container


def voice_service(request: Request) -> VoiceCallService:
    return container(request).voice_service()


def call_service(request: Request) -> CallService:
    return container(request).call_service()


def scenario_service(request: Request) -> ScenarioService:
    return container(request).scenario_service()


def knowledge_service(request: Request) -> KnowledgeService:
    return container(request).knowledge_service()


def speech_service(request: Request) -> SpeechService:
    return container(request).speech_service()
