from __future__ import annotations

import copy
import itertools
from datetime import datetime
from typing import Dict, List, Optional

from app.config import Settings
from app.container import Container
from app.domain.models import (
    FAILED_STATUSES,
    Call,
    CallOutcome,
    CallStatistics,
    CallStatus,
    ChatTurn,
    Customer,
    KnowledgeEntry,
    Message,
    Scenario,
)
from app.domain.ports import (
    CallRepositoryPort,
    CustomerRepositoryPort,
    KnowledgeRepositoryPort,
    LLMPort,
    MessageRepositoryPort,
    ScenarioRepositoryPort,
    TTSPort,
)


class _Store:
    def __init__(self) -> None:
        self.rows: Dict[int, object] = {}
        self._ids = itertools.count(1)

    def insert(self, entity):
        entity = copy.deepcopy(entity)
        entity.id = next(self._ids)
        now = datetime.utcnow()
        for name in ("created_at", "updated_at", "start_time", "timestamp"):
            if hasattr(entity, name) and getattr(entity, name) is None:
                setattr(entity, name, now)
        self.rows[entity.id] = entity
        return copy.deepcopy(entity)

    def update(self, entity):
        if entity.id not in self.rows:
            return entity
        entity = copy.deepcopy(entity)
        entity.updated_at = datetime.utcnow()
        self.rows[entity.id] = entity
        return copy.deepcopy(entity)

    def get(self, row_id):
        row = self.rows.get(row_id)
        return copy.deepcopy(row) if row else None

    def find(self, **kwargs):
        for row in self.rows.values():
            if all(getattr(row, k) == v for k, v in kwargs.items()):
                return copy.deepcopy(row)
        return None


class FakeMessageRepository(MessageRepositoryPort):
    def __init__(self) -> None:
        self.store = _Store()

    def add(self, message: Message) -> Message:
        return self.store.insert(message)

    def list_by_call(self, call_id: int) -> List[Message]:
        return [copy.deepcopy(m) for m in self.store.rows.values() if m.call_id == call_id]


class FakeCallRepository(CallRepositoryPort):
    def __init__(self, messages: FakeMessageRepository) -> None:
        self.store = _Store()
        self._messages = messages

    def add(self, call: Call) -> Call:
        return self.store.insert(call)

    def get(self, call_id: int) -> Optional[Call]:
        return self.store.get(call_id)

    def get_by_sid(self, call_sid: str) -> Optional[Call]:
        return self.store.find(call_sid=call_sid)

    def get_with_messages(self, call_id: int) -> Optional[Call]:
        call = self.store.get(call_id)
        if call:
            call.messages = self._messages.list_by_call(call_id)
        return call

    def list(self, skip, limit, status=None, from_number=None, outcome=None) -> List[Call]:
        calls = [
            c for c in self.store.rows.values()
            if (status is None or c.status == status)
            and (from_number is None or c.from_number == from_number)
            and (outcome is None or c.outcome == outcome)
        ]
        calls.sort(key=lambda c: c.created_at, reverse=True)
        return copy.deepcopy(calls[skip:skip + limit])

    def save(self, call: Call) -> Call:
        return self.store.update(call)

    def delete(self, call_id: int) -> bool:
        if self.store.rows.pop(call_id, None) is None:
            return False
        for mid in [m.id for m in self._messages.store.rows.values() if m.call_id == call_id]:
            del self._messages.store.rows[mid]
        return True

    def statistics(self, since: datetime) -> CallStatistics:
        recent = [c for c in self.store.rows.values() if c.created_at >= since]
        completed = [c for c in recent if c.status == CallStatus.COMPLETED]
        ids = {c.id for c in recent}
        return CallStatistics(
            total_calls=len(recent),
            completed_calls=len(completed),
            failed_calls=len([c for c in recent if c.status in FAILED_STATUSES]),
            average_duration=round(sum(c.duration for c in completed) / len(completed), 2) if completed else 0.0,
            total_messages=len([m for m in self._messages.store.rows.values() if m.call_id in ids]),
            resolved_calls=len([c for c in recent if c.outcome == CallOutcome.RESOLVED]),
            transferred_calls=len([c for c in recent if c.outcome == CallOutcome.TRANSFERRED]),
            callback_requests=len([c for c in recent if c.outcome == CallOutcome.CALLBACK_REQUESTED]),
            abandoned_calls=len([c for c in recent if c.outcome == CallOutcome.ABANDONED]),
        )


