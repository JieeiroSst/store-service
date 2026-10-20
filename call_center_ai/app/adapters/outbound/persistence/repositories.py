from __future__ import annotations

from datetime import datetime
from typing import Any, List, Optional, Type

from sqlalchemy import func, or_, select
from sqlalchemy.orm import Session, selectinload, sessionmaker

from app.adapters.outbound.persistence.orm import CallRow, CustomerRow, KnowledgeRow, MessageRow, ScenarioRow
from app.domain.models import (
    FAILED_STATUSES,
    Call,
    CallOutcome,
    CallStatistics,
    CallStatus,
    Customer,
    KnowledgeEntry,
    Message,
    Scenario,
)
from app.domain.ports import (
    CallRepositoryPort,
    CustomerRepositoryPort,
    KnowledgeRepositoryPort,
    MessageRepositoryPort,
    ScenarioRepositoryPort,
)

CALL_FIELDS = (
    "call_sid", "from_number", "to_number", "status", "duration",
    "start_time", "end_time", "recording_url", "notes", "scenario", "outcome",
)
MESSAGE_FIELDS = ("call_id", "role", "content", "audio_url", "processing_time")
SCENARIO_FIELDS = ("name", "description", "prompt", "greeting", "phone_number", "is_default", "is_active")
KNOWLEDGE_FIELDS = ("question", "answer", "keywords", "category", "scenario", "is_active")
CUSTOMER_FIELDS = ("phone_number", "name", "email", "notes", "total_calls", "last_call_date")


def _copy(source: Any, target: Any, fields, skip_none: bool = False) -> None:
    for name in fields:
        value = getattr(source, name)
        if value is not None or not skip_none:
            setattr(target, name, value)


def _message(row: MessageRow) -> Message:
    return Message(
        id=row.id,
        call_id=row.call_id,
        role=row.role,
        content=row.content,
        audio_url=row.audio_url,
        processing_time=row.processing_time,
        timestamp=row.timestamp,
    )


def _call(row: CallRow, with_messages: bool = False) -> Call:
    return Call(
        id=row.id,
        call_sid=row.call_sid,
        from_number=row.from_number,
        to_number=row.to_number,
        status=row.status,
        duration=row.duration,
        start_time=row.start_time,
        end_time=row.end_time,
        recording_url=row.recording_url,
        notes=row.notes,
        scenario=row.scenario,
        outcome=row.outcome,
        created_at=row.created_at,
        updated_at=row.updated_at,
        messages=[_message(m) for m in row.messages] if with_messages else [],
    )


def _scenario(row: ScenarioRow) -> Scenario:
    return Scenario(
        id=row.id,
        name=row.name,
        description=row.description,
        prompt=row.prompt,
        greeting=row.greeting,
        phone_number=row.phone_number,
        is_default=row.is_default or 0,
        is_active=row.is_active,
        created_at=row.created_at,
        updated_at=row.updated_at,
    )


def _knowledge(row: KnowledgeRow) -> KnowledgeEntry:
    return KnowledgeEntry(
        id=row.id,
        question=row.question,
        answer=row.answer,
        keywords=row.keywords,
        category=row.category,
        scenario=row.scenario,
        is_active=row.is_active,
        created_at=row.created_at,
        updated_at=row.updated_at,
    )


def _customer(row: CustomerRow) -> Customer:
    return Customer(
        id=row.id,
        phone_number=row.phone_number,
        name=row.name,
        email=row.email,
        notes=row.notes,
        total_calls=row.total_calls,
        last_call_date=row.last_call_date,
        created_at=row.created_at,
        updated_at=row.updated_at,
    )


class _Repository:
    def __init__(self, sessions: sessionmaker) -> None:
        self._sessions = sessions

    def _insert(self, row: Any) -> Any:
        with self._sessions() as session, session.begin():
            session.add(row)
            session.flush()
            session.refresh(row)
        return row

    def _update(self, model: Type[Any], entity: Any, fields) -> Optional[Any]:
        with self._sessions() as session, session.begin():
            row = session.get(model, entity.id)
            if row is None:
                return None
            _copy(entity, row, fields)
            row.updated_at = datetime.utcnow()
            session.flush()
            session.refresh(row)
        return row

    def _delete(self, model: Type[Any], row_id: int) -> bool:
        with self._sessions() as session, session.begin():
            row = session.get(model, row_id)
            if row is None:
                return False
            session.delete(row)
        return True

    def _one(self, statement) -> Optional[Any]:
        with self._sessions() as session:
            return session.scalars(statement).first()

    def _all(self, statement) -> List[Any]:
        with self._sessions() as session:
            return list(session.scalars(statement).all())


