"""Dependency Injection container: nối các port (interface) với adapter cụ thể.

Đây là nơi DUY NHẤT trong hệ thống biết cả về domain/application lẫn về
Motor/httpx/FastAPI cùng lúc. Application & domain layer không import container.
"""
from __future__ import annotations

from typing import List, Optional

from motor.motor_asyncio import AsyncIOMotorClient

from app.adapters.outbound.mongo.repository import MongoResultRepository
from app.adapters.outbound.scrapers.mien_bac import MienBacScraper
from app.adapters.outbound.scrapers.mien_nam import MienNamScraper
from app.adapters.outbound.scrapers.mien_trung import MienTrungScraper
from app.application.backfill import BackfillUseCase
from app.application.check_winning import CheckWinningUseCase
from app.application.scrape_results import ScrapeResultsUseCase
from app.config import Settings
from app.domain.ports.result_repository import ResultRepositoryPort
from app.domain.ports.scraper_port import ScraperPort
from app.domain.prize_rules import PrizeRuleConfig


class Container:
    """Khởi tạo và giữ tham chiếu tới các adapter/use case, dùng làm FastAPI dependency."""

    def __init__(self, settings: Settings) -> None:
        self.settings = settings
        self._mongo_client: Optional[AsyncIOMotorClient] = None
        self._repository: Optional[ResultRepositoryPort] = None

    def connect(self) -> None:
        self._mongo_client = AsyncIOMotorClient(self.settings.mongo_uri)
        self._repository = MongoResultRepository(self._mongo_client, self.settings.mongo_db)

    async def disconnect(self) -> None:
        if self._mongo_client is not None:
            self._mongo_client.close()

    async def startup(self) -> None:
        self.connect()
        await self.repository.ensure_indexes()

    @property
    def repository(self) -> ResultRepositoryPort:
        if self._repository is None:
            raise RuntimeError("Container chưa được connect(). Gọi container.startup() trước.")
        return self._repository

    @property
    def has_repository(self) -> bool:
        """True nếu repository đã được gán (qua startup() thật hoặc override_repository() trong test)."""
        return self._repository is not None

    def override_repository(self, repository: ResultRepositoryPort) -> None:
        """Dùng trong test để thay repository thật bằng fake/mock, giữ nguyên use cases."""
        self._repository = repository

    def build_scrapers(self) -> List[ScraperPort]:
        s = self.settings
        common_kwargs = dict(
            base_url=s.scraper_base_url,
            source_name=s.scraper_source_name,
            user_agent=s.http_user_agent,
            timeout_seconds=s.http_timeout_seconds,
            max_retries=s.http_max_retries,
            backoff_seconds=s.http_retry_backoff_seconds,
        )
        return [
            MienNamScraper(**common_kwargs),
            MienTrungScraper(**common_kwargs),
            MienBacScraper(**common_kwargs),
        ]

    def build_scrape_results_use_case(self) -> ScrapeResultsUseCase:
        return ScrapeResultsUseCase(scrapers=self.build_scrapers(), repository=self.repository)

    def build_backfill_use_case(self) -> BackfillUseCase:
        return BackfillUseCase(
            scrapers=self.build_scrapers(),
            repository=self.repository,
            delay_seconds=self.settings.backfill_delay_seconds,
        )

    def build_check_winning_use_case(self) -> CheckWinningUseCase:
        rule_config = PrizeRuleConfig(enable_khuyen_khich_rule=self.settings.enable_khuyen_khich_rule)
        return CheckWinningUseCase(repository=self.repository, rule_config=rule_config)
