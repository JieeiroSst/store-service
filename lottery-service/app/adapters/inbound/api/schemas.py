"""Pydantic schemas cho FastAPI (inbound adapter). KHÔNG dùng lại ở domain/application."""
# Import "date" dưới bí danh "date_type": Pydantic v2 không cho phép annotation
# trùng tên chữ với chính field đó (vd `date: date`) vì gây nhập nhằng khi nó
# tự resolve forward-ref -> xem https://errors.pydantic.dev/.../unevaluable-type-annotation
from datetime import date as date_type
from typing import List, Optional

from pydantic import BaseModel, Field, field_validator


class CheckRequest(BaseModel):
    date: date_type = Field(..., description="Ngày quay số, định dạng YYYY-MM-DD")
    province: str = Field(..., description="Tên tỉnh/đài hoặc province_code, vd 'Đồng Nai' hoặc 'dongnai'")
    ticket_number: str = Field(..., description="Dãy số trên vé (5-6 chữ số tuỳ miền)")

    @field_validator("ticket_number")
    @classmethod
    def validate_ticket_number(cls, v: str) -> str:
        v = v.strip()
        if not v.isdigit():
            raise ValueError("ticket_number chỉ được chứa chữ số")
        if not (2 <= len(v) <= 6):
            raise ValueError("ticket_number phải có từ 2 đến 6 chữ số")
        return v

    @field_validator("province")
    @classmethod
    def validate_province(cls, v: str) -> str:
        v = v.strip()
        if not v:
            raise ValueError("province không được để trống")
        return v


class PrizeDetailResponse(BaseModel):
    prize: str
    matched: str
    rule: str


class CheckResponse(BaseModel):
    won: bool
    prizes_won: List[str]
    detail: List[PrizeDetailResponse]
    result_found: bool
    highest_prize: Optional[str] = None


class ResultResponse(BaseModel):
    date: date_type
    region: str
    province: str
    province_code: str
    prizes: dict
    source: str


class ScrapeRequest(BaseModel):
    date: date_type


class ScrapeResponse(BaseModel):
    date: date_type
    regions_ok: List[str]
    regions_failed: List[str]
    provinces_saved: int


class BackfillRequest(BaseModel):
    start: date_type
    end: date_type
    force: bool = False


class BackfillResponse(BaseModel):
    start: date_type
    end: date_type
    days_processed: int
    days_skipped: int
    errors: List[str]
