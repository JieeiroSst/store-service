from __future__ import annotations

from datetime import datetime
from typing import List, Optional

from pydantic import BaseModel, ConfigDict, Field

from app.domain.models import CallOutcome, CallStatus, MessageRole


class _FromDomain(BaseModel):
    model_config = ConfigDict(from_attributes=True)


class CallResponse(_FromDomain):
    id: int
    call_sid: str
    from_number: str
    to_number: str
    notes: Optional[str] = None
    status: CallStatus
    duration: int
    start_time: datetime
    end_time: Optional[datetime]
    recording_url: Optional[str]
    scenario: Optional[str] = None
    outcome: Optional[CallOutcome] = None
    created_at: datetime
    updated_at: datetime


class MessageResponse(_FromDomain):
    id: int
    call_id: int
    role: MessageRole
    content: str
    audio_url: Optional[str] = None
    timestamp: datetime
    processing_time: Optional[float]


class CallHistoryResponse(CallResponse):
    messages: List[MessageResponse] = []


class ScenarioCreate(BaseModel):
    name: str
    description: Optional[str] = None
    prompt: str
    greeting: Optional[str] = None
    phone_number: Optional[str] = None
    is_default: int = 0
    is_active: int = 1


class ScenarioUpdate(BaseModel):
    name: Optional[str] = None
    description: Optional[str] = None
    prompt: Optional[str] = None
    greeting: Optional[str] = None
    phone_number: Optional[str] = None
    is_default: Optional[int] = None
    is_active: Optional[int] = None


class ScenarioResponse(_FromDomain):
    id: int
    name: str
    description: Optional[str] = None
    prompt: str
    greeting: Optional[str] = None
    phone_number: Optional[str] = None
    is_default: int = 0
    is_active: int
    created_at: datetime
    updated_at: datetime


class KnowledgeCreate(BaseModel):
    question: str
    answer: str
    keywords: Optional[str] = None
    category: Optional[str] = None
    scenario: Optional[str] = None
    is_active: int = 1


class KnowledgeUpdate(BaseModel):
    question: Optional[str] = None
    answer: Optional[str] = None
    keywords: Optional[str] = None
    category: Optional[str] = None
    scenario: Optional[str] = None
    is_active: Optional[int] = None


class KnowledgeResponse(_FromDomain):
    id: int
    question: str
    answer: str
    keywords: Optional[str] = None
    category: Optional[str] = None
    scenario: Optional[str] = None
    is_active: int
    created_at: datetime
    updated_at: datetime


class CustomerResponse(_FromDomain):
    id: int
    phone_number: str
    name: Optional[str] = None
    email: Optional[str] = None
    notes: Optional[str] = None
    total_calls: int
    last_call_date: Optional[datetime]
    created_at: datetime
    updated_at: datetime


class CustomerUpdate(BaseModel):
    name: Optional[str] = None
    email: Optional[str] = None
    notes: Optional[str] = None


class CallStatistics(_FromDomain):
    total_calls: int
    completed_calls: int
    failed_calls: int
    average_duration: float
    total_messages: int
    resolved_calls: int
    transferred_calls: int
    callback_requests: int
    abandoned_calls: int


class SimulateRequest(BaseModel):
    message: str = ""
    call_sid: Optional[str] = None
    scenario: Optional[str] = None
    from_number: str = "simulator"
    to_number: str = "simulator"


class SimulateResponse(BaseModel):
    call_sid: str
    greeting: Optional[str] = None
    reply: Optional[str] = None
    listening: bool
    transfer_to: Optional[str] = None
    ended: bool
    knowledge: List[str] = Field(default_factory=list)
