"""Domain entities. Không import FastAPI, Motor hay bất kỳ framework nào ở đây."""
from __future__ import annotations

import re
import unicodedata
from dataclasses import dataclass, field
from datetime import date, datetime
from enum import Enum
from typing import Dict, List, Optional


class Region(str, Enum):
    MIEN_NAM = "mien-nam"
    MIEN_TRUNG = "mien-trung"
    MIEN_BAC = "mien-bac"


def normalize_province_code(name: str) -> str:
    """Chuẩn hoá tên tỉnh -> mã tỉnh: bỏ dấu, viết thường, bỏ khoảng trắng/ký tự đặc biệt.

    Ví dụ: "Đồng Nai" -> "dongnai", "Bà Rịa - Vũng Tàu" -> "bariavungtau"
    """
    if not name:
        return ""
    s = name.strip().lower()
    s = s.replace("đ", "d")
    s = unicodedata.normalize("NFD", s)
    s = "".join(c for c in s if unicodedata.category(c) != "Mn")
    s = re.sub(r"[^a-z0-9]+", "", s)
    return s


@dataclass
class LotteryResult:
    """Kết quả xổ số của MỘT tỉnh/đài trong MỘT ngày."""

    date: date
    region: Region
    province: str
    province_code: str
    prizes: Dict[str, List[str]]
    source: str
    scraped_at: datetime = field(default_factory=lambda: datetime.utcnow())

    def numbers_for(self, prize_key: str) -> List[str]:
        return self.prizes.get(prize_key, [])


@dataclass
class PrizeMatch:
    """Một giải mà vé số trúng, kèm số đã khớp và lý do (rule) để giải thích cho người dùng."""

    prize: str
    matched: str
    rule: str


@dataclass
class WinningCheckResult:
    won: bool
    prizes_won: List[str]
    detail: List[PrizeMatch]
    result_found: bool
    highest_prize: Optional[str] = None


@dataclass(frozen=True)
class BackfillSummary:
    start: date
    end: date
    days_processed: int
    days_skipped: int
    errors: List[str] = field(default_factory=list)


@dataclass(frozen=True)
class ScrapeSummary:
    date: date
    regions_ok: List[Region] = field(default_factory=list)
    regions_failed: List[Region] = field(default_factory=list)
    provinces_saved: int = 0
