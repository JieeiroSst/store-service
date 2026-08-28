"""Unit test cho domain/prize_rules.py — bao phủ các nhánh: trúng ĐB, trúng phụ ĐB,
trúng khuyến khích, trúng giải thấp, trúng nhiều giải, không trúng, và Miền Bắc.
"""
from app.domain.models import Region
from app.domain.prize_rules import DAC_BIET, KHUYEN_KHICH, PHU_DB, PrizeRuleConfig, check_winning

MN_PRIZES = {
    "dac_biet": ["145917"],
    "giai_1": ["98208"],
    "giai_2": ["69706"],
    "giai_3": ["87474", "73361"],
    "giai_4": ["11234", "22345", "33456", "44567", "55678", "66789", "77890"],
    "giai_5": ["4068"],
    "giai_6": ["6898", "6712", "1527"],
    "giai_7": ["147", "370"],
    "giai_8": ["70"],
}

NO_KK = PrizeRuleConfig(enable_khuyen_khich_rule=False)
WITH_KK = PrizeRuleConfig(enable_khuyen_khich_rule=True)


def test_trung_dac_biet():
    result = check_winning(Region.MIEN_NAM, MN_PRIZES, "145917", NO_KK)
    assert result.won is True
    assert result.result_found is True
    assert result.prizes_won == [DAC_BIET]
    assert result.highest_prize == DAC_BIET
    assert result.detail[0].matched == "145917"


def test_trung_phu_dac_biet_sai_so_dau():
    # "145917" -> đổi số đầu tiên 1 -> 2, giữ nguyên 5 số cuối.
    ticket = "245917"
    result = check_winning(Region.MIEN_NAM, MN_PRIZES, ticket, NO_KK)
    assert result.won is True
    assert PHU_DB in result.prizes_won
    assert KHUYEN_KHICH not in result.prizes_won  # đã tắt cấu hình
    assert DAC_BIET not in result.prizes_won


def test_trung_khuyen_khich_khi_bat_cau_hinh():
    # Sai đúng 1 chữ số ở vị trí KHÔNG PHẢI đầu tiên -> không phải phụ ĐB, chỉ là khuyến khích.
    ticket = "145918"  # đổi chữ số cuối 7 -> 8
    result_on = check_winning(Region.MIEN_NAM, MN_PRIZES, ticket, WITH_KK)
    assert KHUYEN_KHICH in result_on.prizes_won
    assert PHU_DB not in result_on.prizes_won

    result_off = check_winning(Region.MIEN_NAM, MN_PRIZES, ticket, NO_KK)
    assert result_off.won is False
    assert KHUYEN_KHICH not in result_off.prizes_won


def test_trung_giai_thap():
    # Vé chỉ trùng 2 số cuối của giải 8 ("70"), không trùng gì khác.
    ticket = "990070"
    result = check_winning(Region.MIEN_NAM, MN_PRIZES, ticket, NO_KK)
    assert result.won is True
    assert result.prizes_won == ["giai_8"]
    assert result.highest_prize == "giai_8"


def test_trung_nhieu_giai_cung_luc():
    # "370" nằm trong giai_7, và "70" (2 số cuối của "370") nằm trong giai_8
    # => 1 vé trúng đồng thời 2 giải.
    ticket = "123370"
    result = check_winning(Region.MIEN_NAM, MN_PRIZES, ticket, NO_KK)
    assert result.won is True
    assert set(result.prizes_won) == {"giai_7", "giai_8"}
    # giai_7 xếp hạng cao hơn giai_8 trong bảng ưu tiên.
    assert result.highest_prize == "giai_7"
    assert len(result.detail) == 2


def test_khong_trung_giai_nao():
    ticket = "000000"
    result = check_winning(Region.MIEN_NAM, MN_PRIZES, ticket, NO_KK)
    assert result.won is False
    assert result.prizes_won == []
    assert result.detail == []
    assert result.highest_prize is None
    assert result.result_found is True  # có kết quả ngày đó, chỉ là vé không trúng


def test_mien_trung_dung_bo_quy_tac_giong_mien_nam():
    result = check_winning(Region.MIEN_TRUNG, MN_PRIZES, "145917", NO_KK)
    assert result.won is True
    assert result.prizes_won == [DAC_BIET]


MB_PRIZES = {
    "dac_biet": ["12345"],
    "giai_1": ["67890"],
    "giai_2": ["11111"],
    "giai_3": ["22222", "33333"],
    "giai_4": ["4444", "5555", "6666", "7777"],
    "giai_5": ["8888", "9999", "0000", "1111"],
    "giai_6": ["222", "333", "444"],
    "giai_7": ["55", "66"],
}


def test_mien_bac_trung_dac_biet():
    result = check_winning(Region.MIEN_BAC, MB_PRIZES, "12345", NO_KK)
    assert result.won is True
    assert result.prizes_won == [DAC_BIET]
    assert result.highest_prize == DAC_BIET


def test_mien_bac_trung_giai_thap_khong_co_phu_db_hay_khuyen_khich():
    result = check_winning(Region.MIEN_BAC, MB_PRIZES, "00066", WITH_KK)
    assert result.won is True
    assert result.prizes_won == ["giai_7"]
    # Miền Bắc không hỗ trợ phụ ĐB / khuyến khích dù config bật.
    assert PHU_DB not in result.prizes_won
    assert KHUYEN_KHICH not in result.prizes_won


def test_mien_bac_khong_trung():
    result = check_winning(Region.MIEN_BAC, MB_PRIZES, "99900", NO_KK)
    assert result.won is False
    assert result.result_found is True
