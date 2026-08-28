"""FastAPI routes = inbound adapter. Chỉ chuyển đổi HTTP <-> use case, không chứa logic nghiệp vụ."""
from __future__ import annotations

import logging
from datetime import date

from fastapi import APIRouter, HTTPException, Request

from app.adapters.inbound.api.schemas import (
    BackfillRequest,
    BackfillResponse,
    CheckRequest,
    CheckResponse,
    PrizeDetailResponse,
    ResultResponse,
    ScrapeRequest,
    ScrapeResponse,
)
from app.container import Container
from app.domain.models import normalize_province_code

logger = logging.getLogger(__name__)

router = APIRouter(prefix="/api/v1")


def _container(request: Request) -> Container:
    return request.app.state.container


@router.post("/check", response_model=CheckResponse)
async def check_winning(payload: CheckRequest, request: Request) -> CheckResponse:
    """Kiểm tra 1 vé số có trúng thưởng không.

    Trả HTTP 200 kể cả khi chưa có kết quả cho (ngày, tỉnh) đó — đây là một
    kết quả nghiệp vụ hợp lệ ("chưa quay số" / "chưa cào được"), không phải lỗi.
    result_found=false báo hiệu điều đó cho client.
    """
    container = _container(request)
    use_case = container.build_check_winning_use_case()
    province_code = normalize_province_code(payload.province)

    result = await use_case.execute(payload.date, province_code, payload.ticket_number)

    return CheckResponse(
        won=result.won,
        prizes_won=result.prizes_won,
        detail=[PrizeDetailResponse(prize=d.prize, matched=d.matched, rule=d.rule) for d in result.detail],
        result_found=result.result_found,
        highest_prize=result.highest_prize,
    )


@router.get("/results", response_model=ResultResponse)
async def get_result(request: Request, date: date, province: str) -> ResultResponse:
    """Lấy nguyên bản kết quả xổ số của 1 tỉnh/đài trong 1 ngày."""
    container = _container(request)
    province_code = normalize_province_code(province)
    result = await container.repository.find_by_date_province(date, province_code)
    if result is None:
        raise HTTPException(status_code=404, detail="Không tìm thấy kết quả cho ngày/tỉnh này")

    return ResultResponse(
        date=result.date,
        region=result.region.value,
        province=result.province,
        province_code=result.province_code,
        prizes=result.prizes,
        source=result.source,
    )


@router.post("/scrape", response_model=ScrapeResponse)
async def trigger_scrape(payload: ScrapeRequest, request: Request) -> ScrapeResponse:
    """Kích hoạt cào thủ công cho 1 ngày (3 miền). Dùng cho vận hành/kiểm thử."""
    container = _container(request)
    use_case = container.build_scrape_results_use_case()
    summary = await use_case.execute(payload.date)
    return ScrapeResponse(
        date=summary.date,
        regions_ok=[r.value for r in summary.regions_ok],
        regions_failed=[r.value for r in summary.regions_failed],
        provinces_saved=summary.provinces_saved,
    )


@router.post("/backfill", response_model=BackfillResponse)
async def trigger_backfill(payload: BackfillRequest, request: Request) -> BackfillResponse:
    """Kích hoạt backfill lịch sử theo khoảng ngày. Resume-safe (xem BackfillUseCase)."""
    container = _container(request)
    use_case = container.build_backfill_use_case()
    try:
        summary = await use_case.execute(payload.start, payload.end, force=payload.force)
    except ValueError as exc:
        raise HTTPException(status_code=400, detail=str(exc)) from exc

    return BackfillResponse(
        start=summary.start,
        end=summary.end,
        days_processed=summary.days_processed,
        days_skipped=summary.days_skipped,
        errors=summary.errors,
    )
