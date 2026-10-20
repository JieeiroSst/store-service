from __future__ import annotations

import enum
from dataclasses import dataclass, field
from datetime import datetime
from typing import List, Optional


class CallStatus(str, enum.Enum):
    INITIATED = "initiated"
    IN_PROGRESS = "in_progress"
    COMPLETED = "completed"
    FAILED = "failed"
    NO_ANSWER = "no_answer"
    BUSY = "busy"


FAILED_STATUSES = (CallStatus.FAILED, CallStatus.NO_ANSWER, CallStatus.BUSY)
TERMINAL_STATUSES = (CallStatus.COMPLETED,) + FAILED_STATUSES


class CallOutcome(str, enum.Enum):
    RESOLVED = "resolved"
    TRANSFERRED = "transferred"
    CALLBACK_REQUESTED = "callback_requested"
    ABANDONED = "abandoned"


class MessageRole(str, enum.Enum):
    USER = "user"
    ASSISTANT = "assistant"
    SYSTEM = "system"


@dataclass
class Message:
    call_id: int
    role: MessageRole
    content: str
    audio_url: Optional[str] = None
    processing_time: Optional[float] = None
    id: Optional[int] = None
    timestamp: Optional[datetime] = None


@dataclass
class Call:
    call_sid: str
    from_number: str
    to_number: str
    status: CallStatus = CallStatus.INITIATED
    duration: int = 0
    start_time: Optional[datetime] = None
    end_time: Optional[datetime] = None
    recording_url: Optional[str] = None
    notes: Optional[str] = None
    scenario: Optional[str] = None
    outcome: Optional[CallOutcome] = None
    id: Optional[int] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None
    messages: List[Message] = field(default_factory=list)

    @property
    def is_finished(self) -> bool:
        return self.status in TERMINAL_STATUSES

    def mark_in_progress(self) -> None:
        if not self.is_finished:
            self.status = CallStatus.IN_PROGRESS

    def complete(self, outcome: CallOutcome, now: datetime) -> None:
        self.status = CallStatus.COMPLETED
        self.end_time = now
        self.outcome = self.outcome or outcome

    def transfer(self) -> None:
        self.outcome = CallOutcome.TRANSFERRED

    def apply_provider_status(self, status: CallStatus, duration: int, now: datetime) -> None:
        if status not in TERMINAL_STATUSES:
            if not self.is_finished:
                self.status = status
            return
        self.status = status
        self.duration = duration
        self.end_time = self.end_time or now
        if status == CallStatus.COMPLETED and self.outcome is None:
            self.outcome = CallOutcome.ABANDONED


@dataclass
class Scenario:
    name: str
    prompt: str
    description: Optional[str] = None
    greeting: Optional[str] = None
    phone_number: Optional[str] = None
    is_default: int = 0
    is_active: int = 1
    id: Optional[int] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None


@dataclass
class KnowledgeEntry:
    question: str
    answer: str
    keywords: Optional[str] = None
    category: Optional[str] = None
    scenario: Optional[str] = None
    is_active: int = 1
    id: Optional[int] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None


@dataclass
class Customer:
    phone_number: str
    name: Optional[str] = None
    email: Optional[str] = None
    notes: Optional[str] = None
    total_calls: int = 0
    last_call_date: Optional[datetime] = None
    id: Optional[int] = None
    created_at: Optional[datetime] = None
    updated_at: Optional[datetime] = None

    def record_call(self, now: datetime) -> None:
        self.total_calls += 1
        self.last_call_date = now


@dataclass(frozen=True)
class CallStatistics:
    total_calls: int
    completed_calls: int
    failed_calls: int
    average_duration: float
    total_messages: int
    resolved_calls: int = 0
    transferred_calls: int = 0
    callback_requests: int = 0
    abandoned_calls: int = 0


@dataclass(frozen=True)
class ChatTurn:
    role: str
    content: str


@dataclass(frozen=True)
class VoiceReply:
    text: str
    listen: bool
    closing: Optional[str] = None
    transfer_to: Optional[str] = None

    @property
    def ends_call(self) -> bool:
        return not self.listen and self.transfer_to is None
