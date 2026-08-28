"""Use case: cào lịch sử theo khoảng ngày [start, end], resume-safe.

Resume-safe: với mỗi ngày, mỗi miền đã có dữ liệu trong DB sẽ được BỎ QUA
(trừ khi force=True) — cho phép chạy lại backfill nhiều lần / tiếp tục sau khi
bị gián đoạn mà không cào lại từ đầu hay tạo dữ liệu trùng.
"""
from __future__ import annotations

import asyncio
import logging
from datetime import date, timedelta
from typing import Iterable, List

from app.domain.models import BackfillSummary
from app.domain.ports.result_repository import ResultRepositoryPort
from app.domain.ports.scraper_port import ScraperPort

logger = logging.getLogger(__name__)


def _daterange(start: date, end: date) -> Iterable[date]:
    days = (end - start).days
    for i in range(days + 1):
        yield start + timedelta(days=i)


class BackfillUseCase:
    def __init__(
        self,
        scrapers: Iterable[ScraperPort],
        repository: ResultRepositoryPort,
        delay_seconds: float = 1.0,
    ) -> None:
        self._scrapers = list(scrapers)
        self._repository = repository
        self._delay_seconds = delay_seconds

    async def execute(self, start: date, end: date, force: bool = False) -> BackfillSummary:
        if start > end:
            raise ValueError("start date must be <= end date")

        days_processed = 0
        days_skipped = 0
        errors: List[str] = []

        for day in _daterange(start, end):
            existing_regions = set() if force else await self._repository.get_existing_regions(day)
            pending_scrapers = [s for s in self._scrapers if s.region not in existing_regions]

            if not pending_scrapers:
                days_skipped += 1
                logger.info("Backfill skip day=%s (already have all regions)", day)
                continue

            day_had_work = False
            for scraper in pending_scrapers:
                try:
                    results = await scraper.fetch(day)
                except Exception as exc:  # noqa: BLE001
                    msg = f"{day} [{scraper.region}]: {exc}"
                    logger.exception("Backfill scrape failed: %s", msg)
                    errors.append(msg)
                    continue

                for result in results:
                    await self._repository.upsert_result(result)
                day_had_work = True

                await asyncio.sleep(self._delay_seconds)

            if day_had_work:
                days_processed += 1

        return BackfillSummary(
            start=start,
            end=end,
            days_processed=days_processed,
            days_skipped=days_skipped,
            errors=errors,
        )
