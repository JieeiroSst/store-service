"""Cấu hình ứng dụng, đọc từ biến môi trường / .env bằng pydantic-settings."""
from __future__ import annotations

from typing import List

from pydantic_settings import BaseSettings, SettingsConfigDict


class Settings(BaseSettings):
    model_config = SettingsConfigDict(env_file=".env", env_file_encoding="utf-8", extra="ignore")

    app_name: str = "lottery-service"
    log_level: str = "INFO"

    # MongoDB
    mongo_uri: str = "mongodb://localhost:27017"
    mongo_db: str = "lottery"

    # HTTP scraping
    http_timeout_seconds: float = 15.0
    http_max_retries: int = 3
    http_retry_backoff_seconds: float = 2.0
    http_delay_between_requests_seconds: float = 1.5
    http_user_agent: str = (
        "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 "
        "(KHTML, like Gecko) Chrome/124.0.0.0 Safari/537.36"
    )
    scraper_source_name: str = "minhchinh.com"
    scraper_base_url: str = "https://www.minhchinh.com"

    # Scheduler (giờ Việt Nam, chạy nhiều mốc để bắt kịp giờ xổ + phòng lỗi mạng)
    scheduler_enabled: bool = False
    scheduler_timezone: str = "Asia/Ho_Chi_Minh"
    scheduler_run_hours: List[int] = [18, 19, 21]
    scheduler_run_minute: int = 30

    # Prize rules
    enable_khuyen_khich_rule: bool = True

    # Backfill
    backfill_delay_seconds: float = 1.0


settings = Settings()
