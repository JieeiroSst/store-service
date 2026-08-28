"""Test application layer use cases bằng fake scraper/repository (ports thay được cho nhau)."""
from datetime import date, datetime, timezone

import pytest

from app.application.backfill import BackfillUseCase
from app.application.check_winning import CheckWinningUseCase
from app.application.scrape_results import ScrapeResultsUseCase
from app.domain.models import LotteryResult, Region
from tests.fakes import FailingScraper, FakeResultRepository, FakeScraper


def _result(day: date, region: Region, code: str) -> LotteryResult:
    return LotteryResult(
        date=day,
        region=region,
        province=code,
        province_code=code,
        prizes={"dac_biet": ["123456"]},
        source="test",
        scraped_at=datetime.now(timezone.utc),
    )


@pytest.mark.asyncio
async def test_scrape_results_saves_all_regions_and_continues_on_failure():
    day = date(2026, 8, 26)
    repo = FakeResultRepository()
    scrapers = [
        FakeScraper(Region.MIEN_NAM, {day: [_result(day, Region.MIEN_NAM, "dongnai")]}),
        FailingScraper(Region.MIEN_TRUNG),
        FakeScraper(Region.MIEN_BAC, {day: [_result(day, Region.MIEN_BAC, "hanoi")]}),
    ]
    use_case = ScrapeResultsUseCase(scrapers, repo)

    summary = await use_case.execute(day)

    assert Region.MIEN_TRUNG in summary.regions_failed
    assert Region.MIEN_NAM in summary.regions_ok
    assert Region.MIEN_BAC in summary.regions_ok
    assert summary.provinces_saved == 2
    assert await repo.find_by_date_province(day, "dongnai") is not None


@pytest.mark.asyncio
async def test_backfill_is_resume_safe_skips_existing_regions():
    day = date(2026, 8, 26)
    repo = FakeResultRepository()
    await repo.upsert_result(_result(day, Region.MIEN_NAM, "dongnai"))

    scraper_nam = FakeScraper(Region.MIEN_NAM, {day: [_result(day, Region.MIEN_NAM, "dongnai")]})
    scraper_bac = FakeScraper(Region.MIEN_BAC, {day: [_result(day, Region.MIEN_BAC, "hanoi")]})
    use_case = BackfillUseCase([scraper_nam, scraper_bac], repo, delay_seconds=0)

    await use_case.execute(day, day)

    assert scraper_nam.calls == []  # đã có dữ liệu mien-nam ngày này -> bỏ qua
    assert scraper_bac.calls == [day]  # chưa có dữ liệu mien-bac -> vẫn cào


@pytest.mark.asyncio
async def test_backfill_force_re_scrapes_even_if_existing():
    day = date(2026, 8, 26)
    repo = FakeResultRepository()
    await repo.upsert_result(_result(day, Region.MIEN_NAM, "dongnai"))

    scraper_nam = FakeScraper(Region.MIEN_NAM, {day: [_result(day, Region.MIEN_NAM, "dongnai")]})
    use_case = BackfillUseCase([scraper_nam], repo, delay_seconds=0)

    await use_case.execute(day, day, force=True)

    assert scraper_nam.calls == [day]


@pytest.mark.asyncio
async def test_check_winning_use_case_result_not_found():
    repo = FakeResultRepository()
    use_case = CheckWinningUseCase(repo)

    result = await use_case.execute(date(2026, 1, 1), "khongtontai", "123456")

    assert result.result_found is False
    assert result.won is False


@pytest.mark.asyncio
async def test_check_winning_use_case_delegates_to_prize_rules():
    day = date(2026, 8, 26)
    repo = FakeResultRepository()
    await repo.upsert_result(_result(day, Region.MIEN_NAM, "dongnai"))
    use_case = CheckWinningUseCase(repo)

    result = await use_case.execute(day, "dongnai", "123456")

    assert result.result_found is True
    assert result.won is True
    assert result.prizes_won == ["dac_biet"]
