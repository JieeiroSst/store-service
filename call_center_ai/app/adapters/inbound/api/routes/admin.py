from __future__ import annotations

import uuid
from typing import List, Optional

from fastapi import APIRouter, Depends, HTTPException
from fastapi.responses import Response

from app.adapters.inbound.api import schemas
from app.adapters.inbound.api.auth_deps import require_admin
from app.adapters.inbound.api.dependencies import (
    call_service,
    knowledge_service,
    scenario_service,
    speech_service,
    voice_service,
)
from app.application.calls import CallService
from app.application.knowledge import KnowledgeService
from app.application.scenarios import ScenarioService
from app.application.speech import SpeechService
from app.application.voice import VoiceCallService
from app.domain.models import CallOutcome, CallStatus, KnowledgeEntry, Scenario

router = APIRouter(prefix="/api", dependencies=[Depends(require_admin)])


@router.get("/calls", response_model=List[schemas.CallResponse])
def list_calls(
    skip: int = 0,
    limit: int = 100,
    status: Optional[CallStatus] = None,
    from_number: Optional[str] = None,
    outcome: Optional[CallOutcome] = None,
    service: CallService = Depends(call_service),
):
    return service.list_calls(skip=skip, limit=limit, status=status, from_number=from_number, outcome=outcome)


@router.get("/calls/{call_id}", response_model=schemas.CallHistoryResponse)
def get_call(call_id: int, service: CallService = Depends(call_service)):
    return service.get_call(call_id)


@router.delete("/calls/{call_id}")
def delete_call(call_id: int, service: CallService = Depends(call_service)):
    service.delete_call(call_id)
    return {"message": "Call deleted successfully"}


@router.get("/scenarios", response_model=List[schemas.ScenarioResponse])
def list_scenarios(skip: int = 0, limit: int = 100, service: ScenarioService = Depends(scenario_service)):
    return service.list_scenarios(skip=skip, limit=limit)


@router.post("/scenarios", response_model=schemas.ScenarioResponse)
def create_scenario(payload: schemas.ScenarioCreate, service: ScenarioService = Depends(scenario_service)):
    return service.create_scenario(Scenario(**payload.model_dump()))


@router.post("/scenarios/seed")
def seed_scenarios(
    scenarios: ScenarioService = Depends(scenario_service),
    knowledge: KnowledgeService = Depends(knowledge_service),
):
    return {"scenarios": scenarios.seed_defaults(), "knowledge_entries": len(knowledge.seed_defaults())}


@router.get("/scenarios/{scenario_id}", response_model=schemas.ScenarioResponse)
def get_scenario(scenario_id: int, service: ScenarioService = Depends(scenario_service)):
    return service.get_scenario(scenario_id)


@router.put("/scenarios/{scenario_id}", response_model=schemas.ScenarioResponse)
def update_scenario(
    scenario_id: int,
    payload: schemas.ScenarioUpdate,
    service: ScenarioService = Depends(scenario_service),
):
    return service.update_scenario(scenario_id, payload.model_dump(exclude_unset=True))


@router.delete("/scenarios/{scenario_id}")
def delete_scenario(scenario_id: int, service: ScenarioService = Depends(scenario_service)):
    service.delete_scenario(scenario_id)
    return {"message": "Scenario deleted successfully"}


@router.get("/knowledge", response_model=List[schemas.KnowledgeResponse])
def list_knowledge(
    skip: int = 0,
    limit: int = 100,
    category: Optional[str] = None,
    scenario: Optional[str] = None,
    service: KnowledgeService = Depends(knowledge_service),
):
    return service.list_entries(skip=skip, limit=limit, category=category, scenario=scenario)


@router.get("/knowledge/search", response_model=List[schemas.KnowledgeResponse])
def search_knowledge(
    q: str,
    scenario: Optional[str] = None,
    limit: int = 3,
    service: KnowledgeService = Depends(knowledge_service),
):
    return service.search(q, scenario=scenario, limit=limit)


@router.post("/knowledge", response_model=schemas.KnowledgeResponse)
def create_knowledge(payload: schemas.KnowledgeCreate, service: KnowledgeService = Depends(knowledge_service)):
    return service.create_entry(KnowledgeEntry(**payload.model_dump()))


@router.get("/knowledge/{entry_id}", response_model=schemas.KnowledgeResponse)
def get_knowledge(entry_id: int, service: KnowledgeService = Depends(knowledge_service)):
    return service.get_entry(entry_id)


@router.put("/knowledge/{entry_id}", response_model=schemas.KnowledgeResponse)
def update_knowledge(
    entry_id: int,
    payload: schemas.KnowledgeUpdate,
    service: KnowledgeService = Depends(knowledge_service),
):
    return service.update_entry(entry_id, payload.model_dump(exclude_unset=True))


@router.delete("/knowledge/{entry_id}")
def delete_knowledge(entry_id: int, service: KnowledgeService = Depends(knowledge_service)):
    service.delete_entry(entry_id)
    return {"message": "Knowledge entry deleted successfully"}


@router.get("/customers/{phone_number}", response_model=schemas.CustomerResponse)
def get_customer(phone_number: str, service: CallService = Depends(call_service)):
    return service.get_customer(phone_number)


@router.put("/customers/{phone_number}", response_model=schemas.CustomerResponse)
def update_customer(
    phone_number: str,
    payload: schemas.CustomerUpdate,
    service: CallService = Depends(call_service),
):
    return service.update_customer(phone_number, payload.model_dump(exclude_unset=True))


@router.get("/analytics/statistics", response_model=schemas.CallStatistics)
def get_statistics(days: int = 30, service: CallService = Depends(call_service)):
    return service.statistics(days=days)


@router.post("/simulate", response_model=schemas.SimulateResponse)
async def simulate(
    payload: schemas.SimulateRequest,
    voice: VoiceCallService = Depends(voice_service),
    knowledge: KnowledgeService = Depends(knowledge_service),
):
    greeting = None
    call_sid = payload.call_sid
    if not call_sid:
        call_sid = "SIM-" + uuid.uuid4().hex
        greeting = voice.start_call(
            call_sid,
            payload.from_number,
            payload.to_number,
            scenario_name=payload.scenario,
            notes="Simulation",
        ).text

    if not payload.message:
        return schemas.SimulateResponse(call_sid=call_sid, greeting=greeting, listening=True, ended=False)

    reply = await voice.handle_speech(call_sid, payload.message)
    matched = knowledge.search(payload.message, scenario=payload.scenario)
    return schemas.SimulateResponse(
        call_sid=call_sid,
        greeting=greeting,
        reply=reply.text,
        listening=reply.listen,
        transfer_to=reply.transfer_to,
        ended=not reply.listen,
        knowledge=[k.question for k in matched],
    )


@router.post("/test/tts")
async def test_tts(text: str, language: str = "vi-VN", service: SpeechService = Depends(speech_service)):
    audio = await service.synthesize(text, language)
    if not audio:
        raise HTTPException(status_code=500, detail="TTS failed")
    return Response(content=audio, media_type="audio/mpeg")