class SqlCallRepository(_Repository, CallRepositoryPort):
    def add(self, call: Call) -> Call:
        row = CallRow()
        _copy(call, row, CALL_FIELDS, skip_none=True)
        return _call(self._insert(row))

    def get(self, call_id: int) -> Optional[Call]:
        row = self._one(select(CallRow).where(CallRow.id == call_id))
        return _call(row) if row else None

    def get_by_sid(self, call_sid: str) -> Optional[Call]:
        row = self._one(select(CallRow).where(CallRow.call_sid == call_sid))
        return _call(row) if row else None

    def get_with_messages(self, call_id: int) -> Optional[Call]:
        with self._sessions() as session:
            row = session.scalars(
                select(CallRow).options(selectinload(CallRow.messages)).where(CallRow.id == call_id)
            ).first()
            return _call(row, with_messages=True) if row else None

    def list(
        self,
        skip: int,
        limit: int,
        status: Optional[CallStatus] = None,
        from_number: Optional[str] = None,
        outcome: Optional[CallOutcome] = None,
    ) -> List[Call]:
        statement = select(CallRow)
        if status:
            statement = statement.where(CallRow.status == status)
        if from_number:
            statement = statement.where(CallRow.from_number == from_number)
        if outcome:
            statement = statement.where(CallRow.outcome == outcome)
        statement = statement.order_by(CallRow.created_at.desc()).offset(skip).limit(limit)
        return [_call(row) for row in self._all(statement)]

    def save(self, call: Call) -> Call:
        row = self._update(CallRow, call, CALL_FIELDS)
        return _call(row) if row else call

    def delete(self, call_id: int) -> bool:
        return self._delete(CallRow, call_id)

    def statistics(self, since: datetime) -> CallStatistics:
        recent = CallRow.created_at >= since
        with self._sessions() as session:
            total = session.scalar(select(func.count(CallRow.id)).where(recent))
            completed = session.scalar(
                select(func.count(CallRow.id)).where(recent, CallRow.status == CallStatus.COMPLETED)
            )
            failed = session.scalar(
                select(func.count(CallRow.id)).where(recent, CallRow.status.in_(FAILED_STATUSES))
            )
            average = session.scalar(
                select(func.avg(CallRow.duration)).where(recent, CallRow.status == CallStatus.COMPLETED)
            )
            messages = session.scalar(
                select(func.count(MessageRow.id)).join(CallRow, MessageRow.call_id == CallRow.id).where(recent)
            )
            outcomes = dict(
                session.execute(
                    select(CallRow.outcome, func.count(CallRow.id)).where(recent).group_by(CallRow.outcome)
                ).all()
            )
        return CallStatistics(
            total_calls=total or 0,
            completed_calls=completed or 0,
            failed_calls=failed or 0,
            average_duration=round(float(average or 0), 2),
            total_messages=messages or 0,
            resolved_calls=outcomes.get(CallOutcome.RESOLVED, 0),
            transferred_calls=outcomes.get(CallOutcome.TRANSFERRED, 0),
            callback_requests=outcomes.get(CallOutcome.CALLBACK_REQUESTED, 0),
            abandoned_calls=outcomes.get(CallOutcome.ABANDONED, 0),
        )


class SqlMessageRepository(_Repository, MessageRepositoryPort):
    def add(self, message: Message) -> Message:
        row = MessageRow()
        _copy(message, row, MESSAGE_FIELDS, skip_none=True)
        return _message(self._insert(row))

    def list_by_call(self, call_id: int) -> List[Message]:
        statement = select(MessageRow).where(MessageRow.call_id == call_id).order_by(MessageRow.id)
        return [_message(row) for row in self._all(statement)]


