"""Test parser của từng miền bằng HTML mẫu offline — không cần mạng."""
from datetime import date
from pathlib import Path

from app.adapters.outbound.scrapers import mien_bac, mien_nam, mien_trung

FIXTURES_DIR = Path(__file__).parent / "fixtures"


def _read(name: str) -> str:
    return (FIXTURES_DIR / name).read_text(encoding="utf-8")


def test_parse_mien_nam_html():
    html = _read("mien_nam_sample.html")
    results = mien_nam.parse_html(html, date(2026, 8, 26), "test-source")

    assert len(results) == 2
    dong_nai = next(r for r in results if r.province_code == "dongnai")
    assert dong_nai.province == "Đồng Nai"
    assert dong_nai.prizes["dac_biet"] == ["145917"]
    assert dong_nai.prizes["giai_3"] == ["87474", "73361"]
    assert dong_nai.prizes["giai_8"] == ["70"]

    can_tho = next(r for r in results if r.province_code == "cantho")
    # Cần Thơ dùng nhãn theo GIÁ TRỊ (2TỶ, 30TR, ...) thay vì G.ĐB/G.1 -> vẫn phải map đúng.
    assert can_tho.prizes["dac_biet"] == ["805679"]
    assert can_tho.prizes["giai_1"] == ["77777"]
    assert can_tho.prizes["giai_8"] == ["89"]


def test_parse_mien_trung_html():
    html = _read("mien_trung_sample.html")
    results = mien_trung.parse_html(html, date(2026, 8, 26), "test-source")

    assert len(results) == 1
    result = results[0]
    assert result.region.value == "mien-trung"
    assert result.province_code == "daklak"
    assert result.prizes["dac_biet"] == ["512346"]
    assert result.prizes["giai_7"] == ["777"]


def test_parse_mien_bac_html():
    html = _read("mien_bac_sample.html")
    results = mien_bac.parse_html(html, date(2026, 8, 26), "test-source")

    assert len(results) == 1
    result = results[0]
    assert result.region.value == "mien-bac"
    assert result.prizes["dac_biet"] == ["12345"]
    assert result.prizes["giai_7"] == ["55", "66"]
    assert result.prizes["giai_4"] == ["4444", "5555", "6666", "7777"]


def test_parse_html_empty_when_no_province_blocks():
    assert mien_nam.parse_html("<html><body>no data</body></html>", date(2026, 8, 26), "src") == []


def test_fallback_province_for_weekday():
    name, code = mien_bac.fallback_province_for(date(2026, 8, 24))  # Monday
    assert name == "Hà Nội"
    assert code == "hanoi"
