"""Scheduler adapter (APScheduler): chạy use case scrape hằng ngày.

Chạy nhiều mốc giờ chiều/tối (giờ VN) sau khi 3 miền đã quay số xong, để có
thêm cơ hội bắt kịp nếu nguồn dữ liệu cập nhật muộn hoặc lần chạy đầu lỗi mạng.
"""
from __future__ import annotations

import logging
from datetime import date

from apscheduler.schedulers.asyncio import AsyncIOScheduler
from apscheduler.triggers.cron import CronTrigger

from app.config import Settings
from app.container import Container

logger = logging.getLogger(__name__)


async def _run_daily_scrape(container: Container) -> None:
    use_case = container.build_scrape_results_use_case()
    today = date.today()
    logger.info("Scheduled scrape bắt đầu cho ngày %s", today)
    summary = await use_case.execute(today)
    logger.info(
        "Scheduled scrape hoàn tất: ok=%s failed=%s provinces_saved=%d",
        summary.regions_ok,
        summary.regions_failed,
        summary.provinces_saved,
    )


def start_scheduler(container: Container, settings: Settings) -> AsyncIOScheduler:
    scheduler = AsyncIOScheduler(timezone=settings.scheduler_timezone)
    for hour in settings.scheduler_run_hours:
        scheduler.add_job(
            _run_daily_scrape,
            trigger=CronTrigger(hour=hour, minute=settings.scheduler_run_minute),
            args=[container],
            id=f"daily_scrape_{hour}h",
            replace_existing=True,
            misfire_grace_time=3600,
        )
    scheduler.start()
    return scheduler


def stop_scheduler(scheduler: AsyncIOScheduler) -> None:
    scheduler.shutdown(wait=False)
