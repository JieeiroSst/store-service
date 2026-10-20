from __future__ import annotations

from typing import Any, List, Mapping, Optional

from app.domain import knowledge, playbook
from app.domain.errors import NotFoundError
from app.domain.models import KnowledgeEntry
from app.domain.ports import KnowledgeRepositoryPort

EDITABLE_FIELDS = ("question", "answer", "keywords", "category", "scenario", "is_active")


class KnowledgeService:
    def __init__(self, entries: KnowledgeRepositoryPort, default_scenario: str) -> None:
        self._entries = entries
        self._default_scenario = default_scenario

    def list_entries(
        self,
        skip: int = 0,
        limit: int = 100,
        category: Optional[str] = None,
        scenario: Optional[str] = None,
    ) -> List[KnowledgeEntry]:
        return self._entries.list(skip, limit, category=category, scenario=scenario)

    def get_entry(self, entry_id: int) -> KnowledgeEntry:
        entry = self._entries.get(entry_id)
        if entry is None:
            raise NotFoundError("Knowledge entry not found")
        return entry

    def create_entry(self, entry: KnowledgeEntry) -> KnowledgeEntry:
        return self._entries.add(entry)

    def update_entry(self, entry_id: int, changes: Mapping[str, Any]) -> KnowledgeEntry:
        entry = self.get_entry(entry_id)
        for name in EDITABLE_FIELDS:
            if name in changes:
                setattr(entry, name, changes[name])
        return self._entries.save(entry)

    def delete_entry(self, entry_id: int) -> None:
        if not self._entries.delete(entry_id):
            raise NotFoundError("Knowledge entry not found")

    def search(self, query: str, scenario: Optional[str] = None, limit: int = 3) -> List[KnowledgeEntry]:
        candidates = self._entries.list_for_scenario(scenario or self._default_scenario)
        return knowledge.search(query, candidates, limit=limit)

    def seed_defaults(self) -> List[str]:
        created = []
        for entry in playbook.default_knowledge():
            if self._entries.find_by_question(entry.question) is None:
                self._entries.add(entry)
                created.append(entry.question)
        return created
