"""Outbound adapter: implement ResultRepositoryPort bằng MongoDB (Motor - async driver)."""
from __future__ import annotations

from datetime import date, datetime
from typing import Any, Dict, List, Optional, Set

from motor.motor_asyncio import AsyncIOMotorClient, AsyncIOMotorCollection, AsyncIOMotorDatabase

from app.domain.models import LotteryResult, Region
from app.domain.ports.result_repository import ResultRepositoryPort

COLLECTION_NAME = "lottery_results"
DATE_FORMAT = "%Y-%m-%d"


def _to_document(result: LotteryResult) -> Dict[str, Any]:
    return {
        "date": result.date.strftime(DATE_FORMAT),
        "region": result.region.value,
        "province": result.province,
        "province_code": result.province_code,
        "prizes": result.prizes,
        "source": result.source,
        "scraped_at": result.scraped_at,
    }


def _from_document(doc: Dict[str, Any]) -> LotteryResult:
    return LotteryResult(
        date=datetime.strptime(doc["date"], DATE_FORMAT).date(),
        region=Region(doc["region"]),
        province=doc["province"],
        province_code=doc["province_code"],
        prizes=doc["prizes"],
        source=doc["source"],
        scraped_at=doc["scraped_at"],
    )


class MongoResultRepository(ResultRepositoryPort):
    def __init__(self, client: AsyncIOMotorClient, db_name: str) -> None:
        self._client = client
        self._db: AsyncIOMotorDatabase = client[db_name]
        self._collection: AsyncIOMotorCollection = self._db[COLLECTION_NAME]

    async def ensure_indexes(self) -> None:
        await self._collection.create_index(
            [("date", 1), ("region", 1), ("province_code", 1)],
            unique=True,
            name="uniq_date_region_province",
        )
        await self._collection.create_index([("date", 1)], name="idx_date")

    async def upsert_result(self, result: LotteryResult) -> None:
        doc = _to_document(result)
        await self._collection.update_one(
            {
                "date": doc["date"],
                "region": doc["region"],
                "province_code": doc["province_code"],
            },
            {"$set": doc},
            upsert=True,
        )

    async def get_result(self, day: date, region: Region, province_code: str) -> Optional[LotteryResult]:
        doc = await self._collection.find_one(
            {
                "date": day.strftime(DATE_FORMAT),
                "region": region.value,
                "province_code": province_code,
            }
        )
        return _from_document(doc) if doc else None

    async def find_by_date_province(self, day: date, province_code: str) -> Optional[LotteryResult]:
        doc = await self._collection.find_one(
            {"date": day.strftime(DATE_FORMAT), "province_code": province_code}
        )
        return _from_document(doc) if doc else None

    async def list_by_date(self, day: date) -> List[LotteryResult]:
        cursor = self._collection.find({"date": day.strftime(DATE_FORMAT)})
        return [_from_document(doc) async for doc in cursor]

    async def get_existing_regions(self, day: date) -> Set[Region]:
        cursor = self._collection.find(
            {"date": day.strftime(DATE_FORMAT)}, {"region": 1}
        )
        regions: Set[Region] = set()
        async for doc in cursor:
            regions.add(Region(doc["region"]))
        return regions
