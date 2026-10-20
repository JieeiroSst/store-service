from __future__ import annotations

from abc import ABC, abstractmethod
from datetime import datetime
from typing import List, Optional

from app.domain.models import (
    Call,
    CallOutcome,
    CallStatistics,
    CallStatus,
    Customer,
    KnowledgeEntry,
    Message,
    Scenario,
)


class CallRepositoryPort(ABC):
    @abstractmethod
    def add(self, call: Call) -> Call:
        raise NotImplementedError

    @abstractmethod
    def get(self, call_id: int) -> Optional[Call]:
        raise NotImplementedError

    @abstractmethod
    def get_by_sid(self, call_sid: str) -> Optional[Call]:
        raise NotImplementedError

    @abstractmethod
    def get_with_messages(self, call_id: int) -> Optional[Call]:
        raise NotImplementedError

    @abstractmethod
    def list(
        self,
        skip: int,
        limit: int,
        status: Optional[CallStatus] = None,
        from_number: Optional[str] = None,
        outcome: Optional[CallOutcome] = None,
    ) -> List[Call]:
        raise NotImplementedError

    @abstractmethod
    def save(self, call: Call) -> Call:
        raise NotImplementedError

    @abstractmethod
    def delete(self, call_id: int) -> bool:
        raise NotImplementedError

    @abstractmethod
    def statistics(self, since: datetime) -> CallStatistics:
        raise NotImplementedError


class MessageRepositoryPort(ABC):
    @abstractmethod
    def add(self, message: Message) -> Message:
        raise NotImplementedError

    @abstractmethod
    def list_by_call(self, call_id: int) -> List[Message]:
        raise NotImplementedError


class ScenarioRepositoryPort(ABC):
    @abstractmethod
    def add(self, scenario: Scenario) -> Scenario:
        raise NotImplementedError

    @abstractmethod
    def get(self, scenario_id: int) -> Optional[Scenario]:
        raise NotImplementedError

    @abstractmethod
    def get_by_name(self, name: str) -> Optional[Scenario]:
        raise NotImplementedError

    @abstractmethod
    def get_by_phone(self, phone_number: str) -> Optional[Scenario]:
        raise NotImplementedError

    @abstractmethod
    def get_default(self) -> Optional[Scenario]:
        raise NotImplementedError

    @abstractmethod
    def list(self, skip: int, limit: int, active_only: bool = True) -> List[Scenario]:
        raise NotImplementedError

    @abstractmethod
    def save(self, scenario: Scenario) -> Scenario:
        raise NotImplementedError

    @abstractmethod
    def delete(self, scenario_id: int) -> bool:
        raise NotImplementedError


class KnowledgeRepositoryPort(ABC):
    @abstractmethod
    def add(self, entry: KnowledgeEntry) -> KnowledgeEntry:
        raise NotImplementedError

    @abstractmethod
    def get(self, entry_id: int) -> Optional[KnowledgeEntry]:
        raise NotImplementedError

    @abstractmethod
    def find_by_question(self, question: str) -> Optional[KnowledgeEntry]:
        raise NotImplementedError

    @abstractmethod
    def list(
        self,
        skip: int,
        limit: int,
        category: Optional[str] = None,
        scenario: Optional[str] = None,
        active_only: bool = False,
    ) -> List[KnowledgeEntry]:
        raise NotImplementedError

    @abstractmethod
    def list_for_scenario(self, scenario: str) -> List[KnowledgeEntry]:
        raise NotImplementedError

    @abstractmethod
    def save(self, entry: KnowledgeEntry) -> KnowledgeEntry:
        raise NotImplementedError

    @abstractmethod
    def delete(self, entry_id: int) -> bool:
        raise NotImplementedError


class CustomerRepositoryPort(ABC):
    @abstractmethod
    def add(self, customer: Customer) -> Customer:
        raise NotImplementedError

    @abstractmethod
    def get_by_phone(self, phone_number: str) -> Optional[Customer]:
        raise NotImplementedError

    @abstractmethod
    def save(self, customer: Customer) -> Customer:
        raise NotImplementedError
