"""Use case: cào kết quả xổ số của cả 3 miền cho MỘT ngày, lưu qua repository port."""
from __future__ import annotations

import logging
from datetime import date
from typing import Iterable, List

from app.domain.models import ScrapeSummary
from app.domain.ports.result_repository import ResultRepositoryPort
from app.domain.ports.scraper_port import ScraperPort

logger = logging.getLogger(__name__)


class ScrapeResultsUseCase:
    def __init__(self, scrapers: Iterable[ScraperPort], repository: ResultRepositoryPort) -> None:
        self._scrapers = list(scrapers)
        self._repository = repository

    async def execute(self, day: date) -> ScrapeSummary:
        regions_ok = []
        regions_failed = []
        provinces_saved = 0

        for scraper in self._scrapers:
            try:
                results = await scraper.fetch(day)
            except Exception:  # noqa: BLE001 - một miền lỗi không được chặn các miền khác
                logger.exception("Scrape failed for region=%s day=%s", scraper.region, day)
                regions_failed.append(scraper.region)
                continue

            for result in results:
                await self._repository.upsert_result(result)
                provinces_saved += 1

            regions_ok.append(scraper.region)
            logger.info(
                "Scraped region=%s day=%s provinces=%d", scraper.region, day, len(results)
            )

        return ScrapeSummary(
            date=day,
            regions_ok=regions_ok,
            regions_failed=regions_failed,
            provinces_saved=provinces_saved,
        )
