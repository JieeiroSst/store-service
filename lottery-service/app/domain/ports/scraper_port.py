"""Outbound port: nguồn cào dữ liệu. Mỗi miền implement 1 adapter riêng."""
from __future__ import annotations

from abc import ABC, abstractmethod
from datetime import date
from typing import List

from app.domain.models import LotteryResult, Region


class ScraperPort(ABC):
    region: Region

    @abstractmethod
    async def fetch(self, day: date) -> List[LotteryResult]:
        """Cào kết quả của TẤT CẢ các tỉnh/đài thuộc miền này trong ngày `day`.

        Trả về list rỗng nếu ngày đó miền này không xổ (vd. một số tỉnh miền Trung
        không xổ vào một số thứ trong tuần) hoặc chưa có kết quả (xổ muộn / lỗi nguồn).
        """
        raise NotImplementedError
