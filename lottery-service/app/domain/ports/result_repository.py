"""Outbound port: lưu/đọc LotteryResult. Adapter cụ thể (Mongo, fake, ...) implement interface này."""
from __future__ import annotations

from abc import ABC, abstractmethod
from datetime import date
from typing import List, Optional, Set

from app.domain.models import LotteryResult, Region


class ResultRepositoryPort(ABC):
    @abstractmethod
    async def upsert_result(self, result: LotteryResult) -> None:
        """Ghi/cập nhật 1 kết quả theo khoá duy nhất (date, region, province_code)."""
        raise NotImplementedError

    @abstractmethod
    async def get_result(self, day: date, region: Region, province_code: str) -> Optional[LotteryResult]:
        raise NotImplementedError

    @abstractmethod
    async def find_by_date_province(self, day: date, province_code: str) -> Optional[LotteryResult]:
        """Tìm theo (ngày, mã tỉnh) không cần biết trước miền — tiện cho API kiểm tra trúng thưởng."""
        raise NotImplementedError

    @abstractmethod
    async def list_by_date(self, day: date) -> List[LotteryResult]:
        raise NotImplementedError

    @abstractmethod
    async def get_existing_regions(self, day: date) -> Set[Region]:
        """Trả về tập các miền đã có ít nhất 1 kết quả trong ngày — dùng cho backfill resume-safe."""
        raise NotImplementedError

    @abstractmethod
    async def ensure_indexes(self) -> None:
        raise NotImplementedError
