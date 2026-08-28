"""Test API /api/v1/check bằng FastAPI TestClient + fake repository (không cần Mongo/network)."""
from datetime import date, datetime, timezone

import pytest
from fastapi.testclient import TestClient

from app.adapters.inbound.api.main import app
from app.config import settings
from app.container import Container
from app.domain.models import LotteryResult, Region
from tests.fakes import FakeResultRepository


@pytest.fixture
def client():
    container = Container(settings)
    container.override_repository(FakeResultRepository())
    app.state.container = container

    with TestClient(app) as test_client:
        yield test_client, container.repository


@pytest.mark.asyncio
async def test_check_winning_won_dac_biet(client):
    test_client, repository = client
    await repository.upsert_result(
        LotteryResult(
            date=date(2026, 8, 26),
            region=Region.MIEN_NAM,
            province="Đồng Nai",
            province_code="dongnai",
            prizes={"dac_biet": ["145917"]},
            source="test",
            scraped_at=datetime.now(timezone.utc),
        )
    )

    response = test_client.post(
        "/api/v1/check",
        json={"date": "2026-08-26", "province": "Đồng Nai", "ticket_number": "145917"},
    )

    assert response.status_code == 200
    body = response.json()
    assert body["won"] is True
    assert body["result_found"] is True
    assert "dac_biet" in body["prizes_won"]
    assert body["highest_prize"] == "dac_biet"


@pytest.mark.asyncio
async def test_check_winning_by_province_code_instead_of_name(client):
    test_client, repository = client
    await repository.upsert_result(
        LotteryResult(
            date=date(2026, 8, 26),
            region=Region.MIEN_NAM,
            province="Đồng Nai",
            province_code="dongnai",
            prizes={"dac_biet": ["145917"]},
            source="test",
            scraped_at=datetime.now(timezone.utc),
        )
    )

    response = test_client.post(
        "/api/v1/check",
        json={"date": "2026-08-26", "province": "dongnai", "ticket_number": "145917"},
    )
    assert response.status_code == 200
    assert response.json()["won"] is True


@pytest.mark.asyncio
async def test_check_winning_not_won(client):
    test_client, repository = client
    await repository.upsert_result(
        LotteryResult(
            date=date(2026, 8, 26),
            region=Region.MIEN_NAM,
            province="Đồng Nai",
            province_code="dongnai",
            prizes={"dac_biet": ["145917"]},
            source="test",
            scraped_at=datetime.now(timezone.utc),
        )
    )

    response = test_client.post(
        "/api/v1/check",
        json={"date": "2026-08-26", "province": "Đồng Nai", "ticket_number": "000000"},
    )
    assert response.status_code == 200
    body = response.json()
    assert body["won"] is False
    assert body["result_found"] is True


def test_check_winning_result_not_found_returns_200_with_flag(client):
    test_client, _ = client
    response = test_client.post(
        "/api/v1/check",
        json={"date": "2099-01-01", "province": "Đồng Nai", "ticket_number": "145917"},
    )
    assert response.status_code == 200
    body = response.json()
    assert body["result_found"] is False
    assert body["won"] is False


def test_check_winning_invalid_ticket_number_returns_422(client):
    test_client, _ = client
    response = test_client.post(
        "/api/v1/check",
        json={"date": "2026-08-26", "province": "Đồng Nai", "ticket_number": "abc"},
    )
    assert response.status_code == 422
