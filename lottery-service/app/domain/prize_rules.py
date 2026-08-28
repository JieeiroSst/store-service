"""Quy tắc dò thưởng xổ số Việt Nam — pure functions, không phụ thuộc I/O.

Nguyên tắc chung ("so đuôi"): mỗi giải có một độ dài chữ số cố định N.
Vé số trúng giải đó khi N chữ số CUỐI của vé trùng khớp với con số của giải
(có độ dài đúng N). Giải đặc biệt là trường hợp đặc biệt: phải khớp toàn bộ
chiều dài vé.

Bộ quy tắc được tách theo từng miền dưới dạng RegionRuleSet (strategy),
để dễ mở rộng/chỉnh sửa khi thể lệ thay đổi mà không đụng vào logic dò chung.
"""
from __future__ import annotations

from dataclasses import dataclass, field
from typing import Dict, List, Optional

from app.domain.models import LotteryResult, PrizeMatch, Region, WinningCheckResult

DAC_BIET = "dac_biet"
PHU_DB = "phu_db"
KHUYEN_KHICH = "khuyen_khich"


@dataclass(frozen=True)
class PrizeDigitRule:
    prize_key: str
    digit_length: int
    label: str


@dataclass(frozen=True)
class RegionRuleSet:
    region: Region
    ticket_length: int
    # Thứ tự từ giải cao -> giải thấp, dùng để xác định "giải cao nhất" khi trúng nhiều giải.
    digit_rules: List[PrizeDigitRule]
    has_phu_db: bool = False
    supports_khuyen_khich: bool = False


# Miền Nam & Miền Trung: vé 6 chữ số.
_MN_MT_DIGIT_RULES = [
    PrizeDigitRule("giai_1", 5, "giải nhất"),
    PrizeDigitRule("giai_2", 5, "giải nhì"),
    PrizeDigitRule("giai_3", 5, "giải ba"),
    PrizeDigitRule("giai_4", 5, "giải tư"),
    PrizeDigitRule("giai_5", 4, "giải năm"),
    PrizeDigitRule("giai_6", 4, "giải sáu"),
    PrizeDigitRule("giai_7", 3, "giải bảy"),
    PrizeDigitRule("giai_8", 2, "giải tám"),
]

MIEN_NAM_RULES = RegionRuleSet(
    region=Region.MIEN_NAM,
    ticket_length=6,
    digit_rules=_MN_MT_DIGIT_RULES,
    has_phu_db=True,
    supports_khuyen_khich=True,
)

MIEN_TRUNG_RULES = RegionRuleSet(
    region=Region.MIEN_TRUNG,
    ticket_length=6,
    digit_rules=_MN_MT_DIGIT_RULES,
    has_phu_db=True,
    supports_khuyen_khich=True,
)

# Miền Bắc: vé 5 chữ số, hệ giải ĐB, G1..G7.
MIEN_BAC_RULES = RegionRuleSet(
    region=Region.MIEN_BAC,
    ticket_length=5,
    digit_rules=[
        PrizeDigitRule("giai_1", 5, "giải nhất"),
        PrizeDigitRule("giai_2", 5, "giải nhì"),
        PrizeDigitRule("giai_3", 5, "giải ba"),
        PrizeDigitRule("giai_4", 4, "giải tư"),
        PrizeDigitRule("giai_5", 4, "giải năm"),
        PrizeDigitRule("giai_6", 3, "giải sáu"),
        PrizeDigitRule("giai_7", 2, "giải bảy"),
    ],
    has_phu_db=False,
    supports_khuyen_khich=False,
)

REGION_RULES: Dict[Region, RegionRuleSet] = {
    Region.MIEN_NAM: MIEN_NAM_RULES,
    Region.MIEN_TRUNG: MIEN_TRUNG_RULES,
    Region.MIEN_BAC: MIEN_BAC_RULES,
}


@dataclass
class PrizeRuleConfig:
    """Cấu hình bật/tắt các quy tắc "mềm" (khác nhau tuỳ công ty xổ số)."""

    enable_khuyen_khich_rule: bool = True


