"""Fake adapters dùng trong test — minh hoạ tính thay thế được của hexagonal ports."""
from __future__ import annotations

from datetime import date
from typing import Dict, List, Set, Tuple

from app.domain.models import LotteryResult, Region
from app.domain.ports.result_repository import ResultRepositoryPort
from app.domain.ports.scraper_port import ScraperPort


class FakeResultRepository(ResultRepositoryPort):
    def __init__(self) -> None:
        self._store: Dict[Tuple[str, str, str], LotteryResult] = {}

    def _key(self, day: date, region: Region, province_code: str) -> Tuple[str, str, str]:
        return (day.isoformat(), region.value, province_code)

    async def ensure_indexes(self) -> None:
        return None

    async def upsert_result(self, result: LotteryResult) -> None:
        self._store[self._key(result.date, result.region, result.province_code)] = result

    async def get_result(self, day: date, region: Region, province_code: str):
        return self._store.get(self._key(day, region, province_code))

    async def find_by_date_province(self, day: date, province_code: str):
        for result in self._store.values():
            if result.date == day and result.province_code == province_code:
                return result
        return None

    async def list_by_date(self, day: date) -> List[LotteryResult]:
        return [r for r in self._store.values() if r.date == day]

    async def get_existing_regions(self, day: date) -> Set[Region]:
        return {r.region for r in self._store.values() if r.date == day}


class FakeScraper(ScraperPort):
    def __init__(self, region: Region, results_by_day: Dict[date, List[LotteryResult]] | None = None):
        self.region = region
        self._results_by_day = results_by_day or {}
        self.calls: List[date] = []

    async def fetch(self, day: date) -> List[LotteryResult]:
        self.calls.append(day)
        return self._results_by_day.get(day, [])


class FailingScraper(ScraperPort):
    def __init__(self, region: Region):
        self.region = region

    async def fetch(self, day: date) -> List[LotteryResult]:
        raise RuntimeError("boom")
