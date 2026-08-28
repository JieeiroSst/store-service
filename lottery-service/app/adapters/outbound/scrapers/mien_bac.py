"""Scraper adapter cho Xổ số Miền Bắc. Implement ScraperPort.

Miền Bắc mỗi ngày chỉ xổ 1 đài duy nhất, nhưng "đài" (tỉnh phát hành) luân phiên
theo thứ trong tuần. Nếu trang nguồn không nhúng sẵn tên tỉnh trong HTML,
dùng bảng tra `WEEKDAY_PROVINCE` bên dưới làm phương án dự phòng.
"""
from __future__ import annotations

from datetime import date, datetime, timezone
from typing import List

import httpx

from app.domain.models import LotteryResult, Region, normalize_province_code
from app.domain.ports.scraper_port import ScraperPort
from app.adapters.outbound.scrapers.base import (
    LABEL_MAP_MB,
    build_url,
    fetch_html,
    parse_province_blocks,
)

PATH_SLUG = "xo-so-mien-bac"

# date.weekday(): Monday=0 ... Sunday=6
WEEKDAY_PROVINCE = {
    0: "Hà Nội",
    1: "Quảng Ninh",
    2: "Bắc Ninh",
    3: "Hải Phòng",
    4: "Hải Dương",
    5: "Xổ Số Thủ Đô",
    6: "Nam Định",
}


def parse_html(html: str, day: date, source: str) -> List[LotteryResult]:
    blocks = parse_province_blocks(html, LABEL_MAP_MB)
    now = datetime.now(timezone.utc)

    if not blocks:
        return []

    results = []
    for name, code, prizes in blocks:
        results.append(
            LotteryResult(
                date=day,
                region=Region.MIEN_BAC,
                province=name,
                province_code=code,
                prizes=prizes,
                source=source,
                scraped_at=now,
            )
        )
    return results


def fallback_province_for(day: date) -> tuple[str, str]:
    name = WEEKDAY_PROVINCE[day.weekday()]
    return name, normalize_province_code(name)


class MienBacScraper(ScraperPort):
    region = Region.MIEN_BAC

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