def _priority_order(rule_set: RegionRuleSet) -> List[str]:
    order = [DAC_BIET]
    if rule_set.supports_khuyen_khich:
        order.append(KHUYEN_KHICH)
    if rule_set.has_phu_db:
        order.append(PHU_DB)
    order.extend(r.prize_key for r in rule_set.digit_rules)
    return order


def check_winning(
    region: Region,
    prizes: Dict[str, List[str]],
    ticket_number: str,
    config: Optional[PrizeRuleConfig] = None,
) -> WinningCheckResult:
    """Dò một vé số theo kết quả `prizes` của một tỉnh/đài đã biết.

    Giả định `prizes` là kết quả CÓ THẬT (đã tìm thấy trong DB). Trường hợp
    "không tìm thấy kết quả ngày/tỉnh đó" được xử lý ở application layer
    (use case check_winning.py), không phải ở đây.
    """
    config = config or PrizeRuleConfig()
    rule_set = REGION_RULES[region]
    ticket = ticket_number.strip()

    detail: List[PrizeMatch] = []

    # 1) Giải đặc biệt: phải khớp toàn bộ chiều dài vé.
    db_numbers = prizes.get(DAC_BIET, [])
    for db_number in db_numbers:
        if len(db_number) == rule_set.ticket_length and ticket == db_number:
            detail.append(
                PrizeMatch(
                    prize=DAC_BIET,
                    matched=db_number,
                    rule=f"trùng đủ {rule_set.ticket_length} số giải đặc biệt",
                )
            )

    # 2) Các giải 1..8 (hoặc 1..7 ở miền Bắc): so khớp các chữ số cuối của vé.
    for digit_rule in rule_set.digit_rules:
        candidates = prizes.get(digit_rule.prize_key, [])
        n = digit_rule.digit_length
        if len(ticket) < n:
            continue
        ticket_suffix = ticket[-n:]
        for number in candidates:
            if len(number) == n and ticket_suffix == number:
                detail.append(
                    PrizeMatch(
                        prize=digit_rule.prize_key,
                        matched=number,
                        rule=f"trùng {n} số cuối của {digit_rule.label}",
                    )
                )

    # 3) Giải phụ đặc biệt & giải khuyến khích: chỉ áp dụng khi vé cùng độ dài với ĐB
    #    và chỉ sai lệch đúng 1 chữ số so với giải đặc biệt.
    if (rule_set.has_phu_db or rule_set.supports_khuyen_khich) and len(ticket) == rule_set.ticket_length:
        for db_number in db_numbers:
            if len(db_number) != rule_set.ticket_length:
                continue
            diff_positions = [i for i in range(len(ticket)) if ticket[i] != db_number[i]]
            if len(diff_positions) != 1:
                continue
            differing_index = diff_positions[0]

            # Giải phụ đặc biệt: 5 số cuối trùng ĐB, riêng SỐ ĐẦU TIÊN sai khác.
            if rule_set.has_phu_db and differing_index == 0:
                detail.append(
                    PrizeMatch(
                        prize=PHU_DB,
                        matched=db_number,
                        rule="trùng 5 số cuối giải đặc biệt, sai số đầu tiên",
                    )
                )

            # Giải khuyến khích: sai đúng 1 chữ số ở BẤT KỲ vị trí nào (cấu hình được).
            if rule_set.supports_khuyen_khich and config.enable_khuyen_khich_rule:
                detail.append(
                    PrizeMatch(
                        prize=KHUYEN_KHICH,
                        matched=db_number,
                        rule=f"sai đúng 1 chữ số (vị trí {differing_index + 1}) so với giải đặc biệt",
                    )
                )

    prizes_won: List[str] = []
    for m in detail:
        if m.prize not in prizes_won:
            prizes_won.append(m.prize)

    priority = _priority_order(rule_set)
    highest_prize = None
    for p in priority:
        if p in prizes_won:
            highest_prize = p
            break

    return WinningCheckResult(
        won=len(detail) > 0,
        prizes_won=prizes_won,
        detail=detail,
        result_found=True,
        highest_prize=highest_prize,
    )