class FakeScenarioRepository(ScenarioRepositoryPort):
    def __init__(self) -> None:
        self.store = _Store()

    def add(self, scenario: Scenario) -> Scenario:
        return self.store.insert(scenario)

    def get(self, scenario_id: int) -> Optional[Scenario]:
        return self.store.get(scenario_id)

    def get_by_name(self, name: str) -> Optional[Scenario]:
        return self.store.find(name=name)

    def get_by_phone(self, phone_number: str) -> Optional[Scenario]:
        return self.store.find(phone_number=phone_number)

    def get_default(self) -> Optional[Scenario]:
        return self.store.find(is_default=1)

    def list(self, skip, limit, active_only=True) -> List[Scenario]:
        rows = [s for s in self.store.rows.values() if not active_only or s.is_active == 1]
        return copy.deepcopy(rows[skip:skip + limit])

    def save(self, scenario: Scenario) -> Scenario:
        return self.store.update(scenario)

    def delete(self, scenario_id: int) -> bool:
        return self.store.rows.pop(scenario_id, None) is not None


class FakeKnowledgeRepository(KnowledgeRepositoryPort):
    def __init__(self) -> None:
        self.store = _Store()

    def add(self, entry: KnowledgeEntry) -> KnowledgeEntry:
        return self.store.insert(entry)

    def get(self, entry_id: int) -> Optional[KnowledgeEntry]:
        return self.store.get(entry_id)

    def find_by_question(self, question: str) -> Optional[KnowledgeEntry]:
        return self.store.find(question=question)

    def list(self, skip, limit, category=None, scenario=None, active_only=False) -> List[KnowledgeEntry]:
        rows = [
            e for e in self.store.rows.values()
            if (category is None or e.category == category)
            and (scenario is None or e.scenario == scenario)
            and (not active_only or e.is_active == 1)
        ]
        return copy.deepcopy(rows[skip:skip + limit])

    def list_for_scenario(self, scenario: str) -> List[KnowledgeEntry]:
        return copy.deepcopy(
            [e for e in self.store.rows.values() if e.is_active == 1 and e.scenario in (None, "", scenario)]
        )

    def save(self, entry: KnowledgeEntry) -> KnowledgeEntry:
        return self.store.update(entry)

    def delete(self, entry_id: int) -> bool:
        return self.store.rows.pop(entry_id, None) is not None


class FakeCustomerRepository(CustomerRepositoryPort):
    def __init__(self) -> None:
        self.store = _Store()

    def add(self, customer: Customer) -> Customer:
        return self.store.insert(customer)

    def get_by_phone(self, phone_number: str) -> Optional[Customer]:
        return self.store.find(phone_number=phone_number)

    def save(self, customer: Customer) -> Customer:
        return self.store.update(customer)


class FakeLLM(LLMPort):
    def __init__(self, answer: Optional[str] = "Dạ, tôi có thể giúp gì thêm?") -> None:
        self.answer = answer
        self.calls: List[List[ChatTurn]] = []
        self.prompts: List[Optional[str]] = []

    async def generate(self, history, system_prompt=None):
        self.calls.append(list(history))
        self.prompts.append(system_prompt)
        return self.answer


class ScriptedLLM(LLMPort):
    def __init__(self, *answers: Optional[str]) -> None:
        self.answers = list(answers)
        self.prompts: List[Optional[str]] = []

    async def generate(self, history, system_prompt=None):
        self.prompts.append(system_prompt)
        return self.answers.pop(0)


class FailingLLM(LLMPort):
    async def generate(self, history, system_prompt=None):
        raise RuntimeError("boom")


class FakeTTS(TTSPort):
    def __init__(self, audio: Optional[bytes] = b"mp3") -> None:
        self.audio = audio

    async def synthesize(self, text, language="vi-VN"):
        return self.audio


def fake_container(
    llm: Optional[LLMPort] = None,
    tts: Optional[TTSPort] = None,
    **settings,
) -> Container:
    messages = FakeMessageRepository()
    return Container(
        settings=Settings(_env_file=None, **settings),
        calls=FakeCallRepository(messages),
        messages=messages,
        customers=FakeCustomerRepository(),
        scenarios=FakeScenarioRepository(),
        knowledge=FakeKnowledgeRepository(),
        llm=llm or FakeLLM(),
        tts=tts or FakeTTS(),
    )
