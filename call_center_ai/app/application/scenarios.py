from __future__ import annotations

from typing import Any, List, Mapping, Optional

from app.domain import playbook
from app.domain.errors import AlreadyExistsError, NotFoundError
from app.domain.models import Scenario
from app.domain.ports import ScenarioRepositoryPort

EDITABLE_FIELDS = ("name", "description", "prompt", "greeting", "phone_number", "is_default", "is_active")


class ScenarioResolver:
    def __init__(self, scenarios: ScenarioRepositoryPort, fallback: str) -> None:
        self._scenarios = scenarios
        self._fallback = fallback

    def for_number(self, dialed_number: Optional[str]) -> Scenario:
        if dialed_number:
            routed = self._scenarios.get_by_phone(dialed_number)
            if routed is not None and routed.is_active:
                return routed
        default = self._scenarios.get_default()
        if default is not None and default.is_active:
            return default
        return self.by_name(self._fallback)

    def by_name(self, name: Optional[str]) -> Scenario:
        name = name or self._fallback
        stored = self._scenarios.get_by_name(name)
        if stored is not None:
            return stored
        return playbook.builtin_scenario(name)


class ScenarioService:
    def __init__(self, scenarios: ScenarioRepositoryPort) -> None:
        self._scenarios = scenarios

    def list_scenarios(self, skip: int = 0, limit: int = 100) -> List[Scenario]:
        return self._scenarios.list(skip, limit, active_only=True)

    def get_scenario(self, scenario_id: int) -> Scenario:
        scenario = self._scenarios.get(scenario_id)
        if scenario is None:
            raise NotFoundError("Scenario not found")
        return scenario

    def create_scenario(self, scenario: Scenario) -> Scenario:
        if self._scenarios.get_by_name(scenario.name) is not None:
            raise AlreadyExistsError("Scenario with this name already exists")
        created = self._scenarios.add(scenario)
        if created.is_default:
            self._clear_other_defaults(created.id)
        return created

    def update_scenario(self, scenario_id: int, changes: Mapping[str, Any]) -> Scenario:
        scenario = self.get_scenario(scenario_id)
        if "name" in changes and changes["name"] != scenario.name:
            if self._scenarios.get_by_name(changes["name"]) is not None:
                raise AlreadyExistsError("Scenario with this name already exists")
        for name in EDITABLE_FIELDS:
            if name in changes:
                setattr(scenario, name, changes[name])
        saved = self._scenarios.save(scenario)
        if saved.is_default:
            self._clear_other_defaults(saved.id)
        return saved

    def delete_scenario(self, scenario_id: int) -> None:
        if not self._scenarios.delete(scenario_id):
            raise NotFoundError("Scenario not found")

    def seed_defaults(self) -> List[str]:
        has_default = self._scenarios.get_default() is not None
        created = []
        for scenario in playbook.default_scenarios():
            if self._scenarios.get_by_name(scenario.name) is not None:
                continue
            if has_default:
                scenario.is_default = 0
            self._scenarios.add(scenario)
            has_default = has_default or bool(scenario.is_default)
            created.append(scenario.name)
        return created

    def _clear_other_defaults(self, keep_id: int) -> None:
        for other in self._scenarios.list(0, 1000, active_only=False):
            if other.id != keep_id and other.is_default:
                other.is_default = 0
                self._scenarios.save(other)
