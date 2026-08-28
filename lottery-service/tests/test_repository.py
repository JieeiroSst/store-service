"""Test MongoResultRepository bằng mongomock-motor (giả lập Motor, không cần Mongo thật)."""
from datetime import date, datetime, timezone

import pytest
from mongomock_motor import AsyncMongoMockClient

from app.adapters.outbound.mongo.repository import MongoResultRepository
from app.domain.models import LotteryResult, Region


def _make_result(province="Đồng Nai", province_code="dongnai", day=date(2026, 8, 26)) -> LotteryResult:
    return LotteryResult(
        date=day,
        region=Region.MIEN_NAM,
        province=province,
        province_code=province_code,
        prizes={"dac_biet": ["145917"], "giai_8": ["70"]},
        source="test-source",
        scraped_at=datetime.now(timezone.utc),
    )


@pytest.fixture
def repository() -> MongoResultRepository:
    client = AsyncMongoMockClient()
    return MongoResultRepository(client, "lottery_test")


@pytest.mark.asyncio
async def test_upsert_and_get_result(repository: MongoResultRepository):
    await repository.ensure_indexes()
    result = _make_result()
    await repository.upsert_result(result)

    fetched = await repository.get_result(result.date, result.region, result.province_code)
    assert fetched is not None
    assert fetched.province == "Đồng Nai"
    assert fetched.prizes["dac_biet"] == ["145917"]


@pytest.mark.asyncio
async def test_upsert_is_idempotent_no_duplicates(repository: MongoResultRepository):
    await repository.ensure_indexes()
    result = _make_result()
    await repository.upsert_result(result)

    updated = _make_result()
    updated.prizes = {"dac_biet": ["999999"], "giai_8": ["11"]}
    await repository.upsert_result(updated)

    all_results = await repository.list_by_date(result.date)
    assert len(all_results) == 1
    assert all_results[0].prizes["dac_biet"] == ["999999"]


@pytest.mark.asyncio
async def test_find_by_date_province_without_knowing_region(repository: MongoResultRepository):
    await repository.upsert_result(_make_result())
    found = await repository.find_by_date_province(date(2026, 8, 26), "dongnai")
    assert found is not None
    assert found.region == Region.MIEN_NAM


@pytest.mark.asyncio
async def test_get_existing_regions(repository: MongoResultRepository):
    await repository.upsert_result(_make_result())
    regions = await repository.get_existing_regions(date(2026, 8, 26))
    assert regions == {Region.MIEN_NAM}

    other_day_regions = await repository.get_existing_regions(date(2026, 8, 27))
    assert other_day_regions == set()


@pytest.mark.asyncio
async def test_result_not_found_returns_none(repository: MongoResultRepository):
    found = await repository.find_by_date_province(date(2026, 1, 1), "khongtontai")
    assert found is None