class SqlScenarioRepository(_Repository, ScenarioRepositoryPort):
    def add(self, scenario: Scenario) -> Scenario:
        row = ScenarioRow()
        _copy(scenario, row, SCENARIO_FIELDS, skip_none=True)
        return _scenario(self._insert(row))

    def get(self, scenario_id: int) -> Optional[Scenario]:
        row = self._one(select(ScenarioRow).where(ScenarioRow.id == scenario_id))
        return _scenario(row) if row else None

    def get_by_name(self, name: str) -> Optional[Scenario]:
        row = self._one(select(ScenarioRow).where(ScenarioRow.name == name))
        return _scenario(row) if row else None

    def get_by_phone(self, phone_number: str) -> Optional[Scenario]:
        row = self._one(select(ScenarioRow).where(ScenarioRow.phone_number == phone_number))
        return _scenario(row) if row else None

    def get_default(self) -> Optional[Scenario]:
        row = self._one(select(ScenarioRow).where(ScenarioRow.is_default == 1).order_by(ScenarioRow.id))
        return _scenario(row) if row else None

    def list(self, skip: int, limit: int, active_only: bool = True) -> List[Scenario]:
        statement = select(ScenarioRow)
        if active_only:
            statement = statement.where(ScenarioRow.is_active == 1)
        statement = statement.order_by(ScenarioRow.created_at.desc()).offset(skip).limit(limit)
        return [_scenario(row) for row in self._all(statement)]

    def save(self, scenario: Scenario) -> Scenario:
        row = self._update(ScenarioRow, scenario, SCENARIO_FIELDS)
        return _scenario(row) if row else scenario

    def delete(self, scenario_id: int) -> bool:
        return self._delete(ScenarioRow, scenario_id)


class SqlKnowledgeRepository(_Repository, KnowledgeRepositoryPort):
    def add(self, entry: KnowledgeEntry) -> KnowledgeEntry:
        row = KnowledgeRow()
        _copy(entry, row, KNOWLEDGE_FIELDS, skip_none=True)
        return _knowledge(self._insert(row))

    def get(self, entry_id: int) -> Optional[KnowledgeEntry]:
        row = self._one(select(KnowledgeRow).where(KnowledgeRow.id == entry_id))
        return _knowledge(row) if row else None

    def find_by_question(self, question: str) -> Optional[KnowledgeEntry]:
        row = self._one(select(KnowledgeRow).where(KnowledgeRow.question == question))
        return _knowledge(row) if row else None

    def list(
        self,
        skip: int,
        limit: int,
        category: Optional[str] = None,
        scenario: Optional[str] = None,
        active_only: bool = False,
    ) -> List[KnowledgeEntry]:
        statement = select(KnowledgeRow)
        if category:
            statement = statement.where(KnowledgeRow.category == category)
        if scenario:
            statement = statement.where(KnowledgeRow.scenario == scenario)
        if active_only:
            statement = statement.where(KnowledgeRow.is_active == 1)
        statement = statement.order_by(KnowledgeRow.id).offset(skip).limit(limit)
        return [_knowledge(row) for row in self._all(statement)]

    def list_for_scenario(self, scenario: str) -> List[KnowledgeEntry]:
        statement = select(KnowledgeRow).where(
            KnowledgeRow.is_active == 1,
            or_(KnowledgeRow.scenario.is_(None), KnowledgeRow.scenario == "", KnowledgeRow.scenario == scenario),
        )
        return [_knowledge(row) for row in self._all(statement)]

    def save(self, entry: KnowledgeEntry) -> KnowledgeEntry:
        row = self._update(KnowledgeRow, entry, KNOWLEDGE_FIELDS)
        return _knowledge(row) if row else entry

    def delete(self, entry_id: int) -> bool:
        return self._delete(KnowledgeRow, entry_id)


class SqlCustomerRepository(_Repository, CustomerRepositoryPort):
    def add(self, customer: Customer) -> Customer:
        row = CustomerRow()
        _copy(customer, row, CUSTOMER_FIELDS, skip_none=True)
        return _customer(self._insert(row))

    def get_by_phone(self, phone_number: str) -> Optional[Customer]:
        row = self._one(select(CustomerRow).where(CustomerRow.phone_number == phone_number))
        return _customer(row) if row else None

    def save(self, customer: Customer) -> Customer:
        row = self._update(CustomerRow, customer, CUSTOMER_FIELDS)
        return _customer(row) if row else customer
