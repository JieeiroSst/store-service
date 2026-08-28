"""Thành phần dùng chung cho các scraper: HTTP client (retry/delay), parser HTML tổng quát.

Parser cố tình KHÔNG dựa vào tên class CSS cụ thể để xác định một hàng là giải gì.
Thay vào đó nó đọc NHÃN TEXT của mỗi hàng (vd "G.ĐB", "100N", "Đặc Biệt", ...) và
map qua LABEL_MAP để suy ra prize_key chuẩn hoá (dac_biet, giai_1, ...). Nhờ vậy
parser vẫn hoạt động được nếu trang nguồn đổi class/CSS, miễn nhãn text còn giữ
nguyên ý nghĩa.

Ghi chú: cấu trúc HTML cụ thể của một trang tổng hợp thật (vd minhchinh.com) có thể
thay đổi theo thời gian. Hàm `parse_province_blocks` bên dưới kỳ vọng mỗi tỉnh/đài
nằm trong một khối có thuộc tính `data-tinh` (tên tỉnh) và `data-ma` (mã tỉnh, có
thể vắng mặt), bên trong là các hàng <tr><td>nhãn</td><td>số...</td></tr>. Nếu
trang nguồn thực tế có cấu trúc khác, chỉ cần viết lại hàm `parse_province_blocks`
mà không đụng tới phần còn lại của hệ thống (điểm mạnh của kiến trúc hexagonal).
"""
from __future__ import annotations

import asyncio
import logging
import re
from datetime import date
from typing import Dict, List, Optional, Tuple

import httpx
from bs4 import BeautifulSoup

from app.domain.models import normalize_province_code

logger = logging.getLogger(__name__)

_NUMBER_RE = re.compile(r"\d{2,6}")


def extract_numbers(text: str) -> List[str]:
    """Tách các dãy số từ 1 chuỗi (vd "87474 - 73361" -> ["87474", "73361"])."""
    return _NUMBER_RE.findall(text or "")


def normalize_label(text: str) -> str:
    """Chuẩn hoá nhãn giải: bỏ dấu, viết thường, bỏ ký tự không phải chữ/số.

    "G.ĐB" -> "gdb", "Đặc Biệt" -> "dacbiet", "2 Tỷ" -> "2ty"
    """
    return normalize_province_code(text)


def build_url(base_url: str, path_slug: str, day: date) -> str:
    return f"{base_url.rstrip('/')}/{path_slug}/{day.strftime('%d-%m-%Y')}.html"


async def fetch_html(
    client: httpx.AsyncClient,
    url: str,
    max_retries: int = 3,
    backoff_seconds: float = 2.0,
) -> Optional[str]:
    """GET một URL với retry + backoff. Trả None nếu tất cả các lần thử đều lỗi."""
    last_error: Optional[Exception] = None
    for attempt in range(1, max_retries + 1):
        try:
            response = await client.get(url)
            if response.status_code == 404:
                logger.info("URL not found (chưa có kết quả?): %s", url)
                return None
            response.raise_for_status()
            return response.text
        except httpx.HTTPError as exc:  # noqa: PERF203
            last_error = exc
            logger.warning(
                "Fetch attempt %d/%d failed for %s: %s", attempt, max_retries, url, exc
            )
            if attempt < max_retries:
                await asyncio.sleep(backoff_seconds * attempt)

    logger.error("Giving up fetching %s after %d attempts: %s", url, max_retries, last_error)
    return None


def parse_province_blocks(
    html: str,
    label_map: Dict[str, str],
) -> List[Tuple[str, str, Dict[str, List[str]]]]:
    """Parse HTML thành list (province_name, province_code, prizes_dict).

    prizes_dict chỉ chứa các prize_key nhận diện được qua `label_map`; nhãn lạ
    bị bỏ qua (log ở caller nếu cần) thay vì làm hỏng toàn bộ kết quả.
    """
    soup = BeautifulSoup(html, "html.parser")
    blocks = soup.find_all(attrs={"data-tinh": True})

    results: List[Tuple[str, str, Dict[str, List[str]]]] = []
    for block in blocks:
        province_name = block.get("data-tinh", "").strip()
        province_code = (block.get("data-ma") or "").strip() or normalize_province_code(province_name)

        prizes: Dict[str, List[str]] = {}
        for row in block.find_all("tr"):
            cells = row.find_all("td")
            if len(cells) < 2:
                continue
            label_key = label_map.get(normalize_label(cells[0].get_text()))
            if not label_key:
                continue
            value_text = " ".join(c.get_text(" ", strip=True) for c in cells[1:])
            numbers = extract_numbers(value_text)
            if numbers:
                prizes.setdefault(label_key, []).extend(numbers)

        if province_name and prizes:
            results.append((province_name, province_code, prizes))

    return results


# Nhãn dùng chung cho Miền Nam & Miền Trung (vé 6 số).
# Một số trang hiển thị nhãn theo GIÁ TRỊ giải (100N, 200N, ... 2TỶ) thay vì G.1..G.8,
# nên map cả hai kiểu nhãn về cùng prize_key.
LABEL_MAP_MN_MT: Dict[str, str] = {
    "gdb": "dac_biet",
    "db": "dac_biet",
    "dacbiet": "dac_biet",
    "2ty": "dac_biet",
    "g1": "giai_1",
    "giai1": "giai_1",
    "30tr": "giai_1",
    "g2": "giai_2",
    "giai2": "giai_2",
    "15tr": "giai_2",
    "g3": "giai_3",
    "giai3": "giai_3",
    "10tr": "giai_3",
    "g4": "giai_4",
    "giai4": "giai_4",
    "3tr": "giai_4",
    "g5": "giai_5",
    "giai5": "giai_5",
    "1tr": "giai_5",
    "g6": "giai_6",
    "giai6": "giai_6",
    "400n": "giai_6",
    "g7": "giai_7",
    "giai7": "giai_7",
    "200n": "giai_7",
    "g8": "giai_8",
    "giai8": "giai_8",
    "100n": "giai_8",
}

# Nhãn dùng cho Miền Bắc (vé 5 số, ĐB + G1..G7).
LABEL_MAP_MB: Dict[str, str] = {
    "gdb": "dac_biet",
    "db": "dac_biet",
    "dacbiet": "dac_biet",
    "g1": "giai_1",
    "giai1": "giai_1",
    "g2": "giai_2",
    "giai2": "giai_2",
    "g3": "giai_3",
    "giai3": "giai_3",
    "g4": "giai_4",
    "giai4": "giai_4",
    "g5": "giai_5",
    "giai5": "giai_5",
    "g6": "giai_6",
    "giai6": "giai_6",
    "g7": "giai_7",
    "giai7": "giai_7",
}
