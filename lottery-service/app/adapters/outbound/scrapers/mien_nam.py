"""Scraper adapter cho Xổ số Miền Nam. Implement ScraperPort."""
from __future__ import annotations

from datetime import date, datetime, timezone
from typing import List

import httpx

from app.domain.models import LotteryResult, Region
from app.domain.ports.scraper_port import ScraperPort
from app.adapters.outbound.scrapers.base import (
    LABEL_MAP_MN_MT,
    build_url,
    fetch_html,
    parse_province_blocks,
)

PATH_SLUG = "xo-so-mien-nam"


def parse_html(html: str, day: date, source: str) -> List[LotteryResult]:
    """Pure function: parse HTML -> list[LotteryResult]. Test offline được, không cần mạng."""
    blocks = parse_province_blocks(html, LABEL_MAP_MN_MT)
    now = datetime.now(timezone.utc)
    return [
        LotteryResult(
            date=day,
            region=Region.MIEN_NAM,
            province=name,
            province_code=code,
            prizes=prizes,
            source=source,
            scraped_at=now,
        )
        for name, code, prizes in blocks
    ]


class MienNamScraper(ScraperPort):
    region = Region.MIEN_NAM

    def __init__(
        self,
        base_url: str,
        source_name: str,
        user_agent: str,
        timeout_seconds: float = 15.0,
        max_retries: int = 3,
        backoff_seconds: float = 2.0,
    ) -> None:
        self._base_url = base_url
        self._source_name = source_name
        self._user_agent = user_agent
        self._timeout_seconds = timeout_seconds
        self._max_retries = max_retries
        self._backoff_seconds = backoff_seconds

    async def fetch(self, day: date) -> List[LotteryResult]:
        url = build_url(self._base_url, PATH_SLUG, day)
        headers = {"User-Agent": self._user_agent}
        async with httpx.AsyncClient(timeout=self._timeout_seconds, headers=headers) as client:
            html = await fetch_html(client, url, self._max_retries, self._backoff_seconds)

        if not html:
            return []
        return parse_html(html, day, self._source_name)
